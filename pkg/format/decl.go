package format

import "github.com/gookit/goutil/mathutil"

func Declension(n int, forms ...string) string {
	n = mathutil.Abs(n) % 100
	if n >= 11 && n <= 14 {
		return forms[2]
	}

	switch n % 10 {
	case 1:
		return forms[0]
	case 2, 3, 4:
		return forms[1]
	default:
		return forms[2]
	}
}
