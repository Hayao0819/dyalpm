package dyalpm

import (
	stderrors "errors"
	"strings"
	"unsafe"

	"github.com/Hayao0819/dyalpm/internal/lib"
)

// LogLevel represents logging level
type LogLevel uint

const (
	LogError   LogLevel = 1
	LogWarning LogLevel = 1 << 1
	LogDebug   LogLevel = 1 << 2
	LogFunc    LogLevel = 1 << 3
)

type QuestionType int32

const (
	QuestionTypeInstallIgnorepkg QuestionType = 1 << iota
	QuestionTypeReplacePkg
	QuestionTypeConflictPkg
	QuestionTypeCorruptedPkg
	QuestionTypeRemovePkgs
	QuestionTypeSelectProvider
	QuestionTypeImportKey
)

// QuestionAny is a wrapper for question callback data (go-alpm/v2 compatibility)
type QuestionAny struct {
	Question Question
}

// QuestionInstallIgnorepkg returns a QuestionInstallIgnorepkg if this is an install-ignorepkg question
func (qa QuestionAny) QuestionInstallIgnorepkg() (*QuestionInstallIgnorepkg, error) {
	if QuestionType(qa.Question.Type) == QuestionTypeInstallIgnorepkg {
		return &QuestionInstallIgnorepkg{q: qa.Question}, nil
	}
	return nil, stderrors.New("not an install ignorepkg question")
}

// QuestionSelectProvider returns a QuestionSelectProvider if this is a select-provider question
func (qa QuestionAny) QuestionSelectProvider() (*QuestionSelectProvider, error) {
	if QuestionType(qa.Question.Type) == QuestionTypeSelectProvider {
		return &QuestionSelectProvider{q: qa.Question}, nil
	}
	return nil, stderrors.New("not a select provider question")
}

// QuestionInstallIgnorepkg is for handling install-ignorepkg questions
type QuestionInstallIgnorepkg struct {
	q Question
}

// SetInstall sets whether to install the ignored package
func (qi *QuestionInstallIgnorepkg) SetInstall(install bool) {
	answer := 0
	if install {
		answer = 1
	}
	qi.q.SetAnswerInt(answer)
}

// QuestionSelectProvider is for handling select-provider questions
type QuestionSelectProvider struct {
	q Question
}

type cQuestionSelectProvider struct {
	Type      int32
	UseIndex  int32
	Providers uintptr
	Depend    uintptr
}

type cDepend struct {
	Name        uintptr
	Version     uintptr
	Description uintptr
	NameHash    uintptr
	Mod         int32
}

func (qp *QuestionSelectProvider) data() *cQuestionSelectProvider {
	if qp == nil || qp.q.Ptr == 0 ||
		QuestionType(qp.q.Type) != QuestionTypeSelectProvider {
		return nil
	}

	data := (*cQuestionSelectProvider)(unsafe.Pointer(qp.q.Ptr))
	if QuestionType(data.Type) != QuestionTypeSelectProvider {
		return nil
	}
	return data
}

// Dep returns the dependency string being resolved
func (qp *QuestionSelectProvider) Dep() string {
	data := qp.data()
	if data == nil || data.Depend == 0 {
		return ""
	}

	dep := (*cDepend)(unsafe.Pointer(data.Depend))
	return Depend{
		Name:    strings.Clone(lib.PtrToString(dep.Name)),
		Version: strings.Clone(lib.PtrToString(dep.Version)),
		Mod:     DepMod(dep.Mod),
	}.String()
}

// Providers is valid only during the question callback.
func (qp *QuestionSelectProvider) Providers(h Handle) PackageIterator {
	data := qp.data()
	if data == nil || data.Providers == 0 {
		return PackageIterator{}
	}

	handle, _ := h.(*handle)
	return newPackageIterator(data.Providers, handle, false)
}

// SetUseIndex sets which provider to use by its zero-based index.
func (qp *QuestionSelectProvider) SetUseIndex(index int) {
	data := qp.data()
	if data == nil {
		return
	}
	data.UseIndex = clampIntToInt32(index)
}
