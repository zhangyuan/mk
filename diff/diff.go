package diff

import (
	"context"
	"time"

	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Chunk struct {
}

func fileContent(f *object.File) (content string, isBinary bool, err error) {
	if f == nil {
		return
	}

	isBinary, err = f.IsBinary()
	if err != nil || isBinary {
		return
	}

	content, err = f.Contents()

	return
}

type textChunk struct {
	content string
	op      fdiff.Operation
}

func (t *textChunk) Content() string {
	return t.content
}

func (t *textChunk) Type() fdiff.Operation {
	return t.op
}

func Diff(ctx context.Context, from, to *object.File) ([]fdiff.Chunk, error) {
	chunks := []fdiff.Chunk{}

	fromContent, fIsBinary, err := fileContent(from)
	if err != nil {
		return chunks, err
	}

	toContent, tIsBinary, err := fileContent(to)
	if err != nil {
		return chunks, err
	}

	if fIsBinary || tIsBinary {
		return chunks, nil
	}

	diffs := DiffWithTimeout(fromContent, toContent, time.Hour)

	for _, d := range diffs {
		select {
		case <-ctx.Done():
			return nil, object.ErrCanceled
		default:
		}

		var op fdiff.Operation
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			op = fdiff.Equal
		case diffmatchpatch.DiffDelete:
			op = fdiff.Delete
		case diffmatchpatch.DiffInsert:
			op = fdiff.Add
		}

		chunks = append(chunks, &textChunk{d.Text, op})
	}

	return chunks, nil
}

func DiffWithTimeout(src, dst string, timeout time.Duration) (diffs []diffmatchpatch.Diff) {
	dmp := diffmatchpatch.New()
	dmp.DiffTimeout = timeout
	wSrc, wDst, warray := dmp.DiffLinesToRunes(src, dst)
	diffs = dmp.DiffMainRunes(wSrc, wDst, false)
	diffs = dmp.DiffCharsToLines(diffs, warray)
	return diffs
}
