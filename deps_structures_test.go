//go:build linux

package dyalpm

import (
	"testing"
	"unsafe"

	"github.com/Hayao0819/dyalpm/internal/lib"
	"github.com/Hayao0819/dyalpm/internal/testutil/cmem"
)

func TestDependencyDecodingPreservesMetadata(t *testing.T) {
	ptr := cmem.Alloc(t, unsafe.Sizeof(alpmDepend{}))
	raw := (*alpmDepend)(unsafe.Pointer(ptr))
	raw.Name = cmem.String(t, "python")
	raw.Version = cmem.String(t, "3.14")
	raw.Description = cmem.String(t, "runtime")
	raw.NameHash = 0xfeedbeef
	raw.Mod = int32(DepModGE)

	dep := newDependency(ptr)
	if dep.GetName() != "python" ||
		dep.GetVersion() != "3.14" ||
		dep.GetDescription() != "runtime" ||
		dep.GetNameHash() != 0xfeedbeef ||
		dep.GetMod() != DepModGE {
		t.Fatalf("decoded dependency does not match source: %#v", dep)
	}

	values := toDependList([]Dependency{dep})
	if len(values) != 1 {
		t.Fatalf("dependency count = %d, want 1", len(values))
	}
	if got := values[0]; got.Name != "python" ||
		got.Version != "3.14" ||
		got.Description != "runtime" ||
		got.NameHash != 0xfeedbeef ||
		got.Mod != DepModGE {
		t.Fatalf("dependency value = %#v", got)
	}
}

func TestPackageBuildDependencyBindings(t *testing.T) {
	dependPtr := cmem.Alloc(t, unsafe.Sizeof(alpmDepend{}))
	depend := (*alpmDepend)(unsafe.Pointer(dependPtr))
	depend.Name = cmem.String(t, "cmake")
	depend.Description = cmem.String(t, "build system")
	depend.NameHash = 42
	depend.Mod = int32(DepModAny)

	nodePtr := cmem.Alloc(t, unsafe.Sizeof(abiListNode{}))
	(*abiListNode)(unsafe.Pointer(nodePtr)).Data = dependPtr

	oldCheck := lib.AlpmPkgGetCheckdepends
	oldMake := lib.AlpmPkgGetMakedepends
	lib.AlpmPkgGetCheckdepends = func(uintptr) uintptr { return nodePtr }
	lib.AlpmPkgGetMakedepends = func(uintptr) uintptr { return nodePtr }
	t.Cleanup(func() {
		lib.AlpmPkgGetCheckdepends = oldCheck
		lib.AlpmPkgGetMakedepends = oldMake
	})

	pkg := &package_{ptr: 1}
	for name, values := range map[string][]Depend{
		"checkdepends": pkg.CheckDepends(),
		"makedepends":  pkg.MakeDepends(),
	} {
		if len(values) != 1 {
			t.Fatalf("%s count = %d, want 1", name, len(values))
		}
		if got := values[0]; got.Name != "cmake" ||
			got.Description != "build system" ||
			got.NameHash != 42 {
			t.Fatalf("%s = %#v", name, got)
		}
	}
}

func TestConflictDecodingUsesPackageAccessors(t *testing.T) {
	firstPackage := cmem.Alloc(t, 1)
	secondPackage := cmem.Alloc(t, 1)
	firstName := cmem.String(t, "first")
	secondName := cmem.String(t, "second")

	oldGetName := lib.AlpmPkgGetName
	lib.AlpmPkgGetName = func(pkg uintptr) uintptr {
		switch pkg {
		case firstPackage:
			return firstName
		case secondPackage:
			return secondName
		default:
			return 0
		}
	}
	t.Cleanup(func() {
		lib.AlpmPkgGetName = oldGetName
	})

	reasonPtr := cmem.Alloc(t, unsafe.Sizeof(alpmDepend{}))
	reason := (*alpmDepend)(unsafe.Pointer(reasonPtr))
	reason.Name = cmem.String(t, "virtual")
	reason.Mod = int32(DepModAny)

	conflictPtr := cmem.Alloc(t, unsafe.Sizeof(alpmConflict{}))
	raw := (*alpmConflict)(unsafe.Pointer(conflictPtr))
	raw.Package1 = firstPackage
	raw.Package2 = secondPackage
	raw.Reason = reasonPtr

	conflict := newConflict(conflictPtr)
	if got := conflict.GetPackage1(); got != "first" {
		t.Fatalf("package1 = %q, want first", got)
	}
	if got := conflict.GetPackage2(); got != "second" {
		t.Fatalf("package2 = %q, want second", got)
	}
	if got := conflict.GetReason(); got == nil || got.GetName() != "virtual" {
		t.Fatalf("reason = %#v", got)
	}
}

func TestFileConflictDecoding(t *testing.T) {
	ptr := cmem.Alloc(t, unsafe.Sizeof(alpmFileConflict{}))
	raw := (*alpmFileConflict)(unsafe.Pointer(ptr))
	raw.Target = cmem.String(t, "target")
	raw.Type = int32(FileConflictFilesystem)
	raw.File = cmem.String(t, "usr/bin/tool")
	raw.CTarget = cmem.String(t, "owner")

	conflict := newFileConflict(ptr)
	if conflict.GetTarget() != "target" ||
		conflict.GetType() != FileConflictFilesystem ||
		conflict.GetFile() != "usr/bin/tool" ||
		conflict.GetCTarget() != "owner" {
		t.Fatalf("decoded file conflict does not match source")
	}
}
