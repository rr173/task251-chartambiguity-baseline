package model

// 图形稿（Figure）状态机：
//   importing → pending_review → publishable → frozen
//   frozen 可被新规范版本 supersede（经 spec 发布流程）
const (
	FigureStatusImporting     = "importing"
	FigureStatusPendingReview = "pending_review"
	FigureStatusPublishable   = "publishable"
	FigureStatusFrozen        = "frozen"
)

// validFigureStatus 列出图形稿允许的全部状态。
var validFigureStatus = map[string]bool{
	FigureStatusImporting:     true,
	FigureStatusPendingReview: true,
	FigureStatusPublishable:   true,
	FigureStatusFrozen:        true,
}

// IsValidFigureStatus 校验状态字符串是否合法。
func IsValidFigureStatus(s string) bool { return validFigureStatus[s] }

// Figure 表示一篇待复核的学术图表稿件。
type Figure struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	SourceFP   string `json:"source_fp"` // 导入语义的指纹，用于幂等
	LayerCount int    `json:"layer_count"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

// CanEdit 判定图形稿当前是否允许结构性变更（冻结后禁止修改语义）。
func (f Figure) CanEdit() bool { return f.Status != FigureStatusFrozen }
