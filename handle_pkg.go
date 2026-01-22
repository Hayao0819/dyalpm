package dyalpm

import (
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/Jguer/dyalpm/internal/dyerrors"
	"github.com/Jguer/dyalpm/internal/lib"
)

func (h *handle) LoadPackage(filename string, full bool, siglevel int) (Package, error) {
	if h.ptr == 0 {
		return nil, dyerrors.ErrHandleNull
	}

	fn, err := h.registry.GetFunc("alpm_pkg_load")
	if err != nil {
		return nil, err
	}

	cFilename := lib.CString(filename)
	filenamePtr := uintptr(unsafe.Pointer(&cFilename[0]))

	var pkgPtr uintptr

	r1, _, _ := purego.SyscallN(fn, h.ptr, filenamePtr, lib.BoolToInt(full), uintptr(siglevel), uintptr(unsafe.Pointer(&pkgPtr)))
	runtime.KeepAlive(cFilename)

	if r1 != 0 {
		return nil, dyerrors.NewError(h.Errno(), "failed to load package")
	}

	if pkgPtr == 0 {
		return nil, ErrPackageLoadFailed
	}

	return newPackage(pkgPtr, h), nil
}
