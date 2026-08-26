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
	got := CheckAll(encs, nil, legends, nil, excs)
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

// TestCheckAllReuseExemptionTokenScoped 验证复用豁免按 channel+token 精确命中：
// 同一颜色通道上两个不同 token 各自存在复用歧义时，只豁免其中之一，
// 不得把另一个 token 的歧义也连带标成已解决。
func TestCheckAllReuseExemptionTokenScoped(t *testing.T) {
	encs := []model.VisualEncoding{
		{FigureID: "f", Variable: "temp", Channel: model.ChannelColor, Token: "#1f77b4"},
		{FigureID: "f", Variable: "conc", Channel: model.ChannelColor, Token: "#1f77b4"},
		{FigureID: "f", Variable: "pres", Channel: model.ChannelColor, Token: "#ff7f0e"},
		{FigureID: "f", Variable: "flux", Channel: model.ChannelColor, Token: "#ff7f0e"},
	}
	// 两个 token 都有图例覆盖，避免缺图例干扰，只看颜色复用豁免。
	legends := []model.Legend{
		{FigureID: "f", Channel: model.ChannelColor, Token: "#1f77b4", CoversVariable: "temp"},
		{FigureID: "f", Channel: model.ChannelColor, Token: "#ff7f0e", CoversVariable: "pres"},
	}
	// 只豁免 #1f77b4 的复用，#ff7f0e 的复用歧义必须保持未解决。
	excs := []model.Exception{{
		ID: "e1", FigureID: "f", Kind: model.ExceptionReuse,
		TargetChannel: model.ChannelColor, TargetToken: "#1f77b4",
	}}
	got := CheckAll(encs, nil, legends, nil, excs)
	if len(got) != 2 {
		t.Fatalf("expected 2 color reuse ambiguities, got %d", len(got))
	}
	var resolved, unresolved int
	for _, a := range got {
		switch a.Token {
		case "#1f77b4":
			if !a.Resolved || a.ExceptionID != "e1" {
				t.Fatalf("token #1f77b4 应被豁免: resolved=%v id=%s", a.Resolved, a.ExceptionID)
			}
			resolved++
		case "#ff7f0e":
			if a.Resolved {
				t.Fatalf("token #ff7f0e 不应被连带豁免，但被判为已解决")
			}
			unresolved++
		default:
			t.Fatalf("unexpected token: %s", a.Token)
		}
	}
	if resolved != 1 || unresolved != 1 {
		t.Fatalf("expected 1 resolved and 1 unresolved, got %d/%d", resolved, unresolved)
	}
}
