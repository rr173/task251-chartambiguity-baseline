package layer

import (
	"testing"

	"task251-chartambiguity/internal/model"
)

func TestValidateLayerAndSummary(t *testing.T) {
	if err := ValidateLayer(model.Layer{Name: "points", LayerType: "scatter", ZOrder: 1}); err != nil {
		t.Fatalf("valid layer rejected: %v", err)
	}
	if err := ValidateLayer(model.Layer{Name: "", LayerType: "scatter"}); err == nil {
		t.Fatal("empty layer name accepted")
	}

	summary := ComputeLayerSummary([]model.Layer{
		{Name: "background", LayerType: "line", ZOrder: 0, Visible: false},
		{Name: "points", LayerType: "scatter", ZOrder: 2, Visible: true},
		{Name: "labels", LayerType: "text", ZOrder: 3, Visible: true},
	})
	if summary.Total != 3 || summary.Visible != 2 || summary.MaxZOrder != 3 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.ByType["scatter"] != 1 || summary.ByType["text"] != 1 {
		t.Fatalf("unexpected type counts: %+v", summary.ByType)
	}
}

func TestValidateAxisLegendVariable(t *testing.T) {
	if err := ValidateAxis(model.Axis{Name: "x", Variable: "time", Orientation: "x"}); err != nil {
		t.Fatalf("valid axis rejected: %v", err)
	}
	if err := ValidateLegend(model.Legend{Channel: model.ChannelColor, Token: "#1f77b4"}); err != nil {
		t.Fatalf("valid legend rejected: %v", err)
	}
	if err := ValidateVariable(model.Variable{Name: "temperature"}); err != nil {
		t.Fatalf("valid variable rejected: %v", err)
	}
	if err := ValidateAxis(model.Axis{Name: "x", Variable: "time", Orientation: "diagonal"}); err == nil {
		t.Fatal("invalid axis orientation accepted")
	}
}
