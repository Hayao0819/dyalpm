package cmem

import (
	"fmt"
	"sync"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	loadOnce sync.Once
	loadErr  error
	calloc   func(uintptr, uintptr) uintptr
	free     func(uintptr)
)

func loadLibc() {
	libc, err := purego.Dlopen("libc.so.6", purego.RTLD_NOW)
	if err != nil {
		libc, err = purego.Dlopen("libc.so", purego.RTLD_NOW)
	}
	if err != nil {
		loadErr = fmt.Errorf("load libc: %w", err)
		return
	}

	callocSymbol, err := purego.Dlsym(libc, "calloc")
	if err != nil {
		loadErr = fmt.Errorf("resolve calloc: %w", err)
		return
	}
	freeSymbol, err := purego.Dlsym(libc, "free")
	if err != nil {
		loadErr = fmt.Errorf("resolve free: %w", err)
		return
	}

	purego.RegisterFunc(&calloc, callocSymbol)
	purego.RegisterFunc(&free, freeSymbol)
}

func Alloc(t testing.TB, size uintptr) uintptr {
	t.Helper()
	loadOnce.Do(loadLibc)
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	if size == 0 {
		size = 1
	}

	ptr := calloc(1, size)
	if ptr == 0 {
		t.Fatalf("calloc(1, %d) returned null", size)
	}
	t.Cleanup(func() {
		free(ptr)
	})
	return ptr
}

func Bytes(t testing.TB, value []byte) (uintptr, []byte) {
	t.Helper()
	ptr := Alloc(t, uintptr(len(value)))
	buffer := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), len(value))
	copy(buffer, value)
	return ptr, buffer
}

func String(t testing.TB, value string) uintptr {
	t.Helper()
	ptr := Alloc(t, uintptr(len(value)+1))
	copy(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), len(value)), value)
	return ptr
}
