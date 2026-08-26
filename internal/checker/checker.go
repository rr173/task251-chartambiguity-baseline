package checker

import (
	"strings"
	"time"

	"task251-chartambiguity/internal/model"
)

// CheckAll 汇总全部检测器的结果，并依据已登记的例外豁免对歧义做消解判定。
// 返回当前计算出的歧义集合（被豁免的歧义 Resolved=true 并关联 exception_id）。
// figureID 用于把每条歧义归属到对应图形稿，确保落库后可按图形稿检索与计数。
// 入参均为某图形稿的实体切片；空切片表示无数据，不会产生歧义。
func CheckAll(
	figureID string,
	encodings []model.VisualEncoding,
	axes []model.Axis,
	legends []model.Legend,
	mappings []model.VariableMapping,
	exceptions []model.Exception,
) []model.Ambiguity {
	now := time.Now().Unix()
	detected := make([]model.Ambiguity, 0, 16)
	detected = append(detected, DetectColorReuse(encodings)...)
	detected = append(detected, DetectShapeReuse(encodings)...)
	detected = append(detected, DetectAxisUnitConflict(axes)...)
	detected = append(detected, DetectMissingLegend(encodings, legends)...)
	detected = append(detected, DetectMappingConflict(mappings)...)

	for i := range detected {
		// 检测器本身不关心图形稿归属；在此统一打标，避免落库后 figure_id 丢失，
		// 否则按 figure_id 查询歧义或统计未解决歧义数都会误判为空。
		detected[i].FigureID = figureID
		detected[i].CreatedAt = now
		if exc := matchException(detected[i], exceptions); exc != nil {
			detected[i].Resolved = true
			detected[i].ExceptionID = exc.ID
		}
	}
	return detected
}

// matchException 判定某条歧义是否被任一例外覆盖。
func matchException(a model.Ambiguity, exceptions []model.Exception) *model.Exception {
	for i := range exceptions {
		exc := exceptions[i]
		switch a.Type {
		case model.AmbiguityColorReuse, model.AmbiguityShapeReuse:
			if exc.Kind == model.ExceptionReuse &&
				exc.TargetChannel == a.Channel && exc.TargetToken == a.Token {
				return &exc
			}
		case model.AmbiguityAxisUnitConflict:
			if exc.Kind == model.ExceptionAxisUnit && varIn(exc.TargetVariable, a.Variables) {
				return &exc
			}
		case model.AmbiguityMissingLegend:
			if exc.Kind == model.ExceptionMissingLegend &&
				exc.TargetChannel == a.Channel && exc.TargetToken == a.Token {
				return &exc
			}
		case model.AmbiguityMappingConflict:
			if exc.Kind == model.ExceptionMapping &&
				exc.TargetVariable == a.Variables && exc.TargetChannel == a.Channel {
				return &exc
			}
		}
	}
	return nil
}

// varIn 判定 target 是否出现在逗号分隔的变量列表中。
func varIn(target, list string) bool {
	if target == "" {
		return false
	}
	for _, v := range strings.Split(list, ",") {
		if strings.TrimSpace(v) == target {
			return true
		}
	}
	return false
}
