package orders

import (
	"strconv"
	"unicode"
)

func ValidateOrderNumber(orderNumber string) bool {
	var sum int
	alternate := false

	for i := len(orderNumber) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(orderNumber[i]))
		if err != nil || !unicode.IsDigit(rune(orderNumber[i])) {
			return false
		}
		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}
	return sum%10 == 0
}
