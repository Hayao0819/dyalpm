package dyalpm

import (
	"cmp"
	"slices"
	"sync"

	alpmlist "github.com/Hayao0819/dyalpm/internal/list"
)

// PackageIterator reuses borrowed lists and consumes owned lists once.
type PackageIterator struct {
	state  *packageIteratorState
	handle *handle
}

type packageIteratorState struct {
	mu    sync.Mutex
	list  *alpmlist.List
	owned bool
}

func newPackageIterator(listPtr uintptr, h *handle, owned bool) PackageIterator {
	list := alpmlist.NewList(listPtr)
	if list == nil {
		return PackageIterator{handle: h}
	}
	return PackageIterator{
		state:  &packageIteratorState{list: list, owned: owned},
		handle: h,
	}
}

func (s *packageIteratorState) acquire() (*alpmlist.List, bool) {
	if s == nil {
		return nil, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.list
	if list == nil {
		return nil, false
	}
	if s.owned {
		s.list = nil
	}
	return list, s.owned
}

func (it PackageIterator) ForEach(fn func(Package) error) error {
	list, release := it.state.acquire()
	if list == nil {
		return nil
	}
	if release {
		defer list.Free()
	}
	for item := list; item != nil && item.Ptr() != 0; item = item.Next() {
		pkgPtr := item.Data()
		if pkgPtr == 0 {
			continue
		}
		if err := fn(newPackage(pkgPtr, it.handle)); err != nil {
			return err
		}
	}
	return nil
}

// Collect returns all packages in the iterator as a slice.
func (it PackageIterator) Collect() []Package {
	var pkgs []Package
	_ = it.ForEach(func(pkg Package) error {
		pkgs = append(pkgs, pkg)
		return nil
	})
	return pkgs
}

// FindSatisfier finds the first package satisfying a depstring.
func (it PackageIterator) FindSatisfier(depstring string) (Package, error) {
	pkg := FindSatisfier(it.Collect(), depstring)
	if pkg == nil {
		return nil, ErrPackageNotFound
	}
	return pkg, nil
}

// SortBySize returns packages sorted by install size (descending).
func (it PackageIterator) SortBySize() []Package {
	pkgs := it.Collect()
	slices.SortFunc(pkgs, func(a, b Package) int {
		return cmp.Compare(b.ISize(), a.ISize())
	})
	return pkgs
}
