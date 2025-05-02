package utils

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
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

func IsValidOrderNum(num int64) bool {
	s := strconv.FormatInt(num, 10)
	sum := 0

	parity := len(s) % 2

	for i, c := range s {
		digit, err := strconv.Atoi(string(c))
		if err != nil {
			return false
		}

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}
		sum += digit
	}

	return sum%10 == 0
}
