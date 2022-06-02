package authz

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"strings"
)

var nameRegexp = regexp.MustCompile(`^[a-z_][0-9a-z_\-]{0,31}$`)

type UserMap struct {
	user map[string]map[string]struct{}
}

func NewUserMap(cfg_file string) (*UserMap, error) {
	bin, err := ioutil.ReadFile(cfg_file)
	if err != nil {
		return nil, err
	}

	umap := map[string]map[string]struct{}{}

	lines := bytes.Split(bin, []byte{'\n'})
	for _, ln := range lines {
		cs := bytes.SplitN(ln, []byte{':'}, 2)
		user := string(cs[0])
		groups := [][]byte{}
		if len(cs) == 2 {
			groups = bytes.Split(cs[1], []byte{' '})
		}

		gmap := map[string]struct{}{}
		for _, g := range groups {
			gmap[string(g)] = struct{}{}
		}
		umap[user] = gmap
	}

	return &UserMap{user: umap}, nil
}

func (az *UserMap) InUser(user string) bool {
	_, ok := az.user[user]

	return ok
}

func (az *UserMap) InGroup(user string, group string) bool {
	gmap, u_ok := az.user[user]
	if !u_ok {
		return false
	}

	_, g_ok := gmap[group]
	return g_ok
}

func (az *UserMap) Authz(tn_str string, user string) bool {
	for _, tn := range strings.Split(tn_str, "|") {
		if az.one_authz(tn, user) {
			return true
		}
	}

	return false
}

func (az *UserMap) one_authz(tn string, user string) bool {
	switch {
	case tn == "":
		return true
	case tn == "!":
		return false
	case !nameRegexp.MatchString(user):
		return false
	case tn == "*":
		return true
	case tn == "@":
		return az.InUser(user)
	case tn[0] == '@':
		return az.InGroup(user, tn[1:])
	case nameRegexp.MatchString(tn):
		return tn == user
	default:
	}

	return false
}

func VerifyAuthzType(tn_str string) bool {
	for _, tn := range strings.Split(tn_str, "|") {
		if !verify_type(tn) {
			return false
		}
	}
	return true
}

func verify_type(tn string) bool {
	switch {
	case tn == "":
		return true
	case tn == "!":
		return true
	case tn == "*":
		return true
	case tn == "@":
		return true
	case tn[0] == '@':
		return nameRegexp.MatchString(tn[1:])
	case nameRegexp.MatchString(tn):
		return true
	default:
	}

	return false
}
