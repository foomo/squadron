package util

import (
	"os"
	"path"
	"path/filepath"
)

func ValidatePath(wd string, p *string) error {
	if !filepath.IsAbs(*p) {
		*p = path.Join(wd, *p)
	}
	absPath, err := filepath.Abs(*p)
	if err != nil {
		return err
	}
	_, err = os.Stat(absPath)
	if err != nil {
		return err
	}
	*p = absPath
	return nil
}
