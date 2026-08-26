package service

import (
	"testing"

	"task251-chartambiguity/internal/model"
	"task251-chartambiguity/internal/store"
)

func TestTask251Bug04AmbiguityKeepsFigureScope(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := New(st)
	figure, err := svc.CreateFigure("scope")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ImportSemantics(figure.ID, ImportPayload{
		Variables: []VariableInput{{Name: "temp"}, {Name: "conc"}},
		Legends:   []LegendInput{{Channel: model.ChannelColor, Token: "blue", Label: "shared"}},
	}); err != nil {
		t.Fatal(err)
	}
	for _, variable := range []string{"temp", "conc"} {
		if _, err := svc.DeclareEncoding(figure.ID, "", variable, model.ChannelColor, "blue"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.RunCheck(figure.ID); err != nil {
		t.Fatal(err)
	}
	ambiguities, err := st.ListAmbiguities(figure.ID)
	if err != nil || len(ambiguities) == 0 {
		t.Fatalf("stored ambiguities = %v %+v", err, ambiguities)
	}
	open, err := st.CountOpenAmbiguities(figure.ID)
	if err != nil || open == 0 {
		t.Fatalf("open ambiguity count = %d (%v)", open, err)
	}
	if _, err := svc.CreateSpec(figure.ID); err == nil {
		t.Fatal("spec created while persisted ambiguity is open")
	}
}
