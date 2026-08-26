package model

// 例外（Exception）种类，对应可豁免的歧义类型。
const (
	ExceptionReuse         = "reuse_exemption"          // 颜色/形状复用豁免（按 channel+token）
	ExceptionAxisUnit      = "axis_unit_exemption"      // 轴单位冲突豁免（按 variable）
	ExceptionMissingLegend = "missing_legend_exemption" // 缺图例豁免（按 channel+token）
	ExceptionMapping       = "mapping_conflict_exemption" // 映射冲突豁免（按 variable+channel）
)

var validExceptionKind = map[string]bool{
	ExceptionReuse:         true,
	ExceptionAxisUnit:      true,
	ExceptionMissingLegend: true,
	ExceptionMapping:       true,
}

// IsValidExceptionKind 校验例外种类是否合法。
func IsValidExceptionKind(k string) bool { return validExceptionKind[k] }

// Exception 表示编辑对某一歧义的豁免声明。
type Exception struct {
	ID             string `json:"id"`
	FigureID       string `json:"figure_id"`
	Kind           string `json:"kind"`
	TargetChannel  string `json:"target_channel"`
	TargetToken    string `json:"target_token"`
	TargetVariable string `json:"target_variable"`
	Reason         string `json:"reason"`
	CreatedAt      int64  `json:"created_at"`
}
