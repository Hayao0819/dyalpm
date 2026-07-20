package dyalpm

import (
	"testing"

	"github.com/Jguer/dyalpm/internal/lib"
	"github.com/Jguer/dyalpm/internal/testutil/cmem"
)

func TestDependencyComputeStringCopiesAndFrees(t *testing.T) {
	sourcePtr, source := cmem.Bytes(t, []byte{'b', 'a', 's', 'h', 0})

	originalCompute := lib.AlpmDepComputeString
	originalFree := lib.LibcFree
	t.Cleanup(func() {
		lib.AlpmDepComputeString = originalCompute
		lib.LibcFree = originalFree
	})

	lib.AlpmDepComputeString = func(uintptr) uintptr {
		return sourcePtr
	}
	var freed uintptr
	lib.LibcFree = func(ptr uintptr) {
		freed = ptr
	}

	got := newDependency(1).ComputeString()
	copy(source, "xxxx")

	if got != "bash" {
		t.Fatalf("ComputeString() = %q, want %q", got, "bash")
	}
	if freed != sourcePtr {
		t.Fatalf("freed pointer = %#x, want %#x", freed, sourcePtr)
	}
}

func TestPackageSigCopiesAndFrees(t *testing.T) {
	sourcePtr, source := cmem.Bytes(t, []byte{1, 2, 3, 4})

	originalGetSig := lib.AlpmPkgGetSig
	originalFree := lib.LibcFree
	t.Cleanup(func() {
		lib.AlpmPkgGetSig = originalGetSig
		lib.LibcFree = originalFree
	})

	lib.AlpmPkgGetSig = func(_ uintptr, sig *uintptr, sigLen *uintptr) int32 {
		*sig = sourcePtr
		*sigLen = uintptr(len(source))
		return 0
	}
	var freed uintptr
	lib.LibcFree = func(ptr uintptr) {
		freed = ptr
	}

	got, err := (&package_{ptr: 1}).Sig()
	if err != nil {
		t.Fatalf("Sig() error = %v", err)
	}
	clear(source)

	if len(got) != 4 || got[0] != 1 || got[1] != 2 || got[2] != 3 || got[3] != 4 {
		t.Fatalf("Sig() = %v, want [1 2 3 4]", got)
	}
	if freed != sourcePtr {
		t.Fatalf("freed pointer = %#x, want %#x", freed, sourcePtr)
	}
}

func TestChangelogReaderZeroLengthRead(t *testing.T) {
	originalRead := lib.AlpmPkgChangelogRead
	t.Cleanup(func() {
		lib.AlpmPkgChangelogRead = originalRead
	})
	lib.AlpmPkgChangelogRead = nil

	reader := &changelogReader{pkg: &package_{ptr: 1}, fp: 1}
	n, err := reader.Read(nil)
	if n != 0 || err != nil {
		t.Fatalf("Read(nil) = (%d, %v), want (0, nil)", n, err)
	}
}

func TestChangelogReaderCloseReportsFailure(t *testing.T) {
	originalClose := lib.AlpmPkgChangelogClose
	t.Cleanup(func() {
		lib.AlpmPkgChangelogClose = originalClose
	})

	lib.AlpmPkgChangelogClose = func(uintptr, uintptr) int32 {
		return -1
	}
	reader := &changelogReader{pkg: &package_{ptr: 1}, fp: 2}

	if err := reader.Close(); err == nil {
		t.Fatal("Close() error = nil, want close failure")
	}
	if reader.fp != 0 {
		t.Fatalf("reader pointer = %#x after Close(), want 0", reader.fp)
	}
}
