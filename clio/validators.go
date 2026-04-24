package clio

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type ValidatorFn func(string) error

func NotEmpty(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("value cannot be empty")
	}
	return nil
}

func MinLength(min int) ValidatorFn {
	return func(s string) error {
		if len(s) < min {
			return fmt.Errorf("value must be at least %d characters long", min)
		}
		return nil
	}
}

func MaxLength(max int) ValidatorFn {
	return func(s string) error {
		if len(s) > max {
			return fmt.Errorf("value must be at most %d characters long", max)
		}
		return nil
	}
}

func GreaterThan(min float64) ValidatorFn {
	return func(s string) error {
		n, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return errors.New("value must be a number")
		}
		if n <= min {
			return fmt.Errorf("value must be greater than %g", min)
		}
		return nil
	}
}

func LessThan(max float64) ValidatorFn {
	return func(s string) error {
		n, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return errors.New("value must be a number")
		}
		if n >= max {
			return fmt.Errorf("value must be less than %g", max)
		}
		return nil
	}
}

func Between(min, max float64) ValidatorFn {
	return func(s string) error {
		n, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return errors.New("value must be a number")
		}
		if n < min || n > max {
			return fmt.Errorf("value must be between %g and %g", min, max)
		}
		return nil
	}
}

func IsInteger(s string) error {
	if _, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err != nil {
		return errors.New("value must be an integer")
	}
	return nil
}

func IsNumeric(s string) error {
	if _, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err != nil {
		return errors.New("value must be a number")
	}
	return nil
}

func IsEmail(s string) error {
	if _, err := mail.ParseAddress(strings.TrimSpace(s)); err != nil {
		return errors.New("value must be a valid email address")
	}
	return nil
}

func Matches(pattern string) ValidatorFn {
	re := regexp.MustCompile(pattern)
	return func(s string) error {
		if !re.MatchString(s) {
			return fmt.Errorf("value does not match the required format")
		}
		return nil
	}
}

func HasUppercase(s string) error {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return nil
		}
	}
	return errors.New("value must contain at least one uppercase letter")
}

func HasDigit(s string) error {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return nil
		}
	}
	return errors.New("value must contain at least one digit")
}

func Chain(validators ...ValidatorFn) ValidatorFn {
	return func(s string) error {
		for _, v := range validators {
			if err := v(s); err != nil {
				return err
			}
		}
		return nil
	}
}