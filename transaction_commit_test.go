package dyalpm

import (
	"errors"
	"runtime"
	"testing"

	"github.com/Jguer/dyalpm/internal/dyerrors"
)

func TestTransactionCommitDiagnostics(t *testing.T) {
	t.Run("file conflict", func(t *testing.T) {
		log := installTransactionBindings(t)
		var pinner runtime.Pinner
		t.Cleanup(pinner.Unpin)
		target, targetPtr := transactionCString(&pinner, "new-package")
		file, filePtr := transactionCString(&pinner, "usr/bin/tool")
		ctarget, ctargetPtr := transactionCString(&pinner, "installed-package")
		conflict := &transactionTestFileConflict{
			target:  targetPtr,
			kind:    int32(FileConflictFilesystem),
			file:    filePtr,
			ctarget: ctargetPtr,
		}
		conflictPtr := transactionPinnedPointer(&pinner, conflict)
		list := &transactionTestList{data: conflictPtr}
		listPtr := transactionPinnedPointer(&pinner, list)
		stubCommit(t, listPtr, dyerrors.ErrFileConflicts)

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
		runtime.KeepAlive(target)
		runtime.KeepAlive(file)
		runtime.KeepAlive(ctarget)
		runtime.KeepAlive(conflict)
		runtime.KeepAlive(list)
	})

	t.Run("invalid package", func(t *testing.T) {
		log := installTransactionBindings(t)
		var pinner runtime.Pinner
		t.Cleanup(pinner.Unpin)
		name, namePtr := transactionCString(&pinner, "broken.pkg.tar.zst")
		list := &transactionTestList{data: namePtr}
		listPtr := transactionPinnedPointer(&pinner, list)
		stubCommit(t, listPtr, dyerrors.ErrPkgInvalidSig)

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
		runtime.KeepAlive(name)
		runtime.KeepAlive(list)
	})
}
