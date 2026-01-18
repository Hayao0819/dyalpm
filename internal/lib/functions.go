package lib

import (
	"errors"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// FunctionRegistry manages lazy loading of ALPM functions
type FunctionRegistry struct {
	lib   uintptr
	mu    sync.RWMutex
	funcs map[string]uintptr
}

var (
	registryOnce sync.Once
	registry     *FunctionRegistry
	registryErr  error
)

// GetRegistry returns the global function registry
func GetRegistry() (*FunctionRegistry, error) {
	registryOnce.Do(func() {
		lib, loadErr := getLibrary()
		if loadErr != nil {
			registryErr = loadErr
			return
		}
		registry = &FunctionRegistry{
			lib:   lib,
			funcs: make(map[string]uintptr),
		}
	})
	return registry, registryErr
}

// GetFunc lazily loads and returns a function pointer
func (r *FunctionRegistry) GetFunc(name string) (uintptr, error) {
	r.mu.RLock()
	if fn, ok := r.funcs[name]; ok {
		r.mu.RUnlock()
		return fn, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double check after acquiring write lock
	if fn, ok := r.funcs[name]; ok {
		return fn, nil
	}

	fnPtr, err := purego.Dlsym(r.lib, name)
	if err != nil {
		return 0, err
	}

	r.funcs[name] = fnPtr
	return fnPtr, nil
}

// CallFunc calls a function with the given arguments using purego.SyscallN
func (r *FunctionRegistry) CallFunc(name string, args ...uintptr) (uintptr, error) {
	fn, err := r.GetFunc(name)
	if err != nil {
		return 0, err
	}
	r1, _, errno := purego.SyscallN(fn, args...)
	if errno != 0 {
		return 0, errors.New("syscall failed")
	}
	return r1, nil
}

// Helper functions for common conversions

// CString converts a Go string to a null-terminated byte slice.
// The caller must keep the slice alive during the C call and use &buf[0] as the pointer.
func CString(s string) []byte {
	if s == "" {
		return []byte{0}
	}

	buf := make([]byte, len(s)+1)
	copy(buf, s)
	buf[len(s)] = 0
	return buf
}

// PtrToString converts a C string pointer to a Go string
// This function finds the null terminator to determine the length
func PtrToString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	base := unsafe.Pointer(ptr)
	start := (*byte)(base)
	if start == nil {
		return ""
	}
	// Find null terminator
	n := 0
	for *(*byte)(unsafe.Add(base, n)) != 0 {
		n++
		if n > 1024*1024 { // Safety limit
			return ""
		}
	}
	if n == 0 {
		return ""
	}
	return unsafe.String(start, n)
}

// PtrToStringWithLen converts a C string pointer with known length to Go string
func PtrToStringWithLen(ptr uintptr, length int) string {
	if ptr == 0 || length == 0 {
		return ""
	}
	base := unsafe.Pointer(ptr)
	return unsafe.String((*byte)(base), length)
}

// BoolToInt converts a Go bool to C int (0 or 1)
func BoolToInt(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}

// IntToBool converts a C int to Go bool
func IntToBool(i uintptr) bool {
	return i != 0
}
