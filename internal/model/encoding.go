package model

// 视觉编码（VisualEncoding）状态机：
//   parsed → valid | ambiguous | missing_legend
const (
	EncodingStatusParsed       = "parsed"
	EncodingStatusValid        = "valid"
	EncodingStatusAmbiguous   = "ambiguous"
	EncodingStatusMissingLgnd = "missing_legend"
)

var validEncodingStatus = map[string]bool{
	EncodingStatusParsed:       true,
	EncodingStatusValid:        true,
	EncodingStatusAmbiguous:    true,
	EncodingStatusMissingLgnd:  true,
}

// IsValidEncodingStatus 校验编码状态是否合法。
func IsValidEncodingStatus(s string) bool { return validEncodingStatus[s] }

// VisualEncoding 表示「变量 V 经通道 C 以 token T 呈现」的声明。
type VisualEncoding struct {
	ID       string `json:"id"`
	FigureID string `json:"figure_id"`
	LayerID  string `json:"layer_id"`
	Variable string `json:"variable"`
	Channel  string `json:"channel"`
	Token    string `json:"token"`
	Status   string `json:"status"`
}
