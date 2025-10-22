package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type GithubConnector struct {
	bun.BaseModel `bun:"table:github_connectors,alias:gc" swaggerignore:"true"`

	ID             uuid.UUID  `bun:"id,pk,type:uuid" json:"id"`
	AppID          string     `bun:"app_id,notnull" json:"app_id"`
	Slug           string     `bun:"slug,notnull" json:"slug"`
	Pem            string     `bun:"pem,notnull" json:"pem"`
	ClientID       string     `bun:"client_id,notnull" json:"client_id"`
	ClientSecret   string     `bun:"client_secret,notnull" json:"client_secret"`
	WebhookSecret  string     `bun:"webhook_secret,notnull" json:"webhook_secret"`
	InstallationID string     `bun:"installation_id,notnull" json:"installation_id"`
	CreatedAt      time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt      *time.Time `bun:"deleted_at" json:"deleted_at"`
	UserID         uuid.UUID  `json:"user_id" bun:"user_id,notnull,type:uuid"`
}

type GithubRepository struct {
	ID                       uint64    `json:"id"`
	NodeID                   string    `json:"node_id"`
	Name                     string    `json:"name"`
	FullName                 string    `json:"full_name"`
	Private                  bool      `json:"private"`
	Owner                    Owner     `json:"owner"`
	HTMLURL                  string    `json:"html_url"`
	Description              *string   `json:"description,omitempty"`
	Fork                     bool      `json:"fork"`
	URL                      string    `json:"url"`
	ForksURL                 string    `json:"forks_url"`
	KeysURL                  string    `json:"keys_url"`
	CollaboratorsURL         string    `json:"collaborators_url"`
	TeamsURL                 string    `json:"teams_url"`
	HooksURL                 string    `json:"hooks_url"`
	IssueEventsURL           string    `json:"issue_events_url"`
	EventsURL                string    `json:"events_url"`
	AssigneesURL             string    `json:"assignees_url"`
	BranchesURL              string    `json:"branches_url"`
	TagsURL                  string    `json:"tags_url"`
	BlobsURL                 string    `json:"blobs_url"`
	GitTagsURL               string    `json:"git_tags_url"`
	GitRefsURL               string    `json:"git_refs_url"`
	TreesURL                 string    `json:"trees_url"`
	StatusesURL              string    `json:"statuses_url"`
	LanguagesURL             string    `json:"languages_url"`
	StargazersURL            string    `json:"stargazers_url"`
	ContributorsURL          string    `json:"contributors_url"`
	SubscribersURL           string    `json:"subscribers_url"`
	SubscriptionURL          string    `json:"subscription_url"`
	CommitsURL               string    `json:"commits_url"`
	GitCommitsURL            string    `json:"git_commits_url"`
	CommentsURL              string    `json:"comments_url"`
	IssueCommentURL          string    `json:"issue_comment_url"`
	ContentsURL              string    `json:"contents_url"`
	CompareURL               string    `json:"compare_url"`
	MergesURL                string    `json:"merges_url"`
	ArchiveURL               string    `json:"archive_url"`
	DownloadsURL             string    `json:"downloads_url"`
	IssuesURL                string    `json:"issues_url"`
	PullsURL                 string    `json:"pulls_url"`
	MilestonesURL            string    `json:"milestones_url"`
	NotificationsURL         string    `json:"notifications_url"`
	LabelsURL                string    `json:"labels_url"`
	ReleasesURL              string    `json:"releases_url"`
	DeploymentsURL           string    `json:"deployments_url"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
	PushedAt                 time.Time `json:"pushed_at"`
	GitURL                   string    `json:"git_url"`
	SSHURL                   string    `json:"ssh_url"`
	CloneURL                 string    `json:"clone_url"`
	SVNURL                   string    `json:"svn_url"`
	Homepage                 *string   `json:"homepage,omitempty"`
	Size                     uint64    `json:"size"`
	StargazersCount          uint64    `json:"stargazers_count"`
	WatchersCount            uint64    `json:"watchers_count"`
	Language                 *string   `json:"language,omitempty"`
	HasIssues                bool      `json:"has_issues"`
	HasProjects              bool      `json:"has_projects"`
	HasDownloads             bool      `json:"has_downloads"`
	HasWiki                  bool      `json:"has_wiki"`
	HasPages                 bool      `json:"has_pages"`
	HasDiscussions           bool      `json:"has_discussions"`
	ForksCount               uint64    `json:"forks_count"`
	MirrorURL                *string   `json:"mirror_url,omitempty"`
	Archived                 bool      `json:"archived"`
	Disabled                 bool      `json:"disabled"`
	OpenIssuesCount          uint64    `json:"open_issues_count"`
	License                  *License  `json:"license,omitempty"`
	AllowForking             bool      `json:"allow_forking"`
	IsTemplate               bool      `json:"is_template"`
	WebCommitSignoffRequired bool      `json:"web_commit_signoff_required"`
	Topics                   []string  `json:"topics"`
	Visibility               string    `json:"visibility"`
	Forks                    uint64    `json:"forks"`
	OpenIssues               uint64    `json:"open_issues"`
	Watchers                 uint64    `json:"watchers"`
	DefaultBranch            string    `json:"default_branch"`
	Permissions              *struct {
		Admin    bool `json:"admin"`
		Maintain bool `json:"maintain"`
		Push     bool `json:"push"`
		Triage   bool `json:"triage"`
		Pull     bool `json:"pull"`
	} `json:"permissions,omitempty"`
}

type Owner struct {
	Login             string `json:"login"`
	ID                uint64 `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type License struct {
	Key    string  `json:"key"`
	Name   string  `json:"name"`
	SpdxID string  `json:"spdx_id"`
	URL    *string `json:"url,omitempty"`
	NodeID string  `json:"node_id"`
}

type GithubRepositoryBranch struct {
	Name   string `json:"name"`
	Commit struct {
		Sha string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	Protected bool `json:"protected"`
}
