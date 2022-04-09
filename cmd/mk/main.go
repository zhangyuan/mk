package main

import (
	"fmt"
	"log"
	"mk/github"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("repository url is not provided")
	}
	repositoryUrl := os.Args[1]

	if err := run(repositoryUrl); err != nil {
		log.Fatalln(err)
	}
}

func run(repositoryUrl string) error {
	slice := strings.Split(repositoryUrl, "/")

	relativeDirectoryPath := strings.Join(slice[len(slice)-3:len(slice)-1], "/")
	repoName := slice[len(slice)-1]

	absDirectoryPath := "/tmp/mk/" + strings.TrimPrefix(relativeDirectoryPath, "/")

	err := os.MkdirAll(absDirectoryPath, 0750)
	if err != nil {
		return err
	}

	absRepoPath := strings.TrimSuffix(absDirectoryPath, "/") + "/" + repoName

	repo := github.NewRepository(absRepoPath, repositoryUrl)

	issues, err := repo.Inspect()

	if err != nil {
		return err
	}

	for i, ri := range issues {
		fmt.Printf("[%d] %s %s\n", i+1, ri.Url(), ri.FilePath())
		fmt.Printf("%s\n", ri.Message())
	}

	return nil
}
