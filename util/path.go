package util

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	Basepath   = filepath.Dir(b)
)

func GetRootFolderPath() string {
	dir, _ := filepath.Split(Basepath)
	return dir
}
