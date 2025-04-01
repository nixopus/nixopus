package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pkg/sftp"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
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
	MkdirAll(path string) error
	Remove(path string) error
	Stat(path string) (os.FileInfo, error)
	Rename(fromPath string, toPath string) error
	Create(path string) (*sftp.File, error)
	Open(path string) (*sftp.File, error)
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
	f.logger.Log(logger.Info, "creating directory", path)
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

func (f *FileManagerService) DeleteDirectory(path string) error {
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

func (f *FileManagerService) MoveDirectory(fromPath string, toPath string) error {
	err := f.withSFTPClient(func(client SFTPClient) error {
		if err := client.Rename(fromPath, toPath); err != nil {
			return fmt.Errorf("failed to move directory %s to %s: %w", fromPath, toPath, err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// UploadFile handles file upload to the specified path
func (f *FileManagerService) UploadFile(file io.Reader, path string, filename string) error {
	f.logger.Log(logger.Info, "uploading file", filename)

	err := f.withSFTPClient(func(client SFTPClient) error {
		if err := client.MkdirAll(path); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
		targetPath := filepath.Join(path, filename)
		out, err := client.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", targetPath, err)
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			return fmt.Errorf("failed to write file content: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (f *FileManagerService) CopyDirectory(fromPath string, toPath string) error {
	f.logger.Log(logger.Info, "copying directory", fmt.Sprintf("from %s to %s", fromPath, toPath))

	err := f.withSFTPClient(func(client SFTPClient) error {
		sourceInfo, err := client.Stat(fromPath)
		if err != nil {
			return fmt.Errorf("source path %s does not exist: %w", fromPath, err)
		}

		if !sourceInfo.IsDir() {
			return f.copyFile(client, fromPath, toPath)
		}

		if err := client.MkdirAll(toPath); err != nil {
			return fmt.Errorf("failed to create target directory %s: %w", toPath, err)
		}

		files, err := client.ReadDir(fromPath)
		if err != nil {
			return fmt.Errorf("failed to read source directory %s: %w", fromPath, err)
		}

		for _, file := range files {
			sourcePath := filepath.Join(fromPath, file.Name())
			targetPath := filepath.Join(toPath, file.Name())

			if file.IsDir() {
				if err := f.CopyDirectory(sourcePath, targetPath); err != nil {
					return fmt.Errorf("failed to copy directory %s: %w", sourcePath, err)
				}
			} else {
				if err := f.copyFile(client, sourcePath, targetPath); err != nil {
					return fmt.Errorf("failed to copy file %s: %w", sourcePath, err)
				}
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("copy operation failed: %w", err)
	}

	return nil
}

func (f *FileManagerService) copyFile(client SFTPClient, sourcePath, targetPath string) error {
	sourceFile, err := client.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", sourcePath, err)
	}
	defer sourceFile.Close()

	targetFile, err := client.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target file %s: %w", targetPath, err)
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file content from %s to %s: %w", sourcePath, targetPath, err)
	}

	return nil
}

func (f *FileManagerService) DeleteFile(path string) error {
	f.logger.Log(logger.Info, "deleting file/directory", path)

	err := f.withSFTPClient(func(client SFTPClient) error {
		info, err := client.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", path, err)
		}

		if info.IsDir() {
			files, err := client.ReadDir(path)
			if err != nil {
				return fmt.Errorf("failed to read directory %s: %w", path, err)
			}

			for _, file := range files {
				filePath := filepath.Join(path, file.Name())
				if err := f.DeleteFile(filePath); err != nil {
					return err
				}
			}
		}

		if err := client.Remove(path); err != nil {
			return fmt.Errorf("failed to delete %s: %w", path, err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
