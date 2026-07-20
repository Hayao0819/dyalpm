package dyalpm

import (
	"testing"
	"unsafe"

	alpmerrors "github.com/Jguer/dyalpm/errors"
	"github.com/Jguer/dyalpm/internal/lib"
	"github.com/Jguer/dyalpm/internal/testutil/cmem"
)

type transactionTestList struct {
	data uintptr
	_    uintptr
	_    uintptr
}

type transactionTestDepend struct {
	name    uintptr
	version uintptr
	_       uintptr
	_       uintptr
	mod     int32
	_       int32
}

type transactionTestMissing struct {
	target     uintptr
	depend     uintptr
	causingPkg uintptr
}

type transactionTestFileConflict struct {
	target  uintptr
	kind    int32
	_       int32
	file    uintptr
	ctarget uintptr
}

type transactionFreeLog struct {
	lists         []uintptr
	strings       []uintptr
	missing       []uintptr
	conflicts     []uintptr
	fileConflicts []uintptr
}

func installTransactionBindings(t *testing.T) *transactionFreeLog {
	t.Helper()
	oldPrepare := lib.AlpmTransPrepare
	oldCommit := lib.AlpmTransCommit
	oldErrno := lib.AlpmErrno
	oldListFree := lib.AlpmListFree
	oldFree := lib.LibcFree
	oldMissingFree := lib.AlpmDepmissingFree
	oldConflictFree := lib.AlpmConflictFree
	oldFileConflictFree := lib.AlpmFileConflictFree
	oldPkgGetName := lib.AlpmPkgGetName
	t.Cleanup(func() {
		lib.AlpmTransPrepare = oldPrepare
		lib.AlpmTransCommit = oldCommit
		lib.AlpmErrno = oldErrno
		lib.AlpmListFree = oldListFree
		lib.LibcFree = oldFree
		lib.AlpmDepmissingFree = oldMissingFree
		lib.AlpmConflictFree = oldConflictFree
		lib.AlpmFileConflictFree = oldFileConflictFree
		lib.AlpmPkgGetName = oldPkgGetName
	})

	log := &transactionFreeLog{}
	lib.AlpmListFree = func(ptr uintptr) {
		log.lists = append(log.lists, ptr)
	}
	lib.LibcFree = func(ptr uintptr) {
		log.strings = append(log.strings, ptr)
	}
	lib.AlpmDepmissingFree = func(ptr uintptr) {
		log.missing = append(log.missing, ptr)
	}
	lib.AlpmConflictFree = func(ptr uintptr) {
		log.conflicts = append(log.conflicts, ptr)
	}
	lib.AlpmFileConflictFree = func(ptr uintptr) {
		log.fileConflicts = append(log.fileConflicts, ptr)
	}
	return log
}

func stubPrepare(t *testing.T, list uintptr, errno alpmerrors.Errno) {
	t.Helper()
	lib.AlpmTransPrepare = func(_ uintptr, data *uintptr) int32 {
		*data = list
		return -1
	}
	lib.AlpmErrno = func(uintptr) int32 {
		return clampIntToInt32(int(errno))
	}
}

func stubCommit(t *testing.T, list uintptr, errno alpmerrors.Errno) {
	t.Helper()
	lib.AlpmTransCommit = func(_ uintptr, data *uintptr) int32 {
		*data = list
		return -1
	}
	lib.AlpmErrno = func(uintptr) int32 {
		return clampIntToInt32(int(errno))
	}
}

func transactionCString(t *testing.T, value string) ([]byte, uintptr) {
	t.Helper()
	ptr, buffer := cmem.Bytes(t, append([]byte(value), 0))
	return buffer, ptr
}

func transactionPointer[T any](t *testing.T, value *T) uintptr {
	t.Helper()
	ptr := cmem.Alloc(t, unsafe.Sizeof(*value))
	*(*T)(unsafe.Pointer(ptr)) = *value
	return ptr
}

func assertFreed(t *testing.T, values []uintptr, want uintptr) {
	t.Helper()
	if len(values) != 1 || values[0] != want {
		t.Fatalf("freed pointers = %#x, want [%#x]", values, want)
	}
}
