package httpapi

import "net/http"

// selfCheck GET /api/selfcheck  返回服务健康与规模指标。
func (a *API) selfCheck(w http.ResponseWriter, r *http.Request) {
	res, err := a.svc.SelfCheck()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}
