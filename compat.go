package dyalpm

import (
	stderrors "errors"
)

// LogLevel represents logging level
type LogLevel uint

const (
	LogError   LogLevel = 1
	LogWarning LogLevel = 1 << 1
	LogDebug   LogLevel = 1 << 2
	LogFunc    LogLevel = 1 << 3
)

// QuestionAny is a wrapper for question callback data (go-alpm/v2 compatibility)
type QuestionAny struct {
	Question Question
}

// QuestionInstallIgnorepkg returns a QuestionInstallIgnorepkg if this is an install-ignorepkg question
func (qa QuestionAny) QuestionInstallIgnorepkg() (*QuestionInstallIgnorepkg, error) {
	// Question type 1 = install ignorepkg
	if qa.Question.Type == 1 {
		return &QuestionInstallIgnorepkg{q: qa.Question}, nil
	}
	return nil, stderrors.New("not an install ignorepkg question")
}

// QuestionSelectProvider returns a QuestionSelectProvider if this is a select-provider question
func (qa QuestionAny) QuestionSelectProvider() (*QuestionSelectProvider, error) {
	// Question type 4 = select provider
	if qa.Question.Type == 4 {
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

// Dep returns the dependency string being resolved
func (qp *QuestionSelectProvider) Dep() string {
	// TODO: implement proper extraction from question struct
	return ""
}

// Providers returns the list of provider packages
func (qp *QuestionSelectProvider) Providers(h Handle) PackageIterator {
	// TODO: implement proper extraction from question struct
	_ = h
	return PackageIterator{}
}

// SetUseIndex sets which provider index to use (1-based)
func (qp *QuestionSelectProvider) SetUseIndex(index int) {
	qp.q.SetAnswerInt(index)
}
