package httpapi

import (
	"net/http"
)

// listLayers GET /api/figures/{id}/layers
func (a *API) listLayers(w http.ResponseWriter, r *http.Request) {
	ls, err := a.svc.Store().ListLayers(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"layers": ls, "count": len(ls)})
}

// listAxes GET /api/figures/{id}/axes
func (a *API) listAxes(w http.ResponseWriter, r *http.Request) {
	xs, err := a.svc.Store().ListAxes(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"axes": xs, "count": len(xs)})
}

// listLegends GET /api/figures/{id}/legends
func (a *API) listLegends(w http.ResponseWriter, r *http.Request) {
	gs, err := a.svc.Store().ListLegends(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"legends": gs, "count": len(gs)})
}

// listVariables GET /api/figures/{id}/variables
func (a *API) listVariables(w http.ResponseWriter, r *http.Request) {
	vs, err := a.svc.Store().ListVariables(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"variables": vs, "count": len(vs)})
}
