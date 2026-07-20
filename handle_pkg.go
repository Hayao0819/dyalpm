package dyalpm

import (
	stderrors "errors"

	alpmerrors "github.com/Hayao0819/dyalpm/errors"
	"github.com/Hayao0819/dyalpm/internal/lib"
)

func (h *handle) LoadPackage(filename string, full bool, siglevel int) (Package, error) {
	if h.ptr == 0 {
		return nil, alpmerrors.ErrHandleNull
	}

	if lib.AlpmPkgLoad == nil {
		return nil, stderrors.New("missing function: alpm_pkg_load")
	}

	var pkgPtr uintptr
	fullInt := int32(0)
	if full {
		fullInt = 1
	}
	siglevelInt32 := clampIntToInt32(siglevel)
	result := lib.AlpmPkgLoad(h.ptr, filename, fullInt, siglevelInt32, &pkgPtr)

	if result != 0 {
		return nil, alpmerrors.NewError(h.Errno(), "failed to load package")
	}

	if pkgPtr == 0 {
		return nil, ErrPackageLoadFailed
	}

	return newPackage(pkgPtr, h), nil
}
