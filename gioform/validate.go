package gioform

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// MinLength returns a validator that requires at least n characters.
func MinLength(n int, msg string) func(string) error {
	return func(s string) error {
		if utf8.RuneCountInString(s) < n {
			if msg == "" {
				msg = fmt.Sprintf("At least %d characters", n)
			}
			return errors.New(msg)
		}
		return nil
	}
}

// Host returns a validator for a host name with an optional port, such as
// gitea.example.com or 192.0.2.1:3000. A scheme or path is rejected.
func Host(msg string) func(string) error {
	if msg == "" {
		msg = "Enter a host name such as git.example.com"
	}
	return func(s string) error {
		if s == "" {
			return nil // emptiness is Required's job
		}
		if strings.ContainsAny(s, "/ \t@") || strings.Contains(s, "://") || strings.HasPrefix(s, ".") || strings.HasSuffix(s, ".") {
			return errors.New(msg)
		}
		host, port, hasPort := strings.Cut(s, ":")
		if host == "" || (hasPort && (port == "" || strings.Trim(port, "0123456789") != "")) {
			return errors.New(msg)
		}
		return nil
	}
}
