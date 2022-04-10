package git

import (
	"os"
	"os/exec"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
)

func openOrClone(path, repositoryUrl string) (*gogit.Repository, error) {
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
			return nil, errors.Wrap(err, "fail to fetch")
		}

		return repo, nil
	}

	if err := os.RemoveAll(path); err != nil {
		return nil, errors.Wrap(err, "fail to remove directory")
	}

	return clone()
}

type GitRepisotry struct {
	repository *gogit.Repository
}

func (gitRepisotry *GitRepisotry) CommitObjects() (*CommitIter, error) {
	comitIter, err := gitRepisotry.repository.CommitObjects()

	if err != nil {
		return nil, err
	}

	return &CommitIter{
		inner: comitIter,
	}, nil
}

func OpenOrClone(clonePath, repositoryUrl string) (*GitRepisotry, error) {
	if gitRepo, err := openOrClone(clonePath, repositoryUrl); err != nil {
		return nil, errors.Wrap(err, "fail to open or clone go git repo")
	} else {
		if err := gitRepo.Prune(gogit.PruneOptions{
			Handler: gitRepo.DeleteObject,
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		cmd := exec.Command("git", "gc")
		cmd.Dir = clonePath
		_, err = cmd.Output()
		if err != nil {
			return nil, errors.Wrap(err, "fail to run git gc")
		}

		return &GitRepisotry{
			repository: gitRepo,
		}, nil
	}
}

type CommitIter struct {
	inner object.CommitIter
}

func (commitIter *CommitIter) ForEach(f func(*Commit) error) error {
	return commitIter.inner.ForEach(func(c *object.Commit) error {
		commit := &Commit{inner: c}
		return f(commit)
	})
}

type Commit struct {
	inner *object.Commit
}

func (c *Commit) Parents() *CommitIter {
	parents := c.inner.Parents()
	return &CommitIter{inner: parents}
}

func (c *Commit) Hash() string {
	return c.inner.Hash.String()
}

func (c *Commit) NumParents() int {
	return c.inner.NumParents()
}

func (c *Commit) Files() (*FileIter, error) {
	if fileIter, err := c.inner.Files(); err != nil {
		return nil, errors.Wrap(err, "fail to get files from commit")
	} else {
		return &FileIter{innter: *fileIter}, nil
	}
}

func (c *Commit) Patch(to *Commit) (*Patch, error) {
	patch, err := c.inner.Patch(to.inner)

	if err != nil {
		return nil, errors.Wrap(err, "fail to patch")
	}
	return &Patch{
		inner: patch,
	}, nil
}

func (c *Commit) String() string {
	return c.inner.String()
}

type Patch struct {
	inner *object.Patch
}

func (patch *Patch) FilePatches() []FilePatch {
	filePatches := []FilePatch{}
	for _, fp := range patch.inner.FilePatches() {
		filePatches = append(filePatches, FilePatch{
			inner: fp,
		})
	}

	return filePatches
}

type FilePatch struct {
	inner diff.FilePatch
}

func (filePatch *FilePatch) Files() (from DiffFile, to DiffFile) {
	fromFile, toFile := filePatch.inner.Files()

	if fromFile != nil {
		from = DiffFile{path: fromFile.Path()}
	}

	if toFile != nil {
		to = DiffFile{path: toFile.Path()}
	}

	return
}

func (filePatch *FilePatch) AddedChunks() []Chunk {
	chunks := []Chunk{}
	for _, c := range filePatch.inner.Chunks() {
		if c.Type() == diff.Add {
			chunks = append(chunks, Chunk{
				content: c.Content(),
			})
		}
	}
	return chunks
}

type FileIter struct {
	innter object.FileIter
}

func (fileIter *FileIter) ForEach(f func(*ObjectFile) error) error {
	return fileIter.innter.ForEach(func(file *object.File) error {
		return f(&ObjectFile{
			inner: *file,
		})
	})
}

type ObjectFile struct {
	inner object.File
}

func (file *ObjectFile) IsBinary() (bool, error) {
	return file.inner.IsBinary()
}

func (file *ObjectFile) Contents() (string, error) {
	return file.inner.Contents()
}

func (file *ObjectFile) Name() string {
	return file.inner.Name
}

type Chunk struct {
	content string
}

func (chunk *Chunk) Content() string {
	return chunk.content
}

type DiffFile struct {
	path string
}

func (file *DiffFile) Path() string {
	return file.path
}
