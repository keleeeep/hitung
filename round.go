package hitung

import "math/big"

// bpDenominator is 100% in basis points.
const bpDenominator = 10000

// divRound returns num/den rounded half away from zero. den must not be zero.
func divRound(num, den *big.Int) int64 {
	q, r := new(big.Int).QuoRem(num, den, new(big.Int))
	twice := new(big.Int).Abs(r)
	twice.Lsh(twice, 1)
	if twice.Cmp(new(big.Int).Abs(den)) >= 0 {
		if (num.Sign() < 0) != (den.Sign() < 0) {
			q.Sub(q, big.NewInt(1))
		} else {
			q.Add(q, big.NewInt(1))
		}
	}
	return q.Int64()
}

// mulDivRound returns a*num/den rounded half away from zero, without overflow
// in the intermediate product.
func mulDivRound(a, num, den int64) int64 {
	return divRound(new(big.Int).Mul(big.NewInt(a), big.NewInt(num)), big.NewInt(den))
}
