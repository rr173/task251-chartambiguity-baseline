// Package spec 实现「规范模块」：管理例外应用后的图规范版本发布与快照构建。
// 纯逻辑，仅依赖 model 与标准库。
package spec

import (
	"encoding/json"
	"time"

	"task251-chartambiguity/internal/model"
)

// Snapshot 是冻结时保存的不可变视图：编码矩阵 + 图例 + 变量/轴声明 + 例外。
type Snapshot struct {
	FigureID   string                  `json:"figure_id"`
	GeneratedAt int64                  `json:"generated_at"`
	Encodings  []model.VisualEncoding  `json:"encodings"`
	Legends    []model.Legend          `json:"legends"`
	Variables  []model.Variable        `json:"variables"`
	Axes       []model.Axis            `json:"axes"`
	Exceptions []model.Exception       `json:"exceptions"`
}

// BuildSnapshot 将当前图形语义固化为 JSON 字符串，供规范版本冻结保存。
func BuildSnapshot(
	figureID string,
	encodings []model.VisualEncoding,
	legends []model.Legend,
	variables []model.Variable,
	axes []model.Axis,
	exceptions []model.Exception,
) (string, error) {
	snap := Snapshot{
		FigureID:   figureID,
		GeneratedAt: time.Now().Unix(),
		Encodings:  encodings,
		Legends:    legends,
		Variables:  variables,
		Axes:       axes,
		Exceptions: exceptions,
	}
	b, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// NextVersion 由当前最大版本号推导下一版本号。
func NextVersion(maxVersion int) int { return maxVersion + 1 }

// CanPublish 判定图形稿是否可发布规范版本：要求不存在未解决的歧义。
func CanPublish(openAmbiguityCount int) bool { return openAmbiguityCount == 0 }
