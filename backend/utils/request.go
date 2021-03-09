package utils

import "fmt"

func GetCurrentPath(path string) string {
	if path != "" {
		path = fmt.Sprintf("%s/", path)
	}

	return path
}
