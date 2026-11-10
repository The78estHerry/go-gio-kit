package gioform

import (
	"errors"
	"testing"
)

func TestTrimSecret(t *testing.T) {
	for in, want := range map[string]string{
		"  test-token-not-real \r\n": "test-token-not-real",
		"test-\ntoken":               "test-token",
		"pass phrase":                "pass phrase",
	} {
		if got := trimSecret(in); got != want {
			t.Errorf("trimSecret(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStrength(t *testing.T) {
	for p, want := range map[string]int{
		"abc":              0,
		"abcdefgh":         1,
		"abcdefgh1234":     2,
		"abcdefghijklmnop": 2,
		"我的密碼是很長的一句話":      1,
		"今天天氣很好我們去公園散步然後吃冰淇淋": 2,
	} {
		if got := Strength(p); got != want {
			t.Errorf("Strength(%q) = %d, want %d", p, got, want)
		}
	}
}

func TestMode(t *testing.T) {
	if !System.Dark(true) || System.Dark(false) || !Dark.Dark(false) || Light.Dark(true) {
		t.Fatal("Mode.Dark is wrong")
	}
}

func TestHost(t *testing.T) {
	v := Host("")
	for _, ok := range []string{"", "git.example.com", "192.0.2.1:3000", "localhost"} {
		if err := v(ok); err != nil {
			t.Errorf("Host(%q): %v", ok, err)
		}
	}
	for _, bad := range []string{"https://git.example.com", "git.example.com/x", "a b", ".example.com", "host:", "host:abc", "u@host"} {
		if v(bad) == nil {
			t.Errorf("Host(%q) accepted", bad)
		}
	}
}

func TestSecretValue(t *testing.T) {
	s := &SecretValue{b: []byte("test-token-not-real")}
	b := s.Bytes()
	s.Wipe()
	for _, c := range b {
		if c != 0 {
			t.Fatal("Wipe left data in the returned slice")
		}
	}
	if s.Len() != 0 || (*SecretValue)(nil).Len() != 0 {
		t.Fatal("Len after Wipe")
	}
}

func newTestForm() *Form {
	return New(NewTheme(true), []Field{
		{Key: "host", Label: "Server", Required: true, Validate: Host("")},
		{Key: "token", Label: "Token", Type: Secret, Required: true, Trim: true},
		{Key: "pw", Label: "Password", Type: NewSecret},
		{Key: "note", Label: "Note", Validate: func(s string) error { return errors.New("bad note") }},
	})
}

func TestValidateAndSubmit(t *testing.T) {
	f := newTestForm()
	f.trySubmit()
	if _, ok := f.Submitted(); ok {
		t.Fatal("empty required fields submitted")
	}
	if f.field("host").err != DefaultLabels.Required || f.field("token").err != DefaultLabels.Required {
		t.Fatal("missing Required messages")
	}

	f.SetText("host", "git.example.com")
	f.field("token").ed.SetText("test-token-not-real")
	f.field("pw").ed.SetText("一句很長的密碼")
	f.field("pw").confirm.SetText("一句不同的密碼")
	f.trySubmit()
	if f.field("pw").err != DefaultLabels.Mismatch {
		t.Fatalf("mismatch not reported: %q", f.field("pw").err)
	}

	f.field("pw").confirm.SetText("一句很長的密碼")
	f.SetText("note", "x")
	f.trySubmit()
	if f.field("note").err != "bad note" {
		t.Fatal("custom validator ignored")
	}

	f.SetText("note", "")
	f.trySubmit()
	v, ok := f.Submitted()
	if !ok {
		t.Fatalf("valid form not submitted: %+v", f.field("note").err)
	}
	if _, again := f.Submitted(); again {
		t.Fatal("Submitted returned twice")
	}
	if v.Text("host") != "git.example.com" || string(v.Secret("token").Bytes()) != "test-token-not-real" ||
		string(v.Secret("pw").Bytes()) != "一句很長的密碼" {
		t.Fatal("wrong values")
	}
	if v.Text("token") != "" {
		t.Fatal("secret leaked into text values")
	}

	// SetText never fills a secret field.
	f.SetText("token", "other")
	if f.field("token").ed.Text() != "test-token-not-real" {
		t.Fatal("SetText changed a secret field")
	}

	f.Reset()
	for _, fs := range f.fields {
		if fs.ed.Text() != "" || fs.confirm.Text() != "" || fs.err != "" {
			t.Fatalf("Reset left %q", fs.Key)
		}
	}
}

func TestConfirmState(t *testing.T) {
	for _, c := range []struct {
		secret, confirm string
		want            int
	}{
		{"correct horse", "", confirmPending},
		{"correct horse", "corr", confirmPending},
		{"correct horse", "correct horse", confirmMatch},
		{"correct horse", "correct house", confirmMismatch},
		{"correct horse", "correct horse!", confirmMismatch},
		{"", "x", confirmMismatch},
	} {
		if got := confirmState(c.secret, c.confirm); got != c.want {
			t.Errorf("confirmState(%q, %q) = %d, want %d", c.secret, c.confirm, got, c.want)
		}
	}
}

func TestChoice(t *testing.T) {
	f := New(NewTheme(false), []Field{
		{Key: "plugin", Label: "Service", Type: Choice, Required: true, Default: "github",
			Choices: []Option{{"github", "GitHub"}, {"gitea", "Gitea"}}},
		{Key: "kind", Label: "Kind", Type: Choice, Required: true,
			Choices: []Option{{"a", "A"}, {"b", "B"}}},
	})
	if got := f.Choice("plugin"); got != "github" {
		t.Errorf("default choice = %q", got)
	}
	f.SetText("plugin", "gitea")
	f.SetText("plugin", "unknown") // not a choice: ignored
	if got := f.Choice("plugin"); got != "gitea" {
		t.Errorf("after SetText = %q", got)
	}
	if f.validate() {
		t.Fatal("a required choice without a default was accepted")
	}
	f.SetText("kind", "b")
	f.trySubmit()
	v, ok := f.Submitted()
	if !ok || v.Text("plugin") != "gitea" || v.Text("kind") != "b" {
		t.Fatalf("submitted %v %q %q", ok, v.Text("plugin"), v.Text("kind"))
	}
	f.Reset()
	if f.Choice("plugin") != "github" || f.Choice("kind") != "" {
		t.Errorf("Reset did not restore defaults: %q %q", f.Choice("plugin"), f.Choice("kind"))
	}
	if f.Choice("missing") != "" {
		t.Error("Choice of an unknown key")
	}
}
