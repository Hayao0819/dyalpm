package dyalpm

import (
	"testing"

	"github.com/Jguer/dyalpm/internal/lib"
)

type listCallStats struct {
	adds  int
	frees int
}

func installListStubs(t *testing.T) *listCallStats {
	t.Helper()

	oldAdd := lib.AlpmListAdd
	oldFree := lib.AlpmListFree
	stats := &listCallStats{}
	lib.AlpmListAdd = func(list, _ uintptr) uintptr {
		stats.adds++
		if list == 0 {
			return 0x1000
		}
		return list
	}
	lib.AlpmListFree = func(uintptr) {
		stats.frees++
	}
	t.Cleanup(func() {
		lib.AlpmListAdd = oldAdd
		lib.AlpmListFree = oldFree
	})
	return stats
}

func TestDependencyChecksRejectInvalidPackages(t *testing.T) {
	stats := installListStubs(t)
	oldCheckDeps := lib.AlpmCheckDeps
	oldCheckConflicts := lib.AlpmCheckConflicts
	var dependencyCalls, conflictCalls int
	lib.AlpmCheckDeps = func(uintptr, uintptr, uintptr, uintptr, int32) uintptr {
		dependencyCalls++
		return 0
	}
	lib.AlpmCheckConflicts = func(uintptr, uintptr) uintptr {
		conflictCalls++
		return 0
	}
	t.Cleanup(func() {
		lib.AlpmCheckDeps = oldCheckDeps
		lib.AlpmCheckConflicts = oldCheckConflicts
	})

	var typedNil *package_
	invalid := []struct {
		name string
		pkg  Package
	}{
		{"nil", nil},
		{"typed nil", typedNil},
		{"zero internal", &package_{}},
		{"foreign", foreignPackage{}},
	}
	valid := Package(&package_{ptr: 1})

	for _, test := range invalid {
		for _, listName := range []string{"targets", "remove", "upgrade"} {
			t.Run(test.name+"/"+listName, func(t *testing.T) {
				targets := []Package{valid}
				remove := []Package{valid}
				upgrade := []Package{valid}
				switch listName {
				case "targets":
					targets = []Package{valid, test.pkg}
				case "remove":
					remove = []Package{valid, test.pkg}
				case "upgrade":
					upgrade = []Package{valid, test.pkg}
				}

				_, err := (&handle{ptr: 1}).CheckDeps(targets, remove, upgrade, false)
				assertErrorIs(t, err, ErrInvalidPackage)
			})
		}

		t.Run(test.name+"/conflicts", func(t *testing.T) {
			_, err := (&handle{ptr: 1}).CheckConflicts([]Package{valid, test.pkg})
			assertErrorIs(t, err, ErrInvalidPackage)
		})
	}

	if dependencyCalls != 0 || conflictCalls != 0 {
		t.Fatalf("libalpm calls = dependencies:%d conflicts:%d, want zero", dependencyCalls, conflictCalls)
	}
	if stats.frees == 0 {
		t.Fatal("partially built lists were not freed")
	}
}

func TestPackageSearchRejectsWholeInvalidLists(t *testing.T) {
	installListStubs(t)
	oldFind := lib.AlpmPkgFind
	oldSatisfier := lib.AlpmFindSatisfier
	var findCalls, satisfierCalls int
	lib.AlpmPkgFind = func(uintptr, string) uintptr {
		findCalls++
		return 0
	}
	lib.AlpmFindSatisfier = func(uintptr, string) uintptr {
		satisfierCalls++
		return 0
	}
	t.Cleanup(func() {
		lib.AlpmPkgFind = oldFind
		lib.AlpmFindSatisfier = oldSatisfier
	})

	values := []Package{&package_{ptr: 1}, foreignPackage{}}
	if got := PkgFind(values, "pkg"); got != nil {
		t.Fatalf("PkgFind returned %#v", got)
	}
	if got := FindSatisfier(values, "pkg"); got != nil {
		t.Fatalf("FindSatisfier returned %#v", got)
	}
	if findCalls != 0 || satisfierCalls != 0 {
		t.Fatalf("libalpm calls = find:%d satisfier:%d, want zero", findCalls, satisfierCalls)
	}
}
