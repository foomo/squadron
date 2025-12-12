package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/foomo/squadron/cmd/actions"
	"github.com/spf13/cobra/doc"
)

const fmTemplate = `---
title: "%s"
---
# Squadron CLI Reference
`

func main() {
	outputDir := "./docs/reference/cli"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	c := actions.NewRoot()
	c.DisableAutoGenTag = true

	filePrepender := func(filename string) string {
		name := filepath.Base(filename)
		name = strings.TrimSuffix(name, path.Ext(name))

		return fmt.Sprintf(fmTemplate, strings.ReplaceAll(name, "_", " "))
	}
	linkHandler := func(s string) string {
		return "/reference/cli/" + strings.TrimSuffix(s, ".md") + ".html"
	}

	err := doc.GenMarkdownTreeCustom(c, outputDir, filePrepender, linkHandler)
	if err != nil {
		log.Fatal(err)
	}
}
