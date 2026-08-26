package service

import (
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

// buildColorReuseFigure 构造一篇存在颜色复用歧义的图形稿并执行复核，
// 返回图形稿 ID（状态为 pending_review，落库歧义归属正确）。
func buildColorReuseFigure(t *testing.T, svc *Service) string {
	t.Helper()
	fig, err := svc.CreateFigure("颜色复用复核归属")
	if err != nil {
		t.Fatalf("create figure: %v", err)
	}
	payload := ImportPayload{
		Axes:      []AxisInput{{Name: "y", Variable: "temp", Unit: "K", Orientation: "y"}},
		Variables: []VariableInput{{Name: "temp", Unit: "K"}, {Name: "conc", Unit: "mol/L"}},
		// 图例覆盖 #1f77b4 以排除缺图例歧义，聚焦颜色复用。
		Legends: []LegendInput{{Channel: "color", Label: "t", Token: "#1f77b4", CoversVariable: "temp"}},
	}
	if err := svc.ImportSemantics(fig.ID, payload); err != nil {
		t.Fatalf("import: %v", err)
	}
	if _, err := svc.DeclareEncoding(fig.ID, "", "temp", "color", "#1f77b4"); err != nil {
		t.Fatalf("encode temp: %v", err)
	}
	if _, err := svc.DeclareEncoding(fig.ID, "", "conc", "color", "#1f77b4"); err != nil {
		t.Fatalf("encode conc: %v", err)
	}
	if _, err := svc.RunCheck(fig.ID); err != nil {
		t.Fatalf("check: %v", err)
	}
	return fig.ID
}

// TestCheckReportedAmbiguityIsQueryable 验证复核接口报告的颜色复用歧义
// 在重新查询歧义列表时仍可检索到（归属到正确 figure_id）。修复前歧义
// 落库时 figure_id 为空，导致按 figure_id 查询返回空列表。
func TestCheckReportedAmbiguityIsQueryable(t *testing.T) {
	svc := newTestService(t)
	fid := buildColorReuseFigure(t, svc)

	stored, err := svc.store.ListAmbiguities(fid)
	if err != nil {
		t.Fatalf("list ambiguities: %v", err)
	}
	var reuse *model.Ambiguity
	for i := range stored {
		if stored[i].Type == model.AmbiguityColorReuse {
			reuse = &stored[i]
		}
	}
	if reuse == nil {
		t.Fatalf("expected stored color_reuse ambiguity for figure, got %v", stored)
	}
	if reuse.FigureID != fid {
		t.Fatalf("stored ambiguity not attributed to figure: got %q want %q", reuse.FigureID, fid)
	}
	if reuse.Resolved {
		t.Fatalf("expected color reuse to remain open; got resolved=%v", reuse.Resolved)
	}
}

// TestPublishGateBlocksSpecCreationWithOpenAmbiguity 验证即使图形稿状态
// 被人为置为 publishable（模拟过期状态），发布门禁仍依据未解决歧义计数
// 阻止创建规范草稿。修复前 CountOpenAmbiguities 因归属丢失而返回 0，门禁
// 被错误放行，从而允许在存在开放歧义时创建规范草稿。
func TestPublishGateBlocksSpecCreationWithOpenAmbiguity(t *testing.T) {
	svc := newTestService(t)
	fid := buildColorReuseFigure(t, svc)

	// 人为把状态推进到 publishable，模拟「上次复核通过后新增编码未重算」的过期状态，
	// 以单独检验发布门禁是否依据真实的未解决歧义计数而非仅依赖状态。
	if err := svc.store.UpdateFigureStatus(fid, model.FigureStatusPublishable); err != nil {
		t.Fatalf("force publishable: %v", err)
	}

	if _, err := svc.CreateSpec(fid); err == nil {
		t.Fatal("expected spec creation to be blocked by open ambiguity, but it was allowed")
	}
	// 同样校验冻结门禁。
	// （此处无 draft 规范，故期望在更早的状态/计数校验处失败，确保不放行。）
}

// TestPublishGateBlocksPublishWithOpenAmbiguity 单独验证 PublishSpec 门禁：
// 即便存在 draft 规范且图形稿状态被人为置为 publishable，只要未解决歧义计数 > 0
// 就禁止冻结发布。修复前 CountOpenAmbiguities 因归属丢失返回 0，门禁被放行。
func TestPublishGateBlocksPublishWithOpenAmbiguity(t *testing.T) {
	svc := newTestService(t)
	fid := buildColorReuseFigure(t, svc)

	// 豁免颜色复用使状态进入 publishable 并创建一份 draft 规范。
	if _, err := svc.AddException(fid, ExceptionInput{
		Kind:          model.ExceptionReuse,
		TargetChannel: "color",
		TargetToken:   "#1f77b4",
		Reason:        "approved",
	}); err != nil {
		t.Fatalf("add exception: %v", err)
	}
	spec, err := svc.CreateSpec(fid)
	if err != nil {
		t.Fatalf("create spec after exemption: %v", err)
	}
	// 直接落库一条归属正确的开放歧义（模拟重算后新增的未豁免歧义），强制状态为
	// publishable，以单独检验 PublishSpec 门禁依据真实计数而非仅依赖状态。
	amb := model.Ambiguity{
		ID:          "amb-open-1",
		FigureID:    fid,
		Type:        model.AmbiguityAxisUnitConflict,
		Severity:    model.AmbiguitySeverityError,
		Variables:   "temp",
		Description: "人工注入的开放歧义以检验发布门禁",
		Resolved:    false,
	}
	if err := svc.store.InsertAmbiguities([]model.Ambiguity{amb}); err != nil {
		t.Fatalf("insert open ambiguity: %v", err)
	}
	if err := svc.store.UpdateFigureStatus(fid, model.FigureStatusPublishable); err != nil {
		t.Fatalf("force publishable: %v", err)
	}
	if _, err := svc.PublishSpec(fid, spec.ID); err == nil {
		t.Fatal("expected publish to be blocked by open ambiguity, but it was allowed")
	}
}
