package model

// 变量映射（VariableMapping）状态机：
//   candidate → conflict | confirmed | revoked
const (
	MappingDecisionCandidate = "candidate"
	MappingDecisionConflict  = "conflict"
	MappingDecisionConfirmed = "confirmed"
	MappingDecisionRevoked   = "revoked"
)

var validMappingDecision = map[string]bool{
	MappingDecisionCandidate: true,
	MappingDecisionConflict:  true,
	MappingDecisionConfirmed: true,
	MappingDecisionRevoked:   true,
}

// IsValidMappingDecision 校验映射决议是否合法。
func IsValidMappingDecision(s string) bool { return validMappingDecision[s] }

// VariableMapping 表示编辑对「变量→通道」的一致性决议。
// 当同一变量在同一通道出现互相矛盾的 token 赋值时，系统将其判定为 conflict。
type VariableMapping struct {
	ID       string `json:"id"`
	FigureID string `json:"figure_id"`
	Variable string `json:"variable"`
	Channel  string `json:"channel"`
	Decision string `json:"decision"`
	Note     string `json:"note"`
}
