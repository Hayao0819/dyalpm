package dyalpm

import (
	stderrors "errors"
	"runtime"
	"strings"
	"unsafe"

	"github.com/Jguer/dyalpm/internal/lib"
)

// SigStatus represents the status of a signature
type SigStatus int

const (
	SigStatusValid SigStatus = iota
	SigStatusKeyExpired
	SigStatusSigExpired
	SigStatusKeyUnknown
	SigStatusKeyDisabled
	SigStatusInvalid

	// Deprecated: libalpm does not report a distinct revoked-key status.
	SigStatusKeyRevoked
)

// SigValidity represents the validity of a signature
type SigValidity int

const (
	SigValidityFull SigValidity = iota
	SigValidityMarginal
	SigValidityNever
	SigValidityUnknown
)

// SigResult represents a single signature result
type SigResult struct {
	KeyID    string
	Status   SigStatus
	Validity SigValidity
}

// SigList represents a list of signature results
type SigList struct {
	Results []SigResult
}

type alpmSigList struct {
	Count   uintptr
	Results uintptr
}

type alpmPGPKey struct {
	Data        uintptr
	Fingerprint uintptr
	UID         uintptr
	Name        uintptr
	Email       uintptr
	Created     int64
	Expires     int64
	Length      uint32
	Revoked     uint32
}

type alpmSigResult struct {
	Key      alpmPGPKey
	Status   int32
	Validity int32
}

func decodeSigList(sigList *alpmSigList) SigList {
	if sigList == nil {
		return SigList{}
	}

	if sigList.Count == 0 || sigList.Results == 0 {
		return SigList{}
	}

	var results []SigResult
	resultSize := unsafe.Sizeof(alpmSigResult{})
	resultsBase := unsafe.Pointer(sigList.Results)

	for i := uintptr(0); i < sigList.Count; i++ {
		res := (*alpmSigResult)(unsafe.Add(resultsBase, i*resultSize))

		results = append(results, SigResult{
			KeyID:    strings.Clone(lib.PtrToString(res.Key.Fingerprint)),
			Status:   SigStatus(res.Status),
			Validity: SigValidity(res.Validity),
		})
	}

	return SigList{Results: results}
}

func checkPGPSignature(ptr uintptr, handle *handle, funcName string) (SigList, error) {
	var fn func(uintptr, unsafe.Pointer) int32
	switch funcName {
	case "alpm_db_check_pgp_signature":
		fn = lib.AlpmDBCheckPGPSignature
	case "alpm_pkg_check_pgp_signature":
		fn = lib.AlpmPkgCheckPGPSignature
	default:
		return SigList{}, stderrors.New("missing function: " + funcName)
	}

	if fn == nil {
		return SigList{}, stderrors.New("missing function: " + funcName)
	}

	var sigList alpmSigList
	sigListPtr := unsafe.Pointer(&sigList)

	result := fn(ptr, sigListPtr)
	decoded := decodeSigList(&sigList)

	cleanupErr := handle.SigListCleanup(sigListPtr)
	runtime.KeepAlive(&sigList)

	if result != 0 {
		err := stderrors.New("signature check failed")
		if cleanupErr != nil {
			return decoded, stderrors.Join(err, cleanupErr)
		}
		return decoded, err
	}
	if cleanupErr != nil {
		return decoded, cleanupErr
	}

	return decoded, nil
}

func (h *handle) SigListCleanup(siglistPtr unsafe.Pointer) error {
	if h.ptr == 0 {
		return ErrInvalidHandle
	}
	if lib.AlpmSiglistCleanup == nil {
		return stderrors.New("missing function: alpm_siglist_cleanup")
	}

	if lib.AlpmSiglistCleanup(siglistPtr) != 0 {
		return stderrors.New("failed to clean up signature list")
	}
	return nil
}
