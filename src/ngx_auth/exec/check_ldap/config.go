package main

import (
	"os"

	"github.com/naoina/toml"
)

type NgxLdapAuthConfig struct {
	SocketType string
	SocketPath string
	AuthRealm  string `toml:",omitempty"`

	HostUrl        string
	StartTls       int      `toml:",omitempty"`
	SkipCertVerify int      `toml:",omitempty"`
	RootCaFiles    []string `toml:",omitempty"`
	BaseDn         string
	BindDn         string
	UniqFilter     string `toml:",omitempty"`
	Timeout        int    `toml:",omitempty"`
}

type NgxLdapPathAuthConfig struct {
	SocketType string
	SocketPath string
	AuthRealm  string `toml:",omitempty"`
	PathHeader string `toml:",omitempty"`

	Ldap struct {
		HostUrl        string
		StartTls       int      `toml:",omitempty"`
		SkipCertVerify int      `toml:",omitempty"`
		RootCaFiles    []string `toml:",omitempty"`
		BaseDn         string
		BindDn         string
		UniqFilter     string `toml:",omitempty"`
		Timeout        int    `toml:",omitempty"`
	}

	Authz struct {
		UserMap      string
		PathPattern  string
		NomatchRight string            `toml:",omitempty"`
		DefaultRight string            `toml:",omitempty"`
		PathRight    map[string]string `toml:",omitempty"`
	}
}

func load_ldap_auth_config(file string) (*NgxLdapAuthConfig, error) {
	cfg_f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer cfg_f.Close()

	cfg := &NgxLdapAuthConfig{}
	if err := toml.NewDecoder(cfg_f).Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func load_ldap_path_auth_config(file string) (*NgxLdapAuthConfig, error) {
	cfg_f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer cfg_f.Close()

	raw_cfg := &NgxLdapPathAuthConfig{}
	if err := toml.NewDecoder(cfg_f).Decode(&raw_cfg); err != nil {
		return nil, err
	}

	cfg := &NgxLdapAuthConfig{
		SocketType: raw_cfg.SocketType,
		SocketPath: raw_cfg.SocketPath,
		AuthRealm:  raw_cfg.AuthRealm,

		HostUrl:        raw_cfg.Ldap.HostUrl,
		StartTls:       raw_cfg.Ldap.StartTls,
		SkipCertVerify: raw_cfg.Ldap.SkipCertVerify,
		RootCaFiles:    raw_cfg.Ldap.RootCaFiles,
		BaseDn:         raw_cfg.Ldap.BaseDn,
		BindDn:         raw_cfg.Ldap.BindDn,
		UniqFilter:     raw_cfg.Ldap.UniqFilter,
		Timeout:        raw_cfg.Ldap.Timeout,
	}

	return cfg, nil
}

func load_config(file string) (*NgxLdapAuthConfig, error) {
	cfg, err := load_ldap_auth_config(file)
	if err == nil {
		return cfg, nil
	}

	if os.IsNotExist(err) {
		return nil, err
	}

	return load_ldap_path_auth_config(file)
}
