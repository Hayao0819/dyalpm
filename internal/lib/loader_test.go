package lib

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ebitengine/purego"
)

func newTestLibraryLoader(open func(string, int) (uintptr, error)) *libraryLoader {
	return &libraryLoader{
		open: open,
		close: func(uintptr) error {
			return nil
		},
		initialize: func(uintptr) error {
			return nil
		},
	}
}

func TestLibraryLoaderEnsureUsesExactSONAME(t *testing.T) {
	var path string
	var mode int
	loader := newTestLibraryLoader(func(gotPath string, gotMode int) (uintptr, error) {
		path = gotPath
		mode = gotMode
		return 1, nil
	})

	if err := loader.ensure(); err != nil {
		t.Fatalf("ensure() error = %v", err)
	}
	if path != libalpmSONAME {
		t.Fatalf("opened path = %q, want %q", path, libalpmSONAME)
	}
	wantMode := purego.RTLD_NOW | purego.RTLD_GLOBAL
	if mode != wantMode {
		t.Fatalf("open mode = %d, want %d", mode, wantMode)
	}
}

func TestLibraryLoaderExplicitPathDoesNotFallback(t *testing.T) {
	wantErr := errors.New("not found")
	var paths []string
	loader := newTestLibraryLoader(func(path string, _ int) (uintptr, error) {
		paths = append(paths, path)
		return 0, wantErr
	})

	err := loader.load("/opt/lib/libalpm.so.16")
	if !errors.Is(err, wantErr) {
		t.Fatalf("load() error = %v, want %v", err, wantErr)
	}
	if len(paths) != 1 || paths[0] != "/opt/lib/libalpm.so.16" {
		t.Fatalf("opened paths = %v", paths)
	}
}

func TestLibraryLoaderRetriesAfterFailure(t *testing.T) {
	wantErr := errors.New("temporary failure")
	attempts := 0
	loader := newTestLibraryLoader(func(string, int) (uintptr, error) {
		attempts++
		if attempts == 1 {
			return 0, wantErr
		}
		return 2, nil
	})

	if err := loader.ensure(); !errors.Is(err, wantErr) {
		t.Fatalf("first ensure() error = %v, want %v", err, wantErr)
	}
	if err := loader.ensure(); err != nil {
		t.Fatalf("second ensure() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("open attempts = %d, want 2", attempts)
	}
}

func TestLibraryLoaderRetriesAfterInitializationFailure(t *testing.T) {
	wantErr := errors.New("unsupported ABI")
	var opens int
	var initializations int
	var closed uintptr
	loader := newTestLibraryLoader(func(string, int) (uintptr, error) {
		opens++
		if opens == 1 {
			return 1, nil
		}
		return 2, nil
	})
	loader.initialize = func(uintptr) error {
		initializations++
		if initializations == 1 {
			return wantErr
		}
		return nil
	}
	loader.close = func(handle uintptr) error {
		closed = handle
		return nil
	}

	if err := loader.ensure(); !errors.Is(err, wantErr) {
		t.Fatalf("first ensure() error = %v, want %v", err, wantErr)
	}
	if err := loader.ensure(); err != nil {
		t.Fatalf("second ensure() error = %v", err)
	}
	if opens != 2 || initializations != 2 || closed != 1 {
		t.Fatalf("opens = %d, initializations = %d, closed = %d", opens, initializations, closed)
	}
}

func TestLibraryLoaderConcurrentLoad(t *testing.T) {
	const workers = 32
	var opens atomic.Int32
	var initializations atomic.Int32
	loader := newTestLibraryLoader(func(string, int) (uintptr, error) {
		opens.Add(1)
		return 1, nil
	})
	loader.initialize = func(uintptr) error {
		initializations.Add(1)
		return nil
	}

	start := make(chan struct{})
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			<-start
			errs <- loader.ensure()
		})
	}
	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Errorf("ensure() error = %v", err)
		}
	}
	if opens.Load() != 1 || initializations.Load() != 1 {
		t.Fatalf("opens = %d, initializations = %d", opens.Load(), initializations.Load())
	}
}

func TestLibraryLoaderRejectsDifferentPathAfterLoad(t *testing.T) {
	var opens int
	loader := newTestLibraryLoader(func(string, int) (uintptr, error) {
		opens++
		return 1, nil
	})

	if err := loader.load("first.so"); err != nil {
		t.Fatalf("first load() error = %v", err)
	}
	if err := loader.load("second.so"); err == nil {
		t.Fatal("second load() error = nil")
	}
	if opens != 1 {
		t.Fatalf("open calls = %d, want 1", opens)
	}
}

func TestLibraryLoaderEnsureAcceptsExplicitLoad(t *testing.T) {
	var opens int
	loader := newTestLibraryLoader(func(string, int) (uintptr, error) {
		opens++
		return 1, nil
	})

	if err := loader.load("/opt/lib/libalpm.so.16"); err != nil {
		t.Fatalf("load() error = %v", err)
	}
	if err := loader.ensure(); err != nil {
		t.Fatalf("ensure() error = %v", err)
	}
	if opens != 1 {
		t.Fatalf("open calls = %d, want 1", opens)
	}
}

func TestValidateALPMVersion(t *testing.T) {
	tests := []struct {
		version string
		wantErr bool
	}{
		{version: "16.0.1"},
		{version: "16.1"},
		{version: "", wantErr: true},
		{version: "15.0.0", wantErr: true},
		{version: "17.0.0", wantErr: true},
		{version: "16-git", wantErr: true},
	}
	for _, tt := range tests {
		err := validateALPMVersion(tt.version)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateALPMVersion(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
		}
	}
}
