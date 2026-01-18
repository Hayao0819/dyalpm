package lib

// FunctionCaller is an interface for calling C functions.
// This abstraction allows for mocking in unit tests.
type FunctionCaller interface {
	// GetFunc returns a function pointer by name
	GetFunc(name string) (uintptr, error)
	// CallFunc calls a function by name with the given arguments
	CallFunc(name string, args ...uintptr) (uintptr, error)
}

// Ensure FunctionRegistry implements FunctionCaller
var _ FunctionCaller = (*FunctionRegistry)(nil)
