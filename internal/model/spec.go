package model

// 图规范（FigureSpec）状态机：
//   draft → shared → frozen → superseded
const (
	SpecStatusDraft       = "draft"
	SpecStatusShared      = "shared"
	SpecStatusFrozen      = "frozen"
	SpecStatusSuperseded  = "superseded"
)

var validSpecStatus = map[string]bool{
	SpecStatusDraft:      true,
	SpecStatusShared:     true,
	SpecStatusFrozen:     true,
	SpecStatusSuperseded: true,
}

// IsValidSpecStatus 校验规范状态是否合法。
func IsValidSpecStatus(s string) bool { return validSpecStatus[s] }

// FigureSpec 表示一次可发布的图规范版本（冻结时绑定编码与例外快照）。
type FigureSpec struct {
	ID        string `json:"id"`
	FigureID  string `json:"figure_id"`
	Version   int    `json:"version"`
	Status    string `json:"status"`
	Snapshot  string `json:"snapshot"` // JSON：编码矩阵 + 例外 + 变量/轴声明
	CreatedAt int64  `json:"created_at"`
}
