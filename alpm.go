package dyalpm

import "github.com/Jguer/dyalpm/internal/lib"

// Version returns the library version string
func Version() string {
	reg, err := lib.GetRegistry()
	if err != nil {
		return ""
	}

	versionFn, err := reg.GetFunc("alpm_version")
	if err != nil {
		return ""
	}

	ptr := lib.Syscall(versionFn)
	return lib.PtrToString(ptr)
}

// Capabilities returns the library capabilities
func Capabilities() (CapabilitiesMask, error) {
	reg, err := lib.GetRegistry()
	if err != nil {
		return 0, err
	}

	capsFn, err := reg.GetFunc("alpm_capabilities")
	if err != nil {
		return 0, err
	}

	caps := lib.Syscall(capsFn)
	return CapabilitiesMask(caps), nil
}

// CapabilitiesMask represents the library capabilities bitmask
type CapabilitiesMask int

const (
	CapNLS        CapabilitiesMask = 1 << 0
	CapDownloader CapabilitiesMask = 1 << 1
	CapSignatures CapabilitiesMask = 1 << 2
)
