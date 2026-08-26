package spec

import (
	"encoding/json"
	"testing"

	"task251-chartambiguity/internal/model"
)

func TestBuildSnapshotRoundTripsSemanticContent(t *testing.T) {
	snapshot, err := BuildSnapshot(
		"fig-1",
		[]model.VisualEncoding{{FigureID: "fig-1", Variable: "temp", Channel: model.ChannelColor, Token: "blue"}},
		[]model.Legend{{FigureID: "fig-1", Channel: model.ChannelColor, Token: "blue"}},
		[]model.Variable{{FigureID: "fig-1", Name: "temp", Unit: "K"}},
		[]model.Axis{{FigureID: "fig-1", Name: "y", Variable: "temp", Unit: "K"}},
		[]model.Exception{{FigureID: "fig-1", Kind: model.ExceptionReuse, Reason: "approved"}},
	)
	if err != nil {
		t.Fatalf("build snapshot: %v", err)
	}
	var decoded Snapshot
	if err := json.Unmarshal([]byte(snapshot), &decoded); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if decoded.FigureID != "fig-1" || len(decoded.Encodings) != 1 || len(decoded.Variables) != 1 || decoded.Variables[0].Unit != "K" {
		t.Fatalf("snapshot lost semantic content: %+v", decoded)
	}
}

func TestNextVersionAndCanPublish(t *testing.T) {
	if NextVersion(4) != 5 {
		t.Fatalf("unexpected next version: %d", NextVersion(4))
	}
	if !CanPublish(0) || CanPublish(1) {
		t.Fatal("publishability rule is incorrect")
	}
}
