package version_manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type VersionInfo struct {
	Version     string    `json:"version"`
	Status      string    `json:"status"`
	ReleaseDate time.Time `json:"release_date"`
	EndOfLife   time.Time `json:"end_of_life,omitempty"`
	Changes     []string  `json:"changes,omitempty"`
}

type VersionDocumentation struct {
	Versions []VersionInfo `json:"versions"`
}

func NewVersionDocumentation() *VersionDocumentation {
	return &VersionDocumentation{
		Versions: []VersionInfo{
			{
				Version:     CurrentVersion,
				Status:      "active",
				ReleaseDate: time.Now(),
				Changes: []string{
					"Initial API version",
				},
			},
		},
	}
}

func (d *VersionDocumentation) AddVersion(version VersionInfo) error {
	for _, v := range d.Versions {
		if v.Version == version.Version {
			return fmt.Errorf("version %s already exists", version.Version)
		}
	}

	d.Versions = append(d.Versions, version)
	return nil
}

func (d *VersionDocumentation) UpdateVersion(version VersionInfo) error {
	for i, v := range d.Versions {
		if v.Version == version.Version {
			d.Versions[i] = version
			return nil
		}
	}
	return fmt.Errorf("version %s not found", version.Version)
}

func (d *VersionDocumentation) GetVersion(version string) (VersionInfo, error) {
	for _, v := range d.Versions {
		if v.Version == version {
			return v, nil
		}
	}
	return VersionInfo{}, fmt.Errorf("version %s not found", version)
}

func (d *VersionDocumentation) Save(path string) error {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (d *VersionDocumentation) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, d)
}

func (d *VersionDocumentation) GetVersionStatus(version string) (string, error) {
	info, err := d.GetVersion(version)
	if err != nil {
		return "", err
	}
	return info.Status, nil
}

func (d *VersionDocumentation) IsVersionActive(version string) bool {
	status, err := d.GetVersionStatus(version)
	if err != nil {
		return false
	}
	return status == "active"
}

func (d *VersionDocumentation) IsVersionDeprecated(version string) bool {
	status, err := d.GetVersionStatus(version)
	if err != nil {
		return false
	}
	return status == "deprecated"
}

func (d *VersionDocumentation) IsVersionRetired(version string) bool {
	status, err := d.GetVersionStatus(version)
	if err != nil {
		return false
	}
	return status == "retired"
}
