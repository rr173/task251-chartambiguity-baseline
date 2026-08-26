package httpapi

import (
	"net/http"
)

// createSpec POST /api/figures/{id}/specs  创建草稿（要求无未解决歧义）。
func (a *API) createSpec(w http.ResponseWriter, r *http.Request) {
	sp, err := a.svc.CreateSpec(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sp)
}

// publishSpec POST /api/figures/{id}/specs/{sid}/publish  冻结规范版本。
func (a *API) publishSpec(w http.ResponseWriter, r *http.Request) {
	sp, err := a.svc.PublishSpec(r.PathValue("id"), r.PathValue("sid"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sp)
}

// listSpecs GET /api/figures/{id}/specs
func (a *API) listSpecs(w http.ResponseWriter, r *http.Request) {
	ss, err := a.svc.Store().ListSpecs(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"specs": ss, "count": len(ss)})
}

// getSpec GET /api/specs/{sid}
func (a *API) getSpec(w http.ResponseWriter, r *http.Request) {
	sp, err := a.svc.Store().GetSpec(r.PathValue("sid"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sp)
}
