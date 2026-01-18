package lib

import (
	"errors"
	"sync"
)

// MockFunctionRegistry is a mock implementation of FunctionCaller for testing.
type MockFunctionRegistry struct {
	mu          sync.RWMutex
	funcs       map[string]uintptr
	callResults map[string]uintptr
	callErrors  map[string]error
	callCounts  map[string]int
}

// NewMockRegistry creates a new mock function registry.
func NewMockRegistry() *MockFunctionRegistry {
	return &MockFunctionRegistry{
		funcs:       make(map[string]uintptr),
		callResults: make(map[string]uintptr),
		callErrors:  make(map[string]error),
		callCounts:  make(map[string]int),
	}
}

// RegisterFunc registers a mock function pointer.
func (m *MockFunctionRegistry) RegisterFunc(name string, ptr uintptr) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.funcs[name] = ptr
}

// SetCallResult sets the result that CallFunc should return for a function.
func (m *MockFunctionRegistry) SetCallResult(name string, result uintptr) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callResults[name] = result
}

// SetCallError sets an error that CallFunc should return for a function.
func (m *MockFunctionRegistry) SetCallError(name string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callErrors[name] = err
}

// GetCallCount returns how many times a function was called.
func (m *MockFunctionRegistry) GetCallCount(name string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCounts[name]
}

// GetFunc returns a mock function pointer.
func (m *MockFunctionRegistry) GetFunc(name string) (uintptr, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ptr, ok := m.funcs[name]; ok {
		return ptr, nil
	}
	return 0, errors.New("function not found: " + name)
}

// CallFunc returns a mock result for the function call.
func (m *MockFunctionRegistry) CallFunc(name string, args ...uintptr) (uintptr, error) {
	_ = args
	m.mu.Lock()
	m.callCounts[name]++
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, ok := m.callErrors[name]; ok {
		return 0, err
	}
	if result, ok := m.callResults[name]; ok {
		return result, nil
	}
	return 0, nil
}

// Reset clears all mock data.
func (m *MockFunctionRegistry) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.funcs = make(map[string]uintptr)
	m.callResults = make(map[string]uintptr)
	m.callErrors = make(map[string]error)
	m.callCounts = make(map[string]int)
}

// Ensure MockFunctionRegistry implements FunctionCaller
var _ FunctionCaller = (*MockFunctionRegistry)(nil)
