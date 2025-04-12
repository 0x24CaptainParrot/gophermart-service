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

func IsValidOrderNumberLuhn(orderNumber int64) bool {
	digits := make([]int64, 0)
	for v := range orderNumber {
		digits = append(digits, v)
	}

	if len(digits) < 2 {
		return false
	}

	var sum int64
	for i := len(digits) - 2; i >= 0; i -= 2 {
		digits[i] *= 2
		if digits[i] > 9 {
			digits[i] = digits[i]%10 + digits[i]/10
		}
	}

	for _, d := range digits {
		sum += d
	}

	return sum%10 == 0
}
