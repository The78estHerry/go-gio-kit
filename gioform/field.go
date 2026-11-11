package gioform

import (
	"strings"
	"time"
)

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
	// Choice picks one of Field.Choices with radio buttons. Its value is
	// the chosen Option.Value; Field.Default selects one at the start.
	Choice
	// Select picks one of Field.Choices from a drop-down list, for lists
	// too long for radio buttons. Its value is the chosen Option.Value.
	// The arrow keys change the choice; Enter or Space opens the list.
	Select
	// Combo is a text field with a drop-down list of suggestions: typing
	// filters Field.Choices, picking one puts its Value in the field, and
	// any other text is accepted too.
	Combo
	// Duration is a time: arrows step through common values between
	// Field.Min and Field.Max, and a click on the value lets the user type
	// one (see ParseDuration). Field.Default is read with ParseDuration.
	Duration
)

// Option is one option of a Choice, Select or Combo field.
type Option struct {
	Value, Label string
	// Fixed options are built in and can never be removed (see
	// Field.Remove).
	Fixed bool
}

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
	// Choices are the options of Choice, Select and Combo fields. Default
	// is the initial choice, the initial text of a Combo, or the initial
	// time of a Duration ("30s").
	Choices []Option
	Default string

	// Add and Remove, if set, are the labels of buttons next to a Select
	// or Combo field that manage its options, such as "+" and "−".
	//
	// Add in a Combo adds the typed text (after Validate) as an option. In
	// a Select it turns the field into a text box: Enter (or Add again)
	// adds the typed name and selects it, Esc cancels. With AddExternal,
	// Form.AddRequested reports the click instead and the application adds
	// options with SetOptions, for example after its own dialog.
	//
	// Remove removes the option that is the field's current value, unless
	// it is Fixed or CanRemove refuses; then the button is grayed out and
	// hovering it shows the reason with Theme.Icons.Warning. Form.Options
	// and Form.OptionsChanged report the changed list so the application
	// can save it.
	Add, Remove string
	AddExternal bool
	// CanRemove decides whether an option that is not Fixed may be
	// removed, and if not, why (shown to the user). Nil allows it.
	CanRemove func(Option) (ok bool, reason string)
	// ConfirmRemove makes Remove ask first: the button turns into
	// Labels.ConfirmRemove and a second click removes.
	ConfirmRemove bool
	// UndoRemove shows Labels.Removed with an Undo button after a removal,
	// until the next add or remove.
	UndoRemove bool

	// Min and Max limit a Duration; Max 0 means DefaultDurationMax. Steps
	// are the values the arrows step through, by default common times
	// between Min and Max.
	Min, Max time.Duration
	Steps    []time.Duration
}

// Labels holds every word the forms show. DefaultLabels is English;
// replace it to translate the UI.
type Labels struct {
	Submit, Cancel   string
	Show             string
	Confirm          string // label prefix of the NewSecret confirmation field
	Required         string
	Mismatch         string
	Match            string // shown while the NewSecret confirmation matches
	Weak, Fair, Good string // NewSecret strength hints

	// Duration words. The unit formats take one integer (%d).
	// DurationMinutes is used alone ("5 minutes"), DurationMin next to
	// other units ("1 min 30 s"); some languages need both forms.
	DurationNow                                 string
	DurationHours, DurationMinutes, DurationMin string
	DurationSeconds                             string
	DurationTooLong, DurationTooShort           string // take the limit (%s)
	DurationFormat                              string // the text is not a time
	DurationHint                                string // how to type a time, behind "?"
	SelectNone                                  string // a Select with nothing chosen

	// Managing options. ConfirmRemove and Removed take the option label (%s).
	Fixed         string // why a Fixed option cannot be removed
	ConfirmRemove string
	Removed, Undo string
	AddHint       string // placeholder while typing a new Select option
}

// DefaultLabels are the English labels.
var DefaultLabels = Labels{
	Submit: "OK", Cancel: "Cancel", Show: "Show", Confirm: "Confirm",
	Required: "Required", Mismatch: "The two entries do not match",
	Match: "The two entries match",
	Weak:  "Weak: use 4 or more unrelated words, or 12+ characters",
	Fair:  "Fair", Good: "Strong",

	DurationNow: "Now", DurationHours: "%d h", DurationMinutes: "%d min", DurationMin: "%d min",
	DurationSeconds: "%d s",
	DurationTooLong: "At most %s", DurationTooShort: "At least %s",
	DurationFormat: "Not a time. Examples: 90 (minutes), 10s, 3m, 2:30, 1h30m",
	DurationHint: "Click the time to type one: 90 (minutes), 10s, 3m, 3h, 2:30 (h:mm), 1m30s, 2h30m. " +
		"Enter applies, Esc cancels. The arrows step through common times.",
	SelectNone: "Choose…",

	Fixed: "Built in; it cannot be removed", ConfirmRemove: "Remove %s?",
	Removed: "Removed %s", Undo: "Undo", AddHint: "Type a name, then Enter",
}

// Values are the submitted contents of a form.
type Values struct {
	text   map[string]string
	secret map[string]*SecretValue
	dur    map[string]time.Duration
}

// Text returns a text field's value, the chosen value of a Choice or
// Select, or a Duration in Go's form ("1m30s").
func (v Values) Text(key string) string { return v.text[key] }

// Duration returns a Duration field's value.
func (v Values) Duration(key string) time.Duration { return v.dur[key] }

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
