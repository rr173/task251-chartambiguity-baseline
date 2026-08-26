package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"task251-chartambiguity/internal/model"
	"task251-chartambiguity/internal/service"
	"task251-chartambiguity/internal/store"
)

func TestTask251Bug07PublishRequiresCurrentDraft(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := service.New(st)
	figure, err := svc.CreateFigure("publish state")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ImportSemantics(figure.ID, service.ImportPayload{Variables: []service.VariableInput{{Name: "temp"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RunCheck(figure.ID); err != nil {
		t.Fatal(err)
	}
	draft, err := svc.CreateSpec(figure.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateSpecStatus(draft.ID, model.SpecStatusShared); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(New(svc).Handler())
	defer server.Close()
	resp, err := http.Post(server.URL+"/api/figures/"+figure.ID+"/specs/"+draft.ID+"/publish", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("publishing shared spec returned %d", resp.StatusCode)
	}
	got, err := st.GetSpec(draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.SpecStatusShared {
		t.Fatalf("shared spec changed to %s", got.Status)
	}
	current, err := st.GetFigure(figure.ID)
	if err != nil || current.Status != model.FigureStatusPublishable {
		t.Fatalf("figure state changed: %v %+v", err, current)
	}
}
