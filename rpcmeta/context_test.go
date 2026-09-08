package rpcmeta

import (
	"context"
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
