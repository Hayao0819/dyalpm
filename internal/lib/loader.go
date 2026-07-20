package lib

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/ebitengine/purego"
)

const (
	libalpmSONAME   = "libalpm.so.16"
	libalpmABIMajor = "16"
)

type libraryLoader struct {
	mu         sync.Mutex
	handle     uintptr
	path       string
	open       func(string, int) (uintptr, error)
	close      func(uintptr) error
	initialize func(uintptr) error
}

var defaultLibraryLoader = &libraryLoader{
	open:       purego.Dlopen,
	close:      purego.Dlclose,
	initialize: initializeALPMLibrary,
}

func (l *libraryLoader) load(path string) error {
	if path == "" {
		return errors.New("libalpm path is empty")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.handle != 0 {
		if l.path != path {
			return fmt.Errorf("libalpm already loaded from %q", l.path)
		}
		return nil
	}
	return l.loadUnlocked(path)
}

func (l *libraryLoader) loadUnlocked(path string) error {
	handle, err := l.open(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("load libalpm from %q: %w", path, err)
	}
	if handle == 0 {
		return fmt.Errorf("load libalpm from %q: null handle", path)
	}

	if err := l.initialize(handle); err != nil {
		loadErr := fmt.Errorf("initialize libalpm from %q: %w", path, err)
		if closeErr := l.close(handle); closeErr != nil {
			return errors.Join(loadErr, fmt.Errorf("close rejected libalpm: %w", closeErr))
		}
		return loadErr
	}

	l.handle = handle
	l.path = path
	return nil
}

func (l *libraryLoader) ensure() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.handle != 0 {
		return nil
	}
	return l.loadUnlocked(libalpmSONAME)
}

func initializeALPMLibrary(handle uintptr) error {
	version, err := loadedALPMVersion(handle)
	if err != nil {
		return err
	}
	if err := validateALPMVersion(version); err != nil {
		return err
	}

	registerAlpmFuncs(handle)
	registerLibcFuncs()
	return nil
}

func loadedALPMVersion(handle uintptr) (string, error) {
	symbol, err := purego.Dlsym(handle, "alpm_version")
	if err != nil {
		return "", fmt.Errorf("resolve alpm_version: %w", err)
	}

	var versionFunc func() uintptr
	purego.RegisterFunc(&versionFunc, symbol)
	version := PtrToString(versionFunc())
	if version == "" {
		return "", errors.New("alpm_version returned an empty version")
	}
	return version, nil
}

func validateALPMVersion(version string) error {
	major, _, _ := strings.Cut(version, ".")
	if major != libalpmABIMajor {
		return fmt.Errorf("unsupported libalpm version %q: ABI major %s required", version, libalpmABIMajor)
	}
	return nil
}

// LoadLibrary loads and eagerly resolves a libalpm 16 library at path.
func LoadLibrary(path string) error {
	return defaultLibraryLoader.load(path)
}

// EnsureAlpmLoaded loads the exact libalpm 16 SONAME.
func EnsureAlpmLoaded() error {
	return defaultLibraryLoader.ensure()
}
