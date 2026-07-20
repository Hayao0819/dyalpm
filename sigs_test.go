package dyalpm

import (
	"testing"
	"unsafe"

	"github.com/Hayao0819/dyalpm/internal/lib"
)

func TestCheckPGPSignatureKeepsOutputPointer(t *testing.T) {
	previousCheck := lib.AlpmPkgCheckPGPSignature
	previousCleanup := lib.AlpmSiglistCleanup
	t.Cleanup(func() {
		lib.AlpmPkgCheckPGPSignature = previousCheck
		lib.AlpmSiglistCleanup = previousCleanup
	})

	var output unsafe.Pointer
	lib.AlpmPkgCheckPGPSignature = func(_ uintptr, siglist unsafe.Pointer) int32 {
		output = siglist
		return 0
	}
	lib.AlpmSiglistCleanup = func(siglist unsafe.Pointer) int32 {
		if siglist != output {
			t.Errorf("cleanup pointer = %p, check pointer = %p", siglist, output)
		}
		return 0
	}

	sigList, err := checkPGPSignature(
		1,
		&handle{ptr: 1},
		"alpm_pkg_check_pgp_signature",
	)
	if err != nil {
		t.Fatalf("checkPGPSignature() error = %v", err)
	}
	if len(sigList.Results) != 0 {
		t.Fatalf("checkPGPSignature() results = %v, want empty", sigList.Results)
	}
}
