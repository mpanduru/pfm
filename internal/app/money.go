package app

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseRON(input string) (int64, error) {
	input = strings.TrimSpace(input)

	negative := false
	if strings.HasPrefix(input, "-") {
		negative = true
		input = strings.TrimPrefix(input, "-")
	}

	parts := strings.SplitN(input, ".", 2)

	leiPart := parts[0]
	baniPart := "00"

	if len(parts) == 2 {
		baniPart = parts[1]
	}

	if len(baniPart) == 1 {
		baniPart += "0"
	}
	if len(baniPart) > 2 {
		return 0, fmt.Errorf("too many decimal places: %q", input)
	}

	lei, err := strconv.ParseInt(leiPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount: %q", input)
	}

	bani, err := strconv.ParseInt(baniPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount: %q", input)
	}

	total := lei*100 + bani
	if negative {
		total = -total
	}
	return total, nil
}

func FormatRON(bani int64) string {
	sign := ""
	if bani < 0 {
		sign = "-"
		bani = -bani
	}
	return fmt.Sprintf("%s%d.%02d RON", sign, bani/100, bani%100)
}
