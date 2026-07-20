package errors_test

import (
	stderrors "errors"
	"fmt"
	"testing"

	alpmerrors "github.com/Jguer/dyalpm/errors"
)

func TestErrnoValuesAndMessages(t *testing.T) {
	tests := []struct {
		name    string
		errno   alpmerrors.Errno
		value   int32
		message string
	}{
		{"ALPM_ERR_OK", alpmerrors.ErrOK, 0, "No error"},
		{"ALPM_ERR_MEMORY", alpmerrors.ErrMemory, 1, "Failed to allocate memory"},
		{"ALPM_ERR_SYSTEM", alpmerrors.ErrSystem, 2, "A system error occurred"},
		{"ALPM_ERR_BADPERMS", alpmerrors.ErrBadPerms, 3, "Permission denied"},
		{"ALPM_ERR_NOT_A_FILE", alpmerrors.ErrNotAFile, 4, "Should be a file"},
		{"ALPM_ERR_NOT_A_DIR", alpmerrors.ErrNotADir, 5, "Should be a directory"},
		{"ALPM_ERR_WRONG_ARGS", alpmerrors.ErrWrongArgs, 6, "Function was called with invalid arguments"},
		{"ALPM_ERR_DISK_SPACE", alpmerrors.ErrDiskSpace, 7, "Insufficient disk space"},
		{"ALPM_ERR_HANDLE_NULL", alpmerrors.ErrHandleNull, 8, "Handle should be null"},
		{"ALPM_ERR_HANDLE_NOT_NULL", alpmerrors.ErrHandleNotNull, 9, "Handle should not be null"},
		{"ALPM_ERR_HANDLE_LOCK", alpmerrors.ErrHandleLock, 10, "Failed to acquire lock"},
		{"ALPM_ERR_DB_OPEN", alpmerrors.ErrDBOpen, 11, "Failed to open database"},
		{"ALPM_ERR_DB_CREATE", alpmerrors.ErrDBCreate, 12, "Failed to create database"},
		{"ALPM_ERR_DB_NULL", alpmerrors.ErrDBNull, 13, "Database should not be null"},
		{"ALPM_ERR_DB_NOT_NULL", alpmerrors.ErrDBNotNull, 14, "Database should be null"},
		{"ALPM_ERR_DB_NOT_FOUND", alpmerrors.ErrDBNotFound, 15, "The database could not be found"},
		{"ALPM_ERR_DB_INVALID", alpmerrors.ErrDBInvalid, 16, "Database is invalid"},
		{"ALPM_ERR_DB_INVALID_SIG", alpmerrors.ErrDBInvalidSig, 17, "Database has an invalid signature"},
		{"ALPM_ERR_DB_VERSION", alpmerrors.ErrDBVersion, 18, "The localdb is in a newer/older format than libalpm expects"},
		{"ALPM_ERR_DB_WRITE", alpmerrors.ErrDBWrite, 19, "Failed to write to the database"},
		{"ALPM_ERR_DB_REMOVE", alpmerrors.ErrDBRemove, 20, "Failed to remove entry from database"},
		{"ALPM_ERR_SERVER_BAD_URL", alpmerrors.ErrServerBadURL, 21, "Server URL is in an invalid format"},
		{"ALPM_ERR_SERVER_NONE", alpmerrors.ErrServerNone, 22, "The database has no configured servers"},
		{"ALPM_ERR_TRANS_NOT_NULL", alpmerrors.ErrTransNotNull, 23, "A transaction is already initialized"},
		{"ALPM_ERR_TRANS_NULL", alpmerrors.ErrTransNull, 24, "A transaction has not been initialized"},
		{"ALPM_ERR_TRANS_DUP_TARGET", alpmerrors.ErrTransDupTarget, 25, "Duplicate target in transaction"},
		{"ALPM_ERR_TRANS_DUP_FILENAME", alpmerrors.ErrTransDupFilename, 26, "Duplicate filename in transaction"},
		{"ALPM_ERR_TRANS_NOT_INITIALIZED", alpmerrors.ErrTransNotInitialized, 27, "A transaction has not been initialized"},
		{"ALPM_ERR_TRANS_NOT_PREPARED", alpmerrors.ErrTransNotPrepared, 28, "Transaction has not been prepared"},
		{"ALPM_ERR_TRANS_ABORT", alpmerrors.ErrTransAbort, 29, "Transaction was aborted"},
		{"ALPM_ERR_TRANS_TYPE", alpmerrors.ErrTransType, 30, "Failed to interrupt transaction"},
		{"ALPM_ERR_TRANS_NOT_LOCKED", alpmerrors.ErrTransNotLocked, 31, "Tried to commit transaction without locking the database"},
		{"ALPM_ERR_TRANS_HOOK_FAILED", alpmerrors.ErrTransHookFailed, 32, "A hook failed to run"},
		{"ALPM_ERR_PKG_NOT_FOUND", alpmerrors.ErrPkgNotFound, 33, "Package not found"},
		{"ALPM_ERR_PKG_IGNORED", alpmerrors.ErrPkgIgnored, 34, "Package is in ignorepkg"},
		{"ALPM_ERR_PKG_INVALID", alpmerrors.ErrPkgInvalid, 35, "Package is invalid"},
		{"ALPM_ERR_PKG_INVALID_CHECKSUM", alpmerrors.ErrPkgInvalidChecksum, 36, "Package has an invalid checksum"},
		{"ALPM_ERR_PKG_INVALID_SIG", alpmerrors.ErrPkgInvalidSig, 37, "Package has an invalid signature"},
		{"ALPM_ERR_PKG_MISSING_SIG", alpmerrors.ErrPkgMissingSig, 38, "Package does not have a signature"},
		{"ALPM_ERR_PKG_OPEN", alpmerrors.ErrPkgOpen, 39, "Cannot open the package file"},
		{"ALPM_ERR_PKG_CANT_REMOVE", alpmerrors.ErrPkgCantRemove, 40, "Failed to remove package files"},
		{"ALPM_ERR_PKG_INVALID_NAME", alpmerrors.ErrPkgInvalidName, 41, "Package has an invalid name"},
		{"ALPM_ERR_PKG_INVALID_ARCH", alpmerrors.ErrPkgInvalidArch, 42, "Package has an invalid architecture"},
		{"ALPM_ERR_SIG_MISSING", alpmerrors.ErrSigMissing, 43, "Signatures are missing"},
		{"ALPM_ERR_SIG_INVALID", alpmerrors.ErrSigInvalid, 44, "Signatures are invalid"},
		{"ALPM_ERR_UNSATISFIED_DEPS", alpmerrors.ErrUnsatisfiedDeps, 45, "Dependencies could not be satisfied"},
		{"ALPM_ERR_CONFLICTING_DEPS", alpmerrors.ErrConflictingDeps, 46, "Conflicting dependencies"},
		{"ALPM_ERR_FILE_CONFLICTS", alpmerrors.ErrFileConflicts, 47, "Files conflict"},
		{"ALPM_ERR_RETRIEVE_PREPARE", alpmerrors.ErrRetrievePrepare, 48, "Download setup failed"},
		{"ALPM_ERR_RETRIEVE", alpmerrors.ErrRetrieve, 49, "Download failed"},
		{"ALPM_ERR_INVALID_REGEX", alpmerrors.ErrInvalidRegex, 50, "Invalid Regex"},
		{"ALPM_ERR_LIBARCHIVE", alpmerrors.ErrLibArchive, 51, "Error in libarchive"},
		{"ALPM_ERR_LIBCURL", alpmerrors.ErrLibCurl, 52, "Error in libcurl"},
		{"ALPM_ERR_EXTERNAL_DOWNLOAD", alpmerrors.ErrExternalDownload, 53, "Error in external download program"},
		{"ALPM_ERR_GPGME", alpmerrors.ErrGpgme, 54, "Error in gpgme"},
		{"ALPM_ERR_MISSING_CAPABILITY_SIGNATURES", alpmerrors.ErrMissingCapabilitySignatures, 55, "Missing compile-time features"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := int32(tt.errno); got != tt.value {
				t.Errorf("value = %d, want %d", got, tt.value)
			}
			if got := tt.errno.Error(); got != tt.message {
				t.Errorf("Error() = %q, want %q", got, tt.message)
			}
		})
	}
}

func TestErrnoUnknown(t *testing.T) {
	if got := alpmerrors.Errno(999).Error(); got != "" {
		t.Fatalf("Error() = %q, want empty string", got)
	}
}

func TestALPMError(t *testing.T) {
	err := fmt.Errorf("outer: %w", alpmerrors.NewError(alpmerrors.ErrPkgNotFound, "loading target"))

	if !stderrors.Is(err, alpmerrors.ErrPkgNotFound) {
		t.Fatal("errors.Is did not match ErrPkgNotFound")
	}

	var errno alpmerrors.Errno
	if !stderrors.As(err, &errno) || errno != alpmerrors.ErrPkgNotFound {
		t.Fatalf("errors.As Errno = %v", errno)
	}

	var alpmErr *alpmerrors.ALPMError
	if !stderrors.As(err, &alpmErr) {
		t.Fatal("errors.As did not find ALPMError")
	}
	if alpmErr.Errno != alpmerrors.ErrPkgNotFound || alpmErr.Msg != "loading target" {
		t.Fatalf("ALPMError = %#v", alpmErr)
	}
	if got := alpmErr.Error(); got != "Package not found: loading target" {
		t.Fatalf("Error() = %q", got)
	}

	if got := alpmerrors.NewError(alpmerrors.ErrDiskSpace, "").Error(); got != "Insufficient disk space" {
		t.Fatalf("Error() without context = %q", got)
	}
}

func TestErrorInterfaces(t *testing.T) {
	var _ error = alpmerrors.ErrOK
	var _ error = &alpmerrors.ALPMError{}
}
