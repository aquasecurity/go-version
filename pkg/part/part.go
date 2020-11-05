package part

type Part interface {
	Compare(Part) int
	IsNull() bool
}

func NewPart(s string) Part {
	var p Part
	p, err := NewUint64(s)
	if err == nil {
		return p
	}
	return NewString(s)
}
