package dyalpm

import (
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/Jguer/dyalpm/internal/dyerrors"
	"github.com/Jguer/dyalpm/internal/lib"
	"github.com/Jguer/dyalpm/internal/list"
)

// DepMod represents dependency version comparison mode
type DepMod int

const (
	DepModAny DepMod = iota + 1
	DepModEQ
	DepModGE
	DepModLE
	DepModGT
	DepModLT
)

// Dependency represents a package dependency
type Dependency interface {
	GetName() string
	GetVersion() string
	GetMod() DepMod
	ComputeString() string
	Free()
}

// Depend is a value type representation of a dependency for consumer code.
type Depend struct {
	Name        string
	Version     string
	Description string
	NameHash    uint64
	Mod         DepMod
}

// String returns the computed string representation of the dependency.
func (d Depend) String() string {
	if d.Version == "" {
		return d.Name
	}
	modStr := ""
	switch d.Mod {
	case DepModEQ:
		modStr = "="
	case DepModGE:
		modStr = ">="
	case DepModLE:
		modStr = "<="
	case DepModGT:
		modStr = ">"
	case DepModLT:
		modStr = "<"
	}
	return d.Name + modStr + d.Version
}

type dependency struct {
	ptr      uintptr
	registry *lib.FunctionRegistry
	owned    bool
}

func newDependency(ptr uintptr) *dependency {
	reg, _ := lib.GetRegistry()
	return &dependency{
		ptr:      ptr,
		registry: reg,
		owned:    false,
	}
}

func (d *dependency) GetName() string {
	if d.ptr == 0 {
		return ""
	}
	// alpm_depend_t structure: char *name at offset 0
	base := unsafe.Pointer(d.ptr)
	namePtr := *(*uintptr)(base)
	return lib.PtrToString(namePtr)
}

func (d *dependency) GetVersion() string {
	if d.ptr == 0 {
		return ""
	}
	// alpm_depend_t structure: char *version at offset of pointer size
	base := unsafe.Pointer(d.ptr)
	versionPtr := *(*uintptr)(unsafe.Add(base, unsafe.Sizeof(uintptr(0))))
	return lib.PtrToString(versionPtr)
}

func (d *dependency) GetMod() DepMod {
	if d.ptr == 0 {
		return DepModAny
	}
	// alpm_depend_t structure: alpm_depmod_t mod is after name_hash (unsigned long)
	// Offset: name (ptr) + version (ptr) + desc (ptr) + name_hash (ulong) = 3*ptr + ulong
	modOffset := 3*unsafe.Sizeof(uintptr(0)) + unsafe.Sizeof(uint64(0))
	base := unsafe.Pointer(d.ptr)
	mod := *(*int)(unsafe.Add(base, modOffset))
	return DepMod(mod)
}

func (d *dependency) ComputeString() string {
	if d.ptr == 0 {
		return ""
	}

	fn := cachedFunc("alpm_dep_compute_string")
	if fn == 0 {
		return ""
	}

	r1, _, _ := purego.SyscallN(fn, d.ptr)
	if r1 == 0 {
		return ""
	}

	// The string returned by alpm_dep_compute_string MUST be freed by the caller.
	res := lib.PtrToString(r1)
	lib.Free(r1)
	return res
}

func (d *dependency) Free() {
	if d.ptr == 0 || !d.owned {
		return
	}

	fn, err := d.registry.GetFunc("alpm_dep_free")
	if err != nil {
		return
	}

	purego.SyscallN(fn, d.ptr)
	d.ptr = 0
}

// DepFromString creates a dependency from a string
func DepFromString(depstring string) (Dependency, error) {
	reg, err := lib.GetRegistry()
	if err != nil {
		return nil, err
	}

	fn, err := reg.GetFunc("alpm_dep_from_string")
	if err != nil {
		return nil, err
	}

	cStr := lib.CString(depstring)
	r1, _, _ := purego.SyscallN(fn, uintptr(unsafe.Pointer(&cStr[0])))
	runtime.KeepAlive(cStr)

	if r1 == 0 {
		return nil, ErrInvalidDependency
	}

	dep := newDependency(r1)
	dep.owned = true // alpm_dep_from_string returns a newly allocated struct
	return dep, nil
}

// DepMissing represents a missing dependency
type DepMissing interface {
	GetTarget() string
	GetDepend() Dependency
	GetCausingPkg() string
	Free()
}

type depMissing struct {
	ptr      uintptr
	registry *lib.FunctionRegistry
}

func newDepMissing(ptr uintptr) *depMissing {
	reg, _ := lib.GetRegistry()
	return &depMissing{
		ptr:      ptr,
		registry: reg,
	}
}

func (d *depMissing) GetTarget() string {
	if d.ptr == 0 {
		return ""
	}
	// struct _alpm_depmissing_t { char *target; alpm_depend_t *depend; char *causingpkg; }
	base := unsafe.Pointer(d.ptr)
	targetPtr := *(*uintptr)(base)
	return lib.PtrToString(targetPtr)
}

func (d *depMissing) GetDepend() Dependency {
	if d.ptr == 0 {
		return nil
	}
	base := unsafe.Pointer(d.ptr)
	depPtr := *(*uintptr)(unsafe.Add(base, unsafe.Sizeof(uintptr(0))))
	return newDependency(depPtr)
}

func (d *depMissing) GetCausingPkg() string {
	if d.ptr == 0 {
		return ""
	}
	base := unsafe.Pointer(d.ptr)
	pkgPtr := *(*uintptr)(unsafe.Add(base, 2*unsafe.Sizeof(uintptr(0))))
	return lib.PtrToString(pkgPtr)
}

func (d *depMissing) Free() {
	if d.ptr == 0 {
		return
	}
	fn, err := d.registry.GetFunc("alpm_depmissing_free")
	if err != nil {
		return
	}
	purego.SyscallN(fn, d.ptr)
	d.ptr = 0
}

// Conflict represents a package conflict
type Conflict interface {
	GetPackage1() string
	GetPackage2() string
	GetReason() Dependency
	Free()
}

type conflict struct {
	ptr      uintptr
	registry *lib.FunctionRegistry
}

func newConflict(ptr uintptr) *conflict {
	reg, _ := lib.GetRegistry()
	return &conflict{
		ptr:      ptr,
		registry: reg,
	}
}

func (c *conflict) GetPackage1() string {
	if c.ptr == 0 {
		return ""
	}
	// struct _alpm_conflict_t { ulong hash1; ulong hash2; char *pkg1; char *pkg2; alpm_depend_t *reason; }
	base := unsafe.Pointer(c.ptr)
	pkg1Ptr := *(*uintptr)(unsafe.Add(base, 2*unsafe.Sizeof(uintptr(0))))
	return lib.PtrToString(pkg1Ptr)
}

func (c *conflict) GetPackage2() string {
	if c.ptr == 0 {
		return ""
	}
	base := unsafe.Pointer(c.ptr)
	pkg2Ptr := *(*uintptr)(unsafe.Add(base, 3*unsafe.Sizeof(uintptr(0))))
	return lib.PtrToString(pkg2Ptr)
}

func (c *conflict) GetReason() Dependency {
	if c.ptr == 0 {
		return nil
	}
	base := unsafe.Pointer(c.ptr)
	reasonPtr := *(*uintptr)(unsafe.Add(base, 4*unsafe.Sizeof(uintptr(0))))
	return newDependency(reasonPtr)
}

// toDependList converts a Dependency slice into a Depend slice.
func toDependList(deps []Dependency) []Depend {
	result := make([]Depend, len(deps))
	for i, dep := range deps {
		result[i] = Depend{
			Name:    dep.GetName(),
			Version: dep.GetVersion(),
			Mod:     dep.GetMod(),
		}
	}
	return result
}

func (c *conflict) Free() {
	if c.ptr == 0 {
		return
	}
	fn, err := c.registry.GetFunc("alpm_conflict_free")
	if err != nil {
		return
	}
	purego.SyscallN(fn, c.ptr)
	c.ptr = 0
}

// FileConflictType represents the type of file conflict
type FileConflictType int

const (
	FileConflictTarget     FileConflictType = 1
	FileConflictFilesystem FileConflictType = 2
)

// FileConflict represents a file conflict
type FileConflict interface {
	GetTarget() string
	GetType() FileConflictType
	GetFile() string
	GetCTarget() string
	Free()
}

type fileConflict struct {
	ptr      uintptr
	registry *lib.FunctionRegistry
}

func newFileConflict(ptr uintptr) *fileConflict {
	reg, _ := lib.GetRegistry()
	return &fileConflict{
		ptr:      ptr,
		registry: reg,
	}
}

func (f *fileConflict) GetTarget() string {
	if f.ptr == 0 {
		return ""
	}
	// struct _alpm_fileconflict_t { char *target; type; char *file; char *ctarget; }
	base := unsafe.Pointer(f.ptr)
	targetPtr := *(*uintptr)(base)
	return lib.PtrToString(targetPtr)
}

func (f *fileConflict) GetType() FileConflictType {
	if f.ptr == 0 {
		return 0
	}
	base := unsafe.Pointer(f.ptr)
	typeVal := *(*int)(unsafe.Add(base, unsafe.Sizeof(uintptr(0))))
	return FileConflictType(typeVal)
}

func (f *fileConflict) GetFile() string {
	if f.ptr == 0 {
		return ""
	}
	base := unsafe.Pointer(f.ptr)
	filePtr := *(*uintptr)(unsafe.Add(base, 2*unsafe.Sizeof(uintptr(0))))
	return lib.PtrToString(filePtr)
}

func (f *fileConflict) GetCTarget() string {
	if f.ptr == 0 {
		return ""
	}
	base := unsafe.Pointer(f.ptr)
	ctargetPtr := *(*uintptr)(unsafe.Add(base, 3*unsafe.Sizeof(uintptr(0))))
	return lib.PtrToString(ctargetPtr)
}

func (f *fileConflict) Free() {
	if f.ptr == 0 {
		return
	}
	fn, err := f.registry.GetFunc("alpm_fileconflict_free")
	if err != nil {
		return
	}
	purego.SyscallN(fn, f.ptr)
	f.ptr = 0
}

// Resolution functions

func (h *handle) CheckDeps(pkgs []Package, remPkgs []Package, upgradePkgs []Package, reverseDeps bool) ([]DepMissing, error) {
	if h.ptr == 0 {
		return nil, dyerrors.ErrHandleNull
	}

	fn, err := h.registry.GetFunc("alpm_checkdeps")
	if err != nil {
		return nil, err
	}

	var pkgList, remPkgList, upgradePkgList *list.List
	for _, p := range pkgs {
		pkgImpl, _ := p.(*package_)
		pkgList = list.Add(pkgList, pkgImpl.ptr)
	}
	defer pkgList.Free()

	for _, p := range remPkgs {
		pkgImpl, _ := p.(*package_)
		remPkgList = list.Add(remPkgList, pkgImpl.ptr)
	}
	defer remPkgList.Free()

	for _, p := range upgradePkgs {
		pkgImpl, _ := p.(*package_)
		upgradePkgList = list.Add(upgradePkgList, pkgImpl.ptr)
	}
	defer upgradePkgList.Free()

	r1, _, _ := purego.SyscallN(fn, h.ptr, pkgList.Ptr(), remPkgList.Ptr(), upgradePkgList.Ptr(), lib.BoolToInt(reverseDeps))
	if r1 == 0 {
		return []DepMissing{}, nil
	}

	resList := list.NewList(r1)
	defer resList.Free()

	var missings []DepMissing
	for item := resList; item != nil && item.Ptr() != 0; item = item.Next() {
		ptr := item.Data()
		if ptr != 0 {
			missings = append(missings, newDepMissing(ptr))
		}
	}

	return missings, nil
}

func (h *handle) CheckConflicts(pkgs []Package) ([]Conflict, error) {
	if h.ptr == 0 {
		return nil, dyerrors.ErrHandleNull
	}

	fn, err := h.registry.GetFunc("alpm_checkconflicts")
	if err != nil {
		return nil, err
	}

	var pkgList *list.List
	for _, p := range pkgs {
		pkgImpl, _ := p.(*package_)
		pkgList = list.Add(pkgList, pkgImpl.ptr)
	}
	defer pkgList.Free()

	r1, _, _ := purego.SyscallN(fn, h.ptr, pkgList.Ptr())
	if r1 == 0 {
		return []Conflict{}, nil
	}

	resList := list.NewList(r1)
	defer resList.Free()

	var conflicts []Conflict
	for item := resList; item != nil && item.Ptr() != 0; item = item.Next() {
		ptr := item.Data()
		if ptr != 0 {
			conflicts = append(conflicts, newConflict(ptr))
		}
	}

	return conflicts, nil
}

// FindSatisfier finds a package that satisfies a dependency in a list of packages
func FindSatisfier(pkgs []Package, depstring string) Package {
	reg, err := lib.GetRegistry()
	if err != nil {
		return nil
	}

	fn, err := reg.GetFunc("alpm_find_satisfier")
	if err != nil {
		return nil
	}

	var pkgList *list.List
	var h *handle
	for _, p := range pkgs {
		pkgImpl, ok := p.(*package_)
		if ok {
			pkgList = list.Add(pkgList, pkgImpl.ptr)
			if h == nil {
				h = pkgImpl.handle
			}
		}
	}
	defer pkgList.Free()

	cStr := lib.CString(depstring)
	r1, _, _ := purego.SyscallN(fn, pkgList.Ptr(), uintptr(unsafe.Pointer(&cStr[0])))
	runtime.KeepAlive(cStr)

	if r1 == 0 {
		return nil
	}

	return newPackage(r1, h)
}

func (h *handle) FindDBSatisfier(dbs []Database, depstring string) Package {
	if h.ptr == 0 {
		return nil
	}

	fn, err := h.registry.GetFunc("alpm_find_dbs_satisfier")
	if err != nil {
		return nil
	}

	var dbList *list.List
	for _, db := range dbs {
		dbImpl, ok := db.(*database)
		if ok {
			dbList = list.Add(dbList, dbImpl.ptr)
		}
	}
	defer dbList.Free()

	cStr := lib.CString(depstring)
	r1, _, _ := purego.SyscallN(fn, h.ptr, dbList.Ptr(), uintptr(unsafe.Pointer(&cStr[0])))
	runtime.KeepAlive(cStr)

	if r1 == 0 {
		return nil
	}

	return newPackage(r1, h)
}
