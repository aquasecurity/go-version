package part

import (
	"reflect"
	"strings"
)

type Parts []Part

func NewParts(s string) Parts {
	var parts []Part
	for _, p := range strings.Split(s, ".") {
		parts = append(parts, NewPart(p))
	}
	return parts
}

func (parts Parts) Normalize() Parts {
	ret := parts
	for i := len(parts) - 1; i >= 0; i-- {
		lastItem := parts[i]
		if lastItem.IsNull() {
			ret = ret[:i]
			continue
		}
		break
	}
	return ret
}

func (parts Parts) Compare(other Parts, padding Part) int {
	if other == nil {
		return 1
	}

	if reflect.DeepEqual(parts, other) {
		return 0
	}

	iter := parts.Zip(other)
	for tuple := iter(); tuple != nil; tuple = iter() {
		var l, r = tuple.Left, tuple.Right
		if l == nil {
			l = padding
		}
		if r == nil {
			r = padding
		}

		if l == nil {
			if result := -1 * r.Compare(l); result != 0 {
				return result
			}
		} else {
			if result := l.Compare(r); result != 0 {
				return result
			}
		}
	}
	return 0
}

type ZipTuple struct {
	Left  Part
	Right Part
}

func (parts Parts) Zip(other Parts) func() *ZipTuple {
	i := 0
	return func() *ZipTuple {
		var part1, part2 Part
		if i < len(parts) {
			part1 = parts[i]
		}
		if i < len(other) {
			part2 = other[i]
		}
		if part1 == nil && part2 == nil {
			return nil
		}
		i++
		return &ZipTuple{Left: part1, Right: part2}
	}
}
