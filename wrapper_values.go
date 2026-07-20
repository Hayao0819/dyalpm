package dyalpm

import (
	stderrors "errors"

	"github.com/Hayao0819/dyalpm/internal/lib"
	alpmlist "github.com/Hayao0819/dyalpm/internal/list"
)

func internalHandle(value Handle) (*handle, error) {
	h, ok := value.(*handle)
	if !ok || h == nil || h.ptr == 0 {
		return nil, ErrInvalidHandle
	}
	return h, nil
}

func internalPackage(value Package) (*package_, error) {
	pkg, ok := value.(*package_)
	if !ok || pkg == nil || pkg.ptr == 0 {
		return nil, ErrInvalidPackage
	}
	return pkg, nil
}

func internalDatabase(value Database) (*database, error) {
	db, ok := value.(*database)
	if !ok || db == nil || db.ptr == 0 {
		return nil, ErrInvalidDatabase
	}
	return db, nil
}

func internalDependency(value Dependency) (*dependency, error) {
	dep, ok := value.(*dependency)
	if !ok || dep == nil || dep.ptr == 0 {
		return nil, ErrInvalidDependency
	}
	return dep, nil
}

func buildWrapperList[T any](values []T, pointer func(T) (uintptr, error)) (*alpmlist.List, error) {
	if len(values) == 0 {
		return nil, nil
	}

	pointers := make([]uintptr, len(values))
	for i, value := range values {
		ptr, err := pointer(value)
		if err != nil {
			return nil, err
		}
		pointers[i] = ptr
	}

	if lib.AlpmListAdd == nil {
		return nil, stderrors.New("missing function: alpm_list_add")
	}

	var list *alpmlist.List
	for _, ptr := range pointers {
		next := alpmlist.Add(list, ptr)
		if next == nil {
			list.Free()
			return nil, ErrListCreationFailed
		}
		list = next
	}
	return list, nil
}

func packageList(values []Package) (*alpmlist.List, error) {
	return buildWrapperList(values, func(value Package) (uintptr, error) {
		pkg, err := internalPackage(value)
		if err != nil {
			return 0, err
		}
		return pkg.ptr, nil
	})
}

func databaseList(values []Database) (*alpmlist.List, error) {
	return buildWrapperList(values, func(value Database) (uintptr, error) {
		db, err := internalDatabase(value)
		if err != nil {
			return 0, err
		}
		return db.ptr, nil
	})
}

func dependencyList(values []Dependency) (*alpmlist.List, error) {
	return buildWrapperList(values, func(value Dependency) (uintptr, error) {
		dep, err := internalDependency(value)
		if err != nil {
			return 0, err
		}
		return dep.ptr, nil
	})
}
