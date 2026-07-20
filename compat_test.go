//go:build linux

package dyalpm

import (
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/Hayao0819/dyalpm/internal/testutil/cmem"
)

type cListNode struct {
	Data uintptr
	Prev uintptr
	Next uintptr
}

func TestQuestionTypeValues(t *testing.T) {
	tests := []struct {
		name string
		got  QuestionType
		want QuestionType
	}{
		{"install ignorepkg", QuestionTypeInstallIgnorepkg, 1 << 0},
		{"replace package", QuestionTypeReplacePkg, 1 << 1},
		{"conflict package", QuestionTypeConflictPkg, 1 << 2},
		{"corrupted package", QuestionTypeCorruptedPkg, 1 << 3},
		{"remove packages", QuestionTypeRemovePkgs, 1 << 4},
		{"select provider", QuestionTypeSelectProvider, 1 << 5},
		{"import key", QuestionTypeImportKey, 1 << 6},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("question type = %d, want %d", test.got, test.want)
			}
		})
	}
}

func TestQuestionSelectProviderLayout(t *testing.T) {
	var question cQuestionSelectProvider

	if got, want := unsafe.Offsetof(question.UseIndex), uintptr(4); got != want {
		t.Fatalf("use_index offset = %d, want %d", got, want)
	}
	if got, want := unsafe.Offsetof(question.Providers), uintptr(8); got != want {
		t.Fatalf("providers offset = %d, want %d", got, want)
	}
	if got, want := unsafe.Offsetof(question.Depend), uintptr(8)+unsafe.Sizeof(uintptr(0)); got != want {
		t.Fatalf("depend offset = %d, want %d", got, want)
	}
}

func TestQuestionSelectProviderRejectsOtherQuestionTypes(t *testing.T) {
	tests := []QuestionType{
		QuestionTypeInstallIgnorepkg,
		QuestionTypeReplacePkg,
		QuestionTypeConflictPkg,
		QuestionTypeCorruptedPkg,
		QuestionTypeRemovePkgs,
		QuestionTypeImportKey,
	}

	for _, questionType := range tests {
		question := QuestionAny{
			Question: Question{Type: int32(questionType)},
		}
		if _, err := question.QuestionSelectProvider(); err == nil {
			t.Errorf("question type %d was accepted as select-provider", questionType)
		}
	}
}

func TestQuestionSelectProviderDecodesPayload(t *testing.T) {
	dependPtr := cmem.Alloc(t, unsafe.Sizeof(cDepend{}))
	depend := (*cDepend)(unsafe.Pointer(dependPtr))
	depend.Name = cmem.String(t, "java-runtime")
	depend.Version = cmem.String(t, "17")
	depend.Mod = int32(DepModGE)

	firstPackage := cmem.Alloc(t, 1)
	secondPackage := cmem.Alloc(t, 1)
	firstNodePtr := cmem.Alloc(t, unsafe.Sizeof(cListNode{}))
	secondNodePtr := cmem.Alloc(t, unsafe.Sizeof(cListNode{}))
	firstNode := (*cListNode)(unsafe.Pointer(firstNodePtr))
	secondNode := (*cListNode)(unsafe.Pointer(secondNodePtr))
	firstNode.Data = firstPackage
	firstNode.Next = secondNodePtr
	secondNode.Data = secondPackage
	secondNode.Prev = firstNodePtr

	questionPtr := cmem.Alloc(t, unsafe.Sizeof(cQuestionSelectProvider{}))
	rawQuestion := (*cQuestionSelectProvider)(unsafe.Pointer(questionPtr))
	rawQuestion.Type = int32(QuestionTypeSelectProvider)
	rawQuestion.UseIndex = -1
	rawQuestion.Providers = firstNodePtr
	rawQuestion.Depend = dependPtr

	question, err := (QuestionAny{Question: Question{
		Type: rawQuestion.Type,
		Ptr:  questionPtr,
	}}).QuestionSelectProvider()
	if err != nil {
		t.Fatalf("decode select-provider question: %v", err)
	}

	if got, want := question.Dep(), "java-runtime>=17"; got != want {
		t.Fatalf("dependency = %q, want %q", got, want)
	}

	h := &handle{ptr: 1}
	providers := question.Providers(h)
	if providers.state == nil || providers.state.owned {
		t.Fatal("provider list must remain owned by libalpm")
	}
	packages := providers.Collect()
	if got, want := len(packages), 2; got != want {
		t.Fatalf("provider count = %d, want %d", got, want)
	}
	if got := packages[0].(*package_); got.ptr != firstPackage || got.handle != h {
		t.Fatalf("first provider = %#v, want ptr=%#x handle=%p", got, firstPackage, h)
	}
	if got := packages[1].(*package_); got.ptr != secondPackage || got.handle != h {
		t.Fatalf("second provider = %#v, want ptr=%#x handle=%p", got, secondPackage, h)
	}

	question.SetUseIndex(0)
	if rawQuestion.UseIndex != 0 {
		t.Fatalf("zero-based provider index = %d, want 0", rawQuestion.UseIndex)
	}
	question.SetUseIndex(1)
	if rawQuestion.UseIndex != 1 {
		t.Fatalf("provider index = %d, want 1", rawQuestion.UseIndex)
	}
}

func TestQuestionCallbackDecodesSelectProvider(t *testing.T) {
	questionPtr := cmem.Alloc(t, unsafe.Sizeof(cQuestionSelectProvider{}))
	rawQuestion := (*cQuestionSelectProvider)(unsafe.Pointer(questionPtr))
	rawQuestion.Type = int32(QuestionTypeSelectProvider)
	rawQuestion.UseIndex = -1

	const ctx = uintptr(0x5150)
	set := getOrCreateCallbackSet(ctx)
	t.Cleanup(func() {
		unregisterCallbackSet(ctx)
	})

	called := false
	set.mu.Lock()
	set.question = func(raw Question) {
		called = true
		question, err := (QuestionAny{Question: raw}).QuestionSelectProvider()
		if err != nil {
			t.Errorf("decode callback question: %v", err)
			return
		}
		question.SetUseIndex(0)
	}
	set.mu.Unlock()

	questioncbTrampoline(purego.CDecl{}, ctx, questionPtr)

	if !called {
		t.Fatal("question callback was not called")
	}
	if rawQuestion.UseIndex != 0 {
		t.Fatalf("selected provider index = %d, want 0", rawQuestion.UseIndex)
	}
}

func TestQuestionSelectProviderZeroValue(t *testing.T) {
	var question *QuestionSelectProvider

	if got := question.Dep(); got != "" {
		t.Fatalf("nil question dependency = %q, want empty", got)
	}
	if got := question.Providers(nil).Collect(); len(got) != 0 {
		t.Fatalf("nil question providers = %d, want 0", len(got))
	}
	question.SetUseIndex(0)
}
