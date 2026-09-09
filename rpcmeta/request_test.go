package rpcmeta

import (
	"context"
	"github.com/bytedance/gopkg/cloud/metainfo"
	"net/http"
	"testing"
)

// TestRequestInfo 验证五个头原样透传，不归一化或校验字段内容。
func TestRequestInfo(t *testing.T) {
	headers := http.Header{"x-user-id": {" user-1 "}, "X-App-Code": {" product "}, "X-Device-Id": {" device-1 "}, "X-Device-Type": {"unknown-type"}, "X-Device-Name": {" 手机 "}, "Authorization": {"secret"}}
	want := RequestInfo{UserId: " user-1 ", AppCode: " product ", DeviceId: " device-1 ", DeviceType: "unknown-type", DeviceName: " 手机 "}
	info := FromHTTPHeader(headers)
	if info != want {
		t.Fatalf("header changed: %+v", info)
	}
	parent := WithRequestInfo(context.Background(), info)
	if FromContext(parent) != want {
		t.Fatal("request info changed")
	}
	child := WithRequestInfo(parent, RequestInfo{})
	if FromContext(child) != (RequestInfo{}) || FromContext(parent) != want {
		t.Fatal("request metadata leaked or parent changed")
	}
	for _, key := range []string{UserIdKey, AppCodeKey, DeviceIdKey, DeviceTypeKey, DeviceNameKey} {
		if got, _ := metainfo.GetPersistentValue(child, key); got != "" {
			t.Fatalf("empty field %s was not written", key)
		}
	}
	if FromHTTPHeader(nil) != (RequestInfo{}) {
		t.Fatal("missing headers not empty")
	}
}

// TestUserId 验证原有单字段接口仍可使用，并且不会修改字段内容。
func TestUserId(t *testing.T) {
	parent := context.Background()
	ctx := WithUserId(parent, " user-1 ")
	if UserId(ctx) != " user-1 " || UserId(parent) != "" {
		t.Fatal("user id changed or leaked")
	}
	child := WithUserId(ctx, "")
	if value, _ := metainfo.GetPersistentValue(child, UserIdKey); value != "" {
		t.Fatal("empty user id was not propagated")
	}
	if UserId(ctx) != " user-1 " {
		t.Fatal("parent changed")
	}
}
