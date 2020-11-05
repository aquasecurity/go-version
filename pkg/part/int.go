package part

import "strconv"

type Int64 int64

func NewInt64(s string) (Int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return Int64(n), nil
}

func (s Int64) Compare(other Part) int {
	if other == nil {
		return 1
	} else if s == other {
		return 0
	}

	switch o := other.(type) {
	case Int64:
		if s < o {
			return -1
		}
		return 1
	case Uint64:
		panic("not supported")
	case String:
		return -1
	default:
		panic("unknown type")
	}
}

func (s Int64) IsNull() bool {
	return s == 0
}

type Uint64 uint64

func NewUint64(s string) (Uint64, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return Uint64(n), nil
}

func (s Uint64) Compare(other Part) int {
	if other == nil {
		return 1
	} else if s == other {
		return 0
	}

	switch o := other.(type) {
	case Int64:
		panic("not supported")
	case Uint64:
		if s < o {
			return -1
		}
		return 1
	case String:
		return -1
	default:
		panic("unknown type")
	}
}

func (s Uint64) IsNull() bool {
	return s == 0
}
