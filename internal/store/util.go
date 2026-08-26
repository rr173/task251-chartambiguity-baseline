package store

import "strconv"

// boolToInt 将布尔值转换为 SQLite 存储用的整数（0/1）。
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// intToBool 将 SQLite 整数还原为布尔值。
func intToBool(i int) bool { return i != 0 }

// itoa 是整数转字符串的本地封装，避免各 store 文件重复引入 strconv。
func itoa(n int) string { return strconv.Itoa(n) }
