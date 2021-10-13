package utils

import "regexp"

const (
	Regexp_pattern_email = `^([^@]+)@[^\.]+\.([^\.]+)$`
)

func IsEmail(v string) bool { ok, _ := regexp.MatchString(`^([^@]+)@[^\.]+\.([^\.]+)$`, v); return ok }
