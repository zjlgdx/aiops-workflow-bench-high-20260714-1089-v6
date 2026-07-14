package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAddPersistsAndListSurvivesSeparateProcesses(t *testing.T) {
	database := filepath.Join(t.TempDir(), "todos.json")

	stdout, stderr, exitCode := runTodo(t, database, "add", "  Buy milk  ")
	if exitCode != 0 || stdout != "added 1\n" || stderr != "" {
		t.Fatalf("first add = stdout %q, stderr %q, exit %d", stdout, stderr, exitCode)
	}

	stdout, stderr, exitCode = runTodo(t, database, "add", "Walk dog")
	if exitCode != 0 || stdout != "added 2\n" || stderr != "" {
		t.Fatalf("second add = stdout %q, stderr %q, exit %d", stdout, stderr, exitCode)
	}

	stdout, stderr, exitCode = runTodo(t, database, "list")
	if exitCode != 0 || stdout != "1\tactive\tBuy milk\n2\tactive\tWalk dog\n" || stderr != "" {
		t.Fatalf("list = stdout %q, stderr %q, exit %d", stdout, stderr, exitCode)
	}
}

func TestAddRejectsEmptyTitleWithoutChangingDatabase(t *testing.T) {
	database := filepath.Join(t.TempDir(), "todos.json")
	_, _, exitCode := runTodo(t, database, "add", "Keep me")
	if exitCode != 0 {
		t.Fatalf("seed add exited %d", exitCode)
	}
	wantDatabase, err := os.ReadFile(database)
	if err != nil {
		t.Fatal(err)
	}

	stdout, stderr, exitCode := runTodo(t, database, "add", " \t\n ")
	if exitCode == 0 || stdout != "" || stderr != "title must not be empty\n" {
		t.Fatalf("empty add = stdout %q, stderr %q, exit %d", stdout, stderr, exitCode)
	}

	gotDatabase, err := os.ReadFile(database)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotDatabase, wantDatabase) {
		t.Fatalf("database changed after invalid add: got %q, want %q", gotDatabase, wantDatabase)
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	separator := 0
	for i, arg := range os.Args {
		if arg == "--" {
			separator = i
			break
		}
	}
	os.Args = append([]string{"todo"}, os.Args[separator+1:]...)
	main()
	os.Exit(0)
}

func runTodo(t *testing.T, database string, args ...string) (string, string, int) {
	t.Helper()

	commandArgs := append([]string{"-test.run=^TestHelperProcess$", "--"}, args...)
	cmd := exec.Command(os.Args[0], commandArgs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "TODO_DB="+database)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			t.Fatal(err)
		}
		exitCode = exitError.ExitCode()
	}
	return stdout.String(), stderr.String(), exitCode
}
