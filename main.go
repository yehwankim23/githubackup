package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"
)

type RepositoriesJson struct {
	Repositories []string `json:"repositories"`
}

func main() {
	repositoriesFile := "repositories.json"

	if len(os.Args) > 1 && os.Args[1] == "personal" {
		repositoriesFile = "personal.json"
	}

	repositoriesBytes, err := os.ReadFile(repositoriesFile)

	if err != nil {
		fmt.Println("os.ReadFile(repositoriesFile) :", err.Error())
		return
	}

	var repositoriesJson RepositoriesJson
	err = json.Unmarshal(repositoriesBytes, &repositoriesJson)

	if err != nil {
		fmt.Println("json.Unmarshal(repositoriesBytes, &repositoriesJson) :", err.Error())
		return
	}

	repositories := repositoriesJson.Repositories
	length := len(repositories)

	if length == 0 {
		fmt.Println("'repositories.json' is empty")
		return
	}

	if slices.Contains(repositories, "username-or-organization/repository") {
		fmt.Println("Remove 'username-or-organization/repository' from 'repositories.json'")
		return
	}

	dateDirectory := time.Now().Format("2006-01-02 15-04-05 MST")
	err = os.RemoveAll(dateDirectory)

	if err != nil {
		fmt.Println("os.RemoveAll(dateDirectory) :", err.Error())
		return
	}

	err = os.Mkdir(dateDirectory, 0755)

	if err != nil {
		fmt.Println("os.Mkdir(dateDirectory, 0755) :", err.Error())
		return
	}

	err = os.Chdir(dateDirectory)

	if err != nil {
		fmt.Println("os.Chdir(dateDirectory) :", err.Error())
		return
	}

	github := "https://github.com/"

	if len(os.Args) > 1 && os.Args[1] == "ssh" {
		github = "git@github.com:"
	}

	var parentDirectories []string

	for index, repository := range repositories {
		before, after, found := strings.Cut(repository, "/")

		if before == "" || after == "" || !found {
			fmt.Println("Invalid repository: '" + repository + "'")
			return
		}

		parentDirectory := before

		if !slices.Contains(parentDirectories, parentDirectory) {
			err = os.Mkdir(parentDirectory, 0755)

			if err != nil {
				fmt.Println("os.Mkdir(parentDirectory, 0755) :", err.Error())
				return
			}

			parentDirectories = append(parentDirectories, parentDirectory)
		}

		err = os.Chdir(parentDirectory)

		if err != nil {
			fmt.Println("os.Chdir(parentDirectory) :", err.Error())
			return
		}

		fmt.Println("\n[" + strconv.Itoa(index+1) + "/" + strconv.Itoa(length) + "] " + repository)
		childDirectory := before + " " + after
		command := exec.Command("git", "clone", "--recursive", github+repository, childDirectory)
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		err = command.Run()

		if err != nil {
			fmt.Println("command.Run() :", err.Error())
			return
		}

		err = os.Chdir(childDirectory)

		if err != nil {
			fmt.Println("os.Chdir(childDirectory) :", err.Error())
			return
		}

		command = exec.Command("git", "log", "--reverse", "--format=%as")
		command.Stderr = os.Stderr
		date, err := command.Output()

		if err != nil {
			fmt.Println("command.Output() :", err.Error())
			return
		}

		err = os.Chdir("..")

		if err != nil {
			fmt.Println("os.Chdir('..') :", err.Error())
			return
		}

		err = os.Rename(childDirectory, string(date)[:10]+" "+childDirectory)

		if err != nil {
			fmt.Println("os.Rename(childDirectory, string(date)[:10]+' '+childDirectory) :",
				err.Error())

			return
		}

		err = os.Chdir("..")

		if err != nil {
			fmt.Println("os.Chdir('..') :", err.Error())
			return
		}
	}
}
