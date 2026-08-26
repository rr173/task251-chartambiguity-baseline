// Package mapping 实现「映射模块」：从视觉编码派生变量-通道一致性决议，
// 并标记 conflict / confirmed。纯逻辑，仅依赖 model。
package mapping

import (
	"sort"
	"strings"

	"task251-chartambiguity/internal/model"
)

// BuildMappings 根据视觉编码集合，对每个 (variable, channel) 组合给出决议：
//   - 同一变量在同一通道仅出现一种 token → confirmed（一致）
//   - 同一变量在同一通道出现多种互相矛盾的 token → conflict
// 返回的切片按 (variable, channel) 字典序稳定排序。
func BuildMappings(encodings []model.VisualEncoding) []model.VariableMapping {
	type key struct{ variable, channel string }
	tokens := map[key]map[string]struct{}{}
	order := []key{}
	seen := map[key]bool{}
	for _, e := range encodings {
		k := key{e.Variable, e.Channel}
		if !seen[k] {
			seen[k] = true
			order = append(order, k)
		}
		if tokens[k] == nil {
			tokens[k] = map[string]struct{}{}
		}
		tokens[k][e.Token] = struct{}{}
	}

	out := make([]model.VariableMapping, 0, len(order))
	for _, k := range order {
		decision := model.MappingDecisionConfirmed
		if len(tokens[k]) > 1 {
			decision = model.MappingDecisionConflict
		}
		out = append(out, model.VariableMapping{
			ID:       model.NewID("map"),
			FigureID: firstFigureID(encodings),
			Variable: k.variable,
			Channel:  k.channel,
			Decision: decision,
			Note:     describeConflict(tokens[k], decision),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Variable != out[j].Variable {
			return out[i].Variable < out[j].Variable
		}
		return out[i].Channel < out[j].Channel
	})
	return out
}

func describeConflict(toks map[string]struct{}, decision string) string {
	if decision != model.MappingDecisionConflict {
		return "一致的通道赋值"
	}
	parts := make([]string, 0, len(toks))
	for t := range toks {
		parts = append(parts, t)
	}
	sort.Strings(parts)
	return "通道赋值矛盾: " + strings.Join(parts, " / ")
}

func firstFigureID(encodings []model.VisualEncoding) string {
	for _, e := range encodings {
		if e.FigureID != "" {
			return e.FigureID
		}
	}
	return ""
}
