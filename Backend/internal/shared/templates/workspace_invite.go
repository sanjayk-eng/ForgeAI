package templates

import (
	"fmt"
	"time"
)

type WorkspaceInviteTemplateData struct {
	WorkspaceName string
	WorkspaceSlug string
	InviterName   string
	InviteLink    string
	Role          string
	ExpiresAt     time.Time
}

func WorkspaceInviteTemplate(data WorkspaceInviteTemplateData) (subject string, text string, html string) {
	// Default values
	workspaceName := data.WorkspaceName
	if workspaceName == "" {
		workspaceName = "a workspace"
	}
	if data.InviteLink == "" {
		data.InviteLink = "#"
	}
	if data.Role == "" {
		data.Role = "MEMBER"
	}
	inviterName := data.InviterName
	if inviterName == "" {
		inviterName = "Someone"
		
	}

	// Subject
	subject = fmt.Sprintf("%s invited you to join %s on ForgeAI", inviterName, workspaceName)

	// Plain text
	text = fmt.Sprintf(`Hi there,

%s has invited you to join the "%s" workspace on ForgeAI as a %s.

Accept this invitation to start collaborating:
%s

This invitation will expire on %s.

If you have any questions, feel free to reach out.

Best regards,
The ForgeAI Team`,
		inviterName, workspaceName, data.Role, data.InviteLink,
		data.ExpiresAt.Format("January 2, 2006 at 3:04 PM"))

	// Professional HTML
	html = fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background-color:#0f1117;color:#e4e6eb">
    <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#0f1117;padding:40px 20px">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" style="max-width:600px;background-color:#1a1d24;border:1px solid rgba(255,255,255,0.1);border-radius:8px;overflow:hidden">
                    <tr>
                        <td style="background:linear-gradient(135deg,#c6f36a 0%%,#8ed4ff 100%%);padding:32px 40px;text-align:center">
                            <h1 style="margin:0;font-size:28px;font-weight:800;color:#0f1117">Forge<span style="color:#2d5016">AI</span></h1>
                        </td>
                    </tr>
                    <tr>
                        <td style="padding:40px">
                            <h2 style="margin:0 0 16px 0;font-size:24px;font-weight:700;color:#fff">You've been invited! 🎉</h2>
                            <p style="margin:0 0 24px 0;font-size:16px;line-height:1.6;color:#b8bcc8"><strong style="color:#fff">%s</strong> has invited you to join the <strong style="color:#c6f36a">%s</strong> workspace on ForgeAI.</p>
                            <table width="100%%" cellpadding="0" cellspacing="0" style="margin:24px 0;background-color:rgba(198,243,106,0.05);border:1px solid rgba(198,243,106,0.15);border-radius:6px">
                                <tr>
                                    <td style="padding:20px">
                                        <table width="100%%" cellpadding="0" cellspacing="0">
                                            <tr>
                                                <td style="padding:8px 0;font-size:14px;color:#9ca3af"><strong>Workspace:</strong></td>
                                                <td style="padding:8px 0;font-size:14px;color:#fff;text-align:right">%s</td>
                                            </tr>
                                            <tr>
                                                <td style="padding:8px 0;font-size:14px;color:#9ca3af"><strong>Your role:</strong></td>
                                                <td style="padding:8px 0;font-size:14px;color:#c6f36a;text-align:right;font-weight:600">%s</td>
                                            </tr>
                                            <tr>
                                                <td style="padding:8px 0;font-size:14px;color:#9ca3af"><strong>Invited by:</strong></td>
                                                <td style="padding:8px 0;font-size:14px;color:#fff;text-align:right">%s</td>
                                            </tr>
                                            <tr>
                                                <td style="padding:8px 0;font-size:14px;color:#9ca3af"><strong>Expires:</strong></td>
                                                <td style="padding:8px 0;font-size:14px;color:#fbbf24;text-align:right">%s</td>
                                            </tr>
                                        </table>
                                    </td>
                                </tr>
                            </table>
                            <table width="100%%" cellpadding="0" cellspacing="0" style="margin:32px 0">
                                <tr>
                                    <td align="center">
                                        <a href="%s" style="display:inline-block;padding:14px 32px;background-color:#c6f36a;color:#0f1117;text-decoration:none;font-weight:700;font-size:15px;border-radius:6px">Accept Invitation →</a>
                                    </td>
                                </tr>
                            </table>
                            <p style="margin:24px 0 0 0;font-size:14px;line-height:1.6;color:#9ca3af">By accepting this invitation, you'll gain access to collaborate with the team, share resources, and work together on projects.</p>
                        </td>
                    </tr>
                    <tr>
                        <td style="padding:32px 40px;background-color:#141519;border-top:1px solid rgba(255,255,255,0.06)">
                            <p style="margin:0 0 12px 0;font-size:13px;color:#6b7280">If the button doesn't work, copy and paste this link:</p>
                            <p style="margin:0 0 20px 0;font-size:12px;color:#4b5563;word-break:break-all">%s</p>
                            <p style="margin:0;font-size:12px;color:#6b7280">This invitation expires on <strong style="color:#9ca3af">%s</strong>. If you didn't expect this, you can safely ignore this email.</p>
                        </td>
                    </tr>
                    <tr>
                        <td style="padding:24px 40px;text-align:center;background-color:#0d0e12">
                            <p style="margin:0 0 8px 0;font-size:12px;color:#6b7280">© %d ForgeAI. All rights reserved.</p>
                            <p style="margin:0;font-size:11px;color:#4b5563">Building the future of AI-powered collaboration</p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`,
		inviterName, workspaceName,
		workspaceName, data.Role, inviterName, data.ExpiresAt.Format("January 2, 2006"),
		data.InviteLink,
		data.InviteLink, data.ExpiresAt.Format("January 2, 2006 at 3:04 PM"),
		time.Now().Year())

	return subject, text, html
}
