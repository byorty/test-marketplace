package rbac

import (
	"strings"

	"github.com/google/uuid"
)

func NormalizePath(path string) string {
	parts := strings.Split(path, "/")

	if len(parts) == 3 {
		if _, err := uuid.Parse(parts[2]); err == nil {
			return "/products/{id}"
		}
	}

	return path
}