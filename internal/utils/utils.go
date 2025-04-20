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
	n := orderNumber
	if n > 0 {
		digits = append([]int64{n % 10}, digits...)
		n /= 10
	}

	if len(digits) < 2 {
		return false
	}

	workDigits := make([]int64, len(digits))
	copy(workDigits, digits)

	for i := len(workDigits) - 2; i >= 0; i -= 2 {
		workDigits[i] *= 2
		if workDigits[i] > 9 {
			workDigits[i] = workDigits[i]%10 + workDigits[i]/10
		}
	}

	var sum int64
	for _, d := range workDigits {
		sum += d
	}

	return sum%10 == 0
}
