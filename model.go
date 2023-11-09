package main

import "time"

// https://developer.hashicorp.com/terraform/cloud-docs/integrations/run-tasks#integration-details
type TFERunTasksRequest struct {
	PayloadVersion                  int                      `json:"payload_version,omitempty"`
	Stage                           string                   `json:"stage,omitempty"`
	AccessToken                     string                   `json:"access_token,omitempty"`
	Capabilities                    *TFERunTasksCapabilities `json:"capabilities,omitempty"`
	ConfigurationVersionDownloadURL string                   `json:"configuration_version_download_url,omitempty"`
	ConfigurationVersionID          string                   `json:"configuration_version_id,omitempty"`
	IsSpeculative                   bool                     `json:"is_speculative,omitempty"`
	OrganizationName                string                   `json:"organization_nam,omitemptye"`
	PlanJSONAPIURL                  string                   `json:"plan_json_api_url,omitempty"`
	RunAppURL                       string                   `json:"run_app_url,omitempty"`
	RunCreatedAt                    time.Time                `json:"run_created_at,omitempty"`
	RunCreatedBy                    string                   `json:"run_created_by,omitempty"`
	RunID                           string                   `json:"run_id,omitempty"`
	RunMessage                      string                   `json:"run_message,omitempty"`
	TaskResultCallbackURL           string                   `json:"task_result_callback_url,omitempty"`
	TaskResultEnforcementLevel      string                   `json:"task_result_enforcement_level,omitempty"`
	TaskResultID                    string                   `json:"task_result_id,omitempty"`
	VCSBranch                       string                   `json:"vcs_branch,omitempty"`
	VCSCommitURL                    string                   `json:"vcs_commit_url,omitempty"`
	VCSPullRequestURL               string                   `json:"vcs_pull_request_url,omitempty"`
	VCSRepoURL                      string                   `json:"vcs_repo_url,omitempty"`
	WorkspaceAppURL                 string                   `json:"workspace_app_url,omitempty"`
	WorkspaceID                     string                   `json:"workspace_id,omitempty"`
	WorkspaceName                   string                   `json:"workspace_name,omitempty"`
	WorkspaceWorkingDirectory       string                   `json:"workspace_working_directory,omitempty"`
}

type TFERunTasksCapabilities struct {
	Outcomes bool `json:"outcomes,omitempty"`
}

type TFERunTasksResponse struct {
	Data *TFERunTasksResponseData `json:"data,omitempty"`
}

type TFERunTasksResponseData struct {
	Type          string                            `json:"type,omitempty"`
	Attributes    *TFERunTasksResponseAttributes    `json:"attributes,omitempty"`
	Relationships *TFERunTasksResponseRelationships `json:"relationships,omitempty"`
}

type TFERunTasksResponseAttributes struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	URL     string `json:"url,omitempty"`
}

type TFERunTasksResponseRelationships struct {
	Outcomes *TFERunTasksResponseOutcomes `json:"outcomes,omitempty"`
}

type TFERunTasksResponseOutcomes struct {
	Data []*TFERunTasksResponseOutcomesData `json:"data,omitempty"`
}

type TFERunTasksResponseOutcomesData struct {
	OutcomeID   string                                        `json:"outcome-id,omitempty"`
	Description string                                        `json:"description,omitempty"`
	Body        string                                        `json:"body,omitempty"`
	URL         string                                        `json:"url,omitempty"`
	Tags        map[string][]*TFERunTasksResponseOutcomesTags `json:"tags,omitempty"`
}

type TFERunTasksErrorLevel string

const (
	NONE    TFERunTasksErrorLevel = "none"
	INFO    TFERunTasksErrorLevel = "info"
	WARNING TFERunTasksErrorLevel = "warning"
	ERROR   TFERunTasksErrorLevel = "error"
)

type TFERunTasksResponseOutcomesTags struct {
	Label string                `json:"label,omitempty"`
	Level TFERunTasksErrorLevel `json:"level,omitempty"`
}
