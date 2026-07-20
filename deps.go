package dyalpm

import (
	stderrors "errors"
	"unsafe"

	"github.com/Jguer/dyalpm/internal/dyerrors"
	"github.com/Jguer/dyalpm/internal/lib"
	alpmlist "github.com/Jguer/dyalpm/internal/list"
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

type DependencyMetadata interface {
	GetDescription() string
	GetNameHash() uint64
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
	ptr   uintptr
	owned bool
}

// unsigned long is pointer-sized in the Linux alpm_depend_t ABI.
type alpmDepend struct {
	Name        uintptr
	Version     uintptr
	Description uintptr
	NameHash    uintptr
	Mod         int32
}

func newDependency(ptr uintptr) *dependency {
	return &dependency{
		ptr:   ptr,
		owned: false,
	}
}

func (d *dependency) data() *alpmDepend {
	if d == nil || d.ptr == 0 {
		return nil
	}
	return (*alpmDepend)(unsafe.Pointer(d.ptr))
}

func (d *dependency) GetName() string {
	data := d.data()
	if data == nil {
		return ""
	}
	return lib.PtrToString(data.Name)
}

func (d *dependency) GetVersion() string {
	data := d.data()
	if data == nil {
		return ""
	}
	return lib.PtrToString(data.Version)
}

func (d *dependency) GetDescription() string {
	data := d.data()
	if data == nil {
		return ""
	}
	return lib.PtrToString(data.Description)
}

func (d *dependency) GetNameHash() uint64 {
	data := d.data()
	if data == nil {
		return 0
	}
	return uint64(data.NameHash)
}

func (d *dependency) GetMod() DepMod {
	data := d.data()
	if data == nil {
		return DepModAny
	}
	return DepMod(data.Mod)
}

func (d *dependency) ComputeString() string {
	if d.ptr == 0 {
		return ""
	}
	if lib.AlpmDepComputeString == nil {
		return ""
	}

	r1 := lib.AlpmDepComputeString(d.ptr)
	if r1 == 0 {
		return ""
	}

	defer lib.Free(r1)
	return lib.PtrToString(r1)
}

func (d *dependency) Free() {
	if d.ptr == 0 || !d.owned {
		return
	}
	if lib.AlpmDepFree == nil {
		return
	}
	lib.AlpmDepFree(d.ptr)
	d.ptr = 0
}

// DepFromString creates a dependency from a string
func DepFromString(depstring string) (Dependency, error) {
	if err := lib.EnsureAlpmLoaded(); err != nil {
		return nil, err
	}
	if lib.AlpmDepFromString == nil {
		return nil, stderrors.New("missing function: alpm_dep_from_string")
	}

	r1 := lib.AlpmDepFromString(depstring)

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
	ptr uintptr
}

type alpmDepMissing struct {
	Target     uintptr
	Depend     uintptr
	CausingPkg uintptr
}

func newDepMissing(ptr uintptr) *depMissing {
	return &depMissing{
		ptr: ptr,
	}
}

func (d *depMissing) data() *alpmDepMissing {
	if d == nil || d.ptr == 0 {
		return nil
	}
	return (*alpmDepMissing)(unsafe.Pointer(d.ptr))
}

func (d *depMissing) GetTarget() string {
	data := d.data()
	if data == nil {
		return ""
	}
	return lib.PtrToString(data.Target)
}

func (d *depMissing) GetDepend() Dependency {
	data := d.data()
	if data == nil || data.Depend == 0 {
		return nil
	}
	return newDependency(data.Depend)
}

func (d *depMissing) GetCausingPkg() string {
	data := d.data()
	if data == nil {
		return ""
	}
	return lib.PtrToString(data.CausingPkg)
}

func (d *depMissing) Free() {
	if d.ptr == 0 {
		return
	}
	if lib.AlpmDepmissingFree == nil {
		return
	}
	lib.AlpmDepmissingFree(d.ptr)
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
	ptr uintptr
}

type alpmConflict struct {
	Package1 uintptr
	Package2 uintptr
	Reason   uintptr
}

func newConflict(ptr uintptr) *conflict {
	return &conflict{
		ptr: ptr,
	}
}

func (c *conflict) data() *alpmConflict {
	if c == nil || c.ptr == 0 {
		return nil
	}
	return (*alpmConflict)(unsafe.Pointer(c.ptr))
}

func conflictPackageName(ptr uintptr) string {
	if ptr == 0 || lib.AlpmPkgGetName == nil {
		return ""
	}
	return lib.PtrToString(lib.AlpmPkgGetName(ptr))
}

func (c *conflict) GetPackage1() string {
	data := c.data()
	if data == nil {
		return ""
	}
	return conflictPackageName(data.Package1)
}

func (c *conflict) GetPackage2() string {
	data := c.data()
	if data == nil {
		return ""
	}
	return conflictPackageName(data.Package2)
}

func (c *conflict) GetReason() Dependency {
	data := c.data()
	if data == nil || data.Reason == 0 {
		return nil
	}
	return newDependency(data.Reason)
}

func toDependList(deps []Dependency) []Depend {
	result := make([]Depend, len(deps))
	for i, dep := range deps {
		result[i] = Depend{
			Name:    dep.GetName(),
			Version: dep.GetVersion(),
			Mod:     dep.GetMod(),
		}
		if metadata, ok := dep.(DependencyMetadata); ok {
			result[i].Description = metadata.GetDescription()
			result[i].NameHash = metadata.GetNameHash()
		}
	}
	return result
}

func (c *conflict) Free() {
	if c.ptr == 0 {
		return
	}
	if lib.AlpmConflictFree == nil {
		return
	}
	lib.AlpmConflictFree(c.ptr)
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
	ptr uintptr
}

type alpmFileConflict struct {
	Target  uintptr
	Type    int32
	File    uintptr
	CTarget uintptr
}

func newFileConflict(ptr uintptr) *fileConflict {
	return &fileConflict{
		ptr: ptr,
	}
}

func (f *fileConflict) data() *alpmFileConflict {
	if f == nil || f.ptr == 0 {
		return nil
	}
	return (*alpmFileConflict)(unsafe.Pointer(f.ptr))
}

func (f *fileConflict) GetTarget() string {
	data := f.data()
	if data == nil {
		return ""
	}
	return lib.PtrToString(data.Target)
}

func (f *fileConflict) GetType() FileConflictType {
	data := f.data()
	if data == nil {
		return 0
	}
	return FileConflictType(data.Type)
}

func (f *fileConflict) GetFile() string {
	data := f.data()
	if data == nil {
		return ""
	}
	return lib.PtrToString(data.File)
}

func (f *fileConflict) GetCTarget() string {
	data := f.data()
	if data == nil {
		return ""
	}
	return lib.PtrToString(data.CTarget)
}

func (f *fileConflict) Free() {
	if f.ptr == 0 {
		return
	}
	if lib.AlpmFileConflictFree == nil {
		return
	}
	lib.AlpmFileConflictFree(f.ptr)
	f.ptr = 0
}

// Resolution functions

func (h *handle) CheckDeps(pkgs []Package, remPkgs []Package, upgradePkgs []Package, reverseDeps bool) ([]DepMissing, error) {
	if h.ptr == 0 {
		return nil, dyerrors.ErrHandleNull
	}
	if lib.AlpmCheckDeps == nil {
		return nil, stderrors.New("missing function: alpm_checkdeps")
	}

	var pkgList, remPkgList, upgradePkgList *alpmlist.List
	for _, p := range pkgs {
		pkgImpl, _ := p.(*package_)
		pkgList = alpmlist.Add(pkgList, pkgImpl.ptr)
	}
	defer pkgList.Free()

	for _, p := range remPkgs {
		pkgImpl, _ := p.(*package_)
		remPkgList = alpmlist.Add(remPkgList, pkgImpl.ptr)
	}
	defer remPkgList.Free()

	for _, p := range upgradePkgs {
		pkgImpl, _ := p.(*package_)
		upgradePkgList = alpmlist.Add(upgradePkgList, pkgImpl.ptr)
	}
	defer upgradePkgList.Free()

	rev := int32(0)
	if reverseDeps {
		rev = 1
	}
	r1 := lib.AlpmCheckDeps(h.ptr, pkgList.Ptr(), remPkgList.Ptr(), upgradePkgList.Ptr(), rev)
	if r1 == 0 {
		return []DepMissing{}, nil
	}

	resList := alpmlist.NewList(r1)
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
	if lib.AlpmCheckConflicts == nil {
		return nil, stderrors.New("missing function: alpm_checkconflicts")
	}

	var pkgList *alpmlist.List
	for _, p := range pkgs {
		pkgImpl, _ := p.(*package_)
		pkgList = alpmlist.Add(pkgList, pkgImpl.ptr)
	}
	defer pkgList.Free()

	r1 := lib.AlpmCheckConflicts(h.ptr, pkgList.Ptr())
	if r1 == 0 {
		return []Conflict{}, nil
	}

	resList := alpmlist.NewList(r1)
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
	if len(pkgs) == 0 {
		return nil
	}
	if err := lib.EnsureAlpmLoaded(); err != nil {
		return nil
	}
	if lib.AlpmFindSatisfier == nil {
		return nil
	}

	var pkgList *alpmlist.List
	var h *handle
	for _, p := range pkgs {
		pkgImpl, ok := p.(*package_)
		if ok {
			pkgList = alpmlist.Add(pkgList, pkgImpl.ptr)
			if h == nil {
				h = pkgImpl.handle
			}
		}
	}
	defer pkgList.Free()

	r1 := lib.AlpmFindSatisfier(pkgList.Ptr(), depstring)

	if r1 == 0 {
		return nil
	}

	return newPackage(r1, h)
}

func (h *handle) FindDBSatisfier(dbs []Database, depstring string) Package {
	if h.ptr == 0 {
		return nil
	}
	if lib.AlpmFindDBSatisfier == nil {
		return nil
	}

	var dbList *alpmlist.List
	for _, db := range dbs {
		dbImpl, ok := db.(*database)
		if ok {
			dbList = alpmlist.Add(dbList, dbImpl.ptr)
		}
	}
	defer dbList.Free()

	r1 := lib.AlpmFindDBSatisfier(h.ptr, dbList.Ptr(), depstring)

	if r1 == 0 {
		return nil
	}

	return newPackage(r1, h)
}
