package httpapi

import (
	"net/http"

	"task251-chartambiguity/internal/model"
)

// declareEncoding POST /api/figures/{id}/encodings
// body: { "layer_id", "variable", "channel", "token" }
func (a *API) declareEncoding(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		LayerID  string `json:"layer_id"`
		Variable string `json:"variable"`
		Channel  string `json:"channel"`
		Token    string `json:"token"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, model.ErrInvalidArgument)
		return
	}
	e, err := a.svc.DeclareEncoding(id, body.LayerID, body.Variable, body.Channel, body.Token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

// listEncodings GET /api/figures/{id}/encodings
func (a *API) listEncodings(w http.ResponseWriter, r *http.Request) {
	es, err := a.svc.Store().ListEncodings(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"encodings": es, "count": len(es)})
}
