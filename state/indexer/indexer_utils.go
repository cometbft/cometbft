package indexer

import (
	"fmt"
	"math/big"
)

// If the actual event value is a float, we get the condition and parse it as a float
// to compare agains
func compareFloat(op1 *big.Float, op2 interface{}) (int, bool, error) {
	switch opVal := op2.(type) {
	case *big.Int:
		vF, _, err := big.ParseFloat(opVal.String(), 10, op1.Prec(), big.ToNearestEven)
		if err != nil {
			err = fmt.Errorf("failed to convert %s to float", opVal)
		}
		cmp := op1.Cmp(vF)

		return cmp, false, err

	case *big.Float:
		return op1.Cmp(opVal), true, nil
	default:
		return -1, false, fmt.Errorf("unable to parse arguments")
	}
}

// If the event value we compare against the condition (op2) is an integer
// we convert the int to float with a precision equal to the number of bits
// needed to represent the integer to avoid rounding issues with floats
// where 100 would equal to 100.2 because 100.2 is rounded to 100, while 100.7
// would be rounded to 101.
func compareInt(op1 *big.Int, op2 interface{}) (int, bool, error) {
	switch opVal := op2.(type) {
	case *big.Int:
		return op1.Cmp(opVal), false, nil
	case *big.Float:
		vF, _, err := big.ParseFloat(op1.String(), 10, uint(op1.BitLen()), big.ToNearestEven)
		if err != nil {
			return -1, true, fmt.Errorf("failed to convert %f to int", opVal)
		}
		// For Int conditions, if the upper bound is not included we decrease the value of the
		// upper bound by 1 and compare against that. cmp will be 0 in the corner case that the
		// condition is a number with 0 decimals
		// For float we cannot decrease the condition by 1 as it would be wrong
		// example int: x < 100; upperBound = 99; if x.Cmp(99) == 0 the condition holds
		// example float: x < 100.0; upperBound = 100.0; if x.Cmp(100) ==0 then returning x
		// would be wrong. Thus we check whether the float upper bound should be included
		// if cmp == 0 && !include {
		// 	return 1, nil
		// }
		return vF.Cmp(opVal), true, nil
	default:
		return -1, false, fmt.Errorf("unable to parse arguments")
	}
}

func CheckBounds(ranges QueryRange, v interface{}) bool {
	include := true
	lowerBound := ranges.LowerBoundValue()
	upperBound := ranges.UpperBoundValue()
	switch vVal := v.(type) {
	case *big.Int:
		if lowerBound != nil {
			cmp, isFloat, err := compareInt(vVal, lowerBound)
			if err == nil && (cmp == -1 || (isFloat && cmp == 0 && !ranges.IncludeLowerBound)) {
				include = false
			}
		}
		if upperBound != nil {
			cmp, isFloat, err := compareInt(vVal, upperBound)
			if err == nil && (cmp == 1 || (isFloat && cmp == 0 && !ranges.IncludeUpperBound)) {
				include = false
			}
		}

	case *big.Float:
		if lowerBound != nil {
			cmp, isFloat, err := compareFloat(vVal, lowerBound)
			if err == nil && (cmp == -1 || (cmp == 0 && isFloat && !ranges.IncludeLowerBound)) {
				include = false
			}
		}
		if upperBound != nil {
			cmp, isFloat, err := compareFloat(vVal, upperBound)
			if err == nil && (cmp == 1 || (cmp == 0 && isFloat && !ranges.IncludeUpperBound)) {
				include = false
			}
		}
	}
	return include
}
