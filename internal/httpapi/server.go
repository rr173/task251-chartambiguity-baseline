// Package httpapi 暴露基于 net/http 的 JSON API，路由统一以 /api 前缀开头。
// 本包把 service 的高层操作映射为 HTTP 端点，负责请求解析、状态码映射与响应序列化。
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"task251-chartambiguity/internal/model"
	"task251-chartambiguity/internal/service"
)

// API 持有所依赖的 service。
type API struct {
	svc *service.Service
}

// New 构造 API 并注册全部路由。
func New(svc *service.Service) *API { return &API{svc: svc} }

// Handler 返回已注册全部 /api 路由的 http.Handler。
func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/figures", a.createFigure)
	mux.HandleFunc("GET /api/figures", a.listFigures)
	mux.HandleFunc("GET /api/figures/{id}", a.getFigure)
	mux.HandleFunc("GET /api/figures/{id}/status", a.figureStatus)
	mux.HandleFunc("POST /api/figures/{id}/import", a.importSemantics)
	mux.HandleFunc("GET /api/figures/{id}/summary", a.figureSummary)

	mux.HandleFunc("GET /api/figures/{id}/layers", a.listLayers)
	mux.HandleFunc("GET /api/figures/{id}/axes", a.listAxes)
	mux.HandleFunc("GET /api/figures/{id}/legends", a.listLegends)
	mux.HandleFunc("GET /api/figures/{id}/variables", a.listVariables)

	mux.HandleFunc("POST /api/figures/{id}/encodings", a.declareEncoding)
	mux.HandleFunc("GET /api/figures/{id}/encodings", a.listEncodings)

	mux.HandleFunc("POST /api/figures/{id}/check", a.runCheck)
	mux.HandleFunc("GET /api/figures/{id}/ambiguities", a.listAmbiguities)
	mux.HandleFunc("GET /api/figures/{id}/mappings", a.listMappings)

	mux.HandleFunc("POST /api/figures/{id}/exceptions", a.addException)
	mux.HandleFunc("GET /api/figures/{id}/exceptions", a.listExceptions)

	mux.HandleFunc("POST /api/figures/{id}/specs", a.createSpec)
	mux.HandleFunc("POST /api/figures/{id}/specs/{sid}/publish", a.publishSpec)
	mux.HandleFunc("GET /api/figures/{id}/specs", a.listSpecs)
	mux.HandleFunc("GET /api/specs/{sid}", a.getSpec)

	mux.HandleFunc("GET /api/selfcheck", a.selfCheck)
	return mux
}

// writeJSON 序列化 v 为 JSON 并设置 Content-Type。
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError 将领域错误映射为合适的 HTTP 状态码并输出错误体。
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrInvalidArgument):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrInvalidStatus):
		status = http.StatusConflict
	case model.IsConflict(err):
		status = http.StatusConflict
	}
	writeJSON(w, status, map[string]interface{}{"error": err.Error(), "status": status})
}

// decodeBody 解析请求体为 v。
func decodeBody(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return model.ErrInvalidArgument
	}
	dec := json.NewDecoder(r.Body)
	return dec.Decode(v)
}
