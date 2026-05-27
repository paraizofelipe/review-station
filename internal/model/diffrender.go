package model

import (
	"fmt"
	"path/filepath"
	"strings"

	chroma "github.com/alecthomas/chroma/v2"
	chromaLexers "github.com/alecthomas/chroma/v2/lexers"
	chromaStyles "github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"

	"paraizofelipe/review-station/internal/gitlab"
)

// Cores de fundo por tipo de linha, inspirado no tema do delta.
// São levemente distintas do bg0 gruvbox (#282828) para ficarem
// legíveis sem ofuscar o syntax highlight.
const (
	diffBgAdded   = "#1b3220" // verde escuro
	diffBgRemoved = "#3a1a1a" // vermelho escuro
	diffBgContext = "#1e1e1e" // neutro ligeiramente mais escuro
)

// renderDiffBlock renderiza uma slice de DiffLine com syntax highlighting
// (foreground via chroma/monokai) sobre backgrounds coloridos por tipo,
// preenchendo cada linha até contentWidth para a cor cobrir toda a largura.
func renderDiffBlock(lines []gitlab.DiffLine, path string, contentWidth int) string {
	style := chromaStyles.Get("monokai")
	if style == nil {
		style = chromaStyles.Fallback
	}

	// Calcula a largura do gutter de números de linha.
	maxLine := 0
	for _, l := range lines {
		if l.NewLine > maxLine {
			maxLine = l.NewLine
		}
		if l.OldLine > maxLine {
			maxLine = l.OldLine
		}
	}
	lineNumWidth := len(fmt.Sprintf("%d", maxLine))
	if lineNumWidth < 1 {
		lineNumWidth = 1
	}

	// Tokeniza todas as linhas de uma vez para o lexer ter contexto completo.
	tokensByLine := tokenizeForDiff(lines, path)

	var sb strings.Builder
	for i, l := range lines {
		var bgHex string
		switch l.Kind {
		case gitlab.DiffLineKindAdded:
			bgHex = diffBgAdded
		case gitlab.DiffLineKindRemoved:
			bgHex = diffBgRemoved
		default:
			bgHex = diffBgContext
		}

		var toks []chroma.Token
		if i < len(tokensByLine) {
			toks = tokensByLine[i]
		}

		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(renderHighlightedLine(toks, l.Content, l.Kind, l.OldLine, l.NewLine, lineNumWidth, bgHex, style, contentWidth))
	}
	return sb.String()
}

// tokenizeForDiff usa o lexer inferido do path para tokenizar todo o bloco
// de linhas de diff e retorna os tokens agrupados por linha.
func tokenizeForDiff(lines []gitlab.DiffLine, path string) [][]chroma.Token {
	// Constrói o source completo para que o lexer tenha contexto de linguagem.
	var src strings.Builder
	for _, l := range lines {
		src.WriteString(l.Content)
		src.WriteByte('\n')
	}

	lexer := chromaLexers.Match(path)
	if lexer == nil {
		lexer = chromaLexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	it, err := lexer.Tokenise(nil, src.String())
	if err != nil {
		return nil
	}

	all := it.Tokens()
	byLine := chroma.SplitTokensIntoLines(all)

	// chroma.SplitTokensIntoLines inclui o '\n' no final de cada linha;
	// removemos para não duplicar o separador.
	for i, line := range byLine {
		if n := len(line); n > 0 && line[n-1].Value == "\n" {
			byLine[i] = line[:n-1]
		}
	}
	return byLine
}

// lineNumGray é a cor gruvbox gray (#928374) para os números de linha.
const lineNumGray = "\x1b[38;2;146;131;116m"

// renderHighlightedLine emite uma linha com:
//  1. background ANSI cobrindo toda a largura (contentWidth colunas)
//  2. gutter de números de linha (old│new) em cinza discreto
//  3. prefixo "+"/"-"/" " colorido
//  4. tokens com foreground do estilo chroma (sem background — herdamos o nosso)
//  5. reset de foreground (não de background) entre tokens, para manter o bg
//  6. padding de espaços até contentWidth antes do reset final
func renderHighlightedLine(tokens []chroma.Token, fallback string, kind gitlab.DiffLineKind, oldLine, newLine, lineNumWidth int, bgHex string, style *chroma.Style, width int) string {
	bgSeq := hexBgANSI(bgHex)

	var sb strings.Builder

	// 1. Ativa background da linha.
	sb.WriteString(bgSeq)

	// 2. Gutter de números de linha: old e new, alinhados à direita.
	//    Coluna vazia quando a linha não existe para aquele lado do diff.
	if oldLine > 0 {
		sb.WriteString(fmt.Sprintf("%s%*d", lineNumGray, lineNumWidth, oldLine))
	} else {
		sb.WriteString(strings.Repeat(" ", lineNumWidth))
	}
	sb.WriteString(" ")
	if newLine > 0 {
		sb.WriteString(fmt.Sprintf("%s%*d", lineNumGray, lineNumWidth, newLine))
	} else {
		sb.WriteString(strings.Repeat(" ", lineNumWidth))
	}
	// Reset fg antes do prefixo +/-.
	sb.WriteString("\x1b[39m ")

	// 3. Prefixo: símbolo +/- em bold, espaço separador.
	switch kind {
	case gitlab.DiffLineKindAdded:
		// bold bright green
		sb.WriteString("\x1b[1;92m+\x1b[22;39m ")
	case gitlab.DiffLineKindRemoved:
		// bold bright red
		sb.WriteString("\x1b[1;91m-\x1b[22;39m ")
	default:
		sb.WriteString("  ")
	}

	// 4. Tokens com foreground do style chroma. Usa \x1b[39m (reset fg only)
	// entre tokens para não limpar o background que definimos acima.
	if len(tokens) == 0 {
		sb.WriteString(fallback)
	} else {
		prevHadStyle := false
		for _, tok := range tokens {
			entry := style.Get(tok.Type)

			hasStyle := entry.Colour.IsSet() || entry.Bold == chroma.Yes || entry.Italic == chroma.Yes
			if hasStyle {
				if entry.Colour.IsSet() {
					r2, g2, b2 := entry.Colour.Red(), entry.Colour.Green(), entry.Colour.Blue()
					sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r2, g2, b2))
				}
				if entry.Bold == chroma.Yes {
					sb.WriteString("\x1b[1m")
				}
				if entry.Italic == chroma.Yes {
					sb.WriteString("\x1b[3m")
				}
			}

			sb.WriteString(tok.Value)

			if hasStyle {
				// Reset fg + bold + italic mas NÃO background.
				sb.WriteString("\x1b[22;23;39m")
				prevHadStyle = true
			} else if prevHadStyle {
				// Garante reset de estilo anterior antes de conteúdo sem estilo.
				sb.WriteString("\x1b[22;23;39m")
				prevHadStyle = false
			}
		}
	}

	// 5. Mede a largura visível atual e preenche com espaços (bg ainda ativo).
	line := sb.String()
	visible := lipgloss.Width(line)
	if visible < width {
		sb.WriteString(strings.Repeat(" ", width-visible))
	}

	// 6. Reset total ao fim da linha.
	sb.WriteString("\x1b[0m")
	return sb.String()
}

// hexBgANSI converte "#rrggbb" em sequência ANSI TrueColor de background.
func hexBgANSI(hex string) string {
	r, g, b := hexRGB(hex)
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
}

// hexRGB decompõe uma cor "#rrggbb" (com ou sem #) em componentes RGB.
func hexRGB(hex string) (r, g, b uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return
	}
	parse := func(s string) uint8 {
		var v uint64
		fmt.Sscanf(s, "%x", &v)
		return uint8(v)
	}
	return parse(hex[0:2]), parse(hex[2:4]), parse(hex[4:6])
}

// langFromPath infere a linguagem a partir do path do arquivo.
// Retorna o caminho completo para que chromaLexers.Match use o padrão de extensão.
func langFromPath(path string) string {
	return filepath.Base(path)
}
