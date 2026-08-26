package checker

import (
	"testing"

	"task251-chartambiguity/internal/model"
)

func TestDetectColorReuse(t *testing.T) {
	encs := []model.VisualEncoding{
		{FigureID: "f", Variable: "temp", Channel: model.ChannelColor, Token: "#1f77b4"},
		{FigureID: "f", Variable: "conc", Channel: model.ChannelColor, Token: "#1f77b4"},
		{FigureID: "f", Variable: "pres", Channel: model.ChannelColor, Token: "#ff7f0e"},
	}
	got := DetectColorReuse(encs)
	if len(got) != 1 {
		t.Fatalf("expected 1 color reuse ambiguity, got %d", len(got))
	}
	if got[0].Variables != "conc,temp" && got[0].Variables != "temp,conc" {
		t.Fatalf("unexpected variables: %s", got[0].Variables)
	}
}

func TestDetectMissingLegend(t *testing.T) {
	encs := []model.VisualEncoding{
		{FigureID: "f", Variable: "temp", Channel: model.ChannelColor, Token: "#1f77b4"},
	}
	// 图例只覆盖 #ff7f0e，#1f77b4 缺图例。
	legends := []model.Legend{{FigureID: "f", Channel: model.ChannelColor, Token: "#ff7f0e", CoversVariable: "pres"}}
	got := DetectMissingLegend(encs, legends)
	if len(got) != 1 || got[0].Type != model.AmbiguityMissingLegend {
		t.Fatalf("expected 1 missing_legend ambiguity, got %+v", got)
	}
}

func TestDetectAxisUnitConflict(t *testing.T) {
	axes := []model.Axis{
		{FigureID: "f", Variable: "temp", Unit: "K"},
		{FigureID: "f", Variable: "temp", Unit: "C"},
	}
	got := DetectAxisUnitConflict(axes)
	if len(got) != 1 || got[0].Type != model.AmbiguityAxisUnitConflict {
		t.Fatalf("expected 1 axis_unit_conflict, got %+v", got)
	}
}

func TestCheckAllExemption(t *testing.T) {
	encs := []model.VisualEncoding{
		{FigureID: "f", Variable: "temp", Channel: model.ChannelColor, Token: "#1f77b4"},
		{FigureID: "f", Variable: "conc", Channel: model.ChannelColor, Token: "#1f77b4"},
	}
	// 图例覆盖 #1f77b4，避免触发缺图例歧义，仅验证颜色复用豁免。
	legends := []model.Legend{{FigureID: "f", Channel: model.ChannelColor, Token: "#1f77b4", CoversVariable: "temp"}}
	excs := []model.Exception{{
		ID: "e1", FigureID: "f", Kind: model.ExceptionReuse,
		TargetChannel: model.ChannelColor, TargetToken: "#1f77b4",
	}}
	got := CheckAll("f", encs, nil, legends, nil, excs)
	if len(got) != 1 {
		t.Fatalf("expected 1 ambiguity, got %d", len(got))
	}
	if !got[0].Resolved {
		t.Fatalf("expected ambiguity to be resolved by exception")
	}
	if got[0].ExceptionID != "e1" {
		t.Fatalf("expected exception id e1, got %s", got[0].ExceptionID)
	}
}

// TestCheckAllAttributesFigureID 验证 CheckAll 把检测出的歧义归属到传入的图形稿，
// 否则落库后按 figure_id 查询或统计未解决歧义都会误判为空，导致复核结果丢失与
// 发布门禁被错误放行。
func TestCheckAllAttributesFigureID(t *testing.T) {
	encs := []model.VisualEncoding{
		{FigureID: "f", Variable: "temp", Channel: model.ChannelColor, Token: "#1f77b4"},
		{FigureID: "f", Variable: "conc", Channel: model.ChannelColor, Token: "#1f77b4"},
	}
	got := CheckAll("figure-xyz", encs, nil, nil, nil, nil)
	if len(got) == 0 {
		t.Fatal("expected ambiguities to be detected")
	}
	for _, a := range got {
		if a.FigureID != "figure-xyz" {
			t.Fatalf("ambiguity not attributed to figure: got FigureID=%q", a.FigureID)
		}
	}
}
