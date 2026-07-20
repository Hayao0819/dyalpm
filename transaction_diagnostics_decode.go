package dyalpm

import (
	"strings"
	"unsafe"

	"github.com/Jguer/dyalpm/internal/dyerrors"
	"github.com/Jguer/dyalpm/internal/lib"
	alpmlist "github.com/Jguer/dyalpm/internal/list"
)

func decodePrepareDiagnostics(
	errno dyerrors.Errno,
	data *alpmlist.List,
) TransactionDiagnostics {
	var diagnostics TransactionDiagnostics
	switch errno {
	case dyerrors.ErrPkgInvalidArch:
		diagnostics.InvalidArchitecture = copyAndFreeStrings(data)
	case dyerrors.ErrUnsatisfiedDeps:
		forEachData(data, func(ptr uintptr) {
			diagnostics.MissingDependencies = append(
				diagnostics.MissingDependencies,
				copyMissingDependency(ptr),
			)
			if lib.AlpmDepmissingFree != nil {
				lib.AlpmDepmissingFree(ptr)
			}
		})
	case dyerrors.ErrConflictingDeps:
		forEachData(data, func(ptr uintptr) {
			diagnostics.PackageConflicts = append(
				diagnostics.PackageConflicts,
				copyPackageConflict(ptr),
			)
			if lib.AlpmConflictFree != nil {
				lib.AlpmConflictFree(ptr)
			}
		})
	}
	return diagnostics
}

func decodeCommitDiagnostics(
	errno dyerrors.Errno,
	data *alpmlist.List,
) TransactionDiagnostics {
	var diagnostics TransactionDiagnostics
	switch errno {
	case dyerrors.ErrFileConflicts:
		forEachData(data, func(ptr uintptr) {
			diagnostics.FileConflicts = append(
				diagnostics.FileConflicts,
				copyFileConflict(ptr),
			)
			if lib.AlpmFileConflictFree != nil {
				lib.AlpmFileConflictFree(ptr)
			}
		})
	case dyerrors.ErrPkgInvalid,
		dyerrors.ErrPkgInvalidChecksum,
		dyerrors.ErrPkgInvalidSig:
		diagnostics.InvalidPackageFiles = copyAndFreeStrings(data)
	}
	return diagnostics
}

func copyMissingDependency(ptr uintptr) MissingDependency {
	missing := newDepMissing(ptr)
	return MissingDependency{
		Target:         strings.Clone(missing.GetTarget()),
		Dependency:     copyDependency(missing.GetDepend()),
		CausingPackage: strings.Clone(missing.GetCausingPkg()),
	}
}

func copyPackageConflict(ptr uintptr) PackageConflict {
	if ptr == 0 {
		return PackageConflict{}
	}
	raw := (*[3]uintptr)(unsafe.Pointer(ptr))
	return PackageConflict{
		Package1: packageName(raw[0]),
		Package2: packageName(raw[1]),
		Reason:   copyDependency(newDependency(raw[2])),
	}
}

func copyFileConflict(ptr uintptr) FileConflictDetail {
	conflict := newFileConflict(ptr)
	return FileConflictDetail{
		Target:            strings.Clone(conflict.GetTarget()),
		Type:              conflict.GetType(),
		File:              strings.Clone(conflict.GetFile()),
		ConflictingTarget: strings.Clone(conflict.GetCTarget()),
	}
}

func copyDependency(dependency Dependency) Depend {
	if dependency == nil {
		return Depend{}
	}
	return Depend{
		Name:    strings.Clone(dependency.GetName()),
		Version: strings.Clone(dependency.GetVersion()),
		Mod:     dependency.GetMod(),
	}
}

func packageName(ptr uintptr) string {
	if ptr == 0 || lib.AlpmPkgGetName == nil {
		return ""
	}
	return strings.Clone(lib.PtrToString(lib.AlpmPkgGetName(ptr)))
}

func copyAndFreeStrings(data *alpmlist.List) []string {
	values := make([]string, 0)
	forEachData(data, func(ptr uintptr) {
		values = append(values, strings.Clone(lib.PtrToString(ptr)))
		lib.Free(ptr)
	})
	return values
}

func forEachData(data *alpmlist.List, visit func(uintptr)) {
	for item := data; item != nil && item.Ptr() != 0; item = item.Next() {
		if ptr := item.Data(); ptr != 0 {
			visit(ptr)
		}
	}
}

func missingDependencyInterfaces(values []MissingDependency) []DepMissing {
	result := make([]DepMissing, len(values))
	for i := range values {
		result[i] = values[i]
	}
	return result
}

func fileConflictInterfaces(values []FileConflictDetail) []FileConflict {
	result := make([]FileConflict, len(values))
	for i := range values {
		result[i] = values[i]
	}
	return result
}
