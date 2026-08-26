// Package layer 实现「图层模块」：读取并校验图表的图层/轴/图例/变量语义，
// 并派生图层摘要。该包为纯逻辑，仅依赖 model，不接触数据库。
package layer

import (
	"strings"

	"task251-chartambiguity/internal/model"
)

// ValidateLayer 校验单个图层的业务约束。
func ValidateLayer(l model.Layer) error {
	if strings.TrimSpace(l.Name) == "" {
		return model.ErrInvalidArgument
	}
	if !model.IsValidLayerType(l.LayerType) {
		return model.ErrInvalidArgument
	}
	if l.ZOrder < 0 {
		return model.ErrInvalidArgument
	}
	return nil
}

// ValidateAxis 校验单条坐标轴的业务约束。
func ValidateAxis(a model.Axis) error {
	if strings.TrimSpace(a.Name) == "" || strings.TrimSpace(a.Variable) == "" {
		return model.ErrInvalidArgument
	}
	if !model.IsValidAxisOrientation(a.Orientation) {
		return model.ErrInvalidArgument
	}
	return nil
}

// ValidateLegend 校验单个图例条目的业务约束。
func ValidateLegend(l model.Legend) error {
	if !model.IsValidChannel(l.Channel) {
		return model.ErrInvalidArgument
	}
	if strings.TrimSpace(l.Token) == "" {
		return model.ErrInvalidArgument
	}
	return nil
}

// ValidateVariable 校验单个变量声明的业务约束。
func ValidateVariable(v model.Variable) error {
	if strings.TrimSpace(v.Name) == "" {
		return model.ErrInvalidArgument
	}
	return nil
}

// ComputeLayerSummary 由图层集合派生统计摘要（用于自检与响应）。
func ComputeLayerSummary(layers []model.Layer) model.LayerSummary {
	sum := model.LayerSummary{ByType: map[string]int{}}
	for _, l := range layers {
		sum.Total++
		if l.Visible {
			sum.Visible++
		}
		sum.ByType[l.LayerType]++
		if l.ZOrder > sum.MaxZOrder {
			sum.MaxZOrder = l.ZOrder
		}
	}
	return sum
}
