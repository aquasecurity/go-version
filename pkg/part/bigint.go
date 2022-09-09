package part

import (
	"fmt"
	"math/big"
)

var BigIntZero = BigInt(big.Int{})

type BigInt big.Int

func NewBigInt(s string) (BigInt, error) {
	bi := big.Int{}
	_, ok := bi.SetString(s, 10)
	if !ok {
		return BigInt{}, fmt.Errorf("error setting %s to BigInt", s)
	}
	return BigInt(bi), nil
}

func (s BigInt) String() string {
	bi := big.Int(s)
	return bi.String()
}

func (s BigInt) Compare(other Part) int {
	if other == nil {
		return 1
	}

	switch o := other.(type) {
	case Uint64:
		biA := big.Int(s)
		biAP := &biA
		return biAP.Cmp(big.NewInt(int64(o)))
	case BigInt:
		biA := big.Int(s)
		biB := big.Int(o)
		biAP := &biA
		return biAP.Cmp(&biB)
	case String:
		return -1
	case PreString:
		return 1
	case Any:
		return 0
	case Empty:
		if o.IsAny() {
			return 0
		}
		return s.Compare(BigInt{})
	default:
		panic("unknown type")
	}
}

func (s BigInt) IsNull() bool {
	bi := big.Int(s)
	bip := &bi
	return bip.Cmp(big.NewInt(0)) == 0
}

func (s BigInt) IsAny() bool {
	return false
}

func (s BigInt) IsEmpty() bool {
	return false
}
