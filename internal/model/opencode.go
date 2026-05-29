package model

import (
	"strconv"
	"strings"
	"text/template"

	"paraizofelipe/review-station/internal/config"
	"paraizofelipe/review-station/internal/gitlab"
	"paraizofelipe/review-station/internal/launcher"
)

// openCodeData são os campos disponíveis como placeholders no template do
// comando opencode.
type openCodeData struct {
	IID          int
	Title        string
	Description  string
	Author       string
	SourceBranch string
	TargetBranch string
	WebURL       string
	State        string
	ProjectPath  string
	ProjectName  string
	Local        string
}

// openCodeTemplateString retorna o template do comando: override do repo se
// houver, senão o global; "" se nada configurado.
func openCodeTemplateString(repo config.Repo, cfg config.Config) string {
	if repo.OpenCodeCommand != "" {
		return repo.OpenCodeCommand
	}
	return cfg.OpenCode.Command
}

// resolveOpenCodeCommand renderiza o template do comando com os dados do MR.
// Retorna ("", nil) quando não há comando configurado.
func resolveOpenCodeCommand(repo config.Repo, mr *gitlab.MergeRequest, cfg config.Config) (string, error) {
	tmplStr := openCodeTemplateString(repo, cfg)
	if tmplStr == "" || mr == nil {
		return "", nil
	}
	data := openCodeData{
		IID:          mr.IID,
		Title:        mr.Title,
		Description:  mr.Description,
		Author:       mr.Author,
		SourceBranch: mr.SourceBranch,
		TargetBranch: mr.TargetBranch,
		WebURL:       mr.WebURL,
		State:        mr.State,
		ProjectPath:  repo.Path,
		ProjectName:  repo.Name,
		Local:        launcher.ExpandHome(repo.Local),
	}
	tmpl, err := template.New("opencode").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// buildOpenCodeEnv monta as variáveis de ambiente RS_* injetadas no processo.
func buildOpenCodeEnv(repo config.Repo, mr *gitlab.MergeRequest) map[string]string {
	return map[string]string{
		"RS_MR_IID":           strconv.Itoa(mr.IID),
		"RS_MR_TITLE":         mr.Title,
		"RS_MR_DESCRIPTION":   mr.Description,
		"RS_MR_AUTHOR":        mr.Author,
		"RS_MR_SOURCE_BRANCH": mr.SourceBranch,
		"RS_MR_TARGET_BRANCH": mr.TargetBranch,
		"RS_MR_WEB_URL":       mr.WebURL,
		"RS_MR_STATE":         mr.State,
		"RS_PROJECT_PATH":     repo.Path,
		"RS_PROJECT_NAME":     repo.Name,
		"RS_LOCAL":            launcher.ExpandHome(repo.Local),
	}
}
