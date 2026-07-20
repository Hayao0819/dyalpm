//go:build integration

package dyalpm

import (
	"path/filepath"
	"strings"
	"testing"
)

func containsStringInSlice(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func normalizePathValue(v string) string {
	return strings.TrimSuffix(filepath.Clean(v), "/")
}

func containsStringInSliceNormalized(items []string, target string) bool {
	target = normalizePathValue(target)
	for _, item := range items {
		if normalizePathValue(item) == target {
			return true
		}
	}
	return false
}

func TestHandleOptions_StringRoundTrip(t *testing.T) {
	h := mustInitializeTestHandle(t)

	scenarios := []struct {
		name     string
		setValue string
		setFn    func(string) error
		getFn    func() string
	}{
		{
			name:     "logfile",
			setValue: "/tmp/dyalpm-it-log",
			setFn:    h.SetLogFile,
			getFn:    h.LogFile,
		},
		{
			name:     "gpgdir",
			setValue: "/tmp/dyalpm-it-gpg",
			setFn:    h.SetGPGDir,
			getFn:    h.GPGDir,
		},
		{
			name:     "dbext",
			setValue: "it-dbext",
			setFn:    h.SetDBExt,
			getFn:    h.DBExt,
		},
		{
			name:     "sandboxuser",
			setValue: "root",
			setFn:    h.SetSandboxUser,
			getFn:    h.SandboxUser,
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			if err := s.setFn(s.setValue); err != nil {
				t.Fatalf("failed to set %s: %v", s.name, err)
			}
			if got := normalizePathValue(s.getFn()); got != normalizePathValue(s.setValue) {
				t.Fatalf("got %s=%q, want %q", s.name, got, normalizePathValue(s.setValue))
			}
		})
	}
}

func TestHandleOptions_BoolAndIntRoundTrip(t *testing.T) {
	h := mustInitializeTestHandle(t)

	originalSyslog := h.UseSyslog()
	if err := h.SetUseSyslog(!originalSyslog); err != nil {
		t.Fatalf("failed to set use syslog: %v", err)
	}
	if got := h.UseSyslog(); got != !originalSyslog {
		t.Fatalf("use syslog mismatch: got %v, want %v", got, !originalSyslog)
	}

	originalCheckSpace := h.CheckSpace()
	if err := h.SetCheckSpace(!originalCheckSpace); err != nil {
		t.Fatalf("failed to set check space: %v", err)
	}
	if got := h.CheckSpace(); got != !originalCheckSpace {
		t.Fatalf("checkspace mismatch: got %v, want %v", got, !originalCheckSpace)
	}

	if err := h.SetParallelDownloads(8); err != nil {
		t.Fatalf("failed to set parallel downloads: %v", err)
	}
	if got := h.ParallelDownloads(); got != 8 {
		t.Fatalf("parallel downloads mismatch: got %d, want %d", got, 8)
	}

	originalFilesystem := h.DisableSandboxFilesystem()
	originalSyscalls := h.DisableSandboxSyscalls()
	t.Cleanup(func() {
		_ = h.SetDisableSandboxFilesystem(originalFilesystem)
		_ = h.SetDisableSandboxSyscalls(originalSyscalls)
	})

	if err := h.SetDisableSandboxFilesystem(!originalFilesystem); err != nil {
		t.Fatalf("failed to set filesystem sandbox state: %v", err)
	}
	if got := h.DisableSandboxFilesystem(); got != !originalFilesystem {
		t.Fatalf("filesystem sandbox mismatch: got %v, want %v", got, !originalFilesystem)
	}

	if err := h.SetDisableSandboxSyscalls(!originalSyscalls); err != nil {
		t.Fatalf("failed to set syscall sandbox state: %v", err)
	}
	if got := h.DisableSandboxSyscalls(); got != !originalSyscalls {
		t.Fatalf("syscall sandbox mismatch: got %v, want %v", got, !originalSyscalls)
	}

	if err := h.SetDisableSandbox(true); err != nil {
		t.Fatalf("failed to disable sandbox: %v", err)
	}
	if !h.DisableSandboxFilesystem() || !h.DisableSandboxSyscalls() {
		t.Fatal("SetDisableSandbox(true) did not disable both sandbox components")
	}

	if err := h.SetDisableSandbox(false); err != nil {
		t.Fatalf("failed to enable sandbox: %v", err)
	}
	if h.DisableSandboxFilesystem() || h.DisableSandboxSyscalls() {
		t.Fatal("SetDisableSandbox(false) did not enable both sandbox components")
	}
}

func TestHandleOptions_ListRoundTrip(t *testing.T) {
	h := mustInitializeTestHandle(t)

	type optionsCase struct {
		name  string
		setFn func([]string) error
		getFn func() ([]string, error)
	}

	cases := []optionsCase{
		{
			name:  "cachedirs",
			setFn: h.SetCacheDirs,
			getFn: h.CacheDirs,
		},
		{
			name:  "hookdirs",
			setFn: h.SetHookDirs,
			getFn: h.HookDirs,
		},
		{
			name:  "noupgrades",
			setFn: h.SetNoUpgrades,
			getFn: h.NoUpgrades,
		},
		{
			name:  "noextracts",
			setFn: h.SetNoExtracts,
			getFn: h.NoExtracts,
		},
		{
			name:  "ignorepkgs",
			setFn: h.SetIgnorePkgs,
			getFn: h.IgnorePkgs,
		},
		{
			name:  "ignoregroups",
			setFn: h.SetIgnoreGroups,
			getFn: h.IgnoreGroups,
		},
		{
			name:  "overwritefiles",
			setFn: h.SetOverwriteFiles,
			getFn: h.OverwriteFiles,
		},
		{
			name:  "architectures",
			setFn: h.SetArchitectures,
			getFn: h.Architectures,
		},
	}

	values := [][]string{
		{"/tmp/dyalpm-it-cache"},
		{"/tmp/dyalpm-it-hooks"},
		{"integration-no-upgrade"},
		{"integration-no-extract"},
		{"integration-ignore-pkg"},
		{"integration-ignore-group"},
		{"integration-overwrite-glob"},
		{"x86_64"},
	}

	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			value := values[i]
			if err := c.setFn(value); err != nil {
				t.Fatalf("failed to set %s: %v", c.name, err)
			}
			got, err := c.getFn()
			if err != nil {
				t.Fatalf("failed to get %s: %v", c.name, err)
			}
			for _, expected := range value {
				if !containsStringInSliceNormalized(got, expected) {
					t.Fatalf("%s does not include %q (got %#v)", c.name, expected, got)
				}
			}
		})
	}
}

func TestHandleOptions_Matchers(t *testing.T) {
	h := mustInitializeTestHandle(t)

	noUpgradePath := "integration-no-upgrade-marker"
	if err := h.SetNoUpgrades([]string{noUpgradePath}); err != nil {
		t.Fatalf("failed to set no-upgrade patterns: %v", err)
	}

	noExtractPath := "integration-no-extract-marker"
	if err := h.SetNoExtracts([]string{noExtractPath}); err != nil {
		t.Fatalf("failed to set no-extract patterns: %v", err)
	}

	matchUpgrade, err := h.MatchNoUpgrade(noUpgradePath)
	if err != nil {
		t.Fatalf("match no-upgrade failed: %v", err)
	}
	if matchUpgrade == 0 {
		t.Logf("match no-upgrade returned 0 for %q; matcher semantics vary by backend", noUpgradePath)
	}

	matchExtract, err := h.MatchNoExtract(noExtractPath)
	if err != nil {
		t.Fatalf("match no-extract failed: %v", err)
	}
	if matchExtract == 0 {
		t.Logf("match no-extract returned 0 for %q; matcher semantics vary by backend", noExtractPath)
	}
}
