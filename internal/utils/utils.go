package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetMigrationsPath() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working dir for: %v", err)
	}

	root := "gophermart-service"
	idx := strings.LastIndex(wd, root)
	if idx == -1 {
		log.Fatalf("project root '%s' not found in: %s", root, wd)
	}

	rootPath := wd[:idx+len(root)]
	return filepath.Join(rootPath, "internal", "pkg", "repository", "schema")
}
