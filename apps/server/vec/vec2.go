package vec

import "math"

type V2 struct{ X, Y float64 }

func (v V2) Scale(s float64) V2 {
	return V2{X: v.X * s, Y: v.Y * s}
}

func (v V2) Add(u V2) V2 {
	return V2{X: v.X + u.X, Y: v.Y + u.Y}
}

func (v V2) Sub(u V2) V2 {
	return V2{X: v.X - u.X, Y: v.Y - u.Y}
}

func (v V2) Dot(u V2) float64 {
	return v.X*u.X + v.Y*u.Y
}

func (v V2) Length2() float64 {
	return v.X*v.X + v.Y*v.Y
}

func (v V2) Length() float64 {
	return math.Sqrt(v.Length2())
}

func (v V2) Normalize0() V2 {
	length := v.Length()
	if length == 0 {
		return v
	}
	return v.Scale(1.0 / length)
}

func (v V2) Normalize() V2 {
	length := v.Length()
	if length == 0 {
		panic("cannot normalize zero vector")
	}
	return v.Scale(1.0 / length)
}
