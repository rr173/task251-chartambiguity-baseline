package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"task251-chartambiguity/internal/model"
	"task251-chartambiguity/internal/service"
	"task251-chartambiguity/internal/store"
)

func TestTask251Bug05FrozenFigureRejectsRecheck(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := service.New(st)
	figure, err := svc.CreateFigure("frozen")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ImportSemantics(figure.ID, service.ImportPayload{
		Variables: []service.VariableInput{{Name: "temp"}},
		Legends:   []service.LegendInput{{Channel: model.ChannelColor, Token: "blue", Label: "temp"}},
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
	if _, err := svc.PublishSpec(figure.ID, draft.ID); err != nil {
		t.Fatal(err)
	}
	before, _ := st.ListEncodings(figure.ID)
	server := httptest.NewServer(New(svc).Handler())
	defer server.Close()
	resp, err := http.Post(server.URL+"/api/figures/"+figure.ID+"/check", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("frozen recheck returned %d", resp.StatusCode)
	}
	after, _ := st.ListEncodings(figure.ID)
	if len(before) != len(after) || len(after) != 1 || after[0].Status != before[0].Status {
		t.Fatalf("encoding state changed after frozen recheck: before=%+v after=%+v", before, after)
	}
}
