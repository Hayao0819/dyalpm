package dyalpm

import (
	"errors"
	"testing"

	alpmerrors "github.com/Jguer/dyalpm/errors"
	"github.com/Jguer/dyalpm/internal/lib"
)

func TestTransactionPrepareDiagnostics(t *testing.T) {
	t.Run("missing dependency", func(t *testing.T) {
		log := installTransactionBindings(t)
		target, targetPtr := transactionCString(t, "consumer")
		name, namePtr := transactionCString(t, "provider")
		version, versionPtr := transactionCString(t, "2")
		causing, causingPtr := transactionCString(t, "upgrade")
		depend := &transactionTestDepend{
			name:    namePtr,
			version: versionPtr,
			mod:     int32(DepModGE),
		}
		dependPtr := transactionPointer(t, depend)
		missing := &transactionTestMissing{
			target:     targetPtr,
			depend:     dependPtr,
			causingPkg: causingPtr,
		}
		missingPtr := transactionPointer(t, missing)
		list := &transactionTestList{data: missingPtr}
		listPtr := transactionPointer(t, list)
		stubPrepare(t, listPtr, alpmerrors.ErrUnsatisfiedDeps)

		values, err := (&transaction{handle: &handle{ptr: 1}}).Prepare()
		var transactionErr *TransactionError
		if !errors.As(err, &transactionErr) {
			t.Fatalf("Prepare() error = %v, want TransactionError", err)
		}
		if !errors.Is(err, ErrTransactionPrepareFailed) {
			t.Fatalf("Prepare() error = %v, want ErrTransactionPrepareFailed", err)
		}
		if len(values) != 1 {
			t.Fatalf("Prepare() returned %d missing dependencies, want 1", len(values))
		}
		got := values[0]
		if got.GetTarget() != "consumer" ||
			got.GetDepend().ComputeString() != "provider>=2" ||
			got.GetCausingPkg() != "upgrade" {
			t.Fatalf("Prepare() missing dependency = %#v", got)
		}
		clear(target)
		clear(name)
		clear(version)
		clear(causing)
		if got.GetTarget() != "consumer" ||
			got.GetDepend().ComputeString() != "provider>=2" {
			t.Fatal("Prepare() returned borrowed diagnostic strings")
		}
		if len(transactionErr.Diagnostics.MissingDependencies) != 1 {
			t.Fatalf(
				"missing diagnostics = %d, want 1",
				len(transactionErr.Diagnostics.MissingDependencies),
			)
		}
		assertFreed(t, log.lists, listPtr)
		assertFreed(t, log.missing, missingPtr)
	})

	t.Run("invalid architecture", func(t *testing.T) {
		log := installTransactionBindings(t)
		_, namePtr := transactionCString(t, "foreign-1.pkg.tar.zst")
		list := &transactionTestList{data: namePtr}
		listPtr := transactionPointer(t, list)
		stubPrepare(t, listPtr, alpmerrors.ErrPkgInvalidArch)

		values, err := (&transaction{handle: &handle{ptr: 1}}).Prepare()
		var transactionErr *TransactionError
		if !errors.As(err, &transactionErr) {
			t.Fatalf("Prepare() error = %v, want TransactionError", err)
		}
		if len(values) != 0 {
			t.Fatalf("Prepare() returned %d missing dependencies, want 0", len(values))
		}
		got := transactionErr.Diagnostics.InvalidArchitecture
		if len(got) != 1 || got[0] != "foreign-1.pkg.tar.zst" {
			t.Fatalf("invalid architecture diagnostics = %v", got)
		}
		assertFreed(t, log.strings, namePtr)
		assertFreed(t, log.lists, listPtr)
	})

	t.Run("package conflict", func(t *testing.T) {
		log := installTransactionBindings(t)
		_, firstPtr := transactionCString(t, "first")
		_, secondPtr := transactionCString(t, "second")
		_, reasonNamePtr := transactionCString(t, "virtual")
		reason := &transactionTestDepend{
			name: reasonNamePtr,
			mod:  int32(DepModAny),
		}
		reasonPtr := transactionPointer(t, reason)
		const firstPackage = uintptr(0x101)
		const secondPackage = uintptr(0x202)
		conflict := &[3]uintptr{
			firstPackage,
			secondPackage,
			reasonPtr,
		}
		conflictPtr := transactionPointer(t, conflict)
		list := &transactionTestList{data: conflictPtr}
		listPtr := transactionPointer(t, list)
		lib.AlpmPkgGetName = func(ptr uintptr) uintptr {
			switch ptr {
			case firstPackage:
				return firstPtr
			case secondPackage:
				return secondPtr
			default:
				return 0
			}
		}
		stubPrepare(t, listPtr, alpmerrors.ErrConflictingDeps)

		_, err := (&transaction{handle: &handle{ptr: 1}}).Prepare()
		var transactionErr *TransactionError
		if !errors.As(err, &transactionErr) {
			t.Fatalf("Prepare() error = %v, want TransactionError", err)
		}
		got := transactionErr.Diagnostics.PackageConflicts
		if len(got) != 1 ||
			got[0].Package1 != "first" ||
			got[0].Package2 != "second" ||
			got[0].Reason.String() != "virtual" {
			t.Fatalf("package conflict diagnostics = %#v", got)
		}
		assertFreed(t, log.conflicts, conflictPtr)
		assertFreed(t, log.lists, listPtr)
	})
}
