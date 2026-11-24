package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`  "d:/LEARN/my-finder/"  `, `D:\LEARN\my-finder`},
		{`d:`, `D:\`},
		{`c:\`, `C:\`},
		{`\\server\share\folder\`, `\\server\share\folder`},
		{``, ``},
	}
	for _, tc := range cases {
		if got := normalizePath(tc.in); got != tc.want {
			t.Errorf("normalizePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestDefaultNameForPath(t *testing.T) {
	if got := defaultNameForPath(`D:\LEARN\my-finder`); got != "my-finder" {
		t.Fatalf("got %q, want my-finder", got)
	}
	if got := defaultNameForPath(`d:\`); got != "Drive D" {
		t.Fatalf("got %q, want Drive D", got)
	}
}

func TestExpandWindowsEnv(t *testing.T) {
	t.Setenv("MYFINDER_TEST_DIR", `D:\TestRoot`)
	got := resolvePath(`%MYFINDER_TEST_DIR%\Sub`)
	if got != `D:\TestRoot\Sub` {
		t.Fatalf("resolvePath = %q, want %q", got, `D:\TestRoot\Sub`)
	}
}

func TestStoreCRUDAndFilter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "shortcuts.json")

	s := newStore(path)
	if err := s.Load(); err != nil {
		t.Fatalf("Load on missing file failed: %v", err)
	}

	a, err := s.Add("Projects", `D:\LEARN\my-finder`)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	b, err := s.Add("", `C:\Users\Public\Downloads`)
	if err != nil {
		t.Fatalf("Add 2 failed: %v", err)
	}
	if b.Name != "Downloads" {
		t.Fatalf("auto name = %q, want Downloads", b.Name)
	}

	// Reload from disk
	s2 := newStore(path)
	if err := s2.Load(); err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if len(s2.Items) != 2 {
		t.Fatalf("len = %d, want 2", len(s2.Items))
	}

	// Filter multi-token
	if idx := s2.Filter("learn finder"); len(idx) != 1 || s2.Items[idx[0]].ID != a.ID {
		t.Fatalf("unexpected filter result: %v", idx)
	}

	// Reorder
	if !s2.Move(b.ID, -1) {
		t.Fatalf("Move up failed")
	}
	if s2.Items[0].ID != b.ID {
		t.Fatalf("expected %s first, got %s", b.ID, s2.Items[0].ID)
	}

	// Update & Delete
	if err := s2.Update(a.ID, "My Finder Repo", `D:\LEARN\my-finder`); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if err := s2.Delete(b.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if len(s2.Items) != 1 || s2.Items[0].Name != "My Finder Repo" {
		t.Fatalf("unexpected final items: %+v", s2.Items)
	}

	// Ensure file exists on disk
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("shortcuts.json not saved: %v", err)
	}
}
