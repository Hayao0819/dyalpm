package dyalpm

import alpmerrors "github.com/Hayao0819/dyalpm/errors"

type TransactionOperation string

const (
	TransactionPrepare TransactionOperation = "prepare"
	TransactionCommit  TransactionOperation = "commit"
)

type MissingDependency struct {
	Target         string
	Dependency     Depend
	CausingPackage string
}

func (d MissingDependency) GetTarget() string {
	return d.Target
}

func (d MissingDependency) GetDepend() Dependency {
	return dependencyValue{Depend: d.Dependency}
}

func (d MissingDependency) GetCausingPkg() string {
	return d.CausingPackage
}

func (MissingDependency) Free() {}

type PackageConflict struct {
	Package1 string
	Package2 string
	Reason   Depend
}

func (c PackageConflict) GetPackage1() string {
	return c.Package1
}

func (c PackageConflict) GetPackage2() string {
	return c.Package2
}

func (c PackageConflict) GetReason() Dependency {
	return dependencyValue{Depend: c.Reason}
}

func (PackageConflict) Free() {}

type FileConflictDetail struct {
	Target            string
	Type              FileConflictType
	File              string
	ConflictingTarget string
}

func (c FileConflictDetail) GetTarget() string {
	return c.Target
}

func (c FileConflictDetail) GetType() FileConflictType {
	return c.Type
}

func (c FileConflictDetail) GetFile() string {
	return c.File
}

func (c FileConflictDetail) GetCTarget() string {
	return c.ConflictingTarget
}

func (FileConflictDetail) Free() {}

type TransactionDiagnostics struct {
	InvalidArchitecture []string
	MissingDependencies []MissingDependency
	PackageConflicts    []PackageConflict
	FileConflicts       []FileConflictDetail
	InvalidPackageFiles []string
}

type TransactionError struct {
	Operation   TransactionOperation
	Diagnostics TransactionDiagnostics
	Cause       error
	sentinel    error
}

func (e *TransactionError) Error() string {
	if e == nil {
		return ""
	}
	if e.sentinel == nil {
		if e.Cause != nil {
			return e.Cause.Error()
		}
		if e.Operation != "" {
			return "transaction " + string(e.Operation) + " failed"
		}
		return "transaction failed"
	}
	if e.Cause == nil || e.Cause.Error() == "" {
		return e.sentinel.Error()
	}
	return e.sentinel.Error() + ": " + e.Cause.Error()
}

func (e *TransactionError) Unwrap() []error {
	if e == nil {
		return nil
	}
	if e.sentinel == nil {
		if e.Cause == nil {
			return nil
		}
		return []error{e.Cause}
	}
	if e.Cause == nil {
		return []error{e.sentinel}
	}
	return []error{e.sentinel, e.Cause}
}

type dependencyValue struct {
	Depend
}

func (d dependencyValue) GetName() string {
	return d.Name
}

func (d dependencyValue) GetVersion() string {
	return d.Version
}

func (d dependencyValue) GetMod() DepMod {
	return d.Mod
}

func (d dependencyValue) ComputeString() string {
	return d.String()
}

func (dependencyValue) Free() {}

func newTransactionError(
	operation TransactionOperation,
	errno alpmerrors.Errno,
	diagnostics TransactionDiagnostics,
) *TransactionError {
	sentinel := ErrTransactionPrepareFailed
	if operation == TransactionCommit {
		sentinel = ErrTransactionCommitFailed
	}
	return &TransactionError{
		Operation:   operation,
		Diagnostics: diagnostics,
		Cause:       alpmerrors.NewError(errno, ""),
		sentinel:    sentinel,
	}
}
