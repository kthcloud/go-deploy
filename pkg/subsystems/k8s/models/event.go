package models

const (
	EventTypeWarning = "warning"
	EventTypeNormal  = "normal"

	EventReasonMountFailed     = "mountFailed"
	EventReasonCrashLoop       = "crashLoop"
	EventReasonImagePullFailed = "imagePullFailed"

	EventObjectKindDeployment = "deployment"
)

type Event struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Reason      string `json:"reason"`
	Description string `json:"description"`
	ObjectKind  string `json:"objectKind"`
}
