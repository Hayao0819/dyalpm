package dyalpm

import (
	"sync"

	"github.com/Jguer/dyalpm/internal/lib"
)

var fnCache sync.Map

func cachedFunc(name string) uintptr {
	if v, ok := fnCache.Load(name); ok {
		return v.(uintptr)
	}
	reg, err := lib.GetRegistry()
	if err != nil {
		return 0
	}
	fn, err := reg.GetFunc(name)
	if err != nil {
		return 0
	}
	fnCache.Store(name, fn)
	return fn
}
