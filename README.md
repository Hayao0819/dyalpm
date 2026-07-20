# dyalpm

A Go wrapper for the Arch Linux Package Manager (ALPM) library using [purego](https://github.com/ebitengine/purego).

## Features

- **Dynamic FFI**: Calls libalpm through [purego](https://github.com/ebitengine/purego)
- **Eager Symbol Resolution**: Resolves the libalpm 16 bindings on first use
- **Typed Interfaces**: Go wrappers for supported ALPM operations
- **Maintainable Structure**: Well-organized codebase with clear separation of concerns
- **Error Handling**: Go errors with access to libalpm error details

## Requirements

- Go 1.26 or later
- libalpm.so.16
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
	handle, err := alpm.Initialize("/", "/var/lib/pacman")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := handle.Release(); err != nil {
			log.Printf("release ALPM handle: %v", err)
		}
	}()

	localDB, err := handle.LocalDB()
	if err != nil {
		log.Fatal(err)
	}

	pkg := localDB.Pkg("pacman")
	if pkg == nil {
		log.Fatal("package pacman not found")
	}

	fmt.Printf("Package: %s\n", pkg.Name())
	fmt.Printf("Version: %s\n", pkg.Version())
	fmt.Printf("Description: %s\n", pkg.Description())
}
```

### Working with Sync Databases

```go
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
trans := alpm.NewTransaction(handle)

err := trans.Init(0)
if err != nil {
	log.Fatal(err)
}
defer func() {
	if err := trans.Release(); err != nil {
		log.Printf("release transaction: %v", err)
	}
}()

syncDB, err := handle.SyncDBByName("core")
if err != nil {
	log.Fatal(err)
}
pkg := syncDB.Pkg("vim")
if pkg == nil {
	log.Fatal("package vim not found")
}
err = trans.AddPkg(pkg)
if err != nil {
	log.Fatal(err)
}

if _, err := trans.Prepare(); err != nil {
	log.Fatal(err)
}

if _, err := trans.Commit(); err != nil {
	log.Fatal(err)
}
```

## Library Loading

On first use, dyalpm opens the exact `libalpm.so.16` SONAME, verifies that
`alpm_version` reports ABI major 16, and eagerly resolves its bindings. It does
not fall back to an unversioned `libalpm.so`. A failed load is retried by the
next operation.

## Error Handling

Most operations return Go errors directly. The handle also exposes libalpm's
current error number and message:

```go
if errno := handle.Errno(); errno != 0 {
	fmt.Printf("Error: %s\n", handle.StrError(errno))
}
```
