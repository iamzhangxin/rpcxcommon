package errors

// 通用错误按语义分配六位编码，不按业务模块分别定义。
var (
	ErrInvalidArgument       = New(400001, "参数无效")
	ErrIdentityRequired      = New(401003, "用户身份不能为空")
	ErrPermissionDenied      = New(403001, "权限不足")
	ErrNotFound              = New(404001, "资源不存在")
	ErrConflict              = New(409001, "资源冲突")
	ErrInternal              = New(500001, "内部错误")
	ErrInvalidData           = New(500101, "存储数据不合法")
	ErrReadFailed            = New(500102, "读取失败")
	ErrWriteFailed           = New(500103, "写入失败")
	ErrInvalidConfig         = New(500201, "配置无效")
	ErrConfigUnavailable     = New(500202, "配置不可用")
	ErrDependencyUnavailable = New(500203, "依赖不可用")
)
