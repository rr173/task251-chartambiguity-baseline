package service

import (
	"testing"

	"task251-chartambiguity/internal/model"
	"task251-chartambiguity/internal/store"
)

func TestTask251Bug02ImportIsAtomic(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := New(st)
	figure, err := svc.CreateFigure("atomic")
	if err != nil {
		t.Fatal(err)
	}
	err = svc.ImportSemantics(figure.ID, ImportPayload{
		Layers: []LayerInput{{Name: "points", LayerType: model.LayerTypeScatter, ZOrder: 1, Visible: true}},
		Axes:   []AxisInput{{Name: "x", Variable: "time", Unit: "s", Orientation: model.AxisOrientationX}},
		Legends: []LegendInput{{Channel: "invalid", Token: "blue", Label: "bad"}},
	})
	if err == nil {
		t.Fatal("invalid import succeeded")
	}
	if layers, _ := st.ListLayers(figure.ID); len(layers) != 0 {
		t.Fatalf("partial layers persisted: %+v", layers)
	}
	if axes, _ := st.ListAxes(figure.ID); len(axes) != 0 {
		t.Fatalf("partial axes persisted: %+v", axes)
	}
	if legends, _ := st.ListLegends(figure.ID); len(legends) != 0 {
		t.Fatalf("partial legends persisted: %+v", legends)
	}
}
