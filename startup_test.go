package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// seedProject writes one session file for the given real working directory.
func seedProject(t *testing.T, home, realDir string) string {
	t.Helper()
	enc := filepath.Join(home, ".claude", "projects", encodePath(realDir))
	if err := os.MkdirAll(enc, 0o755); err != nil {
		t.Fatal(err)
	}
	line := `{"type":"user","cwd":"` + realDir + `","timestamp":"2026-01-01T00:00:00Z",` +
		`"message":{"role":"user","content":"a prompt from ` + realDir + `"}}` + "\n"
	if err := os.WriteFile(filepath.Join(enc, "aaa11111-2222.jsonl"), []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	return enc
}

// boot runs the startup path: load projects, then decide what to show.
func boot(t *testing.T, launchDir string) model {
	t.Helper()
	m := newModel()
	m.launchDir = launchDir
	m.width, m.height = 100, 24
	m.layout()
	msg := loadProjectsCmd()
	mm, cmd := m.Update(msg)
	m = mm.(model)
	// If it decided to open a project, run that load too. startLoad batches the
	// work with the spinner tick, so the result arrives inside a BatchMsg.
	for _, sub := range drain(cmd) {
		if sm, ok := sub.(sessionsLoadedMsg); ok {
			mm, _ = m.Update(sm)
			m = mm.(model)
		}
	}
	return m
}

// drain runs a command and flattens any batch it produces into plain messages.
func drain(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		var out []tea.Msg
		for _, c := range msg {
			out = append(out, drain(c)...)
		}
		return out
	case nil:
		return nil
	default:
		return []tea.Msg{msg}
	}
}

func TestStartAlwaysAsks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/hasone")
	seedProject(t, home, "/tmp/other")

	// Even when the folder has history, it asks instead of jumping straight in.
	m := boot(t, "/tmp/hasone")
	if m.state != viewStart {
		t.Fatalf("state = %v, want the start chooser", m.state)
	}
	if !m.hasLaunch || m.launchProj.RealPath != "/tmp/hasone" {
		t.Errorf("launch project = %+v, want /tmp/hasone", m.launchProj)
	}
	out := stripANSI(m.View())
	t.Logf("WITH SESSIONS\n%s", out)
	for _, want := range []string{"where do you want to start", "this folder", "/tmp/hasone", "1 session", "all projects", "2 projects", "t this folder", "a all projects", "q quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("chooser missing %q", want)
		}
	}
}

func TestStartSaysWhenFolderHasNoSessions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/elsewhere")

	m := boot(t, "/tmp/nothing-here")
	if m.state != viewStart {
		t.Fatalf("state = %v, want the start chooser", m.state)
	}
	if m.hasLaunch {
		t.Error("claimed the empty folder has a project")
	}
	out := stripANSI(m.View())
	t.Logf("WITHOUT SESSIONS\n%s", out)
	for _, want := range []string{"/tmp/nothing-here", "no Claude Code sessions in this folder", "all projects", "1 project"} {
		if !strings.Contains(out, want) {
			t.Errorf("chooser missing %q", want)
		}
	}
	// The default shifts to the only thing worth showing, and the key that has
	// nothing to open is not advertised.
	if !strings.Contains(out, "enter all projects") {
		t.Errorf("enter should default to all projects here:\n%s", out)
	}
	if strings.Contains(out, "t this folder") {
		t.Errorf("footer offers t with nothing to open:\n%s", out)
	}
}

func TestStartTOpensThisFolder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/hasone")
	seedProject(t, home, "/tmp/other")
	m := boot(t, "/tmp/hasone")

	for _, k := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("t")},
		{Type: tea.KeyEnter},
	} {
		mm, cmd := m.handleKey(k)
		got := mm.(model)
		for _, sub := range drain(cmd) {
			if sm, ok := sub.(sessionsLoadedMsg); ok {
				g, _ := got.Update(sm)
				got = g.(model)
			}
		}
		if got.state != viewSessions {
			t.Errorf("%v did not open the folder's sessions (state %v)", k, got.state)
		}
		if got.curProject.RealPath != "/tmp/hasone" {
			t.Errorf("%v opened %q", k, got.curProject.RealPath)
		}
	}
}

func TestStartAOpensAllProjects(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/hasone")
	seedProject(t, home, "/tmp/other")
	m := boot(t, "/tmp/hasone")

	mm, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	got := mm.(model)
	if got.state != viewProjects {
		t.Fatalf("a did not open the projects list (state %v)", got.state)
	}
	out := stripANSI(got.View())
	if !strings.Contains(out, "/tmp/hasone") || !strings.Contains(out, "/tmp/other") {
		t.Errorf("projects list incomplete:\n%s", out)
	}
}

func TestStartEnterFallsBackToAllWhenFolderEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/elsewhere")
	m := boot(t, "/tmp/nothing-here")

	mm, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if got := mm.(model).state; got != viewProjects {
		t.Errorf("enter went to %v, want the projects list", got)
	}
	for _, sub := range drain(cmd) {
		if _, quit := sub.(tea.QuitMsg); quit {
			t.Error("enter quit the app")
		}
	}
	// "t" has nothing to open, so it must do nothing rather than misfire.
	mm, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	if got := mm.(model).state; got != viewStart {
		t.Errorf("t moved to %v even though the folder has no sessions", got)
	}
}

func TestStartQuits(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/hasone")
	m := boot(t, "/tmp/hasone")

	for _, k := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("q")},
		{Type: tea.KeyEsc},
	} {
		_, cmd := m.handleKey(k)
		quit := false
		for _, sub := range drain(cmd) {
			if _, ok := sub.(tea.QuitMsg); ok {
				quit = true
			}
		}
		if !quit {
			t.Errorf("%v did not quit", k)
		}
	}
}

func TestStartIgnoresStrayKeys(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/hasone")
	m := boot(t, "/tmp/hasone")

	for _, k := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("x")},
		{Type: tea.KeyDown},
		{Type: tea.KeyRunes, Runes: []rune("d")},
	} {
		mm, cmd := m.handleKey(k)
		if got := mm.(model).state; got != viewStart {
			t.Errorf("%v moved off the chooser to %v", k, got)
		}
		for _, sub := range drain(cmd) {
			if _, quit := sub.(tea.QuitMsg); quit {
				t.Errorf("%v quit the app", k)
			}
		}
	}
}

func TestStartWithNoSessionsAnywhere(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".claude", "projects"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := boot(t, "/tmp/nothing-here")
	if m.state != viewStart {
		t.Fatalf("state = %v, want the start chooser", m.state)
	}
	out := stripANSI(m.View())
	t.Logf("NOTHING ANYWHERE\n%s", out)
	if !strings.Contains(out, "no Claude Code sessions in this folder") {
		t.Error("should say this folder is empty")
	}
	if !strings.Contains(out, "nothing recorded yet") {
		t.Errorf("should say there is nothing anywhere:\n%s", out)
	}
	mm, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	_ = mm.(model).View() // empty list must still render
}

func TestSessionsScreenStillReachesAllProjects(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/hasone")
	seedProject(t, home, "/tmp/other")
	m := boot(t, "/tmp/hasone")

	mm, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m = mm.(model)
	for _, sub := range drain(cmd) {
		if sm, ok := sub.(sessionsLoadedMsg); ok {
			g, _ := m.Update(sm)
			m = g.(model)
		}
	}
	mm, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if got := mm.(model).state; got != viewProjects {
		t.Errorf("esc from sessions went to %v, want the projects list", got)
	}
}
