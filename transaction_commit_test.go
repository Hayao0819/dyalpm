package dyalpm

import (
	"errors"
	"testing"

	alpmerrors "github.com/Hayao0819/dyalpm/errors"
)

func TestTransactionCommitDiagnostics(t *testing.T) {
	t.Run("file conflict", func(t *testing.T) {
		log := installTransactionBindings(t)
		_, targetPtr := transactionCString(t, "new-package")
		_, filePtr := transactionCString(t, "usr/bin/tool")
		_, ctargetPtr := transactionCString(t, "installed-package")
		conflict := &transactionTestFileConflict{
			target:  targetPtr,
			kind:    int32(FileConflictFilesystem),
			file:    filePtr,
			ctarget: ctargetPtr,
		}
		conflictPtr := transactionPointer(t, conflict)
		list := &transactionTestList{data: conflictPtr}
		listPtr := transactionPointer(t, list)
		stubCommit(t, listPtr, alpmerrors.ErrFileConflicts)

		values, err := (&transaction{handle: &handle{ptr: 1}}).Commit()
		var transactionErr *TransactionError
		if !errors.As(err, &transactionErr) {
			t.Fatalf("Commit() error = %v, want TransactionError", err)
		}
		if !errors.Is(err, ErrTransactionCommitFailed) {
			t.Fatalf("Commit() error = %v, want ErrTransactionCommitFailed", err)
		}
		if len(values) != 1 ||
			values[0].GetTarget() != "new-package" ||
			values[0].GetType() != FileConflictFilesystem ||
			values[0].GetFile() != "usr/bin/tool" ||
			values[0].GetCTarget() != "installed-package" {
			t.Fatalf("Commit() file conflicts = %#v", values)
		}
		if len(transactionErr.Diagnostics.FileConflicts) != 1 {
			t.Fatalf(
				"file conflict diagnostics = %d, want 1",
				len(transactionErr.Diagnostics.FileConflicts),
			)
		}
		assertFreed(t, log.fileConflicts, conflictPtr)
		assertFreed(t, log.lists, listPtr)
	})

	t.Run("invalid package", func(t *testing.T) {
		log := installTransactionBindings(t)
		_, namePtr := transactionCString(t, "broken.pkg.tar.zst")
		list := &transactionTestList{data: namePtr}
		listPtr := transactionPointer(t, list)
		stubCommit(t, listPtr, alpmerrors.ErrPkgInvalidSig)

		values, err := (&transaction{handle: &handle{ptr: 1}}).Commit()
		var transactionErr *TransactionError
		if !errors.As(err, &transactionErr) {
			t.Fatalf("Commit() error = %v, want TransactionError", err)
		}
		if len(values) != 0 {
			t.Fatalf("Commit() returned %d file conflicts, want 0", len(values))
		}
		got := transactionErr.Diagnostics.InvalidPackageFiles
		if len(got) != 1 || got[0] != "broken.pkg.tar.zst" {
			t.Fatalf("invalid package diagnostics = %v", got)
		}
		assertFreed(t, log.strings, namePtr)
		assertFreed(t, log.lists, listPtr)
	})
}
