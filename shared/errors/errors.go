package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// ErrorType defines the category of an error
type ErrorType string

const (
	// ErrorTypeTransient represents temporary errors that might resolve on retry
	ErrorTypeTransient ErrorType = "transient"
	
	// ErrorTypePermanent represents errors that are unlikely to be resolved by retry
	ErrorTypePermanent ErrorType = "permanent"
	
	// ErrorTypeValidation represents input validation errors
	ErrorTypeValidation ErrorType = "validation"
	
	// ErrorTypeConfiguration represents configuration-related errors
	ErrorTypeConfiguration ErrorType = "configuration"
)

// CustomError is an enhanced error type with additional context
type CustomError struct {
	// Original error
	Err error
	
	// Error type for categorization
	Type ErrorType
	
	// Additional context about the error
	Context map[string]interface{}
	
	// Stack trace for debugging
	StackTrace string
}

// Error implements the error interface
func (e *CustomError) Error() string {
	var sb strings.Builder
	sb.WriteString(e.Err.Error())
	
	if len(e.Context) > 0 {
		sb.WriteString(" | Context: ")
		for k, v := range e.Context {
			sb.WriteString(fmt.Sprintf("%s=%v ", k, v))
		}
	}
	
	return sb.String()
}

// Unwrap allows error unwrapping
func (e *CustomError) Unwrap() error {
	return e.Err
}

// New creates a new CustomError
func New(err error, errorType ErrorType, context map[string]interface{}) *CustomError {
	stackTrace := getStackTrace()
	return &CustomError{
		Err:        err,
		Type:       errorType,
		Context:    context,
		StackTrace: stackTrace,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, message string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}
	
	wrappedErr := fmt.Errorf("%s: %w", message, err)
	
	// If it's already a CustomError, update it
	if customErr, ok := err.(*CustomError); ok {
		return &CustomError{
			Err:        wrappedErr,
			Type:       customErr.Type,
			Context:    mergeContextMaps(customErr.Context, context),
			StackTrace: getStackTrace(),
		}
	}
	
	// Create a new CustomError
	return &CustomError{
		Err:        wrappedErr,
		Type:       ErrorTypeTransient,
		Context:    context,
		StackTrace: getStackTrace(),
	}
}

// Is checks if the error matches a specific type
func Is(err error, errorType ErrorType) bool {
	if customErr, ok := err.(*CustomError); ok {
		return customErr.Type == errorType
	}
	return false
}

// As attempts to extract a CustomError from an error
func As(err error) *CustomError {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr
	}
	return nil
}

// mergeContextMaps combines two context maps
func mergeContextMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	
	for k, v := range map1 {
		merged[k] = v
	}
	
	for k, v := range map2 {
		merged[k] = v
	}
	
	return merged
}

// getStackTrace captures the current stack trace
func getStackTrace() string {
	var trace []string
	
	for i := 2; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		trace = append(trace, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	
	return strings.Join(trace, "\n")
}

// Predefined common errors
var (
	ErrNilArgument       = New(errors.New("nil argument provided"), ErrorTypeValidation, nil)
	ErrInvalidConfig     = New(errors.New("invalid configuration"), ErrorTypeConfiguration, nil)
	ErrResourceNotFound  = New(errors.New("resource not found"), ErrorTypePermanent, nil)
	ErrUnauthorized      = New(errors.New("unauthorized access"), ErrorTypePermanent, nil)
)
