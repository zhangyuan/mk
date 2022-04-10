package github

import (
	"mk/git"
	"mk/inspection"

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

	err = cmmitIter.ForEach(func(commit *git.Commit) error {
		err = commit.Parents().ForEach(func(parentCommit *git.Commit) error {
			patch, err := parentCommit.Patch(commit)

			if err != nil {
				return err
			}

			for _, filePatch := range patch.FilePatches() {
				for _, chunk := range filePatch.AddedChunks() {
					_, to := filePatch.Files()

					exception, ok := inspector.InspectFileContent(chunk.Content())
					if !ok {
						issue := NewRepositoryIssue(
							repo.Url,
							commit.Hash(),
							to.Path(),
							exception.Message,
							"file",
						)
						issues = append(issues, *issue)
					}

				}
			}

			return nil
		})

		if err != nil {
			return err
		}

		if commit.NumParents() == 0 {
			files, err := commit.Files()
			if err != nil {
				return err
			}

			if err := files.ForEach(func(file *git.ObjectFile) error {
				isBinary, err := file.IsBinary()
				if err != nil {
					return err
				}

				if isBinary {
					return nil
				}

				content, err := file.Contents()
				if err != nil {
					return nil
				}

				exception, ok := inspector.InspectFileContent(content)

				if !ok {
					issue := NewRepositoryIssue(
						repo.Url,
						commit.Hash(),
						file.Name(),
						exception.Message,
						"file",
					)
					issues = append(issues, *issue)
				}

				return nil
			}); err != nil {
				return err
			}
		}

		if rule, ok := inspector.InspectCommitMessage(commit.String()); !ok {
			issue := NewRepositoryIssue(
				repo.Url,
				commit.Hash(),
				"",
				rule.Message,
				"commit",
			)
			issues = append(issues, *issue)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return issues, nil
}
