package service

import (
	"testing"

	"task251-chartambiguity/internal/model"
	"task251-chartambiguity/internal/store"
)

func TestTask251Bug09FingerprintTracksSemanticFields(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := New(st)
	figure, err := svc.CreateFigure("fingerprint")
	if err != nil {
		t.Fatal(err)
	}
	first := ImportPayload{
		Variables: []VariableInput{{Name: "temp"}, {Name: "conc"}},
		Legends:   []LegendInput{{Channel: model.ChannelColor, Label: "signal", Token: "blue", CoversVariable: "temp"}},
	}
	if err := svc.ImportSemantics(figure.ID, first); err != nil {
		t.Fatal(err)
	}
	before, err := svc.GetFigure(figure.ID)
	if err != nil {
		t.Fatal(err)
	}
	second := first
	second.Legends = []LegendInput{{Channel: model.ChannelColor, Label: "signal", Token: "blue", CoversVariable: "conc"}}
	if err := svc.ImportSemantics(figure.ID, second); err != nil {
		t.Fatal(err)
	}
	after, err := svc.GetFigure(figure.ID)
	if err != nil {
		t.Fatal(err)
	}
	if before.SourceFP == after.SourceFP {
		t.Fatalf("source fingerprint did not change: %s", after.SourceFP)
	}
	legends, err := st.ListLegends(figure.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, legend := range legends {
		if legend.CoversVariable == "conc" {
			found = true
		}
	}
	if !found {
		t.Fatalf("updated legend coverage was not persisted: %+v", legends)
	}
}
