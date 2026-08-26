package mapping

import (
	"testing"

	"task251-chartambiguity/internal/model"
)

func TestBuildMappingsDetectsConflictAndSorts(t *testing.T) {
	encodings := []model.VisualEncoding{
		{FigureID: "fig-1", Variable: "z", Channel: model.ChannelShape, Token: "circle"},
		{FigureID: "fig-1", Variable: "a", Channel: model.ChannelColor, Token: "red"},
		{FigureID: "fig-1", Variable: "a", Channel: model.ChannelColor, Token: "blue"},
	}
	got := BuildMappings(encodings)
	if len(got) != 2 {
		t.Fatalf("expected two mappings, got %d", len(got))
	}
	if got[0].Variable != "a" || got[0].Decision != model.MappingDecisionConflict {
		t.Fatalf("expected sorted conflict first, got %+v", got[0])
	}
	if got[1].Variable != "z" || got[1].Decision != model.MappingDecisionConfirmed {
		t.Fatalf("expected confirmed mapping second, got %+v", got[1])
	}
	if got[0].FigureID != "fig-1" || got[0].Note == "" {
		t.Fatalf("mapping lost figure or conflict note: %+v", got[0])
	}
}
