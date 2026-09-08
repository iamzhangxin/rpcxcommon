package errors

import "fmt"

type Error struct {
	code    int32
	message string
	cause   error
}

// New 创建具有六位数字编码的错误，扩展编码由使用方统一分配。
func New(code int32, message string) *Error {
	if code < 100000 || code > 999999 {
		panic("error code must contain six digits")
	}
	return &Error{code: code, message: message}
}

// Error 返回错误编码和固定描述且不暴露底层错误中的敏感信息。
func (e *Error) Error() string { return fmt.Sprintf("%06d: %s", e.code, e.message) }

// Code 返回错误的六位数字编码。
func (e *Error) Code() int32 { return e.code }

// Message 返回统一定义的错误描述。
func (e *Error) Message() string { return e.message }

// WithCause 创建保留底层原因的错误副本且不修改共享错误定义。
func (e *Error) WithCause(cause error) *Error {
	if cause == nil {
		return e
	}
	return &Error{code: e.code, message: e.message, cause: cause}
}

// Unwrap 返回底层原因以支持错误链匹配和取消信号识别。
func (e *Error) Unwrap() error { return e.cause }

// Is 按唯一错误编码识别统一错误及其包装副本。
func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && other != nil && e.code == other.code
}
