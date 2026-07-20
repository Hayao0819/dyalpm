package lib

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestPtrToStringCopiesSource(t *testing.T) {
	source := []byte{'a', 'l', 'p', 'm', 0}
	got := PtrToString(uintptr(unsafe.Pointer(&source[0])))

	copy(source, "xxxx")
	runtime.KeepAlive(source)

	if got != "alpm" {
		t.Fatalf("PtrToString() = %q, want %q", got, "alpm")
	}
}

func TestPtrToStringWithLenCopiesSource(t *testing.T) {
	source := []byte{'a', 0, 'b'}
	got := PtrToStringWithLen(uintptr(unsafe.Pointer(&source[0])), len(source))

	clear(source)
	runtime.KeepAlive(source)

	if got != "a\x00b" {
		t.Fatalf("PtrToStringWithLen() = %q, want %q", got, "a\x00b")
	}
}

func TestPtrToStringWithLenRejectsInvalidInput(t *testing.T) {
	value := byte('a')
	if got := PtrToStringWithLen(uintptr(unsafe.Pointer(&value)), -1); got != "" {
		t.Fatalf("PtrToStringWithLen() = %q, want empty string", got)
	}
	if got := PtrToStringWithLen(0, 1); got != "" {
		t.Fatalf("PtrToStringWithLen() = %q, want empty string", got)
	}
}
