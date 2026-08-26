package httpapi

import (
	"net/http"

	"task251-chartambiguity/internal/model"
	"task251-chartambiguity/internal/service"
)

// runCheck POST /api/figures/{id}/check  重算映射与歧义，返回本次结果。
func (a *API) runCheck(w http.ResponseWriter, r *http.Request) {
	ambigs, err := a.svc.RunCheck(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	open := 0
	for _, x := range ambigs {
		if x.IsOpen() {
			open++
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ambiguities": ambigs,
		"total":       len(ambigs),
		"open":        open,
	})
}

// listAmbiguities GET /api/figures/{id}/ambiguities
func (a *API) listAmbiguities(w http.ResponseWriter, r *http.Request) {
	ambigs, err := a.svc.Store().ListAmbiguities(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ambiguities": ambigs, "count": len(ambigs)})
}

// listMappings GET /api/figures/{id}/mappings
func (a *API) listMappings(w http.ResponseWriter, r *http.Request) {
	ms, err := a.svc.Store().ListMappings(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"mappings": ms, "count": len(ms)})
}

// addException POST /api/figures/{id}/exceptions
// body: { "kind", "target_channel", "target_token", "target_variable", "reason" }
func (a *API) addException(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in service.ExceptionInput
	if err := decodeBody(r, &in); err != nil {
		writeError(w, model.ErrInvalidArgument)
		return
	}
	exc, err := a.svc.AddException(id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, exc)
}

// listExceptions GET /api/figures/{id}/exceptions
func (a *API) listExceptions(w http.ResponseWriter, r *http.Request) {
	es, err := a.svc.Store().ListExceptions(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"exceptions": es, "count": len(es)})
}
