package alpm

import (
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/Jguer/dyalpm/internal/dyerrors"
	"github.com/Jguer/dyalpm/internal/lib"
	"github.com/Jguer/dyalpm/internal/list"
)

func (h *handle) SetLogFile(path string) error {
	return h.setOptionStr("alpm_option_set_logfile", path)
}

func (h *handle) LogFile() string {
	return h.getOptionStr("alpm_option_get_logfile")
}

func (h *handle) SetGPGDir(path string) error {
	return h.setOptionStr("alpm_option_set_gpgdir", path)
}

func (h *handle) GPGDir() string {
	return h.getOptionStr("alpm_option_get_gpgdir")
}

func (h *handle) SetUseSyslog(enable bool) error {
	return h.setOptionInt("alpm_option_set_usesyslog", lib.BoolToInt(enable))
}

func (h *handle) UseSyslog() bool {
	val := h.getOptionInt("alpm_option_get_usesyslog")
	return lib.IntToBool(val)
}

func (h *handle) SetCheckSpace(enable bool) error {
	return h.setOptionInt("alpm_option_set_checkspace", lib.BoolToInt(enable))
}

func (h *handle) CheckSpace() bool {
	val := h.getOptionInt("alpm_option_get_checkspace")
	return lib.IntToBool(val)
}

func (h *handle) SetDBExt(ext string) error {
	return h.setOptionStr("alpm_option_set_dbext", ext)
}

func (h *handle) DBExt() string {
	return h.getOptionStr("alpm_option_get_dbext")
}

func (h *handle) SetDefaultSigLevel(level int) error {
	return h.setOptionInt("alpm_option_set_default_siglevel", uintptr(level))
}

func (h *handle) DefaultSigLevel() int {
	return int(h.getOptionInt("alpm_option_get_default_siglevel"))
}

func (h *handle) SetLocalFileSigLevel(level int) error {
	return h.setOptionInt("alpm_option_set_local_file_siglevel", uintptr(level))
}

func (h *handle) LocalFileSigLevel() int {
	return int(h.getOptionInt("alpm_option_get_local_file_siglevel"))
}

func (h *handle) SetRemoteFileSigLevel(level int) error {
	return h.setOptionInt("alpm_option_set_remote_file_siglevel", uintptr(level))
}

func (h *handle) RemoteFileSigLevel() int {
	return int(h.getOptionInt("alpm_option_get_remote_file_siglevel"))
}

func (h *handle) SetParallelDownloads(num int) error {
	return h.setOptionInt("alpm_option_set_parallel_downloads", uintptr(num))
}

func (h *handle) ParallelDownloads() int {
	return int(h.getOptionInt("alpm_option_get_parallel_downloads"))
}

func (h *handle) CacheDirs() ([]string, error) {
	return h.getOptionStrList("alpm_option_get_cachedirs")
}

func (h *handle) SetCacheDirs(dirs []string) error {
	return h.setOptionStrList("alpm_option_set_cachedirs", dirs)
}

func (h *handle) AddCacheDir(dir string) error {
	return h.setOptionStr("alpm_option_add_cachedir", dir)
}

func (h *handle) RemoveCacheDir(dir string) error {
	return h.setOptionStr("alpm_option_remove_cachedir", dir)
}

func (h *handle) HookDirs() ([]string, error) {
	return h.getOptionStrList("alpm_option_get_hookdirs")
}

func (h *handle) SetHookDirs(dirs []string) error {
	return h.setOptionStrList("alpm_option_set_hookdirs", dirs)
}

func (h *handle) AddHookDir(dir string) error {
	return h.setOptionStr("alpm_option_add_hookdir", dir)
}

func (h *handle) RemoveHookDir(dir string) error {
	return h.setOptionStr("alpm_option_remove_hookdir", dir)
}

func (h *handle) NoUpgrades() ([]string, error) {
	return h.getOptionStrList("alpm_option_get_noupgrades")
}

func (h *handle) SetNoUpgrades(paths []string) error {
	return h.setOptionStrList("alpm_option_set_noupgrades", paths)
}

func (h *handle) AddNoUpgrade(path string) error {
	return h.setOptionStr("alpm_option_add_noupgrade", path)
}

func (h *handle) RemoveNoUpgrade(path string) error {
	return h.setOptionStr("alpm_option_remove_noupgrade", path)
}

func (h *handle) NoExtracts() ([]string, error) {
	return h.getOptionStrList("alpm_option_get_noextracts")
}

func (h *handle) SetNoExtracts(paths []string) error {
	return h.setOptionStrList("alpm_option_set_noextracts", paths)
}

func (h *handle) AddNoExtract(path string) error {
	return h.setOptionStr("alpm_option_add_noextract", path)
}

func (h *handle) RemoveNoExtract(path string) error {
	return h.setOptionStr("alpm_option_remove_noextract", path)
}

func (h *handle) IgnorePkgs() ([]string, error) {
	return h.getOptionStrList("alpm_option_get_ignorepkgs")
}

func (h *handle) SetIgnorePkgs(pkgs []string) error {
	return h.setOptionStrList("alpm_option_set_ignorepkgs", pkgs)
}

func (h *handle) AddIgnorePkg(pkg string) error {
	return h.setOptionStr("alpm_option_add_ignorepkg", pkg)
}

func (h *handle) RemoveIgnorePkg(pkg string) error {
	return h.setOptionStr("alpm_option_remove_ignorepkg", pkg)
}

func (h *handle) IgnoreGroups() ([]string, error) {
	return h.getOptionStrList("alpm_option_get_ignoregroups")
}

func (h *handle) SetIgnoreGroups(groups []string) error {
	return h.setOptionStrList("alpm_option_set_ignoregroups", groups)
}

func (h *handle) AddIgnoreGroup(group string) error {
	return h.setOptionStr("alpm_option_add_ignoregroup", group)
}

func (h *handle) RemoveIgnoreGroup(group string) error {
	return h.setOptionStr("alpm_option_remove_ignoregroup", group)
}

func (h *handle) OverwriteFiles() ([]string, error) {
	return h.getOptionStrList("alpm_option_get_overwrite_files")
}

func (h *handle) SetOverwriteFiles(globs []string) error {
	return h.setOptionStrList("alpm_option_set_overwrite_files", globs)
}

func (h *handle) AddOverwriteFile(glob string) error {
	return h.setOptionStr("alpm_option_add_overwrite_file", glob)
}

func (h *handle) RemoveOverwriteFile(glob string) error {
	return h.setOptionStr("alpm_option_remove_overwrite_file", glob)
}

func (h *handle) MatchNoUpgrade(path string) (int, error) {
	return h.matchOption("alpm_option_match_noupgrade", path)
}

func (h *handle) MatchNoExtract(path string) (int, error) {
	return h.matchOption("alpm_option_match_noextract", path)
}

func (h *handle) SetSandboxUser(user string) error {
	return h.setOptionStr("alpm_option_set_sandboxuser", user)
}

func (h *handle) SandboxUser() string {
	return h.getOptionStr("alpm_option_get_sandboxuser")
}

func (h *handle) SetDisableDLTimeout(disable bool) error {
	return h.setOptionInt("alpm_option_set_disable_dl_timeout", lib.BoolToInt(disable))
}

func (h *handle) SetDisableSandbox(disable bool) error {
	return h.setOptionInt("alpm_option_set_disable_sandbox", lib.BoolToInt(disable))
}

func (h *handle) AssumeInstalled() ([]Dependency, error) {
	if h.ptr == 0 {
		return nil, dyerrors.ErrHandleNull
	}

	fn, err := h.registry.GetFunc("alpm_option_get_assumeinstalled")
	if err != nil {
		return nil, err
	}

	r1, _, _ := purego.SyscallN(fn, h.ptr)
	if r1 == 0 {
		return []Dependency{}, nil
	}

	alpmList := list.NewList(r1)
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

func (h *handle) SetAssumeInstalled(deps []Dependency) error {
	return h.setOptionDepList("alpm_option_set_assumeinstalled", deps)
}

func (h *handle) AddAssumeInstalled(dep Dependency) error {
	if h.ptr == 0 {
		return dyerrors.ErrHandleNull
	}

	fn, err := h.registry.GetFunc("alpm_option_add_assumeinstalled")
	if err != nil {
		return err
	}

	d, ok := dep.(*dependency)
	if !ok {
		return ErrInvalidPackage // or something more appropriate
	}

	r1, _, _ := purego.SyscallN(fn, h.ptr, d.ptr)
	if r1 != 0 {
		return dyerrors.NewError(h.Errno(), "failed to add assume-installed dependency")
	}

	return nil
}

func (h *handle) RemoveAssumeInstalled(dep Dependency) error {
	if h.ptr == 0 {
		return dyerrors.ErrHandleNull
	}

	fn, err := h.registry.GetFunc("alpm_option_remove_assumeinstalled")
	if err != nil {
		return err
	}

	d, ok := dep.(*dependency)
	if !ok {
		return ErrInvalidPackage
	}

	r1, _, _ := purego.SyscallN(fn, h.ptr, d.ptr)
	if r1 != 0 {
		return dyerrors.NewError(h.Errno(), "failed to remove assume-installed dependency")
	}

	return nil
}

func (h *handle) Architectures() ([]string, error) {
	return h.getOptionStrList("alpm_option_get_architectures")
}

func (h *handle) SetArchitectures(archs []string) error {
	return h.setOptionStrList("alpm_option_set_architectures", archs)
}

func (h *handle) AddArchitecture(arch string) error {
	return h.setOptionStr("alpm_option_add_architecture", arch)
}

func (h *handle) RemoveArchitecture(arch string) error {
	return h.setOptionStr("alpm_option_remove_architecture", arch)
}

// Helper methods

func (h *handle) matchOption(funcName, path string) (int, error) {
	if h.ptr == 0 {
		return 0, dyerrors.ErrHandleNull
	}

	fn, err := h.registry.GetFunc(funcName)
	if err != nil {
		return 0, err
	}

	cStr := lib.CString(path)
	strPtr := uintptr(unsafe.Pointer(&cStr[0]))
	r1, _, _ := purego.SyscallN(fn, h.ptr, strPtr)
	runtime.KeepAlive(cStr)

	return int(r1), nil
}

func (h *handle) setOptionStr(funcName, value string) error {
	if h.ptr == 0 {
		return dyerrors.ErrHandleNull
	}

	fn, err := h.registry.GetFunc(funcName)
	if err != nil {
		return err
	}

	cStr := lib.CString(value)
	strPtr := uintptr(unsafe.Pointer(&cStr[0]))
	r1, _, _ := purego.SyscallN(fn, h.ptr, strPtr)
	runtime.KeepAlive(cStr)

	if r1 != 0 {
		return dyerrors.NewError(h.Errno(), "failed to set option")
	}
	return nil
}

func (h *handle) getOptionStr(funcName string) string {
	if h.ptr == 0 {
		return ""
	}

	fn, err := h.registry.GetFunc(funcName)
	if err != nil {
		return ""
	}

	r1, _, _ := purego.SyscallN(fn, h.ptr)
	return lib.PtrToString(r1)
}

func (h *handle) setOptionInt(funcName string, value uintptr) error {
	if h.ptr == 0 {
		return dyerrors.ErrHandleNull
	}

	fn, err := h.registry.GetFunc(funcName)
	if err != nil {
		return err
	}

	r1, _, _ := purego.SyscallN(fn, h.ptr, value)
	if r1 != 0 {
		return dyerrors.NewError(h.Errno(), "failed to set option")
	}
	return nil
}

func (h *handle) getOptionInt(funcName string) uintptr {
	if h.ptr == 0 {
		return 0
	}

	fn, err := h.registry.GetFunc(funcName)
	if err != nil {
		return 0
	}

	r1, _, _ := purego.SyscallN(fn, h.ptr)
	return r1
}

func (h *handle) getOptionStrList(funcName string) ([]string, error) {
	if h.ptr == 0 {
		return nil, dyerrors.ErrHandleNull
	}

	fn, err := h.registry.GetFunc(funcName)
	if err != nil {
		return nil, err
	}

	r1, _, _ := purego.SyscallN(fn, h.ptr)
	if r1 == 0 {
		return []string{}, nil
	}

	alpmList := list.NewList(r1)
	if alpmList == nil {
		return []string{}, nil
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

func (h *handle) setOptionStrList(funcName string, values []string) error {
	if h.ptr == 0 {
		return dyerrors.ErrHandleNull
	}

	fn, err := h.registry.GetFunc(funcName)
	if err != nil {
		return err
	}

	var alpmList *list.List
	var cStrings [][]byte

	for _, v := range values {
		cS := lib.CString(v)
		cStrings = append(cStrings, cS)
		alpmList = list.Add(alpmList, uintptr(unsafe.Pointer(&cS[0])))
	}

	var listPtr uintptr
	if alpmList != nil {
		listPtr = alpmList.Ptr()
	}

	r1, _, _ := purego.SyscallN(fn, h.ptr, listPtr)

	runtime.KeepAlive(cStrings)
	runtime.KeepAlive(alpmList)

	if r1 != 0 {
		return dyerrors.NewError(h.Errno(), "failed to set option list")
	}
	return nil
}

func (h *handle) setOptionDepList(funcName string, deps []Dependency) error {
	if h.ptr == 0 {
		return dyerrors.ErrHandleNull
	}

	fn, err := h.registry.GetFunc(funcName)
	if err != nil {
		return err
	}

	var alpmList *list.List
	for _, d := range deps {
		depImpl, ok := d.(*dependency)
		if ok {
			alpmList = list.Add(alpmList, depImpl.ptr)
		}
	}

	var listPtr uintptr
	if alpmList != nil {
		listPtr = alpmList.Ptr()
	}

	r1, _, _ := purego.SyscallN(fn, h.ptr, listPtr)
	runtime.KeepAlive(alpmList)

	if r1 != 0 {
		return dyerrors.NewError(h.Errno(), "failed to set option dependency list")
	}
	return nil
}
