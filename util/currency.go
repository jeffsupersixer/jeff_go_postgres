package util

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

func IsSupportedCurrency(curreny string) bool {
	switch curreny {
	case USD, EUR, CAD:
		return true
	}
	return false
}
