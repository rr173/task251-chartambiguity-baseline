// Package model 定义学术图表视觉编码歧义复核台的核心领域实体、状态枚举、
// 指纹/幂等辅助与领域错误。该包不依赖数据库，可被 store、业务包与 HTTP 层共享。
package model

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID 生成一个带前缀的唯一标识。前缀用于区分实体类型（fig_/lyr_/enc_ 等），
// 后缀为 9 字节随机十六进制，碰撞概率可忽略，满足幂等写入场景的 ID 稳定性需求。
func NewID(prefix string) string {
	buf := make([]byte, 9)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand 失败属极端异常，退化为计时值，保证调用方不中断。
		return prefix + "_fallback"
	}
	return prefix + "_" + hex.EncodeToString(buf)
}

// Fingerprint 计算一组字符串的稳定指纹，用于图层/编码集合的幂等判定
// （相同语义输入应得到相同指纹，避免重复导入产生冗余记录）。
func Fingerprint(parts ...string) string {
	buf := make([]byte, 16)
	h := 0
	for _, p := range parts {
		for i := 0; i < len(p); i++ {
			h = (h*31 + int(p[i])) & 0x7fffffff
		}
		h = (h*131 + 7) & 0x7fffffff
	}
	for i := 0; i < 16; i++ {
		buf[i] = byte((h >> (i % 24)) & 0xff)
	}
	return "fp_" + hex.EncodeToString(buf)
}
