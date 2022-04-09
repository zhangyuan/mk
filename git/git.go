package git

import (
	"os"

	gogit "github.com/go-git/go-git/v5"
)

func OpenOrClone(path, repositoryUrl string) (*gogit.Repository, error) {
	var clone = func() (*gogit.Repository, error) {
		return gogit.PlainClone(path, false, &gogit.CloneOptions{
			URL:      repositoryUrl,
			Progress: os.Stdout,
		})
	}

	repo, err := clone()
	if err == nil {
		return repo, nil
	}

	repo, err = gogit.PlainOpen(path)
	if err == nil {
		err := repo.Fetch(&gogit.FetchOptions{
			Progress: os.Stdout,
			Force:    true,
		})

		if err == gogit.NoErrAlreadyUpToDate {
			return repo, nil
		}

		if err != nil {
			return nil, err
		}

		return repo, nil
	}

	if err := os.RemoveAll(path); err != nil {
		return nil, err
	}

	return clone()
}
