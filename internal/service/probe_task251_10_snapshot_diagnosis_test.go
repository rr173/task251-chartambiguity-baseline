package service

import (
	"strings"
	"testing"

	"task251-chartambiguity/internal/model"
	"task251-chartambiguity/internal/store"
)

func TestTask251Bug10PublishedSnapshotUsesCurrentSemantics(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := New(st)
	figure, err := svc.CreateFigure("snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ImportSemantics(figure.ID, ImportPayload{
		Variables: []VariableInput{{Name: "temp"}},
		Legends:   []LegendInput{{Channel: model.ChannelColor, Label: "temp", Token: "blue"}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DeclareEncoding(figure.ID, "", "temp", model.ChannelColor, "blue"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RunCheck(figure.ID); err != nil {
		t.Fatal(err)
	}
	draft, err := svc.CreateSpec(figure.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddException(figure.ID, ExceptionInput{
		Kind: model.ExceptionReuse, TargetChannel: model.ChannelColor, TargetToken: "blue", Reason: "reviewed-after-draft",
	}); err != nil {
		t.Fatal(err)
	}
	published, err := svc.PublishSpec(figure.ID, draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(published.Snapshot, "reviewed-after-draft") {
		t.Fatalf("published snapshot omitted latest exception: %s", published.Snapshot)
	}
}
