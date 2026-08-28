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

func TestStartupInFolderWithSessionsOpensThem(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/hasone")
	seedProject(t, home, "/tmp/other")

	m := boot(t, "/tmp/hasone")
	if m.state != viewSessions {
		t.Fatalf("state = %v, want sessions", m.state)
	}
	if m.curProject.RealPath != "/tmp/hasone" {
		t.Errorf("opened %q, want /tmp/hasone", m.curProject.RealPath)
	}
	if len(m.curSessions) != 1 {
		t.Errorf("%d sessions loaded, want 1", len(m.curSessions))
	}
	out := stripANSI(m.View())
	if !strings.Contains(out, "/tmp/hasone") || !strings.Contains(out, "a prompt from /tmp/hasone") {
		t.Errorf("did not land on the right project:\n%s", out)
	}
	// esc still steps out to the full list.
	mm, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if mm.(model).state != viewProjects {
		t.Error("esc from the auto-opened project did not reach the projects list")
	}
}

func TestStartupInFolderWithoutSessionsAsks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/elsewhere")

	m := boot(t, "/tmp/nothing-here")
	if m.state != viewNoSessionsHere {
		t.Fatalf("state = %v, want the ask screen", m.state)
	}
	out := stripANSI(m.View())
	t.Logf("ASK\n%s", out)
	for _, want := range []string{"No Claude Code sessions here", "/tmp/nothing-here", "Show all 1 project", "y show all projects", "n quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("ask screen missing %q", want)
		}
	}
}

func TestAskScreenYesShowsAll(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/elsewhere")
	m := boot(t, "/tmp/nothing-here")

	for _, yes := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("y")},
		{Type: tea.KeyEnter},
	} {
		mm, cmd := m.handleKey(yes)
		got := mm.(model)
		if got.state != viewProjects {
			t.Errorf("%v did not open the projects list", yes)
		}
		if cmd != nil {
			if _, quit := cmd().(tea.QuitMsg); quit {
				t.Errorf("%v quit instead of showing all", yes)
			}
		}
		if out := stripANSI(got.View()); !strings.Contains(out, "/tmp/elsewhere") {
			t.Errorf("projects list not shown:\n%s", out)
		}
	}
}

func TestAskScreenNoQuits(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/elsewhere")
	m := boot(t, "/tmp/nothing-here")

	for _, k := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("n")},
		{Type: tea.KeyRunes, Runes: []rune("q")},
		{Type: tea.KeyEsc},
	} {
		_, cmd := m.handleKey(k)
		if cmd == nil {
			t.Errorf("%v did not quit", k)
			continue
		}
		if _, quit := cmd().(tea.QuitMsg); !quit {
			t.Errorf("%v did not quit", k)
		}
	}
}

func TestAskScreenIgnoresStrayKeys(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	seedProject(t, home, "/tmp/elsewhere")
	m := boot(t, "/tmp/nothing-here")

	for _, k := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("x")},
		{Type: tea.KeyDown},
		{Type: tea.KeyRunes, Runes: []rune("d")},
	} {
		mm, cmd := m.handleKey(k)
		if got := mm.(model).state; got != viewNoSessionsHere {
			t.Errorf("%v moved off the ask screen to %v", k, got)
		}
		if cmd != nil {
			if _, quit := cmd().(tea.QuitMsg); quit {
				t.Errorf("%v quit the app", k)
			}
		}
	}
}

func TestStartupWithNoSessionsAnywhere(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".claude", "projects"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := boot(t, "/tmp/nothing-here")
	if m.state != viewNoSessionsHere {
		t.Fatalf("state = %v, want the ask screen", m.state)
	}
	out := stripANSI(m.View())
	t.Logf("EMPTY\n%s", out)
	if !strings.Contains(out, "no Claude Code sessions anywhere") {
		t.Errorf("should say there is nothing anywhere:\n%s", out)
	}
	// Saying yes must still render, not panic on an empty list.
	mm, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	_ = mm.(model).View()
}
