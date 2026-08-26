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

func TestTask251Bug03EncodingRejectsForeignVariable(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	server := httptest.NewServer(New(service.New(st)).Handler())
	defer server.Close()

	ids := make([]string, 2)
	for i, title := range []string{"owner", "consumer"} {
		var figure struct{ ID string `json:"id"` }
		postProbe3(t, server.URL+"/api/figures", map[string]string{"title": title}, &figure)
		ids[i] = figure.ID
	}
	postProbe3(t, server.URL+"/api/figures/"+ids[0]+"/import", service.ImportPayload{
		Variables: []service.VariableInput{{Name: "temperature", Unit: "K"}},
	}, nil)

	body, _ := json.Marshal(map[string]string{"variable": "temperature", "channel": "color", "token": "blue"})
	resp, err := http.Post(server.URL+"/api/figures/"+ids[1]+"/encodings", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode < 400 {
		t.Fatalf("foreign variable encoding returned %d", resp.StatusCode)
	}
	var list struct{ Count int `json:"count"` }
	getProbe3(t, server.URL+"/api/figures/"+ids[1]+"/encodings", &list)
	if list.Count != 0 {
		t.Fatalf("foreign variable encoding persisted, count=%d", list.Count)
	}
}

func postProbe3(t *testing.T, url string, body interface{}, out interface{}) {
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
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatal(err)
		}
	}
}

func getProbe3(t *testing.T, url string, out interface{}) {
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
