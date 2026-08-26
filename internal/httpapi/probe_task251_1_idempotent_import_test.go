package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"task251-chartambiguity/internal/service"
	"task251-chartambiguity/internal/store"
)

func TestTask251Bug01RepeatedImportIsIdempotent(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	server := httptest.NewServer(New(service.New(st)).Handler())
	defer server.Close()

	var figure struct{ ID string `json:"id"` }
	postProbeJSON(t, server.URL+"/api/figures", map[string]string{"title": "retry"}, &figure)
	payload := service.ImportPayload{
		Layers:    []service.LayerInput{{Name: "points", LayerType: "scatter", ZOrder: 1, Visible: true}},
		Axes:      []service.AxisInput{{Name: "x", Variable: "time", Unit: "s", Orientation: "x"}},
		Variables: []service.VariableInput{{Name: "time", Unit: "s"}},
		Legends:   []service.LegendInput{{Channel: "color", Label: "time", Token: "blue"}},
	}
	postProbeJSON(t, server.URL+"/api/figures/"+figure.ID+"/import", payload, nil)
	postProbeJSON(t, server.URL+"/api/figures/"+figure.ID+"/import", payload, nil)

	for _, resource := range []string{"layers", "axes", "variables", "legends"} {
		var response struct{ Count int `json:"count"` }
		getProbeJSON(t, server.URL+"/api/figures/"+figure.ID+"/"+resource, &response)
		if response.Count != 1 {
			t.Fatalf("%s count = %d, want 1", resource, response.Count)
		}
	}
}

func postProbeJSON(t *testing.T, url string, body interface{}, out interface{}) {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("POST %s returned %d", url, resp.StatusCode)
	}
	if out != nil && json.NewDecoder(resp.Body).Decode(out) != nil {
		t.Fatalf("decode POST %s response", url)
	}
}

func getProbeJSON(t *testing.T, url string, out interface{}) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned %d", url, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
}
