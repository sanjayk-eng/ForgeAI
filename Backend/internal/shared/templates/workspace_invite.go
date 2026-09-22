package templates

import "fmt"

type WorkspaceInviteTemplateData struct {
	WorkspaceName string
	WorkspaceSlug string
	InviterName   string
	InviteLink    string
	Role          string
}

func WorkspaceInviteTemplate(data WorkspaceInviteTemplateData) (subject string, text string, html string) {
	workspaceName := data.WorkspaceName
	if workspaceName == "" {
		workspaceName = "your workspace"
	}
	if data.InviteLink == "" {
		data.InviteLink = "#"
	}
	if data.Role == "" {
		data.Role = "member"
	}

	subject = fmt.Sprintf("You are invited to join %s", workspaceName)
	text = fmt.Sprintf("Hi,\n\n%s has invited you to join %s as a %s.\n\nAccept the invite here: %s\n", data.InviterName, workspaceName, data.Role, data.InviteLink)
	html = fmt.Sprintf(`<html><body><h3>You are invited to join %s</h3><p>%s has invited you to join %s as a %s.</p><p><a href="%s">Accept invite</a></p></body></html>`, workspaceName, data.InviterName, workspaceName, data.Role, data.InviteLink)
	return subject, text, html
}
