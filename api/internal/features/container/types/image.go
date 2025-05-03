package types

type Image struct {
	ID          string            `json:"id"`
	RepoTags    []string          `json:"repo_tags"`
	RepoDigests []string          `json:"repo_digests"`
	Created     int64             `json:"created"`
	Size        int64             `json:"size"`
	SharedSize  int64             `json:"shared_size"`
	VirtualSize int64             `json:"virtual_size"`
	Labels      map[string]string `json:"labels"`
}

type ContainerReference struct {
	ID    string   `json:"id"`
	Names []string `json:"names"`
}
