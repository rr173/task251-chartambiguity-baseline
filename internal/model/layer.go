package model

// 图层类型枚举，用于语义校验与图层摘要统计。
const (
	LayerTypeScatter = "scatter"
	LayerTypeLine    = "line"
	LayerTypeBar     = "bar"
	LayerTypeArea    = "area"
	LayerTypeHeatmap = "heatmap"
	LayerTypeText    = "text"
)

var validLayerType = map[string]bool{
	LayerTypeScatter: true,
	LayerTypeLine:    true,
	LayerTypeBar:     true,
	LayerTypeArea:    true,
	LayerTypeHeatmap: true,
	LayerTypeText:    true,
}

// IsValidLayerType 校验图层类型是否合法。
func IsValidLayerType(t string) bool { return validLayerType[t] }

// Layer 表示图表中的一个可视图层。
type Layer struct {
	ID        string `json:"id"`
	FigureID  string `json:"figure_id"`
	Name      string `json:"name"`
	LayerType string `json:"layer_type"`
	ZOrder    int    `json:"z_order"`
	Visible   bool   `json:"visible"`
}

// LayerSummary 是图层集合的派生统计，不入库，仅用于自检与响应。
type LayerSummary struct {
	Total      int            `json:"total"`
	Visible    int            `json:"visible"`
	ByType     map[string]int `json:"by_type"`
	MaxZOrder  int            `json:"max_z_order"`
}
