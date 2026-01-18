package lib

import (
	"sync"

	"github.com/ebitengine/purego"
)

var (
	freeOnce sync.Once
	freeFn   uintptr
)

func getFreeFunc() uintptr {
	freeOnce.Do(func() {
		libc, err := purego.Dlopen("libc.so.6", purego.RTLD_NOW)
		if err != nil {
			libc, err = purego.Dlopen("libc.so", purego.RTLD_NOW)
			if err != nil {
				return
			}
		}
		fn, err := purego.Dlsym(libc, "free")
		if err == nil {
			freeFn = fn
		}
	})
	return freeFn
}

// Free releases memory allocated by libc malloc.
func Free(ptr uintptr) {
	if ptr == 0 {
		return
	}
	fn := getFreeFunc()
	if fn == 0 {
		return
	}
	purego.SyscallN(fn, ptr)
}
