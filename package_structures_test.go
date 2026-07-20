//go:build linux

package dyalpm

import (
	"testing"
	"unsafe"

	"github.com/Jguer/dyalpm/internal/lib"
	"github.com/Jguer/dyalpm/internal/testutil/cmem"
)

func TestSignatureResultDecoding(t *testing.T) {
	statuses := []struct {
		got  SigStatus
		want SigStatus
	}{
		{SigStatusValid, 0},
		{SigStatusKeyExpired, 1},
		{SigStatusSigExpired, 2},
		{SigStatusKeyUnknown, 3},
		{SigStatusKeyDisabled, 4},
		{SigStatusInvalid, 5},
	}
	for _, status := range statuses {
		if status.got != status.want {
			t.Fatalf("signature status = %d, want %d", status.got, status.want)
		}
	}

	fingerprint := cmem.String(t, "0123456789ABCDEF")
	resultsPtr := cmem.Alloc(t, 2*unsafe.Sizeof(alpmSigResult{}))
	results := unsafe.Slice((*alpmSigResult)(unsafe.Pointer(resultsPtr)), 2)
	results[0].Key.Fingerprint = fingerprint
	results[0].Status = int32(SigStatusKeyExpired)
	results[0].Validity = int32(SigValidityMarginal)
	results[1].Status = int32(SigStatusInvalid)
	results[1].Validity = int32(SigValidityNever)

	listPtr := cmem.Alloc(t, unsafe.Sizeof(alpmSigList{}))
	list := (*alpmSigList)(unsafe.Pointer(listPtr))
	list.Count = 2
	list.Results = resultsPtr

	decoded := decodeSigList(list)
	if len(decoded.Results) != 2 {
		t.Fatalf("signature count = %d, want 2", len(decoded.Results))
	}
	if got := decoded.Results[0]; got.KeyID != "0123456789ABCDEF" ||
		got.Status != SigStatusKeyExpired ||
		got.Validity != SigValidityMarginal {
		t.Fatalf("first signature = %#v", got)
	}
	if got := decoded.Results[1]; got.Status != SigStatusInvalid ||
		got.Validity != SigValidityNever {
		t.Fatalf("second signature = %#v", got)
	}

	*(*byte)(unsafe.Pointer(fingerprint)) = 'X'
	if decoded.Results[0].KeyID != "0123456789ABCDEF" {
		t.Fatal("decoded fingerprint aliases libalpm memory")
	}
}

func TestPackageXDataDecoding(t *testing.T) {
	firstName := cmem.String(t, "pkgtype")
	firstValue := cmem.String(t, "split")
	secondName := cmem.String(t, "custom")
	secondValue := cmem.String(t, "value")

	firstDataPtr := cmem.Alloc(t, unsafe.Sizeof(alpmPackageXData{}))
	secondDataPtr := cmem.Alloc(t, unsafe.Sizeof(alpmPackageXData{}))
	*(*alpmPackageXData)(unsafe.Pointer(firstDataPtr)) = alpmPackageXData{
		Name: firstName, Value: firstValue,
	}
	*(*alpmPackageXData)(unsafe.Pointer(secondDataPtr)) = alpmPackageXData{
		Name: secondName, Value: secondValue,
	}

	firstNodePtr := cmem.Alloc(t, unsafe.Sizeof(abiListNode{}))
	secondNodePtr := cmem.Alloc(t, unsafe.Sizeof(abiListNode{}))
	firstNode := (*abiListNode)(unsafe.Pointer(firstNodePtr))
	secondNode := (*abiListNode)(unsafe.Pointer(secondNodePtr))
	firstNode.Data = firstDataPtr
	firstNode.Next = secondNodePtr
	secondNode.Data = secondDataPtr
	secondNode.Prev = firstNodePtr

	oldGetXData := lib.AlpmPkgGetXdata
	lib.AlpmPkgGetXdata = func(uintptr) uintptr { return firstNodePtr }
	t.Cleanup(func() {
		lib.AlpmPkgGetXdata = oldGetXData
	})

	pkg := &package_{ptr: 1}
	values := pkg.XDataValues()
	if len(values) != 2 ||
		values[0] != (PackageXData{Name: "pkgtype", Value: "split"}) ||
		values[1] != (PackageXData{Name: "custom", Value: "value"}) {
		t.Fatalf("xdata values = %#v", values)
	}
	legacy := pkg.XData()
	if len(legacy) != 2 || legacy[0] != "pkgtype=split" || legacy[1] != "custom=value" {
		t.Fatalf("legacy xdata values = %#v", legacy)
	}

	*(*byte)(unsafe.Pointer(firstName)) = 'X'
	if values[0].Name != "pkgtype" {
		t.Fatal("decoded xdata aliases libalpm memory")
	}
}
