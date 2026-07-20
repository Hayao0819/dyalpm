package dyalpm

import (
	"testing"

	"github.com/Hayao0819/dyalpm/internal/lib"
)

func TestHandleReleaseInvalidatesHandleWithoutReadingErrno(t *testing.T) {
	oldRelease := lib.AlpmRelease
	oldErrno := lib.AlpmErrno
	t.Cleanup(func() {
		lib.AlpmRelease = oldRelease
		lib.AlpmErrno = oldErrno
	})

	const handlePtr = uintptr(0x1234)
	released := false
	lib.AlpmRelease = func(ptr uintptr) int32 {
		if ptr != handlePtr {
			t.Fatalf("alpm_release called with %#x, want %#x", ptr, handlePtr)
		}
		released = true
		return 0
	}
	lib.AlpmErrno = func(uintptr) int32 {
		t.Fatal("alpm_errno must not be called after alpm_release")
		return 0
	}

	getOrCreateCallbackSet(handlePtr)
	h := &handle{ptr: handlePtr}
	if err := h.Release(); err != nil {
		t.Fatalf("Release() error = %v", err)
	}
	if !released {
		t.Fatal("alpm_release was not called")
	}
	if h.ptr != 0 {
		t.Fatalf("handle pointer = %#x after release, want 0", h.ptr)
	}
	assertCallbackSetAbsent(t, handlePtr)

	if err := h.Release(); err == nil {
		t.Fatal("second Release() unexpectedly succeeded")
	}
}

func TestHandleReleaseFailureStillInvalidatesHandle(t *testing.T) {
	oldRelease := lib.AlpmRelease
	oldErrno := lib.AlpmErrno
	t.Cleanup(func() {
		lib.AlpmRelease = oldRelease
		lib.AlpmErrno = oldErrno
	})

	const handlePtr = uintptr(0x5678)
	lib.AlpmRelease = func(uintptr) int32 { return -1 }
	lib.AlpmErrno = func(uintptr) int32 {
		t.Fatal("alpm_errno must not be called after a failed alpm_release")
		return 0
	}

	getOrCreateCallbackSet(handlePtr)
	h := &handle{ptr: handlePtr}
	if err := h.Release(); err == nil {
		t.Fatal("Release() unexpectedly succeeded")
	}
	if h.ptr != 0 {
		t.Fatalf("handle pointer = %#x after failed release, want 0", h.ptr)
	}
	assertCallbackSetAbsent(t, handlePtr)
}

func TestHandleReleaseMissingFunctionKeepsHandleValid(t *testing.T) {
	oldRelease := lib.AlpmRelease
	t.Cleanup(func() {
		lib.AlpmRelease = oldRelease
	})

	const handlePtr = uintptr(0x9abc)
	lib.AlpmRelease = nil

	getOrCreateCallbackSet(handlePtr)
	t.Cleanup(func() { unregisterCallbackSet(handlePtr) })
	h := &handle{ptr: handlePtr}
	if err := h.Release(); err == nil {
		t.Fatal("Release() unexpectedly succeeded")
	}
	if h.ptr != handlePtr {
		t.Fatalf("handle pointer = %#x when release is unavailable, want %#x", h.ptr, handlePtr)
	}
}

func assertCallbackSetAbsent(t *testing.T, key uintptr) {
	t.Helper()

	callbackSetsMu.RLock()
	_, exists := callbackSets[key]
	callbackSetsMu.RUnlock()
	if exists {
		t.Fatalf("callback set for %#x was not removed", key)
	}
}
