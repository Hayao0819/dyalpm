//go:build linux

package dyalpm

import (
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

type abiListNode struct {
	Data uintptr
	Prev uintptr
	Next uintptr
}

type abiAllocator struct {
	handle uintptr
	calloc func(uintptr, uintptr) uintptr
	free   func(uintptr)
}

func newABIAllocator(t *testing.T) *abiAllocator {
	t.Helper()

	handle, err := purego.Dlopen("libc.so.6", purego.RTLD_NOW)
	if err != nil {
		handle, err = purego.Dlopen("libc.so", purego.RTLD_NOW)
		if err != nil {
			t.Fatalf("load libc: %v", err)
		}
	}

	var calloc func(uintptr, uintptr) uintptr
	symbol, err := purego.Dlsym(handle, "calloc")
	if err != nil {
		t.Fatalf("resolve calloc: %v", err)
	}
	purego.RegisterFunc(&calloc, symbol)

	var free func(uintptr)
	symbol, err = purego.Dlsym(handle, "free")
	if err != nil {
		t.Fatalf("resolve free: %v", err)
	}
	purego.RegisterFunc(&free, symbol)

	allocator := &abiAllocator{handle: handle, calloc: calloc, free: free}
	t.Cleanup(func() {
		if err := purego.Dlclose(allocator.handle); err != nil {
			t.Errorf("close libc: %v", err)
		}
	})
	return allocator
}

func (a *abiAllocator) alloc(t *testing.T, size uintptr) uintptr {
	t.Helper()

	ptr := a.calloc(1, size)
	if ptr == 0 {
		t.Fatal("calloc returned nil")
	}
	t.Cleanup(func() {
		a.free(ptr)
	})
	return ptr
}

func (a *abiAllocator) string(t *testing.T, value string) uintptr {
	t.Helper()

	ptr := a.alloc(t, uintptr(len(value)+1))
	copy(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), len(value)), value)
	return ptr
}

func TestAlpmStructureLayouts(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("layout assertions target supported 64-bit Linux platforms")
	}

	var depend alpmDepend
	if unsafe.Offsetof(depend.Name) != 0 ||
		unsafe.Offsetof(depend.Version) != 8 ||
		unsafe.Offsetof(depend.Description) != 16 ||
		unsafe.Offsetof(depend.NameHash) != 24 ||
		unsafe.Offsetof(depend.Mod) != 32 ||
		unsafe.Sizeof(depend) != 40 {
		t.Fatalf("unexpected alpm_depend_t layout: size=%d", unsafe.Sizeof(depend))
	}

	var conflict alpmConflict
	if unsafe.Offsetof(conflict.Package1) != 0 ||
		unsafe.Offsetof(conflict.Package2) != 8 ||
		unsafe.Offsetof(conflict.Reason) != 16 ||
		unsafe.Sizeof(conflict) != 24 {
		t.Fatalf("unexpected alpm_conflict_t layout: size=%d", unsafe.Sizeof(conflict))
	}

	var fileConflict alpmFileConflict
	if unsafe.Offsetof(fileConflict.Target) != 0 ||
		unsafe.Offsetof(fileConflict.Type) != 8 ||
		unsafe.Offsetof(fileConflict.File) != 16 ||
		unsafe.Offsetof(fileConflict.CTarget) != 24 ||
		unsafe.Sizeof(fileConflict) != 32 {
		t.Fatalf("unexpected alpm_fileconflict_t layout: size=%d", unsafe.Sizeof(fileConflict))
	}

	var key alpmPGPKey
	if unsafe.Offsetof(key.Fingerprint) != 8 ||
		unsafe.Offsetof(key.Created) != 40 ||
		unsafe.Offsetof(key.Expires) != 48 ||
		unsafe.Offsetof(key.Length) != 56 ||
		unsafe.Offsetof(key.Revoked) != 60 ||
		unsafe.Sizeof(key) != 64 {
		t.Fatalf("unexpected alpm_pgpkey_t layout: size=%d", unsafe.Sizeof(key))
	}

	var signature alpmSigResult
	if unsafe.Offsetof(signature.Status) != 64 ||
		unsafe.Offsetof(signature.Validity) != 68 ||
		unsafe.Sizeof(signature) != 72 {
		t.Fatalf("unexpected alpm_sigresult_t layout: size=%d", unsafe.Sizeof(signature))
	}

	var xdata alpmPackageXData
	if unsafe.Offsetof(xdata.Name) != 0 ||
		unsafe.Offsetof(xdata.Value) != 8 ||
		unsafe.Sizeof(xdata) != 16 {
		t.Fatalf("unexpected alpm_pkg_xdata_t layout: size=%d", unsafe.Sizeof(xdata))
	}
}
