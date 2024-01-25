package testdata

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// RootPath returns the project root directory
func RootPath() string {
	a, err := filepath.Abs(filepath.Join(basepath, ".."))
	if err != nil {
		return ""
	}
	return a
}

// RelPath returns the absolute path relative to the project root.
func RelPath(p ...string) string {
	var a []string
	a = append(a, RootPath())
	a = append(a, p...)
	return filepath.Join(a...)
}
