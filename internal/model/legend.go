package model

// 视觉通道枚举，图例与编码共用。
const (
	ChannelColor   = "color"
	ChannelShape   = "shape"
	ChannelSize    = "size"
	ChannelLinetype = "linetype"
)

var validChannel = map[string]bool{
	ChannelColor:    true,
	ChannelShape:    true,
	ChannelSize:     true,
	ChannelLinetype: true,
}

// IsValidChannel 校验视觉通道是否合法。
func IsValidChannel(c string) bool { return validChannel[c] }

// Legend 表示图例条目，解释某一视觉通道 token 所代表的变量。
type Legend struct {
	ID             string `json:"id"`
	FigureID       string `json:"figure_id"`
	Channel        string `json:"channel"`
	Label          string `json:"label"`
	Token          string `json:"token"` // 实际取值，如 #1f77b4 或 circle
	CoversVariable string `json:"covers_variable"`
}
