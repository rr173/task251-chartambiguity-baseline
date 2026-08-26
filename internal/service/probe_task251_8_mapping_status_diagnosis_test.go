package service

import (
	"testing"

	"task251-chartambiguity/internal/model"
	"task251-chartambiguity/internal/store"
)

func TestTask251Bug08MappingConflictStatus(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := New(st)
	figure, err := svc.CreateFigure("mapping status")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ImportSemantics(figure.ID, ImportPayload{
		Variables: []VariableInput{{Name: "temp"}},
		Legends: []LegendInput{
			{Channel: model.ChannelColor, Token: "blue", Label: "blue"},
			{Channel: model.ChannelColor, Token: "red", Label: "red"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"blue", "red"} {
		if _, err := svc.DeclareEncoding(figure.ID, "", "temp", model.ChannelColor, token); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.RunCheck(figure.ID); err != nil {
		t.Fatal(err)
	}
	encodings, err := st.ListEncodings(figure.ID)
	if err != nil || len(encodings) != 2 {
		t.Fatalf("encodings = %v %+v", err, encodings)
	}
	for _, encoding := range encodings {
		if encoding.Status != model.EncodingStatusAmbiguous {
			t.Fatalf("encoding %s status = %s, want ambiguous", encoding.Token, encoding.Status)
		}
	}
}
