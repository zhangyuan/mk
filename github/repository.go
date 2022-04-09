package github

import (
	"mk/git"
	"mk/inspection"

	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
)

type Issue struct {
	repositoryUrl string
	commit        string
	filePath      string
	message       string
	source        string
}
type Repository struct {
	clonePath string
	Url       string
}

func NewRepository(clonePath string, url string) *Repository {
	return &Repository{
		clonePath: clonePath,
		Url:       url,
	}
}

func (repo *Repository) Inspect() ([]Issue, error) {
	gitRepo, err := git.OpenOrClone(repo.clonePath, repo.Url)
	if err != nil {
		return nil, errors.Wrap(err, "fail to open or clone repo")
	}

	inspector := inspection.NewInspector()

	cmmitIter, err := gitRepo.CommitObjects()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	issues := []Issue{}

	cmmitIter.ForEach(func(commit *object.Commit) error {
		commit.Parents().ForEach(func(parentCommit *object.Commit) error {
			patch, err := parentCommit.Patch(commit)

			if err != nil {
				return err
			}

			for _, filePatch := range patch.FilePatches() {
				for _, chunk := range filePatch.Chunks() {
					_, to := filePatch.Files()

					if chunk.Type() == diff.Add {
						rule, ok := inspector.InspectFileContent(chunk.Content())
						if !ok {
							issue := NewRepositoryIssue(
								repo.Url,
								commit.Hash.String(),
								to.Path(),
								rule.Message,
								"file",
							)
							issues = append(issues, *issue)
						}
					}
				}
			}

			return nil
		})

		if rule, ok := inspector.InspectCommitMessage(commit.String()); !ok {
			issue := NewRepositoryIssue(
				repo.Url,
				commit.Hash.String(),
				"",
				rule.Message,
				"commit",
			)
			issues = append(issues, *issue)
		}
		return nil
	})

	return issues, nil
}
