package dyalpm

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Hayao0819/dyalpm/internal/lib"
)

type iteratorListFixture struct {
	ptr      uintptr
	free     func(uintptr)
	released atomic.Bool
}

func newIteratorListFixture(t *testing.T, packages ...uintptr) *iteratorListFixture {
	t.Helper()
	if err := lib.EnsureAlpmLoaded(); err != nil {
		t.Skipf("libalpm not available: %v", err)
	}
	if lib.AlpmListAdd == nil || lib.AlpmListFree == nil {
		t.Skip("libalpm list functions not available")
	}

	fixture := &iteratorListFixture{free: lib.AlpmListFree}
	t.Cleanup(func() {
		if fixture.ptr != 0 && fixture.released.CompareAndSwap(false, true) {
			fixture.free(fixture.ptr)
		}
	})

	for _, pkg := range packages {
		next := lib.AlpmListAdd(fixture.ptr, pkg)
		if next == 0 {
			t.Fatal("failed to allocate ALPM list")
		}
		fixture.ptr = next
	}
	return fixture
}

func (f *iteratorListFixture) observeFree(t *testing.T) *atomic.Int32 {
	t.Helper()
	original := lib.AlpmListFree
	var calls atomic.Int32

	lib.AlpmListFree = func(ptr uintptr) {
		if ptr != f.ptr {
			original(ptr)
			return
		}
		calls.Add(1)
		if f.released.CompareAndSwap(false, true) {
			f.free(ptr)
		}
	}
	t.Cleanup(func() {
		lib.AlpmListFree = original
	})
	return &calls
}

func packagePointers(t *testing.T, it PackageIterator) []uintptr {
	t.Helper()
	packages := it.Collect()
	var ptrs []uintptr
	for _, pkg := range packages {
		impl, ok := pkg.(*package_)
		if !ok {
			t.Fatalf("package type = %T, want *package_", pkg)
		}
		ptrs = append(ptrs, impl.ptr)
	}
	return ptrs
}

func concurrentPackageVisits(it PackageIterator, workers int) int32 {
	start := make(chan struct{})
	var visits atomic.Int32
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func(current PackageIterator) {
			defer wg.Done()
			<-start
			_ = current.ForEach(func(Package) error {
				visits.Add(1)
				return nil
			})
		}(it)
	}
	close(start)
	wg.Wait()
	return visits.Load()
}

func TestPackageIteratorZeroValue(t *testing.T) {
	var it PackageIterator
	called := false
	if err := it.ForEach(func(Package) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("ForEach() error = %v", err)
	}
	if called || it.Collect() != nil || it.SortBySize() != nil {
		t.Fatal("zero-value iterator produced packages")
	}
	if pkg, err := it.FindSatisfier("missing"); pkg != nil || !errors.Is(err, ErrPackageNotFound) {
		t.Fatalf("FindSatisfier() = (%v, %v), want ErrPackageNotFound", pkg, err)
	}
}

func TestPackageIteratorBorrowedIsReusable(t *testing.T) {
	fixture := newIteratorListFixture(t, 11, 0, 22)
	freeCalls := fixture.observeFree(t)
	it := newPackageIterator(fixture.ptr, nil, false)
	copied := it

	for i, current := range []PackageIterator{it, copied, it} {
		got := packagePointers(t, current)
		if len(got) != 2 || got[0] != 11 || got[1] != 22 {
			t.Fatalf("iteration %d = %v, want [11 22]", i, got)
		}
	}
	if got := freeCalls.Load(); got != 0 {
		t.Fatalf("free calls = %d, want 0", got)
	}
}

func TestPackageIteratorOwnedCopyReleasesOnce(t *testing.T) {
	fixture := newIteratorListFixture(t, 11, 22)
	freeCalls := fixture.observeFree(t)
	it := newPackageIterator(fixture.ptr, nil, true)
	copied := it

	got := packagePointers(t, copied)
	if len(got) != 2 || got[0] != 11 || got[1] != 22 {
		t.Fatalf("first iteration = %v, want [11 22]", got)
	}
	if got := packagePointers(t, it); len(got) != 0 {
		t.Fatalf("second iteration = %v, want nil", got)
	}
	if got := freeCalls.Load(); got != 1 {
		t.Fatalf("free calls = %d, want 1", got)
	}
}

func TestPackageIteratorOwnedReleasesOnCallbackError(t *testing.T) {
	fixture := newIteratorListFixture(t, 11, 22)
	freeCalls := fixture.observeFree(t)
	it := newPackageIterator(fixture.ptr, nil, true)
	wantErr := errors.New("stop")

	seen := 0
	err := it.ForEach(func(Package) error {
		seen++
		return wantErr
	})
	if !errors.Is(err, wantErr) || seen != 1 {
		t.Fatalf("ForEach() = (%d callbacks, %v), want (1, %v)", seen, err, wantErr)
	}
	if got := freeCalls.Load(); got != 1 {
		t.Fatalf("free calls = %d, want 1", got)
	}

	called := false
	if err := it.ForEach(func(Package) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("second ForEach() error = %v", err)
	}
	if called || freeCalls.Load() != 1 {
		t.Fatal("consumed iterator was traversed or freed again")
	}
}

func TestPackageIteratorConcurrentUse(t *testing.T) {
	const workers = 16
	tests := []struct {
		name       string
		owned      bool
		wantVisits int32
		wantFrees  int32
	}{
		{name: "owned", owned: true, wantVisits: 3, wantFrees: 1},
		{name: "borrowed", wantVisits: workers * 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newIteratorListFixture(t, 11, 22, 33)
			freeCalls := fixture.observeFree(t)
			it := newPackageIterator(fixture.ptr, nil, tt.owned)

			if got := concurrentPackageVisits(it, workers); got != tt.wantVisits {
				t.Fatalf("callback calls = %d, want %d", got, tt.wantVisits)
			}
			if got := freeCalls.Load(); got != tt.wantFrees {
				t.Fatalf("free calls = %d, want %d", got, tt.wantFrees)
			}
		})
	}
}
