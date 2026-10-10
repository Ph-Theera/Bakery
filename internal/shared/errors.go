package shared

import (
	"errors"
	"strings"
)

var ErrInvalid = errors.New("invalid input")
var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
var ErrForbidden = errors.New("forbidden")

func UUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return false
		}
	}
	return true
}
