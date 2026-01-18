package lib

import "github.com/ebitengine/purego"

// Syscall calls a function and returns only the first return value (r1)
// This is a convenience wrapper for purego.SyscallN
func Syscall(fn uintptr, args ...uintptr) uintptr {
	r1, _, _ := purego.SyscallN(fn, args...)
	return r1
}

// SyscallWithError calls a function and returns r1 and errno
func SyscallWithError(fn uintptr, args ...uintptr) (uintptr, uintptr) {
	r1, _, errno := purego.SyscallN(fn, args...)
	return r1, errno
}
