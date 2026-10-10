package agent

type StatusResponse struct {
	Configured bool   `json:"configured"`
	Model      string `json:"model,omitempty"`
}

type TaskRequest struct {
	Prompt string `json:"prompt" binding:"required" validate:"required,min=1,max=10000"`
}

type TaskResponse struct {
	Message      string   `json:"message"`
	ChangedFiles []string `json:"changed_files"`
}
