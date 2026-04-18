package utils

import "regexp"

func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]*@[a-zA-Z]+(?:\\.[a-zA-Z]+)*$")
	return emailRegex.MatchString(email)
}
