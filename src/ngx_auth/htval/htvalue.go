package htval

import (
	"errors"
	"net/http"
	"strings"

	"ngx_auth/val_fmt"
)

var ErrInvalidValueParameter = errors.New("Invalud HTTP value parameter")

type HtValueFormat map[rune]any

func NewHtValueArgs(user, pass string, header http.Header, matchs []string) HtValueFormat {
	trans_tbl := map[rune]any{
		'u': user,
		'p': pass,
		'h': func(name string) string {
			return header.Get(name)
		},
	}
	for i, m := range matchs {
		if i > 9 {
			break
		}
		trans_tbl[rune('0'+i)] = m
	}

	return trans_tbl
}

type HtValue string

func NewHtValue(str string) (HtValue, error) {
	if err := verify_src(str); err != nil {
		return "", err
	}
	return HtValue(str), nil
}

var testHtValueFormat = val_fmt.Must(map[rune]any{
	'0': "n",
	'1': "n",
	'2': "n",
	'3': "n",
	'4': "n",
	'5': "n",
	'6': "n",
	'7': "n",
	'8': "n",
	'9': "n",
	'u': "name",
	'p': "pass",
	'h': func(h string) (string, error) {
		if strings.ContainsFunc(h, not_header_char) {
			return "", ErrInvalidHeader
		}
		return "head", nil
	},
})

func verify_src(str string) error {
	if _, err := testHtValueFormat.ToString(str); err != nil {
		return err
	}

	if strings.ContainsAny(str, "\n\x00") {
		return ErrInvalidValueParameter
	}

	return nil
}

func (hv *HtValue) ToString(hv_fmt HtValueFormat) string {
	vfmt := val_fmt.ValuesFmt(hv_fmt)
	str, err := vfmt.ToString(string(*hv))
	if err != nil {
		panic("Unexpected error: " + err.Error())
	}
	return str
}

func (hv *HtValue) UnmarshalTOML(decode func(interface{}) error) error {
	var str string
	if err := decode(&str); err != nil {
		return err
	}

	val, err := NewHtValue(str)
	if err != nil {
		return err
	}

	*hv = val
	return nil
}
