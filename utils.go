package gmeter

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
)

func ErrPrintln(s string) {
	if len(s) != 0 {
		os.Stderr.WriteString(s)
	}
	os.Stderr.WriteString("\n")
}

func ErrPrintf(format string, items ...any) {
	os.Stderr.WriteString(fmt.Sprintf(format, items...))
}

func GainError(errs []error) error {
	var endErrs []error
	for _, err := range errs {
		if err != nil {
			endErrs = append(endErrs, err)
		}
	}
	errs = endErrs
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	var strs []string
	for _, err := range errs {
		strs = append(strs, err.Error())
	}
	return errors.New(strings.Join(strs, " "))
}

func sum(nums []int64) int64 {
	sum := int64(0)
	for _, num := range nums {
		sum += num
	}
	return sum
}

func div(num1 int64, num2 int64) float64 {
	if num2 == 0 {
		return 0
	}
	return float64(num1) / float64(num2)
}

func max(nums ...int64) int64 {
	m := int64(math.MinInt64)
	for _, num := range nums {
		if num > m {
			m = num
		}
	}
	return m
}

func maxWithDefault(defaultNum int64, nums ...int64) int64 {
	if len(nums) == 0 {
		return defaultNum
	}
	return max(nums...)
}

func minWithDefault(defaultNum int64, nums ...int64) int64 {
	if len(nums) == 0 {
		return defaultNum
	}
	return min(nums...)
}

func min(nums ...int64) int64 {
	m := int64(math.MaxInt64)
	for _, num := range nums {
		if num < m {
			m = num
		}
	}
	return m
}
