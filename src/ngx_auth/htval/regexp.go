package htval

import (
	"regexp"
)

type HtPattern struct {
	Regexp *regexp.Regexp
}

func (rv *HtPattern) UnmarshalTOML(decode func(interface{}) error) error {
	var pattern string
	if err := decode(&pattern); err != nil {
		return err
	}

	reg, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	*rv = HtPattern{Regexp: reg}
	return nil
}
