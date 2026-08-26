package service

import (
	"errors"
	"testing"

	"task251-chartambiguity/internal/model"
	"task251-chartambiguity/internal/store"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return New(st)
}

func TestCreateSpecRequiresExistingPublishableFigure(t *testing.T) {
	svc := newTestService(t)
	figure, err := svc.CreateFigure("spec gate")
	if err != nil {
		t.Fatalf("create figure: %v", err)
	}
	if _, err := svc.CreateSpec(figure.ID); err == nil {
		t.Fatal("spec created before figure was checked")
	}
}

// TestDeclareEncodingRejectsUndeclaredVariable 验证变量归属校验：当编码引用
// 一个当前图形稿未声明的变量时，声明应被拒绝，而非返回 201 创建成功。
// 否则系统会为不存在的变量建立「变量-通道」映射，使复核结果出现无法解释的变量。
func TestDeclareEncodingRejectsUndeclaredVariable(t *testing.T) {
	svc := newTestService(t)
	figure, err := svc.CreateFigure("undeclared variable")
	if err != nil {
		t.Fatalf("create figure: %v", err)
	}
	// 只声明 temp 变量；conc 未导入。
	if err := svc.ImportSemantics(figure.ID, ImportPayload{
		Layers:    []LayerInput{{Name: "scatter", LayerType: "scatter", ZOrder: 1, Visible: true}},
		Variables: []VariableInput{{Name: "temp", Unit: "K"}},
	}); err != nil {
		t.Fatalf("import: %v", err)
	}
	// temp 已声明 → 编码创建成功。
	if _, err := svc.DeclareEncoding(figure.ID, "", "temp", "color", "#1f77b4"); err != nil {
		t.Fatalf("declared variable should be accepted: %v", err)
	}
	// conc 未声明 → 编码创建应失败，且应映射为 ErrInvalidArgument/ErrNotFound。
	_, err = svc.DeclareEncoding(figure.ID, "", "conc", "color", "#1f77b4")
	if err == nil {
		t.Fatal("encoding for undeclared variable was accepted")
	}
	if !errors.Is(err, model.ErrNotFound) && !errors.Is(err, model.ErrInvalidArgument) {
		t.Fatalf("expected ErrNotFound or ErrInvalidArgument, got %v", err)
	}
}
