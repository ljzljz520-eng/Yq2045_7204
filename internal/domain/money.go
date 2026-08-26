package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const centsScale = int64(100)

func CentsFromFloat(value float64) int64 {
	if value <= 0 {
		return 0
	}
	return int64(math.Round(value * float64(centsScale)))
}

func FloatFromCents(value int64) float64 { return float64(value) / float64(centsScale) }

func FormatCents(value int64) string {
	negative := value < 0
	if negative {
		value = -value
	}
	whole := value / centsScale
	part := value % centsScale
	result := fmt.Sprintf("%d.%02d", whole, part)
	if negative {
		return "-" + result
	}
	return result
}

func ParseMoney(value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("amount is empty")
	}
	if strings.ContainsAny(trimmed, "eE") {
		return 0, fmt.Errorf("amount must be a decimal")
	}
	negative := strings.HasPrefix(trimmed, "-")
	if negative {
		trimmed = strings.TrimPrefix(trimmed, "-")
	}
	parts := strings.Split(trimmed, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("invalid amount %q", value)
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 {
		return 0, fmt.Errorf("invalid amount %q", value)
	}
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	if len(frac) > 2 {
		if len(frac) != 3 || frac[2] != '5' {
			return 0, fmt.Errorf("amount has more than two decimals")
		}
	}
	for len(frac) < 2 {
		frac += "0"
	}
	fracValue := int64(0)
	if frac != "" {
		fracValue, err = strconv.ParseInt(frac[:2], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid amount %q", value)
		}
		if len(parts) == 2 && len(parts[1]) == 3 && parts[1][2] >= '5' {
			fracValue++
		}
	}
	result := whole*centsScale + fracValue
	if negative && result != 0 {
		result = -result
	}
	return result, nil
}

func AddCents(values ...int64) int64 {
	var total int64
	for _, value := range values {
		total += value
	}
	return total
}

func SubCents(base int64, deductions ...int64) int64 {
	result := base
	for _, deduction := range deductions {
		result -= deduction
	}
	if result < 0 {
		return 0
	}
	return result
}

func MultiplyCents(cents int64, quantity int) int64 {
	if quantity <= 0 || cents <= 0 {
		return 0
	}
	return cents * int64(quantity)
}

func PercentageCents(cents int64, percent float64) int64 {
	if cents <= 0 || percent <= 0 {
		return 0
	}
	return int64(math.Round(float64(cents) * percent / 100))
}

func SplitCents(total int64, parts int) []int64 {
	if parts <= 0 {
		return nil
	}
	result := make([]int64, parts)
	base := total / int64(parts)
	remainder := total % int64(parts)
	for i := range result {
		result[i] = base
		if int64(i) < remainder {
			result[i]++
		}
	}
	return result
}
