package realtime

type Event struct {
	Version     int    `json:"version"`
	Event       string `json:"event"`
	organizationID string `json:"organization_id,omitempty"`
	ProjectID   string `json:"project_id"`
	SandboxID   string `json:"sandbox_id,omitempty"`
	Path        string `json:"path,omitempty"`
	OldPath     string `json:"old_path,omitempty"`
	ChangeType  string `json:"change_type,omitempty"`
	IsDirectory bool   `json:"is_directory,omitempty"`
	Status      string `json:"status,omitempty"`
	Message     string `json:"message,omitempty"`
}

type Publisher interface {
	Publish(Event)
}
