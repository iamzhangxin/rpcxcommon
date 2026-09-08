package errors

import (
	"context"
	stderrors "errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/cloudwego/kitex/pkg/kerrors"
)

// TestCatalogCodes 验证集中定义的错误编码均为唯一的六位数字。
func TestCatalogCodes(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[int64]string{}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "New" {
				return true
			}
			literal, ok := call.Args[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.INT {
				t.Errorf("error code must be a numeric literal at %s", set.Position(call.Pos()))
				return true
			}
			code, err := strconv.ParseInt(literal.Value, 10, 32)
			if err != nil || code < 100000 || code > 999999 {
				t.Errorf("invalid six-digit code: %s", literal.Value)
			}
			where := set.Position(call.Pos()).String()
			if previous, ok := seen[code]; ok {
				t.Errorf("duplicate error code %d at %s and %s", code, previous, where)
			}
			seen[code] = where
			return true
		})
	}
	if len(seen) == 0 {
		t.Fatal("no central error definitions found")
	}
}

// TestErrorCause 验证包装保留错误链且共享错误定义不被修改。
func TestErrorCause(t *testing.T) {
	cause := stderrors.New("private database detail")
	err := ErrNotFound.WithCause(cause)
	if !stderrors.Is(err, ErrNotFound) || !stderrors.Is(err, cause) {
		t.Fatal("error chain was lost")
	}
	if ErrNotFound.Unwrap() != nil {
		t.Fatal("shared error definition was mutated")
	}
	if strings.Contains(err.Error(), "private") {
		t.Fatal("cause leaked into public error text")
	}
	if err.Code() != 404001 {
		t.Fatal("code changed while wrapping")
	}
}

// TestToRpc 验证统一状态转换、内部信息隐藏及上下游错误兼容。
func TestToRpc(t *testing.T) {
	for _, test := range []struct {
		name    string
		err     error
		code    int32
		message string
	}{
		{"argument", ErrInvalidArgument, 400001, "invalid argument"},
		{"identity", ErrIdentityRequired, 401001, "user identity is required"},
		{"wrapped", fmt.Errorf("wrapped: %w", ErrNotFound), 404001, "resource not found"},
		{"conflict", ErrConflict, 409001, "resource conflict"},
		{"internal", ErrReadFailed.WithCause(stderrors.New("private sql detail")), 500001, "internal error"},
		{"unknown", stderrors.New("private unknown detail"), 500001, "internal error"},
		{"downstream", kerrors.NewBizStatusError(404002, "downstream item missing"), 404002, "downstream item missing"},
		{"legacy-code", kerrors.NewBizStatusError(40101, "legacy detail"), 500001, "internal error"},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := ToRpc(context.Background(), test.err)
			biz, ok := kerrors.FromBizStatusError(err)
			if !ok || biz.BizStatusCode() != test.code || biz.BizMessage() != test.message {
				t.Fatalf("unexpected Rpc error: %v", err)
			}
		})
	}
	if ToRpc(context.Background(), nil) != nil {
		t.Fatal("nil error became a failure")
	}
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		err := ErrReadFailed.WithCause(cause)
		if ToRpc(context.Background(), err) != err {
			t.Fatal("cancellation semantics changed")
		}
	}
}

// TestFromStorageWrite 验证数据库唯一约束错误集中映射且底层原因可追溯。
func TestFromStorageWrite(t *testing.T) {
	unique := &storageFailure{state: "23505"}
	err := FromStorageWrite(unique)
	if !stderrors.Is(err, ErrConflict) || !stderrors.Is(err, unique) {
		t.Fatal("unique constraint was not classified")
	}
	other := &storageFailure{state: "42501"}
	if !stderrors.Is(FromStorageWrite(other), ErrWriteFailed) {
		t.Fatal("database failure was not classified")
	}
	if FromStorageWrite(nil) != nil {
		t.Fatal("nil database error became a failure")
	}
}

type storageFailure struct{ state string }

// Error 返回测试使用的数据库内部信息。
func (e *storageFailure) Error() string { return "private constraint detail" }

// SQLState 实现数据库驱动约定的状态码接口。
func (e *storageFailure) SQLState() string { return e.state }

// TestNewCodeValidation 验证扩展错误编码边界及非法编码拒绝行为。
func TestNewCodeValidation(t *testing.T) {
	for _, code := range []int32{100000, 999999} {
		if New(code, "custom error").Code() != code {
			t.Fatal("valid code changed")
		}
	}
	for _, code := range []int32{0, 99999, 1000000} {
		t.Run(strconv.Itoa(int(code)), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid code accepted")
				}
			}()
			New(code, "invalid code")
		})
	}
}
