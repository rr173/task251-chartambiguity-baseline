package model

import "errors"

// 领域错误。业务包与 store 统一返回这些错误，service/HTTP 层据此映射到状态码。
var (
	// ErrNotFound 表示请求的实体不存在。
	ErrNotFound = errors.New("entity not found")
	// ErrInvalidStatus 表示状态机流转非法。
	ErrInvalidStatus = errors.New("invalid status transition")
	// ErrConflict 表示出现业务冲突（如冻结后修改、单位冲突等）。
	ErrConflict = errors.New("business conflict")
	// ErrFrozen 表示实体已冻结，禁止变更。
	ErrFrozen = errors.New("entity is frozen")
	// ErrDuplicate 表示唯一键冲突。
	ErrDuplicate = errors.New("duplicate entry")
	// ErrInvalidArgument 表示入参非法。
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrHasOpenAmbiguity 表示存在未解决的歧义，禁止发布规范。
	ErrHasOpenAmbiguity = errors.New("figure has open ambiguities")
)

// IsConflict 判定错误是否源于业务冲突，便于 HTTP 层返回 409。
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict) || errors.Is(err, ErrFrozen) || errors.Is(err, ErrHasOpenAmbiguity)
}
