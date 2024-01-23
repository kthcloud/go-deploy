package body

type HarborWebhook struct {
	Type      string `json:"type"`
	OccurAt   int    `json:"occur_at"`
	Operator  string `json:"operator"`
	EventData struct {
		Resources []struct {
			Digest      string `json:"digest"`
			Tag         string `json:"tag"`
			ResourceURL string `json:"resource_url"`
		} `json:"resources"`
		Repository struct {
			DateCreated  int    `json:"date_created"`
			Name         string `json:"name"`
			Namespace    string `json:"namespace"`
			RepoFullName string `json:"repo_full_name"`
			RepoType     string `json:"repo_type"`
		} `json:"repository"`
	} `json:"event_data"`
}

type GitHubWebhookPing struct {
	Hook struct {
		ID     int64    `json:"id"`
		Type   string   `json:"type"`
		Events []string `json:"events"`
		Config struct {
			URL         string `json:"url"`
			ContentType string `json:"content_type"`
			Token       string `json:"token"`
		}
	} `json:"hook"`

	Repository struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Owner struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		}
	} `json:"repository"`
}

type GithubWebhookPayloadPush struct {
	Ref        string `json:"ref" binding:"required"`
	Repository struct {
		ID    int64  `json:"id" binding:"required"`
		Name  string `json:"name" binding:"required"`
		Owner struct {
			ID    int64  `json:"id" binding:"required"`
			Login string `json:"login" binding:"required"`
		}
		CloneURL      string `json:"clone_url" binding:"required"`
		DefaultBranch string `json:"default_branch" binding:"required"`
	} `json:"repository" binding:"required"`
}

type GitHubWebhookPush struct {
	ID        int64
	Event     string
	Signature string
	Payload   GithubWebhookPayloadPush
}
