package model

// 歧义（Ambiguity）类型，对应 checker 的各检测项。
const (
	AmbiguityColorReuse       = "color_reuse"        // 同色被两个不同变量复用
	AmbiguityShapeReuse       = "shape_reuse"        // 同形状被两个不同变量复用
	AmbiguityAxisUnitConflict = "axis_unit_conflict" // 同一变量轴单位冲突
	AmbiguityMissingLegend    = "missing_legend"     // 编码通道无图例覆盖
	AmbiguityMappingConflict  = "mapping_conflict"   // 变量-通道映射矛盾
)

// 歧义严重度。
const (
	AmbiguitySeverityError   = "error"
	AmbiguitySeverityWarning = "warning"
)

// Ambiguity 表示一次检测出的视觉编码歧义（可能已被例外豁免）。
type Ambiguity struct {
	ID          string `json:"id"`
	FigureID    string `json:"figure_id"`
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Channel     string `json:"channel"`
	Token       string `json:"token"`
	Variables   string `json:"variables"` // 涉及变量，逗号分隔
	Description string `json:"description"`
	Resolved    bool   `json:"resolved"`
	ExceptionID string `json:"exception_id,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

// IsOpen 判定该歧义是否仍未解决（未被豁免）。
func (a Ambiguity) IsOpen() bool { return !a.Resolved }
