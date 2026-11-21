package httpapi

import "strings"

func publicURL(path string) any {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	path = strings.TrimPrefix(path, "/storage/")
	path = strings.TrimPrefix(path, "storage/")
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return nil
	}
	return "/storage/" + path
}
