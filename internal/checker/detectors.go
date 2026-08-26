// Package checker 实现「检查模块」：发现视觉编码歧义。包含各独立检测器，
// 由 checker.go 的 CheckAll 统一编排并应用例外豁免。纯逻辑，仅依赖 model。
package checker

import (
	"sort"
	"strings"

	"task251-chartambiguity/internal/model"
)

// detectChannelReuse 检测某一视觉通道上「同一 token 被多个不同变量复用」的歧义。
// 例如同色 #1f77b4 同时表示「温度」与「浓度」，读者无法区分。
func detectChannelReuse(encodings []model.VisualEncoding, channel, ambType string) []model.Ambiguity {
	byToken := map[string]map[string]struct{}{}
	for _, e := range encodings {
		if e.Channel != channel {
			continue
		}
		if byToken[e.Token] == nil {
			byToken[e.Token] = map[string]struct{}{}
		}
		byToken[e.Token][e.Variable] = struct{}{}
	}
	var out []model.Ambiguity
	for token, vars := range byToken {
		if len(vars) <= 1 {
			continue
		}
		names := make([]string, 0, len(vars))
		for v := range vars {
			names = append(names, v)
		}
		sort.Strings(names)
		out = append(out, model.Ambiguity{
			ID:          model.NewID("amb"),
			Type:        ambType,
			Severity:    model.AmbiguitySeverityError,
			Channel:     channel,
			Token:       token,
			Variables:   strings.Join(names, ","),
			Description: "通道 " + channel + " 的取值 " + token + " 被多个变量复用: " + strings.Join(names, ", "),
		})
	}
	return out
}

// DetectColorReuse 检测颜色复用歧义。
func DetectColorReuse(encodings []model.VisualEncoding) []model.Ambiguity {
	return detectChannelReuse(encodings, model.ChannelColor, model.AmbiguityColorReuse)
}

// DetectShapeReuse 检测形状复用歧义。
func DetectShapeReuse(encodings []model.VisualEncoding) []model.Ambiguity {
	return detectChannelReuse(encodings, model.ChannelShape, model.AmbiguityShapeReuse)
}

// DetectAxisUnitConflict 检测同一变量在坐标轴上使用冲突单位。
func DetectAxisUnitConflict(axes []model.Axis) []model.Ambiguity {
	unitsByVar := map[string]map[string]struct{}{}
	for _, a := range axes {
		if strings.TrimSpace(a.Variable) == "" {
			continue
		}
		u := strings.TrimSpace(a.Unit)
		if u == "" {
			continue
		}
		if unitsByVar[a.Variable] == nil {
			unitsByVar[a.Variable] = map[string]struct{}{}
		}
		unitsByVar[a.Variable][u] = struct{}{}
	}
	var out []model.Ambiguity
	for variable, units := range unitsByVar {
		if len(units) <= 1 {
			continue
		}
		ul := make([]string, 0, len(units))
		for u := range units {
			ul = append(ul, u)
		}
		sort.Strings(ul)
		out = append(out, model.Ambiguity{
			ID:          model.NewID("amb"),
			Type:        model.AmbiguityAxisUnitConflict,
			Severity:    model.AmbiguitySeverityError,
			Channel:     "",
			Token:       "",
			Variables:   variable,
			Description: "变量 " + variable + " 在坐标轴上使用了冲突单位: " + strings.Join(ul, ", "),
		})
	}
	return out
}

// legendCovers 判定图例是否覆盖某条编码（按 channel+token，或 channel+变量）。
func legendCovers(legends []model.Legend, e model.VisualEncoding) bool {
	for _, l := range legends {
		if l.Channel != e.Channel {
			continue
		}
		if l.Token == e.Token {
			return true
		}
		if l.CoversVariable != "" && l.CoversVariable == e.Variable {
			return true
		}
	}
	return false
}

// DetectMissingLegend 检测有编码但无图例覆盖的视觉通道。
func DetectMissingLegend(encodings []model.VisualEncoding, legends []model.Legend) []model.Ambiguity {
	var out []model.Ambiguity
	for _, e := range encodings {
		if legendCovers(legends, e) {
			continue
		}
		out = append(out, model.Ambiguity{
			ID:          model.NewID("amb"),
			Type:        model.AmbiguityMissingLegend,
			Severity:    model.AmbiguitySeverityError,
			Channel:     e.Channel,
			Token:       e.Token,
			Variables:   e.Variable,
			Description: "通道 " + e.Channel + " 的取值 " + e.Token + "（变量 " + e.Variable + "）缺少图例说明",
		})
	}
	return out
}

// DetectMappingConflict 把决议为 conflict 的变量-通道映射转为歧义。
func DetectMappingConflict(mappings []model.VariableMapping) []model.Ambiguity {
	var out []model.Ambiguity
	for _, m := range mappings {
		if m.Decision != model.MappingDecisionConflict {
			continue
		}
		out = append(out, model.Ambiguity{
			ID:          model.NewID("amb"),
			Type:        model.AmbiguityMappingConflict,
			Severity:    model.AmbiguitySeverityError,
			Channel:     m.Channel,
			Token:       "",
			Variables:   m.Variable,
			Description: "变量 " + m.Variable + " 在通道 " + m.Channel + " 存在矛盾赋值: " + m.Note,
		})
	}
	return out
}
