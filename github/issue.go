package github

import (
	"fmt"
	"strings"
)

func (issue *Issue) Url() string {
	return fmt.Sprintf("%s/commit/%s",
		strings.TrimSuffix(issue.repositoryUrl, "/"),
		issue.commit)
}

func (issue *Issue) FilePath() string {
	return issue.filePath
}

func (issue *Issue) Message() string {
	return issue.message
}

func NewRepositoryIssue(repositoryName, commit, filePath, message, source string) *Issue {
	return &Issue{
		repositoryName,
		commit,
		filePath,
		message,
		source,
	}
}
