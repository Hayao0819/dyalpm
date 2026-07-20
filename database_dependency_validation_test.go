package dyalpm

import (
	"testing"

	"github.com/Hayao0819/dyalpm/internal/lib"
)

type foreignDatabase struct {
	Database
}

type foreignDependency struct {
	Dependency
}

func TestDatabaseConsumersRejectInvalidDatabases(t *testing.T) {
	installListStubs(t)
	oldGroup := lib.AlpmFindGroupPkgs
	oldSatisfier := lib.AlpmFindDBSatisfier
	oldNewVersion := lib.AlpmPkgSyncGetNewVersion
	var groupCalls, satisfierCalls, versionCalls int
	lib.AlpmFindGroupPkgs = func(uintptr, string) uintptr {
		groupCalls++
		return 0
	}
	lib.AlpmFindDBSatisfier = func(uintptr, uintptr, string) uintptr {
		satisfierCalls++
		return 0
	}
	lib.AlpmPkgSyncGetNewVersion = func(uintptr, uintptr) uintptr {
		versionCalls++
		return 0
	}
	t.Cleanup(func() {
		lib.AlpmFindGroupPkgs = oldGroup
		lib.AlpmFindDBSatisfier = oldSatisfier
		lib.AlpmPkgSyncGetNewVersion = oldNewVersion
	})

	var typedNil *database
	invalid := []struct {
		name string
		db   Database
	}{
		{"nil", nil},
		{"typed nil", typedNil},
		{"zero internal", &database{}},
		{"foreign", foreignDatabase{}},
	}
	h := &handle{ptr: 1}
	pkg := &package_{ptr: 1}
	valid := Database(&database{ptr: 1})

	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			values := []Database{valid, test.db}
			_, err := h.FindGroupPkgs(values, "group")
			assertErrorIs(t, err, ErrInvalidDatabase)
			if got := h.FindDBSatisfier(values, "dep"); got != nil {
				t.Fatalf("FindDBSatisfier returned %#v", got)
			}
			if got := pkg.SyncGetNewVersion(values); got != nil {
				t.Fatalf("SyncGetNewVersion returned %#v", got)
			}
		})
	}

	if groupCalls != 0 || satisfierCalls != 0 || versionCalls != 0 {
		t.Fatalf("libalpm calls = group:%d satisfier:%d version:%d, want zero",
			groupCalls, satisfierCalls, versionCalls)
	}
}

func TestAssumeInstalledRejectsInvalidDependencies(t *testing.T) {
	installListStubs(t)
	oldSet := lib.AlpmOptionSetNoassumeInstalled
	oldAdd := lib.AlpmOptionAddAssumeInstalled
	oldRemove := lib.AlpmOptionRemoveAssumeInstalled
	var setCalls, addCalls, removeCalls int
	lib.AlpmOptionSetNoassumeInstalled = func(uintptr, uintptr) int32 {
		setCalls++
		return 0
	}
	lib.AlpmOptionAddAssumeInstalled = func(uintptr, uintptr) int32 {
		addCalls++
		return 0
	}
	lib.AlpmOptionRemoveAssumeInstalled = func(uintptr, uintptr) int32 {
		removeCalls++
		return 0
	}
	t.Cleanup(func() {
		lib.AlpmOptionSetNoassumeInstalled = oldSet
		lib.AlpmOptionAddAssumeInstalled = oldAdd
		lib.AlpmOptionRemoveAssumeInstalled = oldRemove
	})

	var typedNil *dependency
	invalid := []struct {
		name string
		dep  Dependency
	}{
		{"nil", nil},
		{"typed nil", typedNil},
		{"zero internal", &dependency{}},
		{"foreign", foreignDependency{}},
	}
	h := &handle{ptr: 1}
	valid := Dependency(&dependency{ptr: 1})

	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			assertErrorIs(t, h.SetAssumeInstalled([]Dependency{valid, test.dep}), ErrInvalidDependency)
			assertErrorIs(t, h.AddAssumeInstalled(test.dep), ErrInvalidDependency)
			assertErrorIs(t, h.RemoveAssumeInstalled(test.dep), ErrInvalidDependency)
		})
	}

	if setCalls != 0 || addCalls != 0 || removeCalls != 0 {
		t.Fatalf("libalpm calls = set:%d add:%d remove:%d, want zero",
			setCalls, addCalls, removeCalls)
	}
}

func TestWrapperListReportsAllocationFailure(t *testing.T) {
	oldAdd := lib.AlpmListAdd
	oldFree := lib.AlpmListFree
	var calls, frees int
	lib.AlpmListAdd = func(uintptr, uintptr) uintptr {
		calls++
		if calls == 1 {
			return 0x1000
		}
		return 0
	}
	lib.AlpmListFree = func(uintptr) {
		frees++
	}
	t.Cleanup(func() {
		lib.AlpmListAdd = oldAdd
		lib.AlpmListFree = oldFree
	})

	_, err := packageList([]Package{&package_{ptr: 1}, &package_{ptr: 2}})
	assertErrorIs(t, err, ErrListCreationFailed)
	if frees != 1 {
		t.Fatalf("partial list frees = %d, want 1", frees)
	}
}
