package utils

import "strings"

// BuildObjectName builds the object name based on provided object path
func BuildObjectName(path string) string {
	objects := strings.Split(path, "/")

	return objects[len(objects)-1]
}
