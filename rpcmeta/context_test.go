package rpcmeta

import (
	"context"
	"net/http"
	"testing"

	"github.com/bytedance/gopkg/cloud/metainfo"
)

// TestUserId 验证身份归一化、持续元信息和上下文隔离。
func TestUserId(t *testing.T) {
	parent := context.Background()
	if UserId(parent) != "" {
		t.Fatal("unexpected identity")
	}
	ctx := WithUserId(parent, "  user-1  ")
	if UserId(ctx) != "user-1" {
		t.Fatal("identity was not normalized")
	}
	if value, ok := metainfo.GetPersistentValue(ctx, UserIdKey); !ok || value != "user-1" {
		t.Fatal("identity is not persistent")
	}
	child := WithUserId(ctx, "user-2")
	if UserId(child) != "user-2" || UserId(ctx) != "user-1" || UserId(parent) != "" {
		t.Fatal("context identity leaked")
	}
	if UserId(WithUserId(ctx, "  ")) != "" {
		t.Fatal("empty identity was not cleared")
	}
	incoming := metainfo.WithPersistentValue(parent, UserIdKey, " user-3 ")
	if UserId(incoming) != "user-3" {
		t.Fatal("incoming identity was not normalized")
	}
}

// TestRequestInfo 验证五个字段的 HTTP 读取、归一化、持续透传和空值清理。
func TestRequestInfo(t *testing.T) {
	headers := http.Header{"x-user-id": {" user-1 "}, "X-App-Code": {" product "}, "X-Device-Id": {" device-1 "}, "X-Device-Type": {" mobile "}, "X-Device-Name": {" 手机 "}, "Authorization": {"secret"}}
	want := RequestInfo{UserId: "user-1", AppCode: "product", DeviceId: "device-1", DeviceType: "mobile", DeviceName: "手机"}
	info := FromHTTPHeader(headers)
	if info != want {
		t.Fatalf("header mapping: %+v", info)
	}
	parent := WithRequestInfo(context.Background(), info)
	if FromContext(parent) != want {
		t.Fatal("request info lost")
	}
	for key, want := range map[string]string{UserIdKey: info.UserId, AppCodeKey: info.AppCode, DeviceIdKey: info.DeviceId, DeviceTypeKey: info.DeviceType, DeviceNameKey: info.DeviceName} {
		if got, _ := metainfo.GetPersistentValue(parent, key); got != want {
			t.Fatalf("key %s lost", key)
		}
	}
	child := WithRequestInfo(parent, RequestInfo{AppCode: "next"})
	if FromContext(child) != (RequestInfo{AppCode: "next"}) || FromContext(parent) != want {
		t.Fatal("request metadata leaked or parent changed")
	}
	headers["X-User-Id"] = []string{"forged"}
	if FromHTTPHeader(headers).UserId != "" {
		t.Fatal("duplicate identity accepted")
	}
	headers["X-App-Code"] = []string{"one", "two"}
	if FromHTTPHeader(headers).AppCode != "" {
		t.Fatal("duplicate app accepted")
	}
	if FromHTTPHeader(nil) != (RequestInfo{}) {
		t.Fatal("missing headers not empty")
	}
}
