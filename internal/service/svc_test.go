package service

import (
	"testing"

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

func baseImportPayload() ImportPayload {
	return ImportPayload{
		Layers:    []LayerInput{{Name: "scatter", LayerType: "scatter", ZOrder: 1, Visible: true}},
		Axes:      []AxisInput{{Name: "x", Variable: "time", Unit: "s", Orientation: "x"}, {Name: "y", Variable: "temp", Unit: "K", Orientation: "y"}},
		Variables: []VariableInput{{Name: "temp", Unit: "K"}, {Name: "conc", Unit: "mol/L"}},
		Legends:   []LegendInput{{Channel: "color", Label: "temperature", Token: "#1f77b4", CoversVariable: "temp"}},
	}
}

// TestImportSemanticsIdempotentOnRetry 断言：客户端因重试提交两次完全相同的
// 语义导入请求时，相同语义只保留一套数据，不会产生重复的图层/轴/变量/图例。
func TestImportSemanticsIdempotentOnRetry(t *testing.T) {
	svc := newTestService(t)
	fig, err := svc.CreateFigure("retry figure")
	if err != nil {
		t.Fatalf("create figure: %v", err)
	}
	payload := baseImportPayload()
	if err := svc.ImportSemantics(fig.ID, payload); err != nil {
		t.Fatalf("first import: %v", err)
	}
	// 模拟客户端重试：完全相同的导入再次提交。
	if err := svc.ImportSemantics(fig.ID, payload); err != nil {
		t.Fatalf("retry import: %v", err)
	}

	layers, err := svc.store.ListLayers(fig.ID)
	if err != nil {
		t.Fatalf("list layers: %v", err)
	}
	axes, err := svc.store.ListAxes(fig.ID)
	if err != nil {
		t.Fatalf("list axes: %v", err)
	}
	vars, err := svc.store.ListVariables(fig.ID)
	if err != nil {
		t.Fatalf("list variables: %v", err)
	}
	legends, err := svc.store.ListLegends(fig.ID)
	if err != nil {
		t.Fatalf("list legends: %v", err)
	}
	if len(layers) != 1 || len(axes) != 2 || len(vars) != 2 || len(legends) != 1 {
		t.Fatalf("duplicate data after identical retry: layers=%d axes=%d vars=%d legends=%d",
			len(layers), len(axes), len(vars), len(legends))
	}
}

// TestImportSemanticsDistinctSemanticsNotShortCircuited 断言：语义不同的导入
// 不会因指纹短路而被忽略，正常追加产生新数据（导入不清理既有语义，故累加）。
func TestImportSemanticsDistinctSemanticsNotShortCircuited(t *testing.T) {
	svc := newTestService(t)
	fig, err := svc.CreateFigure("distinct figure")
	if err != nil {
		t.Fatalf("create figure: %v", err)
	}
	if err := svc.ImportSemantics(fig.ID, baseImportPayload()); err != nil {
		t.Fatalf("first import: %v", err)
	}
	// 改变语义：新增一个图层，指纹不同，应当正常写入。
	p2 := baseImportPayload()
	p2.Layers = append(p2.Layers, LayerInput{Name: "line", LayerType: "line", ZOrder: 2, Visible: true})
	if err := svc.ImportSemantics(fig.ID, p2); err != nil {
		t.Fatalf("second distinct import: %v", err)
	}
	layers, err := svc.store.ListLayers(fig.ID)
	if err != nil {
		t.Fatalf("list layers: %v", err)
	}
	// 首次导入 1 层 + 第二次新增 2 层（语义导入为追加语义，不清理既有项）。
	if len(layers) != 3 {
		t.Fatalf("distinct semantics should not be short-circuited, expected 3 layers, got %d", len(layers))
	}
}

