package gioform

import "strings"

// FieldType selects the kind of input.
type FieldType int

const (
	// Text is an ordinary single-line text field.
	Text FieldType = iota
	// Secret is a masked field for tokens and passwords. Its content can be
	// pasted but never copied or cut.
	Secret
	// NewSecret is a Secret with a second confirmation field and a strength
	// hint, for setting a new password. Any Unicode text is accepted.
	NewSecret
)

// Field describes one input of a form.
type Field struct {
	Key         string
	Label       string
	Placeholder string
	Help        string
	Type        FieldType
	Required    bool
	// Trim removes leading and trailing spaces and all line breaks, which
	// copying a token from a web page often adds. Leave it off for
	// passwords that may contain spaces.
	Trim bool
	// Validate is an optional extra check; its error is shown under the field.
	Validate func(string) error
}

// Labels holds every word the forms show. DefaultLabels is English;
// replace it to translate the UI.
type Labels struct {
	Submit, Cancel   string
	Show             string
	Confirm          string // label prefix of the NewSecret confirmation field
	Required         string
	Mismatch         string
	Weak, Fair, Good string // NewSecret strength hints
}

// DefaultLabels are the English labels.
var DefaultLabels = Labels{
	Submit: "OK", Cancel: "Cancel", Show: "Show", Confirm: "Confirm",
	Required: "Required", Mismatch: "The two entries do not match",
	Weak: "Weak: use 4 or more unrelated words, or 12+ characters",
	Fair: "Fair", Good: "Strong",
}

// Values are the submitted contents of a form.
type Values struct {
	text   map[string]string
	secret map[string]*SecretValue
}

// Text returns a text field's value.
func (v Values) Text(key string) string { return v.text[key] }

// Secret returns a secret field's value; call Wipe on it when done.
func (v Values) Secret(key string) *SecretValue { return v.secret[key] }

// Wipe clears every secret value.
func (v Values) Wipe() {
	for _, s := range v.secret {
		s.Wipe()
	}
}

// SecretValue holds a secret as bytes so it can be wiped after use.
//
// Go strings cannot be cleared and the Gio editor keeps its text as a
// string, so wiping is best effort; keep secrets in the OS credential store
// rather than in memory.
type SecretValue struct{ b []byte }

// Bytes returns the secret. The slice is wiped by Wipe.
func (s *SecretValue) Bytes() []byte {
	if s == nil {
		return nil
	}
	return s.b
}

// Len returns the length in bytes.
func (s *SecretValue) Len() int { return len(s.Bytes()) }

// Wipe overwrites the secret with zeros.
func (s *SecretValue) Wipe() {
	if s != nil {
		clear(s.b)
		s.b = nil
	}
}

// trimSecret removes line breaks anywhere and spaces at both ends.
func trimSecret(s string) string {
	s = strings.NewReplacer("\r", "", "\n", "").Replace(s)
	return strings.TrimSpace(s)
}

// Strength rates a new password: 0 weak, 1 fair, 2 strong. It counts
// characters, not bytes, so any Unicode text works; a long phrase scores
// well even in one script.
func Strength(p string) int {
	n := 0
	var lower, upper, digit, other bool
	for _, r := range p {
		n++
		switch {
		case r >= 'a' && r <= 'z':
			lower = true
		case r >= 'A' && r <= 'Z':
			upper = true
		case r >= '0' && r <= '9':
			digit = true
		default:
			other = true
		}
	}
	classes := 0
	for _, c := range []bool{lower, upper, digit, other} {
		if c {
			classes++
		}
	}
	switch {
	case n >= 16, n >= 12 && classes >= 2:
		return 2
	case n >= 8:
		return 1
	}
	return 0
}
