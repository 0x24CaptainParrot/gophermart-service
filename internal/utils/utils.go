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

func IsValidOrderNumberLuhn(orderNumber int64) bool {
	s := strconv.FormatInt(orderNumber, 10)

	if len(s) < 5 {
		return false
	}

	sum := 0
	alternate := len(s)%2 == 0

	for i, c := range s {
		digit, err := strconv.Atoi(string(c))
		if err != nil {
			return false
		}

		if (i%2 == 0) == alternate {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}
		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
