package utils

import (
	"fmt"
	"strings"
)

type Breadcrumb struct {
	Label string
	Path  string
}

// BuildBreadcrumbs builds breadcrumb for any view
func BuildBreadcrumbs(path string) []Breadcrumb {
	breadcrumbs := []Breadcrumb{}

	// base path
	breadcrumbs = append(breadcrumbs, Breadcrumb{
		Label: "Home",
		Path:  "/",
	})

	if path != "" {
		pathSplit := strings.Split(path, "/")
		for index, object := range pathSplit {
			if object != "" {
				breadcrumbs = append(breadcrumbs, Breadcrumb{
					Label: object,
					Path:  fmt.Sprintf("?path=%s", strings.Join(pathSplit[0:index+1], "/")),
				})
			}
		}
	}

	return breadcrumbs
}
