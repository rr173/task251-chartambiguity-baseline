package httpapi

import (
	"net/http"
	"strconv"

	"task251-chartambiguity/internal/model"
	"task251-chartambiguity/internal/service"
)

// createFigure POST /api/figures  { "title": "..." }
func (a *API) createFigure(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeError(w, model.ErrInvalidArgument)
		return
	}
	f, err := a.svc.CreateFigure(body.Title)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

// listFigures GET /api/figures?limit=20
func (a *API) listFigures(w http.ResponseWriter, r *http.Request) {
	limit := 0
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
		}
	}
	figs, err := a.svc.ListFigures(limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"figures": figs, "count": len(figs)})
}

// getFigure GET /api/figures/{id}
func (a *API) getFigure(w http.ResponseWriter, r *http.Request) {
	f, err := a.svc.GetFigure(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

// figureStatus GET /api/figures/{id}/status
func (a *API) figureStatus(w http.ResponseWriter, r *http.Request) {
	f, err := a.svc.GetFigure(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"id": f.ID, "status": f.Status, "source_fp": f.SourceFP})
}

// importSemantics POST /api/figures/{id}/import
func (a *API) importSemantics(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var p service.ImportPayload
	if err := decodeBody(r, &p); err != nil {
		writeError(w, model.ErrInvalidArgument)
		return
	}
	if err := a.svc.ImportSemantics(id, p); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "figure_id": id})
}

// figureSummary GET /api/figures/{id}/summary
func (a *API) figureSummary(w http.ResponseWriter, r *http.Request) {
	s, err := a.svc.Summary(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s)
}
