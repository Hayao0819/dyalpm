package alpm

import (
	stderrors "errors"
	"unsafe"

	"github.com/Jguer/dyalpm/internal/lib"
)

// SigStatus represents the status of a signature
type SigStatus int

const (
	SigStatusValid SigStatus = iota
	SigStatusInvalid
	SigStatusSigExpired
	SigStatusKeyExpired
	SigStatusKeyUnknown
	SigStatusKeyDisabled
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

// Internal structure for alpm_siglist_t
type alpmSiglistT struct {
	Count   uintptr
	Results uintptr
}

// Internal structure for alpm_sigresult_t
type alpmSigresultT struct {
	KeyID    uintptr
	Status   int
	Validity int
}

func decodeSigList(ptr uintptr) SigList {
	if ptr == 0 {
		return SigList{}
	}

	base := unsafe.Pointer(ptr)
	sigList := (*alpmSiglistT)(base)
	if sigList.Count == 0 || sigList.Results == 0 {
		return SigList{}
	}

	var results []SigResult
	resultSize := unsafe.Sizeof(alpmSigresultT{})
	resultsBase := unsafe.Pointer(sigList.Results)

	for i := uintptr(0); i < sigList.Count; i++ {
		res := (*alpmSigresultT)(unsafe.Add(resultsBase, i*resultSize))

		results = append(results, SigResult{
			KeyID:    lib.PtrToString(res.KeyID),
			Status:   SigStatus(res.Status),
			Validity: SigValidity(res.Validity),
		})
	}

	return SigList{Results: results}
}

func checkPGPSignature(ptr uintptr, registry *lib.FunctionRegistry, handle *handle, funcName string) (SigList, error) {
	fn, err := registry.GetFunc(funcName)
	if err != nil {
		return SigList{}, err
	}

	var sigList alpmSiglistT
	sigListPtr := uintptr(unsafe.Pointer(&sigList))

	result := lib.Syscall(fn, ptr, sigListPtr)
	decoded := decodeSigList(sigListPtr)

	// We should clean up the siglist after use
	cleanupErr := handle.SigListCleanup(sigListPtr)

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

// SigListCleanup cleans up an ALPM signature list
func (h *handle) SigListCleanup(siglistPtr uintptr) error {
	if h.ptr == 0 {
		return ErrInvalidHandle
	}

	fn, err := h.registry.GetFunc("alpm_siglist_cleanup")
	if err != nil {
		return err
	}

	lib.Syscall(fn, siglistPtr)
	return nil
}
