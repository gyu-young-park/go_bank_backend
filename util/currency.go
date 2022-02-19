package util

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
	WON = "WON"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, CAD, WON:
		return true
	}
	return false
}
