package git

import (
	"context"
	"errors"
	"os"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

type Info struct {
	URL    string
	Ref    string
	Commit string
}

func GetInfo(ctx context.Context) (Info, error) {
	ret := Info{}

	dir := "."
	for _, s := range []string{"GIT_DIR", "PROJECT_ROOT"} {
		if v := os.Getenv(s); v != "" {
			dir = v
			break
		}
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return ret, err
	}

	// Get remote "origin" URL
	remote, err := repo.Remote("origin")
	if err != nil {
		return ret, err
	}

	if urls := remote.Config().URLs; len(urls) > 0 {
		ret.URL = OriginURL(urls[0])
	}

	// Get HEAD reference to find the current branch or commit
	ref, err := repo.Head()
	if err != nil {
		return ret, err
	}

	ret.Ref = ref.Name().Short()
	ret.Commit = ref.Hash().String()

	if t, err := repo.Tags(); err == nil {
		_ = t.ForEach(func(reference *plumbing.Reference) error {
			if ref.Hash() == reference.Hash() {
				ret.Ref = reference.Name().Short()
				return errors.New("break")
			}

			return nil
		})
	}

	return ret, nil
}
