package val_fmt

import (
	"errors"
	"strings"
)

var ErrInvalidInitalValue = errors.New("Invalid initial value")

type ValuesFmt map[rune]any

func New(vals map[rune]any) (ValuesFmt, error) {
	if !ValuesFmt(vals).VerifySyntax() {
		return nil, ErrInvalidInitalValue
	}

	return vals, nil
}

func Must(vals map[rune]any) ValuesFmt {
	res, err := New(vals)
	if err != nil {
		panic("Error: " + err.Error())
	}

	return res
}

func (rfmt ValuesFmt) Must() ValuesFmt {
	if !rfmt.VerifySyntax() {
		panic("Error: " + ErrInvalidInitalValue.Error())
	}

	return rfmt
}

func (rfmt ValuesFmt) VerifySyntax() bool {
	for _, iv := range rfmt {
		switch iv.(type) {
		case string:
		case map[string]string:
		case func(string) string:
		case func(string) (string, error):
		default:
			return false
		}
	}

	return true
}

func (rfmt ValuesFmt) ToString(fstr string) (string, error) {
	var ret_b strings.Builder
	var pkey_b strings.Builder

	var mode rune = 0
	var val_func func(string) (string, error)
	for _, c := range fstr {
		switch mode {
		case '{':
			if c != '{' {
				str, err := val_func("")
				if err != nil {
					return "", err
				}
				ret_b.WriteString(str)
				mode = 0
				continue
			}
			mode = '}'

		case '}':
			if c == '}' {
				str, err := val_func(pkey_b.String())
				if err != nil {
					return "", err
				}
				ret_b.WriteString(str)
				mode = 0
				continue
			}
			pkey_b.WriteRune(c)

		case '$':
			if c == '$' {
				ret_b.WriteRune('$')
				continue
			}

			val, ok := rfmt[c]
			if !ok {
				continue
			}

			switch v := val.(type) {
			case string:
				ret_b.WriteString(v)
				mode = 0
			case map[string]string:
				mode = '{'
				val_func = func(k string) (string, error) { return v[k], nil }
				pkey_b.Reset()
			case func(string) string:
				mode = '{'
				val_func = func(k string) (string, error) { return v(k), nil }
				pkey_b.Reset()
			case func(string) (string, error):
				mode = '{'
				val_func = v
				pkey_b.Reset()
			default:
				panic("bad ValuesFmt")
			}

		default:
			switch c {
			case '$':
				mode = '$'
			default:
				ret_b.WriteRune(c)
			}
		}
	}
	if mode == '}' {
		str, err := val_func(pkey_b.String())
		if err != nil {
			return "", err
		}
		ret_b.WriteString(str)
		mode = 0
	}

	return ret_b.String(), nil
}

func (rfmt ValuesFmt) Verify(fmt_str string) bool {
	var mode rune = 0
	for _, c := range fmt_str {
		switch mode {
		case '{':
			if c != '{' {
				return false
			}
			mode = '}'

		case '}':
			if c == '}' {
				mode = 0
				continue
			}

		case '$':
			if c == '$' {
				continue
			}

			val, ok := rfmt[c]
			if !ok {
				continue
			}

			switch val.(type) {
			case string:
				mode = 0
			case map[string]string:
				mode = '{'
			case func(string) string:
				mode = '{'
			case func(string) (string, error):
				mode = '{'
			default:
				return false
			}

		default:
			switch c {
			case '$':
				mode = '$'
			default:
			}
		}
	}

	if mode != 0 {
		return false
	}

	return true
}
