package model

// 轴朝向枚举。
const (
	AxisOrientationX      = "x"
	AxisOrientationY      = "y"
	AxisOrientationColor  = "color"
	AxisOrientationRadius = "radius"
	AxisOrientationNone   = "none"
)

var validAxisOrientation = map[string]bool{
	AxisOrientationX:      true,
	AxisOrientationY:      true,
	AxisOrientationColor:  true,
	AxisOrientationRadius: true,
	AxisOrientationNone:   true,
}

// IsValidAxisOrientation 校验轴朝向是否合法。
func IsValidAxisOrientation(o string) bool { return validAxisOrientation[o] }

// Axis 表示图表的一条坐标轴，承载变量与单位语义。
type Axis struct {
	ID          string `json:"id"`
	FigureID    string `json:"figure_id"`
	Name        string `json:"name"`
	Variable    string `json:"variable"`
	Unit        string `json:"unit"`
	Orientation string `json:"orientation"`
}
