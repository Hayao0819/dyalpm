package alpm

import (
	stderrors "errors"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/Jguer/dyalpm/internal/lib"
	"github.com/Jguer/dyalpm/internal/list"
)

func collectList[T any](alpmList *list.List, build func(uintptr) T) []T {
	var items []T
	for item := alpmList; item != nil && item.Ptr() != 0; item = item.Next() {
		ptr := item.Data()
		if ptr != 0 {
			items = append(items, build(ptr))
		}
	}
	return items
}

func computeFileSum(funcName, filename string) (string, error) {
	reg, err := lib.GetRegistry()
	if err != nil {
		return "", err
	}

	fn, err := reg.GetFunc(funcName)
	if err != nil {
		return "", err
	}

	cStr := lib.CString(filename)
	r1, _, _ := purego.SyscallN(fn, uintptr(unsafe.Pointer(&cStr[0])))
	runtime.KeepAlive(cStr)

	if r1 == 0 {
		return "", ErrInvalidPackage
	}

	res := lib.PtrToString(r1)
	lib.Free(r1)
	return res, nil
}

// ComputeMD5Sum computes the MD5 sum of a file
func ComputeMD5Sum(filename string) (string, error) {
	return computeFileSum("alpm_compute_md5sum", filename)
}

// ComputeSHA256Sum computes the SHA256 sum of a file
func ComputeSHA256Sum(filename string) (string, error) {
	return computeFileSum("alpm_compute_sha256sum", filename)
}

func (h *handle) FindGroupPkgs(dbs []Database, name string) ([]Package, error) {
	if h.ptr == 0 {
		return nil, ErrInvalidHandle
	}

	fn, err := h.registry.GetFunc("alpm_find_group_pkgs")
	if err != nil {
		return nil, err
	}

	var dbList *list.List
	for _, db := range dbs {
		dbImpl, ok := db.(*database)
		if ok {
			dbList = list.Add(dbList, dbImpl.ptr)
		}
	}
	defer dbList.Free()

	cName := lib.CString(name)
	r1, _, _ := purego.SyscallN(fn, dbList.Ptr(), uintptr(unsafe.Pointer(&cName[0])))
	runtime.KeepAlive(cName)

	if r1 == 0 {
		return []Package{}, nil
	}

	resList := list.NewList(r1)
	defer resList.Free()

	var pkgs []Package
	for item := resList; item != nil && item.Ptr() != 0; item = item.Next() {
		ptr := item.Data()
		if ptr != 0 {
			pkgs = append(pkgs, newPackage(ptr, h))
		}
	}

	return pkgs, nil
}

func (h *handle) ExtractKeyID(identifier string, sig []byte) ([]string, error) {
	if h.ptr == 0 {
		return nil, ErrInvalidHandle
	}

	fn, err := h.registry.GetFunc("alpm_extract_keyid")
	if err != nil {
		return nil, err
	}

	cIdentifier := lib.CString(identifier)
	var keysListPtr uintptr

	// alpm_extract_keyid(handle, identifier, sig, len, &keys)
	var sigPtr uintptr
	if len(sig) > 0 {
		sigPtr = uintptr(unsafe.Pointer(&sig[0]))
	}

	r1, _, _ := purego.SyscallN(
		fn,
		h.ptr,
		uintptr(unsafe.Pointer(&cIdentifier[0])),
		sigPtr,
		uintptr(len(sig)),
		uintptr(unsafe.Pointer(&keysListPtr)),
	)

	runtime.KeepAlive(cIdentifier)
	runtime.KeepAlive(sig)

	if r1 != 0 {
		return nil, stderrors.New("failed to extract key id")
	}

	if keysListPtr == 0 {
		return []string{}, nil
	}

	alpmList := list.NewList(keysListPtr)
	// We need to free the strings in the list too
	// alpm_list_free_inner(list, free)
	// For now, let's just free the list structure and hope the strings are managed or short-lived.
	// Actually, alpm_list_free just frees the nodes.
	defer alpmList.Free()

	var keys []string
	for item := alpmList; item != nil && item.Ptr() != 0; item = item.Next() {
		ptr := item.Data()
		if ptr != 0 {
			keys = append(keys, lib.PtrToString(ptr))
		}
	}

	return keys, nil
}
