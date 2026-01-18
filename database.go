package alpm

import (
	stderrors "errors"
	"runtime"
	"unsafe"

	"github.com/Jguer/dyalpm/internal/lib"
	"github.com/Jguer/dyalpm/internal/list"
)

// Usage type for database usage flags
type Usage int

const (
	UsageSync    Usage = 1 << 0
	UsageSearch  Usage = 1 << 1
	UsageInstall Usage = 1 << 2
	UsageUpgrade Usage = 1 << 3
	UsageAll     Usage = (1 << 4) - 1
)

// Database represents an ALPM database (minimal surface used by yay).
type Database interface {
	Name() string
	Pkg(name string) Package
	PkgCache() PackageIterator
	Search(needles []string) PackageIterator
	SetServers(servers []string) error
	SetUsage(usage int) error
}

type database struct {
	ptr      uintptr
	handle   *handle
	registry *lib.FunctionRegistry
}

func newDatabase(ptr uintptr, h *handle) *database {
	reg, _ := lib.GetRegistry()
	return &database{
		ptr:      ptr,
		handle:   h,
		registry: reg,
	}
}

func (d *database) getList(funcName string) (*list.List, error) {
	fn, err := d.registry.GetFunc(funcName)
	if err != nil {
		return nil, err
	}

	listPtr := lib.Syscall(fn, d.ptr)
	if listPtr == 0 {
		return nil, nil
	}

	return list.NewList(listPtr), nil
}

func (d *database) getServers(funcName string) []string {
	if d.ptr == 0 {
		return nil
	}

	listPtr, err := d.getList(funcName)
	if err != nil || listPtr == nil {
		return nil
	}

	return collectList(listPtr, func(ptr uintptr) string {
		return lib.PtrToString(ptr)
	})
}

func (d *database) GetName() string {
	if d.ptr == 0 {
		return ""
	}

	getNameFn := cachedFunc("alpm_db_get_name")
	if getNameFn == 0 {
		return ""
	}

	result := lib.Syscall(getNameFn, d.ptr)
	return lib.PtrToString(result)
}

func (d *database) GetHandle() Handle {
	return d.handle
}

func (d *database) GetPkg(name string) (Package, error) {
	if d.ptr == 0 {
		return nil, ErrInvalidDatabase
	}

	getPkgFn := cachedFunc("alpm_db_get_pkg")
	if getPkgFn == 0 {
		return nil, stderrors.New("missing function: alpm_db_get_pkg")
	}

	nameBytes := lib.CString(name)
	namePtr := uintptr(unsafe.Pointer(&nameBytes[0]))
	pkgPtr := lib.Syscall(getPkgFn, d.ptr, namePtr)
	runtime.KeepAlive(nameBytes)
	if pkgPtr == 0 {
		return nil, ErrPackageNotFound
	}

	return newPackage(pkgPtr, d.handle), nil
}

func (d *database) GetPkgCache() ([]Package, error) {
	if d.ptr == 0 {
		return nil, ErrInvalidDatabase
	}

	alpmList, err := d.getList("alpm_db_get_pkgcache")
	if err != nil || alpmList == nil {
		return []Package{}, err
	}

	pkgs := collectList(alpmList, func(ptr uintptr) Package {
		return newPackage(ptr, d.handle)
	})

	return pkgs, nil
}

// go-alpm/v2 compatible methods
func (d *database) Name() string { return d.GetName() }

func (d *database) Pkgs() []Package {
	pkgs, _ := d.GetPkgCache()
	return pkgs
}

func (d *database) Pkg(name string) Package {
	pkg, err := d.GetPkg(name)
	if err != nil {
		return nil
	}
	return pkg
}

func (d *database) PkgCache() PackageIterator {
	if d.ptr == 0 {
		return PackageIterator{}
	}

	getPkgCacheFn := cachedFunc("alpm_db_get_pkgcache")
	if getPkgCacheFn == 0 {
		return PackageIterator{}
	}

	listPtr := lib.Syscall(getPkgCacheFn, d.ptr)
	if listPtr == 0 {
		return PackageIterator{}
	}

	return newPackageIterator(listPtr, d.handle, false)
}

func (d *database) Search(needles []string) PackageIterator {
	if d.ptr == 0 {
		return PackageIterator{}
	}

	if len(needles) == 0 {
		return PackageIterator{}
	}

	searchFn := cachedFunc("alpm_db_search")
	if searchFn == 0 {
		return PackageIterator{}
	}

	var alpmList *list.List
	var cStrings [][]byte

	for _, s := range needles {
		cs := lib.CString(s)
		cStrings = append(cStrings, cs)
		alpmList = list.Add(alpmList, uintptr(unsafe.Pointer(&cs[0])))
	}

	if alpmList == nil {
		return PackageIterator{}
	}

	// alpm_db_search signature: int alpm_db_search(db, needles, &result)
	// Returns error code and writes result list to output parameter
	var resultListPtr uintptr
	ret := lib.Syscall(searchFn, d.ptr, alpmList.Ptr(), uintptr(unsafe.Pointer(&resultListPtr)))
	runtime.KeepAlive(cStrings)
	runtime.KeepAlive(alpmList)

	if ret != 0 || resultListPtr == 0 {
		return PackageIterator{}
	}

	// alpm_db_search returns a list that must be freed by the caller.
	return newPackageIterator(resultListPtr, d.handle, true)
}

func (d *database) GetGroup(name string) (Group, error) {
	if d.ptr == 0 {
		return nil, ErrInvalidDatabase
	}

	getGroupFn, err := d.registry.GetFunc("alpm_db_get_group")
	if err != nil {
		return nil, err
	}

	nameBytes := lib.CString(name)
	namePtr := uintptr(unsafe.Pointer(&nameBytes[0]))
	groupPtr := lib.Syscall(getGroupFn, d.ptr, namePtr)
	runtime.KeepAlive(nameBytes)
	if groupPtr == 0 {
		return nil, ErrGroupNotFound
	}

	return newGroup(groupPtr, d.handle), nil
}

func (d *database) GetGroupCache() ([]Group, error) {
	if d.ptr == 0 {
		return nil, ErrInvalidDatabase
	}

	alpmList, err := d.getList("alpm_db_get_groupcache")
	if err != nil || alpmList == nil {
		return []Group{}, err
	}

	groups := collectList(alpmList, func(ptr uintptr) Group {
		return newGroup(ptr, d.handle)
	})

	return groups, nil
}

func (d *database) Update(force bool) error {
	if d.ptr == 0 {
		return ErrInvalidDatabase
	}

	updateFn, err := d.registry.GetFunc("alpm_db_update")
	if err != nil {
		return err
	}

	// alpm_db_update expects a list of databases
	dbList := list.Add(nil, d.ptr)
	if dbList == nil {
		return ErrDatabaseUpdateFailed
	}
	defer dbList.Free()

	forceInt := lib.BoolToInt(force)
	result := lib.Syscall(updateFn, d.handle.ptr, dbList.Ptr(), forceInt)
	if result != 0 {
		return ErrDatabaseUpdateFailed
	}

	return nil
}

func (d *database) Unregister() error {
	if d.ptr == 0 {
		return ErrInvalidDatabase
	}

	unregisterFn, err := d.registry.GetFunc("alpm_db_unregister")
	if err != nil {
		return err
	}

	result := lib.Syscall(unregisterFn, d.ptr)
	if result != 0 {
		return ErrDatabaseUnregisterFailed
	}

	d.ptr = 0
	return nil
}

func (d *database) GetServers() []string {
	return d.getServers("alpm_db_get_servers")
}

func (d *database) SetServers(servers []string) error {
	return d.setServers("alpm_db_set_servers", servers)
}

func (d *database) setServers(funcName string, servers []string) error {
	if d.ptr == 0 {
		return ErrInvalidDatabase
	}

	setServersFn, err := d.registry.GetFunc(funcName)
	if err != nil {
		return err
	}

	var alpmList *list.List
	var cStrings [][]byte

	for _, s := range servers {
		cs := lib.CString(s)
		cStrings = append(cStrings, cs)
		alpmList = list.Add(alpmList, uintptr(unsafe.Pointer(&cs[0])))
	}

	result := lib.Syscall(setServersFn, d.ptr, alpmList.Ptr())

	// Keep strings alive during the call
	runtime.KeepAlive(cStrings)
	runtime.KeepAlive(alpmList)

	if result != 0 {
		return ErrDatabaseUpdateFailed
	}

	return nil
}

func (d *database) AddServer(url string) error {
	if d.ptr == 0 {
		return ErrInvalidDatabase
	}

	addServerFn, err := d.registry.GetFunc("alpm_db_add_server")
	if err != nil {
		return err
	}

	urlBytes := lib.CString(url)
	result := lib.Syscall(addServerFn, d.ptr, uintptr(unsafe.Pointer(&urlBytes[0])))
	runtime.KeepAlive(urlBytes)

	if result != 0 {
		return ErrDatabaseUpdateFailed
	}

	return nil
}

func (d *database) RemoveServer(url string) error {
	if d.ptr == 0 {
		return ErrInvalidDatabase
	}

	removeServerFn, err := d.registry.GetFunc("alpm_db_remove_server")
	if err != nil {
		return err
	}

	urlBytes := lib.CString(url)
	result := lib.Syscall(removeServerFn, d.ptr, uintptr(unsafe.Pointer(&urlBytes[0])))
	runtime.KeepAlive(urlBytes)

	if result != 0 {
		return ErrDatabaseUpdateFailed
	}

	return nil
}

func (d *database) GetCacheServers() []string {
	return d.getServers("alpm_db_get_cache_servers")
}

func (d *database) SetCacheServers(servers []string) error {
	return d.setServers("alpm_db_set_cache_servers", servers)
}

func (d *database) AddCacheServer(url string) error {
	if d.ptr == 0 {
		return ErrInvalidDatabase
	}

	addCacheServerFn, err := d.registry.GetFunc("alpm_db_add_cache_server")
	if err != nil {
		return err
	}

	urlBytes := lib.CString(url)
	result := lib.Syscall(addCacheServerFn, d.ptr, uintptr(unsafe.Pointer(&urlBytes[0])))
	runtime.KeepAlive(urlBytes)

	if result != 0 {
		return ErrDatabaseUpdateFailed
	}

	return nil
}

func (d *database) RemoveCacheServer(url string) error {
	if d.ptr == 0 {
		return ErrInvalidDatabase
	}

	removeCacheServerFn, err := d.registry.GetFunc("alpm_db_remove_cache_server")
	if err != nil {
		return err
	}

	urlBytes := lib.CString(url)
	result := lib.Syscall(removeCacheServerFn, d.ptr, uintptr(unsafe.Pointer(&urlBytes[0])))
	runtime.KeepAlive(urlBytes)

	if result != 0 {
		return ErrDatabaseUpdateFailed
	}

	return nil
}

func (d *database) SetUsage(usage int) error {
	if d.ptr == 0 {
		return ErrInvalidDatabase
	}

	setUsageFn, err := d.registry.GetFunc("alpm_db_set_usage")
	if err != nil {
		return err
	}

	result := lib.Syscall(setUsageFn, d.ptr, uintptr(usage))
	if result != 0 {
		return ErrDatabaseUpdateFailed
	}

	return nil
}

func (d *database) GetUsage() (int, error) {
	if d.ptr == 0 {
		return 0, ErrInvalidDatabase
	}

	getUsageFn, err := d.registry.GetFunc("alpm_db_get_usage")
	if err != nil {
		return 0, err
	}

	var usage int
	result := lib.Syscall(getUsageFn, d.ptr, uintptr(unsafe.Pointer(&usage)))
	if result != 0 {
		return 0, ErrDatabaseUpdateFailed
	}

	return usage, nil
}

func (d *database) IsValid() bool {
	if d.ptr == 0 {
		return false
	}

	getValidFn, err := d.registry.GetFunc("alpm_db_get_valid")
	if err != nil {
		return false
	}

	result := lib.Syscall(getValidFn, d.ptr)
	return result == 0
}

func (d *database) GetSigLevel() int {
	if d.ptr == 0 {
		return 0
	}

	getSigLevelFn, err := d.registry.GetFunc("alpm_db_get_siglevel")
	if err != nil {
		return 0
	}

	result := lib.Syscall(getSigLevelFn, d.ptr)
	return int(result)
}

func (d *database) GetNativeHandle() Handle {
	if d.ptr == 0 {
		return nil
	}

	fn, err := d.registry.GetFunc("alpm_db_get_handle")
	if err != nil {
		return nil
	}

	result := lib.Syscall(fn, d.ptr)
	if result == 0 {
		return nil
	}

	// This should match d.handle.ptr if everything is correct
	return d.handle
}

func (d *database) CheckPGPSignature() (SigList, error) {
	if d.ptr == 0 {
		return SigList{}, ErrInvalidDatabase
	}

	return checkPGPSignature(d.ptr, d.registry, d.handle, "alpm_db_check_pgp_signature")
}

// Group represents a package group
type Group interface {
	GetName() string
	GetPackages() ([]Package, error)
}

type group struct {
	ptr      uintptr
	handle   *handle
	registry *lib.FunctionRegistry
}

func newGroup(ptr uintptr, h *handle) *group {
	reg, _ := lib.GetRegistry()
	return &group{
		ptr:      ptr,
		handle:   h,
		registry: reg,
	}
}

func (g *group) GetName() string {
	if g.ptr == 0 {
		return ""
	}
	// Group structure: char *name at offset 0
	base := unsafe.Pointer(g.ptr)
	namePtr := *(*uintptr)(base)
	return lib.PtrToString(namePtr)
}

func (g *group) GetPackages() ([]Package, error) {
	if g.ptr == 0 {
		return nil, ErrInvalidGroup
	}
	// Group structure: alpm_list_t *packages at offset of pointer size
	base := unsafe.Pointer(g.ptr)
	packagesPtr := *(*uintptr)(unsafe.Add(base, unsafe.Sizeof(uintptr(0))))
	if packagesPtr == 0 {
		return []Package{}, nil
	}

	alpmList := list.NewList(packagesPtr)
	if alpmList == nil {
		return []Package{}, nil
	}

	var pkgs []Package
	for item := alpmList; item != nil && item.Ptr() != 0; item = item.Next() {
		pkgPtr := item.Data()
		if pkgPtr != 0 {
			pkgs = append(pkgs, newPackage(pkgPtr, g.handle))
		}
	}

	return pkgs, nil
}
