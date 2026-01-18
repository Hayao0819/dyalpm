# dyalpm

A Go wrapper for the Arch Linux Package Manager (ALPM) library using [purego](https://github.com/ebitengine/purego).

## Features

- **Lazy Loading**: Functions are loaded and registered only when needed
- **Type-Safe Interfaces**: Clean Go interfaces for all ALPM operations
- **Maintainable Structure**: Well-organized codebase with clear separation of concerns
- **Error Handling**: Comprehensive error types matching ALPM error codes

## Requirements

- Go 1.21 or later
- libalpm.so.15 (Arch Linux package manager library)
- Linux system with ALPM installed

## Installation

```bash
go get github.com/Jguer/dyalpm
```

## Usage

### Basic Example

```go
package main

import (
	"fmt"
	"log"

	alpm "github.com/Jguer/dyalpm"
)

func main() {
	// Initialize ALPM handle
	handle, err := alpm.Initialize("/", "/var/lib/pacman")
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Release()

	// Get local database
	localDB, err := handle.LocalDB()
	if err != nil {
		log.Fatal(err)
	}

	// Get a package
	pkg := localDB.Pkg("pacman")
	if pkg == nil {
		log.Fatal(err)
	}

	fmt.Printf("Package: %s\n", pkg.Name())
	fmt.Printf("Version: %s\n", pkg.Version())
	fmt.Printf("Description: %s\n", pkg.Description())
}
```

### Working with Sync Databases

```go
// List registered sync databases
syncDBs, err := handle.SyncDBs()
if err != nil {
	log.Fatal(err)
}

for _, db := range syncDBs {
	fmt.Printf("%s\n", db.Name())
}
```

### Transactions

```go
// Create a transaction
trans := alpm.NewTransaction(handle)

// Initialize transaction
err := trans.Init(0)
if err != nil {
	log.Fatal(err)
}
defer trans.Release()

// Add a package to install
pkg, _ := syncDB.GetPkg("vim")
err = trans.AddPkg(pkg)
if err != nil {
	log.Fatal(err)
}

// Prepare the transaction
missing, err := trans.Prepare()
if err != nil {
	log.Fatal(err)
}

if len(missing) > 0 {
	fmt.Println("Missing dependencies:")
	for _, dep := range missing {
		fmt.Printf("  %s requires %s\n", dep.GetTarget(), dep.GetDepend().GetName())
	}
}

// Commit the transaction
conflicts, err := trans.Commit()
if err != nil {
	log.Fatal(err)
}

if len(conflicts) > 0 {
	fmt.Println("File conflicts detected!")
}
```

## Architecture

The wrapper is structured as follows:

- **`internal/lib`**: Core library loading and function registry with lazy loading
- **`internal/list`**: ALPM list operations wrapper
- **`internal/errors`**: Error types matching ALPM error codes
- **`handle.go`**: Handle interface and implementation
- **`database.go`**: Database interface and implementation
- **`package.go`**: Package interface and implementation
- **`transaction.go`**: Transaction interface and implementation

## Lazy Loading

All C functions are loaded lazily - they are only resolved from the library when first needed. This improves startup time and allows the library to work even if some functions are unavailable.

## Error Handling

Errors are represented using the `errors.Errno` type which matches ALPM's error codes. You can check for specific errors:

```go
if errno := handle.Errno(); errno != errors.ErrOK {
	fmt.Printf("Error: %s\n", handle.StrError(errno))
}
```

## License

This project is licensed under the same license as ALPM (GPL).
