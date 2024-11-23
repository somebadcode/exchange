package converter

import (
	"math/big"
	"strconv"
	"strings"
)

func IntPow[T ~int | ~int16 | ~int32 | ~int64](n T, m T) T {
	if m == 0 {
		return 1
	}

	if m == 1 {
		return n
	}

	result := n
	for i := T(2); i <= m; i++ {
		result *= n
	}

	return result
}

func ParseRat(s string) (*big.Rat, error) {
	s = strings.TrimSpace(s)

	var denom int64 = 1

	// Use the position of the decimal point to calculate the denominator.
	pos := strings.IndexByte(s, '.')
	if pos != -1 {
		denom = IntPow(10, max(1, int64(len(s[pos:]))-1))
	}

	num, err := strconv.ParseInt(strings.ReplaceAll(s, ".", ""), 10, 64)
	if err != nil {
		return nil, err
	}

	qty := big.NewRat(num, denom)

	return qty, nil
}
