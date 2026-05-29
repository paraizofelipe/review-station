// Package launcher detecta o ambiente de terminal e dispara um processo
// externo (opencode) num buffer do neovim, numa janela do tmux ou numa
// janela nova do ghostty.
package launcher

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// detection guarda os valores das variáveis de ambiente usadas para escolher
// o alvo de spawn. Valor vazio significa "não setada".
type detection struct {
	Nvim string
	Tmux string
}

// ExpandHome resolve um "~" ou "~/" inicial para o home do usuário.
func ExpandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return path
}

// shellQuote envolve um caminho em aspas simples para embuti-lo com segurança
// numa string de shell, escapando aspas simples internas.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// buildInvocation monta (binário, args) para o spawn, conforme a detecção do
// ambiente. Precedência: NVIM > TMUX > ghostty. Função pura: não lê env nem
// executa nada.
//
// Nota: `command` é tratado como fragmento de shell confiável vindo da config
// do usuário — pode conter aspas e espaços intencionais, portanto não é citado.
// Apenas `local` (caminho de diretório) é citado via shellQuote.
func buildInvocation(det detection, command, local, windowName string) (string, []string) {
	switch {
	case det.Nvim != "":
		send := `<C-\><C-n>:botright vsplit | terminal cd ` + shellQuote(local) + " && " + command + "<CR>"
		return "nvim", []string{"--server", det.Nvim, "--remote-send", send}
	case det.Tmux != "":
		// tmux passa local como argumento separado (-c), sem interpolação de shell — seguro.
		return "tmux", []string{"new-window", "-n", windowName, "-c", local, "sh", "-c", command + "; exec $SHELL"}
	default:
		return "ghostty", []string{"-e", "sh", "-c", "cd " + shellQuote(local) + " && " + command + "; exec $SHELL"}
	}
}

// Launch resolve o ambiente atual, monta a invocação e dispara o processo de
// forma fire-and-forget. Retorna erro se o binário do alvo não puder ser
// iniciado.
func Launch(command, local, windowName string, extraEnv map[string]string) error {
	expanded := ExpandHome(local)
	det := detection{Nvim: os.Getenv("NVIM"), Tmux: os.Getenv("TMUX")}
	bin, args := buildInvocation(det, command, expanded, windowName)

	cmd := exec.Command(bin, args...)
	cmd.Dir = expanded
	cmd.Env = os.Environ()
	for k, v := range extraEnv {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }() // reap para evitar zumbis
	return nil
}
