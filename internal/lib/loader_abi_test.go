package lib

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestAlpmFuncSignatures(t *testing.T) {
	var _ func() int32 = AlpmCapabilities
	var _ func(uintptr) uintptr = AlpmListCount
	var _ func(uintptr) int32 = AlpmSiglistCleanup
	var _ func(uintptr, string, uintptr, uintptr, *uintptr) int32 = AlpmExtractKeyID
	var _ func(uintptr) int64 = AlpmPkgGetBuildDate
	var _ func(uintptr) int64 = AlpmPkgGetInstallDate
	var _ func(uintptr, string, string, bool) int32 = AlpmSandboxSetupChild
	var _ func(uintptr, string) uintptr = AlpmPkgGetFilesContains
	var _ func(uintptr, uintptr) int32 = AlpmPkgChangelogClose
	var _ func(uintptr) = AlpmDepFree
	var _ func(uintptr) = AlpmDepmissingFree
	var _ func(uintptr) = AlpmConflictFree
	var _ func(uintptr) = AlpmFileConflictFree
}

func TestAlpmFuncResolution(t *testing.T) {
	if err := EnsureAlpmLoaded(); err != nil {
		t.Skipf("libalpm not available: %v", err)
	}
	if AlpmVersion == nil {
		t.Error("expected alpm_version to resolve")
	}
	VersionPtr := AlpmVersion()
	if VersionPtr == 0 {
		t.Error("expected alpm_version to return a non-empty pointer")
	} else {
		version := PtrToString(VersionPtr)
		if version == "" {
			t.Error("expected alpm_version to return a non-empty version string")
		}
	}
	if AlpmRelease == nil {
		t.Error("expected alpm_release to resolve")
	}
	if AlpmPkgGetName == nil {
		t.Error("expected alpm_pkg_get_name to resolve")
	}
	if AlpmInitialize == nil {
		t.Error("expected alpm_initialize to resolve")
	}
	if AlpmGetLocalDB == nil {
		t.Error("expected alpm_get_localdb to resolve")
	}
	if AlpmGetSyncDBS == nil {
		t.Error("expected alpm_get_syncdbs to resolve")
	}
	if AlpmOptionGetRoot == nil {
		t.Error("expected alpm_option_get_root to resolve")
	}
	if AlpmOptionSetLogfile == nil {
		t.Error("expected alpm_option_set_logfile to resolve")
	}
	if AlpmOptionGetLogfile == nil {
		t.Error("expected alpm_option_get_logfile to resolve")
	}
	if AlpmOptionSetNoUpgrades == nil {
		t.Error("expected alpm_option_set_noupgrades to resolve")
	}
	if AlpmOptionMatchNoUpgrade == nil {
		t.Error("expected alpm_option_match_noupgrade to resolve")
	}
	if AlpmPkgLoad == nil {
		t.Error("expected alpm_pkg_load to resolve")
	}
	if AlpmDBCheckPGPSignature == nil {
		t.Error("expected alpm_db_check_pgp_signature to resolve")
	}
	if AlpmPkgCheckPGPSignature == nil {
		t.Error("expected alpm_pkg_check_pgp_signature to resolve")
	}
	if AlpmPkgGetCheckdepends == nil {
		t.Error("expected alpm_pkg_get_checkdepends to resolve")
	}
	if AlpmPkgGetMakedepends == nil {
		t.Error("expected alpm_pkg_get_makedepends to resolve")
	}
	if LibcFree == nil {
		t.Error("expected libc free to resolve")
	}
	if LibcVsnprintf == nil {
		t.Error("expected libc vsnprintf to resolve")
	}
	if AlpmCapabilities == nil {
		t.Error("expected alpm_capabilities to resolve")
	} else {
		caps := AlpmCapabilities()
		if caps == 0 {
			t.Logf("alpm_capabilities returned zero bitmask")
		}
	}
}

func TestAlpmListCountABI(t *testing.T) {
	if err := EnsureAlpmLoaded(); err != nil {
		t.Skipf("libalpm not available: %v", err)
	}
	if AlpmListAdd == nil || AlpmListCount == nil || AlpmListFree == nil {
		t.Fatal("required alpm list functions did not resolve")
	}

	var list uintptr
	for _, value := range []uintptr{1, 2, 3, 4} {
		list = AlpmListAdd(list, value)
		if list == 0 {
			t.Fatal("alpm_list_add returned nil")
		}
	}
	defer AlpmListFree(list)

	if got := AlpmListCount(list); got != 4 {
		t.Fatalf("alpm_list_count returned %d, want 4", got)
	}
}

func TestAlpmSiglistCleanupABI(t *testing.T) {
	if err := EnsureAlpmLoaded(); err != nil {
		t.Skipf("libalpm not available: %v", err)
	}
	if AlpmSiglistCleanup == nil {
		t.Fatal("alpm_siglist_cleanup did not resolve")
	}

	siglist := [2]uintptr{}
	if result := AlpmSiglistCleanup(uintptr(unsafe.Pointer(&siglist[0]))); result != 0 {
		t.Fatalf("alpm_siglist_cleanup returned %d", result)
	}
	runtime.KeepAlive(siglist)
}

func TestAlpmVoidFreeABI(t *testing.T) {
	if err := EnsureAlpmLoaded(); err != nil {
		t.Skipf("libalpm not available: %v", err)
	}
	if AlpmDepFromString == nil || AlpmDepFree == nil {
		t.Fatal("dependency functions did not resolve")
	}

	dep := AlpmDepFromString("bash>=5")
	if dep == 0 {
		t.Fatal("alpm_dep_from_string returned nil")
	}
	AlpmDepFree(dep)
}
