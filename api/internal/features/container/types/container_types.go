package types

type Container struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	Status     string            `json:"status"`
	State      string            `json:"state"`
	Created    string            `json:"created"`
	Labels     map[string]string `json:"labels"`
	Ports      []Port            `json:"ports"`
	Mounts     []Mount           `json:"mounts"`
	Networks   []Network         `json:"networks"`
	Command    string            `json:"command"`
	IPAddress  string            `json:"ip_address"`
	HostConfig HostConfig        `json:"host_config"`
}

type Port struct {
	PrivatePort int    `json:"private_port"`
	PublicPort  int    `json:"public_port"`
	Type        string `json:"type"`
}

type Mount struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
}

type Network struct {
	Name       string   `json:"name"`
	IPAddress  string   `json:"ip_address"`
	Gateway    string   `json:"gateway"`
	MacAddress string   `json:"mac_address"`
	Aliases    []string `json:"aliases"`
}

type HostConfig struct {
	Memory     int64 `json:"memory"`
	MemorySwap int64 `json:"memory_swap"`
	CPUShares  int64 `json:"cpu_shares"`
}

type ContainerStats struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage int64   `json:"memory_usage"`
	NetworkIO   struct {
		RxBytes int64 `json:"rx_bytes"`
		TxBytes int64 `json:"tx_bytes"`
	} `json:"network_io"`
	BlockIO struct {
		Read  int64 `json:"read"`
		Write int64 `json:"write"`
	} `json:"block_io"`
}

type ContainerListOptions struct {
	All     bool   `json:"all"`
	Limit   int    `json:"limit"`
	Since   string `json:"since"`
	Before  string `json:"before"`
	Size    bool   `json:"size"`
	Filters map[string][]string
}

type ContainerIDRequest struct {
	ID string `json:"id"`
}

type ContainerLogsRequest struct {
	ID     string `json:"id"`
	Follow bool   `json:"follow"`
	Tail   int    `json:"tail"`
	Since  string `json:"since"`
	Until  string `json:"until"`
	Stdout bool   `json:"stdout"`
	Stderr bool   `json:"stderr"`
}

type ContainerExecRequest struct {
	ID      string   `json:"id"`
	Command []string `json:"command"`
	User    string   `json:"user"`
	WorkDir string   `json:"work_dir"`
}

type ContainerExecResponse struct {
	ID string `json:"id"`
}

type Volume struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	Labels     map[string]string `json:"labels"`
	Options    map[string]string `json:"options"`
}

type VolumeCreateRequest struct {
	Name    string            `json:"name"`
	Driver  string            `json:"driver"`
	Labels  map[string]string `json:"labels"`
	Options map[string]string `json:"options"`
}

type VolumeListOptions struct {
	Filters map[string][]string `json:"filters"`
}
