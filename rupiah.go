package hitung

import (
	"errors"
	"strconv"
	"strings"
)

// Rupiah formats n the way Indonesian receipts and invoices print it:
// Rupiah(1234567) is "Rp 1.234.567" and Rupiah(-5000) is "-Rp 5.000".
func Rupiah(n int64) string {
	neg := n < 0
	u := uint64(n)
	if neg {
		u = uint64(-(n + 1)) + 1 // safe for math.MinInt64
	}
	digits := strconv.FormatUint(u, 10)
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	b.WriteString("Rp ")
	for i, c := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// ErrNotRupiah is returned by ParseRupiah for text that is not a whole rupiah amount.
var ErrNotRupiah = errors.New("hitung: not a rupiah amount")

// ParseRupiah reads an amount typed or printed in Indonesian style.
// It accepts "Rp 1.234.567", "Rp1.234.567", "1.234.567", "1234567" and a
// leading minus sign. Dots must group thousands; decimals are rejected.
func ParseRupiah(s string) (int64, error) {
	s = strings.TrimSpace(s)
	neg := false
	if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "−") {
		neg = true
		s = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(s, "-"), "−"))
	}
	if len(s) >= 2 && strings.EqualFold(s[:2], "rp") {
		s = strings.TrimSpace(s[2:])
	}
	if s == "" {
		return 0, ErrNotRupiah
	}
	if strings.Contains(s, ".") {
		groups := strings.Split(s, ".")
		if len(groups[0]) == 0 || len(groups[0]) > 3 {
			return 0, ErrNotRupiah
		}
		for _, g := range groups[1:] {
			if len(g) != 3 {
				return 0, ErrNotRupiah
			}
		}
		s = strings.Join(groups, "")
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, ErrNotRupiah
		}
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, ErrNotRupiah
	}
	if neg {
		n = -n
	}
	return n, nil
}
