package alpm

import "errors"

var (
	ErrInvalidDatabase          = errors.New("invalid database")
	ErrInvalidPackage           = errors.New("invalid package")
	ErrInvalidGroup             = errors.New("invalid group")
	ErrInvalidHandle            = errors.New("invalid handle")
	ErrPackageNotFound          = errors.New("package not found")
	ErrGroupNotFound            = errors.New("group not found")
	ErrDatabaseUpdateFailed     = errors.New("database update failed")
	ErrDatabaseUnregisterFailed = errors.New("database unregister failed")
	ErrTransactionInitFailed    = errors.New("transaction init failed")
	ErrTransactionPrepareFailed = errors.New("transaction prepare failed")
	ErrTransactionCommitFailed  = errors.New("transaction commit failed")
	ErrTransactionReleaseFailed = errors.New("transaction release failed")
	ErrAddPackageFailed         = errors.New("add package failed")
	ErrRemovePackageFailed      = errors.New("remove package failed")
	ErrSysupgradeFailed         = errors.New("sysupgrade failed")
	ErrPackageFreeFailed        = errors.New("package free failed")
	ErrPackageLoadFailed        = errors.New("package load failed")
	ErrInvalidDependency        = errors.New("invalid dependency")
)
