package service

import (
	"testing"

	"task251-chartambiguity/internal/model"
)

// TestImportSemanticsAllOrNothingOnBadLegend 验证整批导入的一致性：
// 当批次中后面的图例格式错误时，导入请求应当整体失败，
// 且前面已校验通过的图层、坐标轴、变量、图例都不得落库，
// 避免请求失败后图形稿中残留半套语义。
func TestImportSemanticsAllOrNothingOnBadLegend(t *testing.T) {
	svc := newTestService(t)
	figure, err := svc.CreateFigure("bad legend rollback")
	if err != nil {
		t.Fatalf("create figure: %v", err)
	}

	// 第三条图例 channel 非法，应触发整批失败。
	payload := ImportPayload{
		Layers: []LayerInput{
			{Name: "scatter", LayerType: "scatter", ZOrder: 1, Visible: true},
		},
		Axes: []AxisInput{
			{Name: "x", Variable: "time", Unit: "s", Orientation: "x"},
		},
		Variables: []VariableInput{
			{Name: "temp", Unit: "K"},
		},
		Legends: []LegendInput{
			{Channel: "color", Label: "temperature", Token: "#1f77b4", CoversVariable: "temp"},
			{Channel: "color", Label: "concentration", Token: "#ff7f0e", CoversVariable: "conc"},
			{Channel: "not-a-channel", Label: "broken", Token: "#000000", CoversVariable: "conc"},
		},
	}

	if err := svc.ImportSemantics(figure.ID, payload); err == nil {
		t.Fatal("import with invalid legend accepted; want error")
	}

	// 请求失败后，图层/轴/变量/图例均不得残留在图形稿中。
	st := svc.Store()
	if ls, err := st.ListLayers(figure.ID); err != nil || len(ls) != 0 {
		t.Errorf("layers leaked after failed import: len=%d err=%v", len(ls), err)
	}
	if xs, err := st.ListAxes(figure.ID); err != nil || len(xs) != 0 {
		t.Errorf("axes leaked after failed import: len=%d err=%v", len(xs), err)
	}
	if vs, err := st.ListVariables(figure.ID); err != nil || len(vs) != 0 {
		t.Errorf("variables leaked after failed import: len=%d err=%v", len(vs), err)
	}
	if gs, err := st.ListLegends(figure.ID); err != nil || len(gs) != 0 {
		t.Errorf("legends leaked after failed import: len=%d err=%v", len(gs), err)
	}

	// 图形稿指纹与状态不得被推进，便于用户修正后整批重导。
	f, err := svc.GetFigure(figure.ID)
	if err != nil {
		t.Fatalf("get figure: %v", err)
	}
	if f.SourceFP != "" {
		t.Errorf("source_fp advanced after failed import: %q", f.SourceFP)
	}
	if f.Status != model.FigureStatusImporting {
		t.Errorf("status advanced after failed import: %q", f.Status)
	}
}

// TestImportSemanticsRejectsEmptyLegendToken 确认图例 token 缺失在写入前被拦下。
func TestImportSemanticsRejectsEmptyLegendToken(t *testing.T) {
	svc := newTestService(t)
	figure, err := svc.CreateFigure("empty legend token")
	if err != nil {
		t.Fatalf("create figure: %v", err)
	}
	payload := ImportPayload{
		Layers:  []LayerInput{{Name: "scatter", LayerType: "scatter", ZOrder: 1, Visible: true}},
		Legends: []LegendInput{{Channel: "color", Label: "temperature", Token: ""}},
	}
	if err := svc.ImportSemantics(figure.ID, payload); err == nil {
		t.Fatal("import with empty legend token accepted; want error")
	}
	if ls, err := svc.Store().ListLayers(figure.ID); err != nil || len(ls) != 0 {
		t.Errorf("layers leaked after failed import: len=%d err=%v", len(ls), err)
	}
}

// TestImportSemanticsAcceptsValidBatch 确认合法整批仍能完整写入，回归保护。
func TestImportSemanticsAcceptsValidBatch(t *testing.T) {
	svc := newTestService(t)
	figure, err := svc.CreateFigure("valid batch")
	if err != nil {
		t.Fatalf("create figure: %v", err)
	}
	payload := ImportPayload{
		Layers:    []LayerInput{{Name: "scatter", LayerType: "scatter", ZOrder: 1, Visible: true}},
		Axes:      []AxisInput{{Name: "x", Variable: "time", Unit: "s", Orientation: "x"}},
		Variables: []VariableInput{{Name: "temp", Unit: "K"}},
		Legends:   []LegendInput{{Channel: "color", Label: "temperature", Token: "#1f77b4", CoversVariable: "temp"}},
	}
	if err := svc.ImportSemantics(figure.ID, payload); err != nil {
		t.Fatalf("import valid batch: %v", err)
	}
	f, err := svc.GetFigure(figure.ID)
	if err != nil {
		t.Fatalf("get figure: %v", err)
	}
	if f.Status != model.FigureStatusPendingReview {
		t.Errorf("status=%q, want pending_review", f.Status)
	}
	if f.SourceFP == "" {
		t.Errorf("source_fp not set after valid import")
	}
	if f.LayerCount != 1 {
		t.Errorf("layer_count=%d, want 1", f.LayerCount)
	}
	if gs, err := svc.Store().ListLegends(figure.ID); err != nil || len(gs) != 1 {
		t.Errorf("legends not persisted on valid import: len=%d err=%v", len(gs), err)
	}
}
