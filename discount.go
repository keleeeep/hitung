package hitung

import "math/big"

// Discount applies percentage discounts one after another and rounds once at
// the end. Discount(100000, 500, 300) is 5% then 3% off: 92150, not 92000.
func Discount(amount int64, discountsBP ...int64) int64 {
	num := big.NewInt(amount)
	den := big.NewInt(1)
	for _, d := range discountsBP {
		num.Mul(num, big.NewInt(bpDenominator-d))
		den.Mul(den, big.NewInt(bpDenominator))
	}
	return divRound(num, den)
}
