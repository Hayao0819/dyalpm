package dyalpm

import (
	"errors"
	"testing"

	"github.com/Jguer/dyalpm/internal/lib"
)

type foreignHandle struct {
	Handle
}

type foreignPackage struct {
	Package
}

func TestNewTransactionRejectsInvalidHandlesWithoutPanicking(t *testing.T) {
	var typedNil *handle
	tests := []struct {
		name   string
		handle Handle
	}{
		{"nil", nil},
		{"typed nil", typedNil},
		{"zero internal", &handle{}},
		{"foreign", foreignHandle{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := NewTransaction(test.handle)
			if tx == nil {
				t.Fatal("NewTransaction returned nil")
			}

			assertErrorIs(t, tx.Init(0), ErrInvalidHandle)
			_, err := tx.Prepare()
			assertErrorIs(t, err, ErrInvalidHandle)
			_, err = tx.Commit()
			assertErrorIs(t, err, ErrInvalidHandle)
			assertErrorIs(t, tx.Release(), ErrInvalidHandle)
			assertErrorIs(t, tx.AddPkg(&package_{ptr: 1}), ErrInvalidHandle)
			assertErrorIs(t, tx.RemovePkg(&package_{ptr: 1}), ErrInvalidHandle)
			assertErrorIs(t, tx.SyncSysupgrade(false), ErrInvalidHandle)
			if flags := tx.GetFlags(); flags != 0 {
				t.Fatalf("GetFlags() = %d, want 0", flags)
			}
			_, err = tx.GetAdd()
			assertErrorIs(t, err, ErrInvalidHandle)
			_, err = tx.GetRemove()
			assertErrorIs(t, err, ErrInvalidHandle)
		})
	}
}

func TestTransactionRejectsInvalidPackagesWithoutCallingLibalpm(t *testing.T) {
	oldAdd := lib.AlpmAddPkg
	oldRemove := lib.AlpmRemovePkg
	var addCalls, removeCalls int
	lib.AlpmAddPkg = func(uintptr, uintptr) int32 {
		addCalls++
		return 0
	}
	lib.AlpmRemovePkg = func(uintptr, uintptr) int32 {
		removeCalls++
		return 0
	}
	t.Cleanup(func() {
		lib.AlpmAddPkg = oldAdd
		lib.AlpmRemovePkg = oldRemove
	})

	var typedNil *package_
	tests := []struct {
		name string
		pkg  Package
	}{
		{"nil", nil},
		{"typed nil", typedNil},
		{"zero internal", &package_{}},
		{"foreign", foreignPackage{}},
	}

	tx := &transaction{handle: &handle{ptr: 1}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertErrorIs(t, tx.AddPkg(test.pkg), ErrInvalidPackage)
			assertErrorIs(t, tx.RemovePkg(test.pkg), ErrInvalidPackage)
		})
	}

	if addCalls != 0 || removeCalls != 0 {
		t.Fatalf("libalpm calls = add:%d remove:%d, want zero", addCalls, removeCalls)
	}
}

func assertErrorIs(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("error = %v, want %v", got, want)
	}
}
