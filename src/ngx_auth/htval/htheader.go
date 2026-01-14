package htval

import (
	"errors"
	"net/http"
	"strings"
)

var ErrDuplicateSetHeaderKey = errors.New("duplicate HTTP set_header table key")
var ErrInvalidHeader = errors.New("Invalid HTTP header name")
var ErrInvalidSetHeaderKey = errors.New("Invalid HTTP set_header table key")

type HtHeader string
type HtSetHeader map[HtHeader]HtValue

func not_header_char(r rune) bool {
	return !(r != ':' && r >= '!' && r < '~')
}

func NewHtHeader(str string) (HtHeader, error) {
	if strings.ContainsFunc(str, not_header_char) {
		return "", ErrInvalidHeader
	}

	return HtHeader(http.CanonicalHeaderKey(str)), nil
}

func (hh *HtHeader) UnmarshalTOML(decode func(interface{}) error) error {
	var str string
	if err := decode(&str); err != nil {
		return err
	}

	new_hh, err := NewHtHeader(str)
	if err != nil {
		return err
	}

	*hh = new_hh
	return nil
}

func (hf *HtSetHeader) UnmarshalTOML(decode func(interface{}) error) error {
	var raw map[string]HtValue

	if err := decode(&raw); err != nil {
		return err
	}

	new_hf := HtSetHeader{}
	for k, v := range raw {
		hk, err := NewHtHeader(k)
		if err != nil {
			return ErrInvalidSetHeaderKey
		}
		if _, ok := new_hf[hk]; ok {
			return ErrDuplicateSetHeaderKey
		}

		new_hf[hk] = v
	}

	*hf = new_hf
	return nil
}
