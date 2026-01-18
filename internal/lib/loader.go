package lib

import (
	"errors"
	"sync"

	"github.com/ebitengine/purego"
)

const (
	libalpmPath         = "libalpm.so.16"
	libalpmPathFallback = "libalpm.so"
)

var (
	libalpm     uintptr
	libalpmOnce sync.Once
	libalpmErr  error
)

// LoadLibrary loads the ALPM library. It's safe to call multiple times.
func LoadLibrary(libalpmPath string) error {
	libalpmOnce.Do(func() {
		handle, err := purego.Dlopen(libalpmPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			libalpmErr = err
			return
		}
		libalpm = handle
	})
	return libalpmErr
}

// getLibrary returns the loaded library handle. It will attempt to load if not already loaded.
func getLibrary() (uintptr, error) {
	if err := LoadLibrary(libalpmPath); err != nil {
		if err := LoadLibrary(libalpmPathFallback); err != nil {
			return 0, err
		}
		return 0, err
	}
	if libalpm == 0 {
		return 0, errors.New("library not loaded")
	}
	return libalpm, nil
}
