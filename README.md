# rpcxcommon

服务端与 Gateway 共用的上下文和错误库。

```go
import (
    "github.com/iamzhangxin/rpcxcommon/errors"
    "github.com/iamzhangxin/rpcxcommon/rpcmeta"
)
```

## 上下文

`rpcmeta.RequestInfo` 统一表达用户、应用与设备请求上下文：

| 字段 | HTTP Header | 单字段读取 |
| --- | --- | --- |
| UserId | x-user-id | rpcmeta.UserId(ctx) |
| AppCode | x-app-code | rpcmeta.AppCode(ctx) |
| DeviceId | x-device-id | rpcmeta.DeviceId(ctx) |
| DeviceType | x-device-type | rpcmeta.DeviceType(ctx) |
| DeviceName | x-device-name | rpcmeta.DeviceName(ctx) |

```go
info := rpcmeta.FromHTTPHeader(request.Header)
ctx := rpcmeta.WithRequestInfo(request.Context(), info)
// 在下游读取完整请求信息：
info = rpcmeta.FromContext(ctx)
// 需要用户身份的接口可使用：
if info.UserId == "" {
    return nil, errors.ToRpc(ctx, errors.ErrIdentityRequired)
}
```

所有字段均去除首尾空白并写入 persistent metainfo。整体写入会清除未提供的字段，防止旧上下文信息残留；缺失、空白或重复 HTTP 头作为空值处理。只有用户身份在需要登录的接口上要求非空，设备字段不默认必填。

保留 `WithUserId` / `UserId`，并提供 `WithAppCode`、`WithDeviceId`、`WithDeviceType`、`WithDeviceName` 进行单字段更新。调用后续服务时继续传递当前 Context。

入口负责身份校验；本包只传递身份，不验证 Token。Kitex 客户端需同时启用 TTHeader 和 `transmeta.ClientTTHeaderHandler`，服务端需启用 `transmeta.ServerTTHeaderHandler`，才能通过网络持续透传。

## 错误

错误按通用语义集中定义，不按 Category、User 等业务模块重复分配。

| 定义 | 六位编码 | 含义 |
| --- | --- | --- |
| ErrInvalidArgument | 400001 | 参数无效 |
| ErrIdentityRequired | 401003 | 用户身份不能为空 |
| ErrPermissionDenied | 403001 | 权限不足 |
| ErrNotFound | 404001 | 资源不存在 |
| ErrConflict | 409001 | 资源冲突 |
| ErrInternal | 500001 | 内部错误 |
| ErrInvalidData | 500101 | 存储数据不合法 |
| ErrReadFailed | 500102 | 读取失败 |
| ErrWriteFailed | 500103 | 写入失败 |
| ErrInvalidConfig | 500201 | 配置无效 |
| ErrConfigUnavailable | 500202 | 配置不可用 |
| ErrDependencyUnavailable | 500203 | 依赖不可用 |

`WithCause(err)` 创建错误副本并保留错误链，不修改共享定义；标准库 `errors.Is`、`errors.As` 可继续使用。错误文本不包含底层原因，原因可通过 `Unwrap` 获取。

`ToRpc(ctx, err)` 转为 Kitex 业务状态：本地编码小于 500000 时返回其编码和固定描述，本地内部错误及未知错误统一返回 500001；下游合法六位业务状态直接透传，取消和超时保留错误链。日志不输出底层错误中的配置或数据库细节。Http 响应格式由 Gateway 决定。

`FromStorageWrite(err)` 通过数据库驱动已有的 `SQLState()` 接口识别唯一约束冲突并转换为 409001，其余写入失败转换为 500103。不依赖具体数据库驱动；标准数据库状态码 23505 不属于本库业务编码。

确有通用定义无法表达的业务语义时，可通过 `New(code, message)` 扩展，编码必须为六位数字且由使用方统一分配，不能与已有编码重复；`Is` 按编码匹配。传给客户端的描述应是固定、安全的文本。

## 验证

```sh
go test -race ./...
go vet ./...
```

每个手写方法附一行中文注释，自定义缩写采用 `Rpc`、`Id` 等形式；第三方接口名保留既有契约。
