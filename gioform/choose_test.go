package gioform

import (
	"slices"
	"testing"

	"gioui.org/layout"
)

var services = []Option{
	{Value: "github", Label: "GitHub"}, {Value: "gitea", Label: "Gitea"}, {Value: "gitlab", Label: "GitLab"}, {Value: "forgejo", Label: "Forgejo"}, {Value: "bitbucket", Label: "Bitbucket"},
}

func TestFilterOptions(t *testing.T) {
	for q, want := range map[string][]int{
		"":     {0, 1, 2, 3, 4},
		"  ":   {0, 1, 2, 3, 4},
		"git":  {0, 1, 2},
		"GIT":  {0, 1, 2},
		"lab":  {2},
		"jo":   {3},
		"none": {},
	} {
		if got := filterOptions(services, q); !slices.Equal(got, want) {
			t.Errorf("filterOptions(%q) = %v, want %v", q, got, want)
		}
	}
}

func TestMoveHighlight(t *testing.T) {
	for _, c := range []struct{ hl, n, dir, want int }{
		{-1, 3, 1, 0}, {-1, 3, -1, 0}, {0, 3, 1, 1}, {2, 3, 1, 2}, {0, 3, -1, 0}, {1, 0, 1, -1},
	} {
		if got := moveHighlight(c.hl, c.n, c.dir); got != c.want {
			t.Errorf("moveHighlight(%d, %d, %d) = %d, want %d", c.hl, c.n, c.dir, got, c.want)
		}
	}
}

func TestSelectAndCombo(t *testing.T) {
	f := New(NewTheme(false), []Field{
		{Key: "service", Label: "Service", Type: Select, Required: true, Choices: services, Add: "+ Add"},
		{Key: "host", Label: "Server", Type: Combo, Required: true, Default: "git.example.com",
			Choices: []Option{{Value: "git.example.com", Label: "git.example.com"}, {Value: "code.example.com", Label: "code.example.com"}}},
	})
	if f.validate() {
		t.Fatal("a required Select without a choice was accepted")
	}
	f.SetText("service", "gitlab")
	if f.Choice("service") != "gitlab" {
		t.Fatalf("Select choice = %q", f.Choice("service"))
	}

	sel := f.field("service")
	sel.openList(false)
	if !sel.open || sel.hl != 2 {
		t.Fatalf("open list highlights the choice: open %v hl %d", sel.open, sel.hl)
	}

	combo := f.field("host")
	if combo.open {
		t.Fatal("the default text opened the list")
	}
	combo.ed.SetText("code")
	combo.comboChanged()
	if !combo.shown() || !slices.Equal(combo.visible(), []int{1}) || combo.hl != -1 {
		t.Fatalf("typing should filter: shown %v visible %v hl %d", combo.shown(), combo.visible(), combo.hl)
	}
	combo.ed.SetText("free text")
	combo.comboChanged()
	if combo.shown() {
		t.Fatal("no match should show no list")
	}

	f.trySubmit()
	v, ok := f.Submitted()
	if !ok || v.Text("service") != "gitlab" || v.Text("host") != "free text" {
		t.Fatalf("submitted %v %q %q", ok, v.Text("service"), v.Text("host"))
	}

	if f.AddRequested("service") {
		t.Fatal("AddRequested without a click")
	}
	sel.addReq = true
	if !f.AddRequested("service") || f.AddRequested("service") {
		t.Fatal("AddRequested should report a click once")
	}

	f.Reset()
	if f.Choice("service") != "" || combo.ed.Text() != "git.example.com" || sel.open {
		t.Fatalf("Reset: %q %q %v", f.Choice("service"), combo.ed.Text(), sel.open)
	}
}

func TestManageOptions(t *testing.T) {
	inUse := "code.example.com"
	f := New(NewTheme(false), []Field{
		{Key: "host", Label: "Server", Type: Combo, Add: "+", Remove: "−", UndoRemove: true,
			Validate: Host(""),
			Choices: []Option{
				{Value: "git.example.com", Label: "git.example.com", Fixed: true},
				{Value: "code.example.com", Label: "code.example.com"},
				{Value: "old.example.com", Label: "old.example.com"},
			},
			CanRemove: func(o Option) (bool, string) {
				if o.Value == inUse {
					return false, "A saved token uses this server"
				}
				return true, ""
			}},
		{Key: "service", Label: "Service", Type: Select, Remove: "−", ConfirmRemove: true,
			Choices: []Option{{Value: "github", Label: "GitHub"}, {Value: "gitea", Label: "Gitea"}}},
	})
	host := f.field("host")
	values := func(fs *fieldState) []string {
		var v []string
		for _, o := range fs.Choices {
			v = append(v, o.Value)
		}
		return v
	}

	// Add the typed text; invalid or existing text is not added.
	host.ed.SetText("https://bad")
	f.addTyped(layout.Context{}, host)
	if host.err == "" || len(host.Choices) != 3 {
		t.Fatalf("invalid host added: err %q, %v", host.err, values(host))
	}
	host.ed.SetText(" new.example.com ")
	if !host.canAdd() {
		t.Fatal("new text cannot be added")
	}
	f.addTyped(layout.Context{}, host)
	if !slices.Equal(values(host), []string{"git.example.com", "code.example.com", "old.example.com", "new.example.com"}) ||
		host.ed.Text() != "new.example.com" || len(host.items) != 4 {
		t.Fatalf("after add: %v %q", values(host), host.ed.Text())
	}
	if host.canAdd() {
		t.Fatal("an existing option can be added again")
	}
	if !f.OptionsChanged("host") || f.OptionsChanged("host") {
		t.Fatal("OptionsChanged should report once")
	}

	// Fixed, refused and free options.
	for _, c := range []struct {
		text, reason string
		ok           bool
	}{
		{"git.example.com", DefaultLabels.Fixed, false},
		{"code.example.com", "A saved token uses this server", false},
		{"old.example.com", "", true},
		{"not.in.list", "", false},
	} {
		host.ed.SetText(c.text)
		if _, ok, reason := f.removable(host); ok != c.ok || reason != c.reason {
			t.Errorf("removable(%q) = %v %q", c.text, ok, reason)
		}
	}

	// Remove and undo.
	host.ed.SetText("old.example.com")
	i, _, _ := f.removable(host)
	f.removeOption(host, i)
	if slices.Contains(values(host), "old.example.com") || host.ed.Text() != "" || host.removed == nil {
		t.Fatalf("after remove: %v %q", values(host), host.ed.Text())
	}
	f.undoRemove(host)
	if !slices.Equal(values(host), []string{"git.example.com", "code.example.com", "old.example.com", "new.example.com"}) ||
		host.ed.Text() != "old.example.com" || host.removed != nil {
		t.Fatalf("after undo: %v %q", values(host), host.ed.Text())
	}
	if !f.OptionsChanged("host") {
		t.Fatal("undo did not report a change")
	}

	// SetOptions from the application is not a user change and clears a
	// Select value that is gone.
	f.SetText("service", "gitea")
	f.SetOptions("service", []Option{{Value: "github", Label: "GitHub"}})
	if f.OptionsChanged("service") || f.Choice("service") != "" || len(f.Options("service")) != 1 {
		t.Fatalf("SetOptions: changed %v choice %q", f.OptionsChanged("service"), f.Choice("service"))
	}
	// The caller's slice is never changed.
	opts := f.Options("service")
	opts[0].Label = "changed"
	if f.Options("service")[0].Label != "GitHub" {
		t.Fatal("Options returned the internal slice")
	}
}

func TestSelectAdd(t *testing.T) {
	f := New(NewTheme(false), []Field{
		{Key: "service", Label: "Service", Type: Select, Add: "+",
			Choices: []Option{{Value: "github", Label: "GitHub"}}},
		{Key: "ext", Label: "External", Type: Select, Add: "+", AddExternal: true},
	})
	var gtx layout.Context
	fs := f.field("service")
	fs.adding = true
	fs.ed.SetText(" codeberg ")
	f.addTyped(gtx, fs)
	if fs.adding || f.Choice("service") != "codeberg" || len(f.Options("service")) != 2 || !f.OptionsChanged("service") {
		t.Fatalf("after add: adding %v choice %q options %v", fs.adding, f.Choice("service"), f.Options("service"))
	}
	// Typing an existing name selects it without adding.
	fs.adding = true
	fs.ed.SetText("github")
	f.addTyped(gtx, fs)
	if f.Choice("service") != "github" || len(f.Options("service")) != 2 || f.OptionsChanged("service") {
		t.Fatalf("existing name: choice %q options %v", f.Choice("service"), f.Options("service"))
	}
	// Empty text just closes the box.
	fs.adding = true
	fs.ed.SetText("")
	f.addTyped(gtx, fs)
	if fs.adding || f.Choice("service") != "github" {
		t.Fatal("empty add changed the field")
	}
	if f.AddRequested("ext") {
		t.Fatal("AddRequested without a click")
	}
}
