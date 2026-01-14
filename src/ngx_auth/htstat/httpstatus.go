package htstat

import (
	"net/http"
)

type HttpStatusTbl struct {
	Ok             HttpStatusMsg `toml:",omitempty"`
	Unauth         HttpStatusMsg `toml:",omitempty"`
	Forbidden      HttpStatusMsg `toml:",omitempty"`
	Nopath         HttpStatusMsg `toml:",omitempty"`
	Nouser         HttpStatusMsg `toml:",omitempty"`
	BadRequest     HttpStatusMsg `toml:",omitempty"`
	BadGateway     HttpStatusMsg `toml:",omitempty"`
	Timeout        HttpStatusMsg `toml:",omitempty"`
	InvalidSetting HttpStatusMsg `toml:",omitempty"`
}

func (st *HttpStatusTbl) SetDefault() {
	st.Ok.SetDefault(http.StatusOK, "Authorized")
	st.Unauth.SetDefault(http.StatusUnauthorized, "Not authenticated")
	st.Forbidden.SetDefault(http.StatusForbidden, "Forbidden")
	st.Nopath.SetDefault(http.StatusForbidden, "No path header")
	st.Nouser.SetDefault(http.StatusForbidden, "No user header")
	st.BadRequest.SetDefault(http.StatusBadRequest, "Bad request")
	st.BadGateway.SetDefault(http.StatusBadGateway, "Bad gateway")
	st.Timeout.SetDefault(http.StatusGatewayTimeout, "Timeout")
	st.InvalidSetting.SetDefault(http.StatusInternalServerError, "Invalid setting")
}

func (st *HttpStatusTbl) SetDefaultByTbl(def *HttpStatusTbl) {
	st.Ok.SetDefault(def.Ok.Code, def.Ok.Message)
	st.Unauth.SetDefault(def.Unauth.Code, def.Unauth.Message)
	st.Forbidden.SetDefault(def.Forbidden.Code, def.Forbidden.Message)
	st.Nopath.SetDefault(def.Nopath.Code, def.Nopath.Message)
	st.Nouser.SetDefault(def.Nouser.Code, def.Nouser.Message)
	st.BadRequest.SetDefault(def.BadRequest.Code, def.BadRequest.Message)
	st.BadGateway.SetDefault(def.BadGateway.Code, def.BadGateway.Message)
	st.Timeout.SetDefault(def.Timeout.Code, def.Timeout.Message)
	st.InvalidSetting.SetDefault(def.InvalidSetting.Code, def.InvalidSetting.Message)
}

func (st *HttpStatusTbl) IsValid() bool {
	return st.Ok.IsValid() &&
		st.Unauth.IsValid() &&
		st.Forbidden.IsValid() &&
		st.Nopath.IsValid() &&
		st.Nouser.IsValid() &&
		st.BadRequest.IsValid() &&
		st.BadGateway.IsValid() &&
		st.Timeout.IsValid() &&
		st.InvalidSetting.IsValid()
}

type HttpStatusMsg struct {
	Code    int    `toml:",omitempty"`
	Message string `toml:",omitempty"`
}

func (em *HttpStatusMsg) IsValid() bool {
	return em.Code >= 100 && em.Code < 600
}

func (em *HttpStatusMsg) SetDefault(code int, msg string) {
	if em.Code == 0 {
		em.Code = code
	}
	if em.Message == "" {
		em.Message = msg
	}
}

func (em *HttpStatusMsg) Error(w http.ResponseWriter) {
	http.Error(w, em.Message, em.Code)
}
