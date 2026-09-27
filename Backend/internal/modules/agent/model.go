package agent

type StatusResponse struct {
	Configured bool   `json:"configured"`
	Model      string `json:"model,omitempty"`
}

type TaskRequest struct {
	Prompt string `json:"prompt"`
}

type TaskResponse struct {
	Message      string   `json:"message"`
	ChangedFiles []string `json:"changed_files"`
}

type modelTaskResult struct {
	Message string       `json:"message"`
	Changes []fileChange `json:"changes"`
}

type fileChange struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	MaxTokens      int            `json:"max_tokens"`
	ResponseFormat responseFormat `json:"response_format"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}
