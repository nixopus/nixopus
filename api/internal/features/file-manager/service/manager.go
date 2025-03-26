package service

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type FileInfo struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Mode     string    `json:"mode"`
	ModTime  time.Time `json:"mod_time"`
	IsDir    bool      `json:"is_dir"`
	Path     string    `json:"path"`
	IsHidden bool      `json:"is_hidden"`
}

type SSHClient interface {
	NewSftp() (SFTPClient, error)
}

type SFTPClient interface {
	Close() error
	ReadDir(path string) ([]os.FileInfo, error)
	Mkdir(path string) error
	Remove(path string) error
	Stat(path string) (os.FileInfo, error)
}

type SFTPFileInfo interface {
	Name() string
	Size() int64
	Mode() SFTPFileMode
	ModTime() time.Time
	IsDir() bool
}

type SFTPFileMode interface {
	String() string
}

// withSFTPClient safely executes an operation with an SFTP client
func (f *FileManagerService) withSFTPClient(operation func(SFTPClient) error) error {
	if f == nil {
		return fmt.Errorf("file manager service is nil")
	}

	if f.sshpkg == nil {
		return fmt.Errorf("ssh client is nil")
	}

	client, err := f.sshpkg.NewSftp()
	if err != nil {
		return err
	}
	defer client.Close()

	return operation(client)
}

// ListFiles returns a list of files in the given path
func (f *FileManagerService) ListFiles(path string) ([]FileData, error) {
	var fileData []FileData

	err := f.withSFTPClient(func(client SFTPClient) error {
		sftpFileInfos, err := client.ReadDir(path)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", path, err)
		}

		fileData = make([]FileData, 0, len(sftpFileInfos))
		for _, info := range sftpFileInfos {
			fileType := getFileType(info)

			var extension *string
			if !info.IsDir() {
				ext := filepath.Ext(info.Name())
				if ext != "" {
					extension = &ext
				}
			}

			sysInfo := info.Sys()
			var ownerId, groupId int64
			var permissions int64

			if statInfo, ok := sysInfo.(*syscall.Stat_t); ok {
				ownerId = int64(statInfo.Uid)
				groupId = int64(statInfo.Gid)
				permissions = int64(statInfo.Mode & 0777)
			}

			fullPath := filepath.Join(path, info.Name())

			fileData = append(fileData, FileData{
				Path:        fullPath,
				Name:        info.Name(),
				Size:        info.Size(),
				CreatedAt:   "",
				UpdatedAt:   info.ModTime().Format(time.RFC3339),
				FileType:    fileType,
				Permissions: permissions,
				IsHidden:    info.Name()[0] == '.',
				Extension:   extension,
				OwnerId:     ownerId,
				GroupId:     groupId,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileData, nil
}

// Helper function to determine file type
func getFileType(info os.FileInfo) string {
	if info.IsDir() {
		return "Directory"
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "Symlink"
	}
	if info.Mode().IsRegular() {
		return "File"
	}
	return "Other"
}

// FileData structure that matches the TypeScript interface
type FileData struct {
	Path        string  `json:"path"`
	Name        string  `json:"name"`
	Size        int64   `json:"size"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	FileType    string  `json:"file_type"`
	Permissions int64   `json:"permissions"`
	IsHidden    bool    `json:"is_hidden"`
	Extension   *string `json:"extension"`
	OwnerId     int64   `json:"owner_id"`
	GroupId     int64   `json:"group_id"`
}

// CreateDirectory creates a new directory at the given path and returns its contents
func (f *FileManagerService) CreateDirectory(path string) error {
	err := f.withSFTPClient(func(client SFTPClient) error {
		if err := client.Mkdir(path); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (f *FileManagerService) DeleteFile(path string) error {
	err := f.withSFTPClient(func(client SFTPClient) error {
		if err := client.Remove(path); err != nil {
			return fmt.Errorf("failed to delete file %s: %w", path, err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
