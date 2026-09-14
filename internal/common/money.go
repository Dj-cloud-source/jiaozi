package common

import "regexp"

var pricePattern = regexp.MustCompile(`^[0-9]+(\.[0-9]{1,2})?$`)

func IsMoney(value string) bool {
	return pricePattern.MatchString(value)
}
