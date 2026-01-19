package plugstep

import (
	"testing"
)

func TestCreatePlugstep_SetsFields(t *testing.T) {
	args := []string{"install", "--verbose"}
	serverDir := "/path/to/server"

	ps := CreatePlugstep(args, serverDir)

	if ps == nil {
		t.Fatal("expected non-nil Plugstep")
	}
	if len(ps.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(ps.Args))
	}
	if ps.Args[0] != "install" {
		t.Errorf("expected first arg %q, got %q", "install", ps.Args[0])
	}
	if ps.ServerDirectory != serverDir {
		t.Errorf("expected server directory %q, got %q", serverDir, ps.ServerDirectory)
	}
	if ps.Config != nil {
		t.Error("expected nil config before Init()")
	}
}

func TestCreatePlugstep_EmptyArgs(t *testing.T) {
	ps := CreatePlugstep([]string{}, "/server")

	if ps == nil {
		t.Fatal("expected non-nil Plugstep")
	}
	if len(ps.Args) != 0 {
		t.Errorf("expected 0 args, got %d", len(ps.Args))
	}
}

func TestCreatePlugstep_NilArgs(t *testing.T) {
	ps := CreatePlugstep(nil, "/server")

	if ps == nil {
		t.Fatal("expected non-nil Plugstep")
	}
	if ps.Args != nil {
		t.Error("expected nil args to remain nil")
	}
}

func TestCreatePlugstep_EmptyServerDirectory(t *testing.T) {
	ps := CreatePlugstep([]string{}, "")

	if ps == nil {
		t.Fatal("expected non-nil Plugstep")
	}
	if ps.ServerDirectory != "" {
		t.Errorf("expected empty server directory, got %q", ps.ServerDirectory)
	}
}
