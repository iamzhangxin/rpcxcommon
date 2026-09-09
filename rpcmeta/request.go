// Package rpcmeta 定义在 Rpc 调用链中持续透传的请求上下文。
package rpcmeta

import (
	"context"
	"net/http"
	"strings"

	"github.com/bytedance/gopkg/cloud/metainfo"
)

const (
	UserIdKey     = "x-user-id"
	AppCodeKey    = "x-app-code"
	DeviceIdKey   = "x-device-id"
	DeviceTypeKey = "x-device-type"
	DeviceNameKey = "x-device-name"
)

// RequestInfo 描述已由可信入口处理的用户、应用与设备请求信息。
type RequestInfo struct {
	UserId     string
	AppCode    string
	DeviceId   string
	DeviceType string
	DeviceName string
}

// WithRequestInfo 整体替换请求上下文，空字段清除旧值，原上下文不变。
func WithRequestInfo(ctx context.Context, info RequestInfo) context.Context {
	ctx = WithUserId(ctx, info.UserId)
	ctx = WithAppCode(ctx, info.AppCode)
	ctx = WithDeviceId(ctx, info.DeviceId)
	ctx = WithDeviceType(ctx, info.DeviceType)
	return WithDeviceName(ctx, info.DeviceName)
}

// FromContext 读取当前 Rpc 调用链中的完整请求信息。
func FromContext(ctx context.Context) RequestInfo {
	return RequestInfo{UserId: UserId(ctx), AppCode: AppCode(ctx), DeviceId: DeviceId(ctx), DeviceType: DeviceType(ctx), DeviceName: DeviceName(ctx)}
}

// FromHTTPHeader 仅提取约定的五个请求头，缺失、空白或重复字段作为空值处理。
func FromHTTPHeader(header http.Header) RequestInfo {
	return RequestInfo{UserId: headerValue(header, UserIdKey), AppCode: headerValue(header, AppCodeKey), DeviceId: headerValue(header, DeviceIdKey), DeviceType: headerValue(header, DeviceTypeKey), DeviceName: headerValue(header, DeviceNameKey)}
}

// headerValue 忽略头名大小写，避免多个值或不同大小写键造成身份歧义。
func headerValue(header http.Header, key string) string {
	count, value := 0, ""
	for name, values := range header {
		if strings.EqualFold(name, key) {
			count += len(values)
			if len(values) == 1 {
				value = values[0]
			}
		}
	}
	if count != 1 {
		return ""
	}
	return strings.TrimSpace(value)
}

// value 读取并归一化持久请求元信息。
func value(ctx context.Context, key string) string {
	v, _ := metainfo.GetPersistentValue(ctx, key)
	return strings.TrimSpace(v)
}

// withValue 写入持久请求元信息，空值清除当前上下文中的对应字段。
func withValue(ctx context.Context, key, v string) context.Context {
	v = strings.TrimSpace(v)
	if v == "" {
		return metainfo.DelPersistentValue(ctx, key)
	}
	return metainfo.WithPersistentValue(ctx, key, v)
}

// UserId 读取已由入口校验的用户身份。
func UserId(ctx context.Context) string { return value(ctx, UserIdKey) }

// WithUserId 写入用户身份；保留既有调用方式。
func WithUserId(ctx context.Context, id string) context.Context { return withValue(ctx, UserIdKey, id) }

// AppCode 读取应用编码。
func AppCode(ctx context.Context) string { return value(ctx, AppCodeKey) }

// WithAppCode 写入应用编码。
func WithAppCode(ctx context.Context, code string) context.Context {
	return withValue(ctx, AppCodeKey, code)
}

// DeviceId 读取设备标识。
func DeviceId(ctx context.Context) string { return value(ctx, DeviceIdKey) }

// WithDeviceId 写入设备标识。
func WithDeviceId(ctx context.Context, id string) context.Context {
	return withValue(ctx, DeviceIdKey, id)
}

// DeviceType 读取设备类型。
func DeviceType(ctx context.Context) string { return value(ctx, DeviceTypeKey) }

// WithDeviceType 写入设备类型。
func WithDeviceType(ctx context.Context, kind string) context.Context {
	return withValue(ctx, DeviceTypeKey, kind)
}

// DeviceName 读取设备名称。
func DeviceName(ctx context.Context) string { return value(ctx, DeviceNameKey) }

// WithDeviceName 写入设备名称。
func WithDeviceName(ctx context.Context, name string) context.Context {
	return withValue(ctx, DeviceNameKey, name)
}
