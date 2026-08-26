package model

// Variable 表示图表声明的一个数据变量。
type Variable struct {
	ID          string `json:"id"`
	FigureID    string `json:"figure_id"`
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	Description string `json:"description"`
}
