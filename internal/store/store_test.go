package store

import (
	"testing"

	"task251-chartambiguity/internal/model"
)

func TestFigureAndSemanticPersistence(t *testing.T) {
	st, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	figure := &model.Figure{ID: "fig-1", Title: "test", Status: model.FigureStatusImporting}
	if err := st.CreateFigure(figure); err != nil {
		t.Fatalf("create figure: %v", err)
	}
	layer := &model.Layer{ID: "layer-1", FigureID: figure.ID, Name: "points", LayerType: "scatter", ZOrder: 1, Visible: true}
	if err := st.CreateLayer(layer); err != nil {
		t.Fatalf("create layer: %v", err)
	}
	encoding := &model.VisualEncoding{ID: "enc-1", FigureID: figure.ID, Variable: "temp", Channel: model.ChannelColor, Token: "blue", Status: model.EncodingStatusParsed}
	if err := st.CreateEncoding(encoding); err != nil {
		t.Fatalf("create encoding: %v", err)
	}
	if err := st.SetFigureSourceFP(figure.ID, "fp-1", 1); err != nil {
		t.Fatalf("set source fingerprint: %v", err)
	}
	if err := st.UpdateFigureStatus(figure.ID, model.FigureStatusPendingReview); err != nil {
		t.Fatalf("update figure status: %v", err)
	}

	got, err := st.GetFigure(figure.ID)
	if err != nil {
		t.Fatalf("get figure: %v", err)
	}
	if got.Status != model.FigureStatusPendingReview || got.SourceFP != "fp-1" || got.LayerCount != 1 {
		t.Fatalf("figure persistence mismatch: %+v", got)
	}
	encodings, err := st.ListEncodings(figure.ID)
	if err != nil || len(encodings) != 1 || encodings[0].Token != "blue" {
		t.Fatalf("encoding persistence mismatch: %v %+v", err, encodings)
	}
}

func TestMissingFigureReturnsDomainError(t *testing.T) {
	st, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if _, err := st.GetFigure("does-not-exist"); err != model.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
