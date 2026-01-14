package http_auth

import (
	"errors"
	"net/url"
	"path"
	"strings"
)

const UNIX_SOCKET_URL_PREFIX = "unix:"

type HtAuthAddr struct {
	Url  *url.URL
	Unix *HtUnixAddr
}

var ErrInvalidUrl = errors.New("invalid URL")
var ErrUnsupportedUrlScheme = errors.New("unsupported URL scheme")
var ErrUnsupportedUrlUserinfo = errors.New("unsupported URL userinfo")

func NewAuthAddr(url_str string) (*HtAuthAddr, error) {
	var aurl *url.URL
	var aunix *HtUnixAddr = nil

	switch {
	case strings.HasPrefix(url_str, "http:"):
		fallthrough
	case strings.HasPrefix(url_str, "https:"):
		u, e := url.Parse(url_str)
		if e != nil {
			return nil, e
		}
		if u.Host == "" {
			return nil, ErrInvalidUrl
		}

		if u.User != nil {
			return nil, ErrUnsupportedUrlUserinfo
		}

		aurl = u
		aunix = nil

	case strings.HasPrefix(url_str, UNIX_SOCKET_URL_PREFIX):
		ux, err := NewHtUnixAddr(url_str)
		if err != nil {
			return nil, err
		}
		if ux.User != nil {
			return nil, ErrUnsupportedUrlUserinfo
		}

		u, e := url.Parse("http://localhost/")
		if e != nil {
			panic("must parse: " + e.Error())
		}
		u.Path = ux.Path

		aurl = u
		aunix = ux
	default:
		return nil, ErrUnsupportedUrlScheme
	}

	return &HtAuthAddr{
		Url:  aurl,
		Unix: aunix,
	}, nil
}

func (aa *HtAuthAddr) IsZero() bool {
	return aa.Url == nil
}

func (aa *HtAuthAddr) UnmarshalTOML(decode func(interface{}) error) error {
	var url_str string
	if err := decode(&url_str); err != nil {
		return err
	}

	new_aa, err := NewAuthAddr(url_str)
	if err != nil {
		return err
	}

	*aa = *new_aa
	return nil
}

type HtUnixAddr struct {
	User   *url.Userinfo
	Socket string
	Path   string
}

func NewHtUnixAddr(url_str string) (*HtUnixAddr, error) {
	if !strings.HasPrefix(url_str, UNIX_SOCKET_URL_PREFIX) {
		return nil, ErrUnsupportedUrlScheme
	}
	pm_str := url_str[len(UNIX_SOCKET_URL_PREFIX):]

	var userinfo *url.Userinfo = nil
	if ui_str, af, found := strings.Cut(pm_str, "@"); found {
		pm_str = af
		ui, err := url.PathUnescape(ui_str)
		if err != nil {
			return nil, ErrInvalidUrl
		}

		u, p, has_pw := strings.Cut(ui, ":")
		if has_pw {
			userinfo = url.UserPassword(u, p)
		} else {
			userinfo = url.User(u)
		}
	}

	pm := strings.SplitN(pm_str, ":", 3)
	if len(pm) > 2 {
		return nil, ErrInvalidUrl
	}

	sock_str, err := url.PathUnescape(pm[0])
	if err != nil {
		return nil, ErrInvalidUrl
	}

	path_str := "/"
	if len(pm) >= 2 {
		var err error
		path_str, err = url.PathUnescape(pm[1])
		if err != nil {
			return nil, ErrInvalidUrl
		}
		if path_str == "" || path_str[0] != '/' {
			return nil, ErrInvalidUrl
		}
	}
	path_str = path.Clean("/" + path_str)

	return &HtUnixAddr{
		Socket: sock_str,
		Path:   path_str,
		User:   userinfo,
	}, nil
}
