package lib

import (
	"testing"

	"github.com/Jguer/dyalpm/internal/testutil/cmem"
)

func TestPtrToStringCopiesSource(t *testing.T) {
	sourcePtr, source := cmem.Bytes(t, []byte{'a', 'l', 'p', 'm', 0})
	got := PtrToString(sourcePtr)

	copy(source, "xxxx")

	if got != "alpm" {
		t.Fatalf("PtrToString() = %q, want %q", got, "alpm")
	}
}

func TestPtrToStringWithLenCopiesSource(t *testing.T) {
	sourcePtr, source := cmem.Bytes(t, []byte{'a', 0, 'b'})
	got := PtrToStringWithLen(sourcePtr, len(source))

	clear(source)

	if got != "a\x00b" {
		t.Fatalf("PtrToStringWithLen() = %q, want %q", got, "a\x00b")
	}
}

func TestPtrToStringWithLenRejectsInvalidInput(t *testing.T) {
	value, _ := cmem.Bytes(t, []byte{'a'})
	if got := PtrToStringWithLen(value, -1); got != "" {
		t.Fatalf("PtrToStringWithLen() = %q, want empty string", got)
	}
	if got := PtrToStringWithLen(0, 1); got != "" {
		t.Fatalf("PtrToStringWithLen() = %q, want empty string", got)
	}
}
