package email

// JobType represents the type of email to send
type JobType string

const (
	JobTypeWelcome         JobType = "welcome"
	JobTypeVerification    JobType = "verification"
	JobTypePasswordReset   JobType = "password_reset"
	JobTypeWorkspaceInvite JobType = "workspace_invite"
	JobTypeNotification    JobType = "notification"
	JobTypeGeneric         JobType = "generic"
)

// Job represents an email job to be processed
type Job struct {
	Type    JobType
	To      string
	Subject string
	Text    string
	HTML    string
	Data    map[string]interface{}
}
