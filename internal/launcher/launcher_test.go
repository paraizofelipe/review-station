package launcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("sem home dir")
	}
	if got, want := ExpandHome("~/projects/app"), filepath.Join(home, "projects/app"); got != want {
		t.Errorf("ExpandHome(~/projects/app) = %q, want %q", got, want)
	}
	if got, want := ExpandHome("/abs/path"), "/abs/path"; got != want {
		t.Errorf("ExpandHome(/abs/path) = %q, want %q", got, want)
	}
	if got, want := ExpandHome("relative"), "relative"; got != want {
		t.Errorf("ExpandHome(relative) = %q, want %q", got, want)
	}
}

func TestBuildInvocationNvim(t *testing.T) {
	det := detection{Nvim: "/tmp/nvim.sock", Tmux: ""}
	bin, args := buildInvocation(det, "opencode run", "/home/u/app", "rs-review-7")

	if bin != "nvim" {
		t.Errorf("bin = %q, want nvim", bin)
	}
	want := []string{"--server", "/tmp/nvim.sock", "--remote-send", `<C-\><C-n>:botright vsplit | terminal cd '/home/u/app' && opencode run<CR>`}
	if strings.Join(args, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("args = %v, want %v", args, want)
	}
}

func TestBuildInvocationTmux(t *testing.T) {
	det := detection{Nvim: "", Tmux: "/tmp/tmux-1000/default,123,0"}
	bin, args := buildInvocation(det, "opencode run", "/home/u/app", "rs-review-7")

	if bin != "tmux" {
		t.Errorf("bin = %q, want tmux", bin)
	}
	want := []string{"new-window", "-n", "rs-review-7", "-c", "/home/u/app", "sh", "-c", "opencode run; exec $SHELL"}
	if strings.Join(args, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("args = %v, want %v", args, want)
	}
}

func TestBuildInvocationGhosttyDefault(t *testing.T) {
	det := detection{Nvim: "", Tmux: ""}
	bin, args := buildInvocation(det, "opencode run", "/home/u/app", "rs-review-7")

	if bin != "ghostty" {
		t.Errorf("bin = %q, want ghostty", bin)
	}
	want := []string{"-e", "sh", "-c", "cd '/home/u/app' && opencode run; exec $SHELL"}
	if strings.Join(args, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("args = %v, want %v", args, want)
	}
}

func TestBuildInvocationGhosttyQuotesLocalWithSpaces(t *testing.T) {
	det := detection{}
	_, args := buildInvocation(det, "opencode run", "/home/u/My Projects/app", "rs-review-1")
	want := []string{"-e", "sh", "-c", "cd '/home/u/My Projects/app' && opencode run; exec $SHELL"}
	if strings.Join(args, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("args = %v, want %v", args, want)
	}
}

func TestBuildInvocationNvimTakesPrecedenceOverTmux(t *testing.T) {
	det := detection{Nvim: "/tmp/nvim.sock", Tmux: "/tmp/tmux"}
	bin, _ := buildInvocation(det, "opencode run", "/home/u/app", "rs-review-7")
	if bin != "nvim" {
		t.Errorf("bin = %q, want nvim (NVIM tem precedência sobre TMUX)", bin)
	}
}
