package main

import (
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
)

// EffectiveParam searches for the effective value.
// First among the POST fields.
// Then among the URL "path" parameters.
// Then among the URL GET parameters.
// Then inside the session.
// It might be smarter, to condense all levels down to session level
// at the begin of each request.
// We then would only ask the session and flash messages.
func EffectiveParam(ctx iris.Context, key string, defaultVal ...string) string {
	// Form data and url query parameters for POST or PUT HTTP methods.
	if v := ctx.FormValue(key); v != "" {
		return v
	}

	// Path Param.
	if v := ctx.Params().Get(key); v != "" {
		return v
	}

	// URL Get Param.
	if v := ctx.URLParam(key); v != "" {
		return v
	}

	// Session.
	sess := sessions.Get(ctx)
	if sess != nil {
		if v := sess.GetString(key); v != "" {
			return v
		}

		if v := sess.GetFlashString(key); v != "" {
			return v
		}
	}

	def := ""
	if len(defaultVal) > 0 {
		def = defaultVal[0]
	}

	return def
}

// EffectiveParamInt is a wrapper around EffectiveParam
// with subsequent parsing into an int
func EffectiveParamInt(c iris.Context, key string, defaultVal ...int) int {
	s := EffectiveParam(c, key)
	if s == "" {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0

	}
	i, _ := strconv.Atoi(s)
	return i
}

// EffectiveParamFloat is a wrapper around EffectiveParam
// with subsequent parsing into float
func EffectiveParamFloat(c iris.Context, key string, defaultVal ...float64) (float64, error) {
	s := EffectiveParam(c, key)
	if s == "" {
		if len(defaultVal) > 0 {
			return defaultVal[0], nil
		}
		return 0.0, nil

	}

	fl, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, err
	}
	return fl, nil

}
