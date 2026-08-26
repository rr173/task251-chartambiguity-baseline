package checker

import (
	"testing"

	"task251-chartambiguity/internal/model"
)

func TestTask251Bug06ExceptionMatchesExactToken(t *testing.T) {
	encodings := []model.VisualEncoding{
		{FigureID: "fig-1", Variable: "temp", Channel: model.ChannelColor, Token: "blue"},
		{FigureID: "fig-1", Variable: "conc", Channel: model.ChannelColor, Token: "blue"},
		{FigureID: "fig-1", Variable: "pressure", Channel: model.ChannelColor, Token: "red"},
		{FigureID: "fig-1", Variable: "flow", Channel: model.ChannelColor, Token: "red"},
	}
	legends := []model.Legend{
		{FigureID: "fig-1", Channel: model.ChannelColor, Token: "blue"},
		{FigureID: "fig-1", Channel: model.ChannelColor, Token: "red"},
	}
	exceptions := []model.Exception{{
		ID: "exc-blue", FigureID: "fig-1", Kind: model.ExceptionReuse,
		TargetChannel: model.ChannelColor, TargetToken: "blue",
	}}
	got := CheckAll(encodings, nil, legends, nil, exceptions)
	if len(got) != 2 {
		t.Fatalf("ambiguity count = %d, want 2", len(got))
	}
	resolved := 0
	for _, ambiguity := range got {
		if ambiguity.Token == "blue" && !ambiguity.Resolved {
			t.Fatalf("blue ambiguity was not resolved: %+v", ambiguity)
		}
		if ambiguity.Token == "red" && ambiguity.Resolved {
			t.Fatalf("red ambiguity was incorrectly resolved: %+v", ambiguity)
		}
		if ambiguity.Resolved {
			resolved++
		}
	}
	if resolved != 1 {
		t.Fatalf("resolved count = %d, want 1", resolved)
	}
}
