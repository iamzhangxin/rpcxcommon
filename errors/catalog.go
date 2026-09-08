package errors

// 通用错误按语义分配六位编码，不按业务模块分别定义。
var (
	ErrInvalidArgument       = New(400001, "invalid argument")
	ErrIdentityRequired      = New(401001, "user identity is required")
	ErrPermissionDenied      = New(403001, "permission denied")
	ErrNotFound              = New(404001, "resource not found")
	ErrConflict              = New(409001, "resource conflict")
	ErrInternal              = New(500001, "internal error")
	ErrInvalidData           = New(500101, "invalid stored data")
	ErrReadFailed            = New(500102, "read failed")
	ErrWriteFailed           = New(500103, "write failed")
	ErrInvalidConfig         = New(500201, "invalid configuration")
	ErrConfigUnavailable     = New(500202, "configuration unavailable")
	ErrDependencyUnavailable = New(500203, "dependency unavailable")
)
