package alpm

import (
	stderrors "errors"
	"io"
	"runtime"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/Jguer/dyalpm/internal/lib"
	"github.com/Jguer/dyalpm/internal/list"
)

// Package represents an ALPM package (minimal surface used by yay).
type Package interface {
	Name() string
	Version() string
	Description() string
	Architecture() string
	Size() int64
	ISize() int64
	DB() Database
	Depends() []Depend
	Provides() []Depend
	OptionalDepends() []Depend
	Groups() []string
	BuildDate() time.Time
	Reason() PkgReason
	Base() string
	ShouldIgnore() bool
}

// Validation represents package validation status
type Validation int

// PkgValidation represents package validation status
type PkgValidation int

const (
	PkgValidationUnknown   PkgValidation = 0
	PkgValidationNone      PkgValidation = (1 << 0)
	PkgValidationMD5Sum    PkgValidation = (1 << 1)
	PkgValidationSHA256Sum PkgValidation = (1 << 2)
	PkgValidationSignature PkgValidation = (1 << 3)
)

//nolint:revive // legacy name kept for compatibility
type package_ struct {
	ptr      uintptr
	handle   *handle
	registry *lib.FunctionRegistry
}

func newPackage(ptr uintptr, h *handle) *package_ {
	reg, _ := lib.GetRegistry()
	return &package_{
		ptr:      ptr,
		handle:   h,
		registry: reg,
	}
}

func (p *package_) Name() string {
	if p.ptr == 0 {
		return ""
	}

	getNameFn := cachedFunc("alpm_pkg_get_name")
	if getNameFn == 0 {
		return ""
	}

	result := lib.Syscall(getNameFn, p.ptr)
	return lib.PtrToString(result)
}

func (p *package_) Version() string {
	if p.ptr == 0 {
		return ""
	}

	getVersionFn := cachedFunc("alpm_pkg_get_version")
	if getVersionFn == 0 {
		return ""
	}

	result := lib.Syscall(getVersionFn, p.ptr)
	return lib.PtrToString(result)
}

func (p *package_) Description() string {
	if p.ptr == 0 {
		return ""
	}

	getDescFn := cachedFunc("alpm_pkg_get_desc")
	if getDescFn == 0 {
		return ""
	}

	result := lib.Syscall(getDescFn, p.ptr)
	return lib.PtrToString(result)
}

func (p *package_) Architecture() string {
	if p.ptr == 0 {
		return ""
	}

	getArchFn := cachedFunc("alpm_pkg_get_arch")
	if getArchFn == 0 {
		return ""
	}

	result := lib.Syscall(getArchFn, p.ptr)
	return lib.PtrToString(result)
}

func (p *package_) Size() int64 {
	if p.ptr == 0 {
		return 0
	}

	getSizeFn := cachedFunc("alpm_pkg_get_size")
	if getSizeFn == 0 {
		return 0
	}

	result := lib.Syscall(getSizeFn, p.ptr)
	return int64(result)
}

func (p *package_) ISize() int64 {
	if p.ptr == 0 {
		return 0
	}

	getISizeFn := cachedFunc("alpm_pkg_get_isize")
	if getISizeFn == 0 {
		return 0
	}

	result := lib.Syscall(getISizeFn, p.ptr)
	return int64(result)
}

func (p *package_) DB() Database {
	if p.ptr == 0 {
		return nil
	}

	getDBFn := cachedFunc("alpm_pkg_get_db")
	if getDBFn == 0 {
		return nil
	}

	dbPtr := lib.Syscall(getDBFn, p.ptr)
	if dbPtr == 0 {
		return nil
	}

	return newDatabase(dbPtr, p.handle)
}

func (p *package_) Depends() []Depend {
	deps, _ := p.getDependencyList("alpm_pkg_get_depends")
	return toDependList(deps)
}

func (p *package_) Conflicts() []Depend {
	deps, _ := p.getDependencyList("alpm_pkg_get_conflicts")
	return toDependList(deps)
}

func (p *package_) Provides() []Depend {
	deps, _ := p.getDependencyList("alpm_pkg_get_provides")
	return toDependList(deps)
}

func (p *package_) getDependencyList(funcName string) ([]Dependency, error) {
	if p.ptr == 0 {
		return nil, ErrInvalidPackage
	}

	getFn := cachedFunc(funcName)
	if getFn == 0 {
		return nil, stderrors.New("missing function: " + funcName)
	}

	listPtr := lib.Syscall(getFn, p.ptr)
	if listPtr == 0 {
		return []Dependency{}, nil
	}

	alpmList := list.NewList(listPtr)
	if alpmList == nil {
		return []Dependency{}, nil
	}

	var deps []Dependency
	for item := alpmList; item != nil && item.Ptr() != 0; item = item.Next() {
		depPtr := item.Data()
		if depPtr != 0 {
			deps = append(deps, newDependency(depPtr))
		}
	}

	return deps, nil
}

// PkgFind finds a package in a list by name
func PkgFind(pkgs []Package, name string) Package {
	if len(pkgs) == 0 {
		return nil
	}

	reg, err := lib.GetRegistry()
	if err != nil {
		return nil
	}

	fn, err := reg.GetFunc("alpm_pkg_find")
	if err != nil {
		return nil
	}

	var alpmList *list.List
	for _, p := range pkgs {
		pkgImpl, ok := p.(*package_)
		if ok {
			alpmList = list.Add(alpmList, pkgImpl.ptr)
		}
	}
	defer alpmList.Free()

	nameBytes := lib.CString(name)
	pkgPtr, _, _ := purego.SyscallN(fn, alpmList.Ptr(), uintptr(unsafe.Pointer(&nameBytes[0])))
	runtime.KeepAlive(nameBytes)

	if pkgPtr == 0 {
		return nil
	}

	// We need a handle to create a Package.
	// This is a bit tricky since PkgFind doesn't have a handle.
	// But the packages in the input list should have handles.
	var h *handle
	if len(pkgs) > 0 {
		pkgImpl, ok := pkgs[0].(*package_)
		if ok {
			h = pkgImpl.handle
		}
	}

	return newPackage(pkgPtr, h)
}

// VerCmp compares two version strings according to libalpm version comparison rules.
// Returns <0 if v1 < v2, 0 if v1 == v2, >0 if v1 > v2.
func VerCmp(v1, v2 string) int {
	return vercmpPureGo(v1, v2)
}

// vercmpPureGo is a pure Go implementation of version comparison
// following pacman/libalpm version comparison conventions.
func vercmpPureGo(a, b string) int {
	if a == b {
		return 0
	}

	// Split into epoch:version-rel
	ae, av, ar := splitVersion(a)
	be, bv, br := splitVersion(b)

	// Compare epochs
	if ae != be {
		return compareNumericString(ae, be)
	}

	// Compare versions
	ret := compareVersionSegment(av, bv)
	if ret != 0 {
		return ret
	}

	// Compare releases
	return compareVersionSegment(ar, br)
}

// splitVersion splits a version string into epoch, version, and release components
func splitVersion(v string) (epoch, version, rel string) {
	epoch = "0"
	version = v
	rel = ""

	// Find epoch (before ':')
	for i, c := range v {
		if c == ':' {
			epoch = v[:i]
			version = v[i+1:]
			break
		}
	}

	// Find release (after last '-')
	for i := len(version) - 1; i >= 0; i-- {
		if version[i] == '-' {
			rel = version[i+1:]
			version = version[:i]
			break
		}
	}

	return
}

// compareVersionSegment compares two version segments (like "1.2.3" vs "1.2.4")
func compareVersionSegment(a, b string) int {
	if a == b {
		return 0
	}

	i, j := 0, 0
	for i < len(a) || j < len(b) {
		// Skip non-alphanumeric
		for i < len(a) && !isAlnum(a[i]) {
			i++
		}
		for j < len(b) && !isAlnum(b[j]) {
			j++
		}

		// If either is exhausted
		if i >= len(a) && j >= len(b) {
			return 0
		}
		if i >= len(a) {
			return -1
		}
		if j >= len(b) {
			return 1
		}

		// Determine segment type and extract
		var segA, segB string
		var numA, numB bool

		if isDigit(a[i]) {
			numA = true
			start := i
			for i < len(a) && isDigit(a[i]) {
				i++
			}
			segA = a[start:i]
		} else {
			start := i
			for i < len(a) && isAlpha(a[i]) {
				i++
			}
			segA = a[start:i]
		}

		if isDigit(b[j]) {
			numB = true
			start := j
			for j < len(b) && isDigit(b[j]) {
				j++
			}
			segB = b[start:j]
		} else {
			start := j
			for j < len(b) && isAlpha(b[j]) {
				j++
			}
			segB = b[start:j]
		}

		// If types differ, numeric wins
		if numA && !numB {
			return 1
		}
		if !numA && numB {
			return -1
		}

		// Compare same types
		if numA {
			ret := compareNumericString(segA, segB)
			if ret != 0 {
				return ret
			}
		} else {
			if segA < segB {
				return -1
			}
			if segA > segB {
				return 1
			}
		}
	}

	return 0
}

// compareNumericString compares two numeric strings
func compareNumericString(a, b string) int {
	// Strip leading zeros
	a = stripLeadingZeros(a)
	b = stripLeadingZeros(b)

	// Compare lengths first
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}

	// Same length, compare lexicographically
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func stripLeadingZeros(s string) string {
	i := 0
	for i < len(s)-1 && s[i] == '0' {
		i++
	}
	return s[i:]
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
func isAlpha(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isAlnum(c byte) bool { return isDigit(c) || isAlpha(c) }

func (p *package_) CheckDepends() []Depend {
	deps, _ := p.getDependencyList("alpm_pkg_get_checkdepends")
	return toDependList(deps)
}

func (p *package_) MakeDepends() []Depend {
	deps, _ := p.getDependencyList("alpm_pkg_get_makedepends")
	return toDependList(deps)
}

func (p *package_) OptionalDepends() []Depend {
	deps, _ := p.getDependencyList("alpm_pkg_get_optdepends")
	return toDependList(deps)
}

func (p *package_) Replaces() []Depend {
	deps, _ := p.getDependencyList("alpm_pkg_get_replaces")
	return toDependList(deps)
}

func (p *package_) Groups() []string {
	groups, _ := p.getStringList("alpm_pkg_get_groups")
	return groups
}

func (p *package_) Licenses() []string {
	licenses, _ := p.getStringList("alpm_pkg_get_licenses")
	return licenses
}

func (p *package_) getStringList(funcName string) ([]string, error) {
	return p.getStringListWithFree(funcName, false)
}

func (p *package_) getStringListWithFree(funcName string, freeList bool) ([]string, error) {
	if p.ptr == 0 {
		return nil, ErrInvalidPackage
	}

	fn := cachedFunc(funcName)
	if fn == 0 {
		return nil, stderrors.New("missing function: " + funcName)
	}

	r1, _, _ := purego.SyscallN(fn, p.ptr)
	if r1 == 0 {
		return []string{}, nil
	}

	alpmList := list.NewList(r1)
	if alpmList == nil {
		return []string{}, nil
	}
	if freeList {
		defer alpmList.Free()
	}

	var items []string
	for item := alpmList; item != nil && item.Ptr() != 0; item = item.Next() {
		strPtr := item.Data()
		if strPtr != 0 {
			items = append(items, lib.PtrToString(strPtr))
		}
	}

	return items, nil
}

// Backup represents a backup file
type Backup interface {
	Name() string
	Hash() string
}

type backup struct {
	ptr uintptr
}

func newBackup(ptr uintptr) *backup {
	return &backup{ptr: ptr}
}

func (b *backup) Name() string {
	if b.ptr == 0 {
		return ""
	}
	base := unsafe.Pointer(b.ptr)
	namePtr := *(*uintptr)(base)
	return lib.PtrToString(namePtr)
}

func (b *backup) Hash() string {
	if b.ptr == 0 {
		return ""
	}
	// hash is second pointer
	base := unsafe.Pointer(b.ptr)
	hashPtr := *(*uintptr)(unsafe.Add(base, unsafe.Sizeof(uintptr(0))))
	return lib.PtrToString(hashPtr)
}

func (p *package_) Backup() []Backup {
	if p.ptr == 0 {
		return []Backup{}
	}

	fn, err := p.registry.GetFunc("alpm_pkg_get_backup")
	if err != nil {
		return []Backup{}
	}

	r1, _, _ := purego.SyscallN(fn, p.ptr)
	if r1 == 0 {
		return []Backup{}
	}

	alpmList := list.NewList(r1)
	if alpmList == nil {
		return []Backup{}
	}

	var backups []Backup
	for item := alpmList; item != nil && item.Ptr() != 0; item = item.Next() {
		dataPtr := item.Data()
		if dataPtr != 0 {
			backups = append(backups, newBackup(dataPtr))
		}
	}

	return backups
}

// File represents a file in a package
type File interface {
	Name() string
	Size() int64
	Mode() uint32
}

type file struct {
	name string
	size int64
	mode uint32
}

func (f *file) Name() string { return f.name }
func (f *file) Size() int64  { return f.size }
func (f *file) Mode() uint32 { return f.mode }

func (p *package_) Files() []File {
	if p.ptr == 0 {
		return []File{}
	}

	fn, err := p.registry.GetFunc("alpm_pkg_get_files")
	if err != nil {
		return []File{}
	}

	// alpm_filelist_t* returned
	r1, _, _ := purego.SyscallN(fn, p.ptr)
	if r1 == 0 {
		return []File{}
	}

	// Read count (size_t)
	// assuming size_t is uintptr size (safe assumption usually)
	base := unsafe.Pointer(r1)
	count := *(*uintptr)(base)
	if count == 0 {
		return []File{}
	}

	// Read files pointer (alpm_file_t*)
	filesPtr := *(*uintptr)(unsafe.Add(base, unsafe.Sizeof(uintptr(0))))
	if filesPtr == 0 {
		return []File{}
	}

	// Need alpm_file_t size
	// struct { char *name; off_t size; mode_t mode; }
	ptrSize := unsafe.Sizeof(uintptr(0))
	offSize := unsafe.Sizeof(int64(0))   // assuming off_t is 64bit
	modeSize := unsafe.Sizeof(uint32(0)) // assuming mode_t is 32bit

	structSize := ptrSize + offSize + modeSize
	// Add padding for alignment if needed
	if structSize%ptrSize != 0 {
		structSize += ptrSize - (structSize % ptrSize)
	}

	var files []File
	filesBase := unsafe.Pointer(filesPtr)
	for i := 0; i < int(count); i++ {
		current := unsafe.Add(filesBase, uintptr(i)*structSize)

		namePtr := *(*uintptr)(current)
		name := lib.PtrToString(namePtr)

		size := *(*int64)(unsafe.Add(current, ptrSize))
		mode := *(*uint32)(unsafe.Add(current, ptrSize+offSize))

		files = append(files, &file{
			name: name,
			size: size,
			mode: mode,
		})
	}

	return files
}

// PkgFrom represents where the package came from
type PkgFrom int

const (
	PkgFromFile    PkgFrom = 1
	PkgFromLocalDB PkgFrom = 2
	PkgFromSyncDB  PkgFrom = 3
)

// PkgReason represents why the package was installed
type PkgReason int

const (
	PkgReasonExplicit PkgReason = 0
	PkgReasonDepend   PkgReason = 1
	PkgReasonUnknown  PkgReason = 2
)

func (p *package_) Origin() PkgFrom {
	if p.ptr == 0 {
		return PkgFromFile // Default or error?
	}

	fn, err := p.registry.GetFunc("alpm_pkg_get_origin")
	if err != nil {
		return PkgFromFile
	}

	r1, _, _ := purego.SyscallN(fn, p.ptr)
	return PkgFrom(r1)
}

func (p *package_) BuildDate() time.Time {
	if p.ptr == 0 {
		return time.Time{}
	}

	fn := cachedFunc("alpm_pkg_get_builddate")
	if fn == 0 {
		return time.Time{}
	}

	r1, _, _ := purego.SyscallN(fn, p.ptr)
	return toTime(int64(r1))
}

func (p *package_) InstallDate() time.Time {
	if p.ptr == 0 {
		return time.Time{}
	}

	fn, err := p.registry.GetFunc("alpm_pkg_get_installdate")
	if err != nil {
		return time.Time{}
	}

	r1, _, _ := purego.SyscallN(fn, p.ptr)
	return toTime(int64(r1))
}

func (p *package_) Reason() PkgReason {
	if p.ptr == 0 {
		return PkgReasonExplicit
	}

	fn := cachedFunc("alpm_pkg_get_reason")
	if fn == 0 {
		return PkgReasonExplicit
	}

	r1, _, _ := purego.SyscallN(fn, p.ptr)
	return PkgReason(r1)
}

func (p *package_) HasScriptlet() bool {
	if p.ptr == 0 {
		return false
	}

	fn, err := p.registry.GetFunc("alpm_pkg_has_scriptlet")
	if err != nil {
		return false
	}

	r1, _, _ := purego.SyscallN(fn, p.ptr)
	return r1 != 0
}

func (p *package_) DownloadSize() int64 {
	if p.ptr == 0 {
		return 0
	}

	fn, err := p.registry.GetFunc("alpm_pkg_download_size")
	if err != nil {
		return 0
	}

	r1, _, _ := purego.SyscallN(fn, p.ptr)
	return int64(r1)
}

func (p *package_) Free() error {
	if p.ptr == 0 {
		return nil
	}

	// Only free if origin is FILE
	if p.Origin() != PkgFromFile {
		return nil
	}

	fn, err := p.registry.GetFunc("alpm_pkg_free")
	if err != nil {
		return err
	}

	r1, _, _ := purego.SyscallN(fn, p.ptr)
	if r1 != 0 {
		return ErrPackageFreeFailed
	}

	p.ptr = 0
	return nil
}

func (p *package_) ComputeRequiredBy() ([]string, error) {
	return p.getStringListWithFree("alpm_pkg_compute_requiredby", true)
}

func (p *package_) ComputeOptionalFor() ([]string, error) {
	return p.getStringListWithFree("alpm_pkg_compute_optionalfor", true)
}

func (p *package_) ShouldIgnore() bool {
	if p.ptr == 0 || p.handle == nil || p.handle.ptr == 0 {
		return false
	}

	fn := cachedFunc("alpm_pkg_should_ignore")
	if fn == 0 {
		return false
	}

	r1, _, _ := purego.SyscallN(fn, p.handle.ptr, p.ptr)
	return r1 != 0
}

func (p *package_) CheckMD5Sum() error {
	if p.ptr == 0 {
		return ErrInvalidPackage
	}

	fn, err := p.registry.GetFunc("alpm_pkg_checkmd5sum")
	if err != nil {
		return err
	}

	r1, _, _ := purego.SyscallN(fn, p.ptr)
	if r1 != 0 {
		return stderrors.New("MD5 sum mismatch")
	}

	return nil
}

func (p *package_) NativeHandle() Handle {
	if p.ptr == 0 {
		return nil
	}

	fn, err := p.registry.GetFunc("alpm_pkg_get_handle")
	if err != nil {
		return nil
	}

	result := lib.Syscall(fn, p.ptr)
	if result == 0 {
		return nil
	}

	return p.handle
}

func (p *package_) CheckPGPSignature() (SigList, error) {
	if p.ptr == 0 {
		return SigList{}, ErrInvalidPackage
	}

	return checkPGPSignature(p.ptr, p.registry, p.handle, "alpm_pkg_check_pgp_signature")
}

func (p *package_) Sig() ([]byte, error) {
	if p.ptr == 0 {
		return nil, ErrInvalidPackage
	}

	fn, err := p.registry.GetFunc("alpm_pkg_get_sig")
	if err != nil {
		return nil, err
	}

	// alpm_pkg_get_sig signature: int alpm_pkg_get_sig(pkg, &sig, &sig_len)
	// Returns error code and writes signature bytes to output parameters
	var sigPtr uintptr
	var sigLen uintptr
	result := lib.Syscall(fn, p.ptr, uintptr(unsafe.Pointer(&sigPtr)), uintptr(unsafe.Pointer(&sigLen)))
	if result != 0 || sigPtr == 0 || sigLen == 0 {
		return nil, nil // No signature or error
	}

	// Copy the signature bytes to a Go slice
	sig := make([]byte, sigLen)
	base := unsafe.Pointer(sigPtr)
	for i := uintptr(0); i < sigLen; i++ {
		sig[i] = *(*byte)(unsafe.Add(base, i))
	}

	return sig, nil
}

// Base64Sig returns the base64-encoded package signature.
func (p *package_) Base64Sig() string {
	if p.ptr == 0 {
		return ""
	}

	fn, err := p.registry.GetFunc("alpm_pkg_get_base64_sig")
	if err != nil {
		return ""
	}

	result := lib.Syscall(fn, p.ptr)
	return lib.PtrToString(result)
}

func (p *package_) PkgValidation() PkgValidation {
	if p.ptr == 0 {
		return PkgValidationUnknown
	}

	fn, err := p.registry.GetFunc("alpm_pkg_get_validation")
	if err != nil {
		return PkgValidationUnknown
	}

	result := lib.Syscall(fn, p.ptr)
	return PkgValidation(result)
}

func (p *package_) XData() []string {
	if p.ptr == 0 {
		return nil
	}

	// alpm_pkg_get_xdata returns a list owned by the package
	result, _ := p.getStringList("alpm_pkg_get_xdata")
	return result
}

func (p *package_) Contains(path string) bool {
	if p.ptr == 0 {
		return false
	}

	getFilesFn, err := p.registry.GetFunc("alpm_pkg_get_files")
	if err != nil {
		return false
	}

	filelistPtr := lib.Syscall(getFilesFn, p.ptr)
	if filelistPtr == 0 {
		return false
	}

	containsFn, err := p.registry.GetFunc("alpm_filelist_contains")
	if err != nil {
		return false
	}

	cPath := lib.CString(path)
	result := lib.Syscall(containsFn, filelistPtr, uintptr(unsafe.Pointer(&cPath[0])))
	runtime.KeepAlive(cPath)

	return result != 0
}

func (p *package_) Changelog() (io.ReadCloser, error) {
	if p.ptr == 0 {
		return nil, ErrInvalidPackage
	}

	fn, err := p.registry.GetFunc("alpm_pkg_changelog_open")
	if err != nil {
		return nil, err
	}

	fp := lib.Syscall(fn, p.ptr)
	if fp == 0 {
		return nil, stderrors.New("no changelog found")
	}

	return &changelogReader{
		pkg: p,
		fp:  fp,
	}, nil
}

func (p *package_) SyncGetNewVersion(dbsSync []Database) Package {
	if p.ptr == 0 {
		return nil
	}

	fn, err := p.registry.GetFunc("alpm_sync_get_new_version")
	if err != nil {
		return nil
	}

	var dbList *list.List
	for _, db := range dbsSync {
		dbImpl, ok := db.(*database)
		if ok {
			dbList = list.Add(dbList, dbImpl.ptr)
		}
	}
	defer dbList.Free()

	r1 := lib.Syscall(fn, p.ptr, dbList.Ptr())
	if r1 == 0 {
		return nil
	}

	return newPackage(r1, p.handle)
}

type changelogReader struct {
	pkg *package_
	fp  uintptr
}

func (r *changelogReader) Read(p []byte) (n int, err error) {
	if r.fp == 0 {
		return 0, io.EOF
	}

	fn, err := r.pkg.registry.GetFunc("alpm_pkg_changelog_read")
	if err != nil {
		return 0, err
	}

	// size_t alpm_pkg_changelog_read(void *ptr, size_t size, const alpm_pkg_t *pkg, void *fp);
	res, _, _ := purego.SyscallN(fn, uintptr(unsafe.Pointer(&p[0])), uintptr(len(p)), r.pkg.ptr, r.fp)
	if res == 0 {
		return 0, io.EOF
	}
	return int(res), nil
}

func (r *changelogReader) Close() error {
	if r.fp == 0 {
		return nil
	}

	fn, err := r.pkg.registry.GetFunc("alpm_pkg_changelog_close")
	if err != nil {
		return err
	}

	purego.SyscallN(fn, r.pkg.ptr, r.fp)
	r.fp = 0
	return nil
}

func (p *package_) Validation() Validation {
	return Validation(p.PkgValidation())
}

func (p *package_) Base() string {
	if p.ptr == 0 {
		return ""
	}
	fn := cachedFunc("alpm_pkg_get_base")
	if fn == 0 {
		return ""
	}
	result := lib.Syscall(fn, p.ptr)
	return lib.PtrToString(result)
}

func (p *package_) FileName() string {
	if p.ptr == 0 {
		return ""
	}
	fn, err := p.registry.GetFunc("alpm_pkg_get_filename")
	if err != nil {
		return ""
	}
	result := lib.Syscall(fn, p.ptr)
	return lib.PtrToString(result)
}

func (p *package_) Base64Signature() string {
	return p.Base64Sig()
}

func (p *package_) SHA256Sum() string {
	if p.ptr == 0 {
		return ""
	}
	fn, err := p.registry.GetFunc("alpm_pkg_get_sha256sum")
	if err != nil {
		return ""
	}
	result := lib.Syscall(fn, p.ptr)
	return lib.PtrToString(result)
}

func (p *package_) Packager() string {
	if p.ptr == 0 {
		return ""
	}
	fn, err := p.registry.GetFunc("alpm_pkg_get_packager")
	if err != nil {
		return ""
	}
	result := lib.Syscall(fn, p.ptr)
	return lib.PtrToString(result)
}

func (p *package_) URL() string {
	if p.ptr == 0 {
		return ""
	}
	fn, err := p.registry.GetFunc("alpm_pkg_get_url")
	if err != nil {
		return ""
	}
	result := lib.Syscall(fn, p.ptr)
	return lib.PtrToString(result)
}

func (p *package_) Type() string {
	// Returns "pkg" for regular packages
	return "pkg"
}

func (p *package_) ContainsFile(path string) (File, error) {
	if p.Contains(path) {
		files := p.Files()
		for _, f := range files {
			if f.Name() == path {
				return f, nil
			}
		}
	}
	return nil, stderrors.New("file not found")
}

func (p *package_) SyncNewVersion(dbs []Database) Package {
	return p.SyncGetNewVersion(dbs)
}

// toTime converts Unix timestamp to time.Time
func toTime(ts int64) time.Time {
	if ts == 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}
