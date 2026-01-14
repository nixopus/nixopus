package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/extension/parser"
	"github.com/raghavyuva/nixopus-api/internal/features/extension/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

type ExtensionService struct {
	store   *shared_storage.Store
	storage storage.ExtensionStorageInterface
	ctx     context.Context
	logger  logger.Logger
}

func NewExtensionService(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	storage storage.ExtensionStorageInterface,
) *ExtensionService {
	return &ExtensionService{
		store:   store,
		storage: storage,
		ctx:     ctx,
		logger:  l,
	}
}

func (s *ExtensionService) CreateExtension(extension *types.Extension) error {
	if err := s.storage.CreateExtension(extension); err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return err
	}
	return nil
}

func (s *ExtensionService) GetExtension(id string) (*types.Extension, error) {
	extension, err := s.storage.GetExtension(id)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	return extension, nil
}

func (s *ExtensionService) GetExtensionByID(extensionID string) (*types.Extension, error) {
	extension, err := s.storage.GetExtensionByID(extensionID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	return extension, nil
}

func (s *ExtensionService) UpdateExtension(extension *types.Extension) error {
	if err := s.storage.UpdateExtension(extension); err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return err
	}
	return nil
}

func (s *ExtensionService) DeleteExtension(id string) error {
	if err := s.storage.DeleteExtension(id); err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return err
	}
	return nil
}

func (s *ExtensionService) DeleteFork(id string) error {
	ext, err := s.storage.GetExtension(id)
	if err != nil {
		return err
	}
	if ext.ParentExtensionID == nil {
		return fmt.Errorf("only forked extensions can be removed")
	}
	return s.storage.DeleteExtension(id)
}

func (s *ExtensionService) ForkExtension(extensionID string, yamlOverride string, authorName string) (*types.Extension, error) {
	src, err := s.storage.GetExtensionByID(extensionID)
	if err != nil {
		return nil, err
	}
	// Disallow forking a forked extension
	if src.ParentExtensionID != nil {
		return nil, fmt.Errorf("forking a forked extension is not allowed")
	}
	fork := *src
	fork.ID = uuid.UUID{}
	fork.ParentExtensionID = &src.ID
	fork.Name = src.Name + " (Fork)"
	// ensure unique extension_id for fork
	fork.ExtensionID = src.ExtensionID + "-fork-" + time.Now().Format("20060102150405")
	fork.CreatedAt = time.Now()
	fork.UpdatedAt = time.Now()
	fork.DeletedAt = nil

	if yamlOverride != "" {
		p := parser.NewParser()
		ext, variables, err := p.ParseExtensionContent(yamlOverride)
		if err != nil {
			return nil, err
		}
		fork.Name = ext.Name
		fork.Description = ext.Description
		if authorName != "" {
			fork.Author = authorName
		} else {
			fork.Author = ext.Author
		}
		fork.Icon = ext.Icon
		fork.Category = ext.Category
		fork.ExtensionType = ext.ExtensionType
		fork.Version = ext.Version
		fork.IsVerified = false
		fork.Featured = false
		fork.YAMLContent = yamlOverride
		fork.ParsedContent = ext.ParsedContent
		fork.ContentHash = ext.ContentHash

		if authorName != "" {
			fork.Author = authorName
		}
		if err := s.storage.CreateExtension(&fork); err != nil {
			return nil, err
		}
		if len(variables) > 0 {
			for i := range variables {
				variables[i].ExtensionID = fork.ID
			}
			if err := s.storage.CreateExtensionVariables(variables); err != nil {
				return nil, err
			}
		}
		return &fork, nil
	}

	if err := s.storage.CreateExtension(&fork); err != nil {
		return nil, err
	}

	// copy variables
	if len(src.Variables) > 0 {
		vars := make([]types.ExtensionVariable, 0, len(src.Variables))
		for _, v := range src.Variables {
			vars = append(vars, types.ExtensionVariable{
				ExtensionID:       fork.ID,
				VariableName:      v.VariableName,
				VariableType:      v.VariableType,
				Description:       v.Description,
				DefaultValue:      v.DefaultValue,
				IsRequired:        v.IsRequired,
				ValidationPattern: v.ValidationPattern,
				CreatedAt:         time.Now(),
			})
		}
		if err := s.storage.CreateExtensionVariables(vars); err != nil {
			return nil, err
		}
	}
	return &fork, nil
}

func (s *ExtensionService) ListExtensions(params types.ExtensionListParams) (*types.ExtensionListResponse, error) {
	response, err := s.storage.ListExtensions(params)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	return response, nil
}

func (s *ExtensionService) ListCategories() ([]types.ExtensionCategory, error) {
	cats, err := s.storage.ListCategories()
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	return cats, nil
}

func (s *ExtensionService) ParseMultipartRunRequest(r *http.Request) (map[string]interface{}, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	vars := map[string]interface{}{}
	if raw := r.FormValue("variables"); raw != "" {
		_ = json.Unmarshal([]byte(raw), &vars)
	}
	file, header, err := r.FormFile("file")
	if err == nil && file != nil {
		defer file.Close()
		tmpDir := os.TempDir()
		tmpPath := filepath.Join(tmpDir, header.Filename)
		out, err := os.Create(tmpPath)
		if err != nil {
			return nil, err
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			return nil, err
		}
		vars["uploaded_file_path"] = tmpPath
	}
	return vars, err
}

func (s *ExtensionService) CancelExecution(id string) error {
	exec, err := s.storage.GetExecutionByID(id)
	if err != nil {
		return err
	}
	if exec.Status == types.ExecutionStatusCompleted || exec.Status == types.ExecutionStatusFailed {
		return nil
	}
	now := time.Now()
	exec.Status = types.ExecutionStatusCancelled
	exec.CompletedAt = &now
	return s.storage.UpdateExecution(exec)
}

func (s *ExtensionService) GetExecutionByID(id string) (*types.ExtensionExecution, error) {
	exec, err := s.storage.GetExecutionByID(id)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	return exec, nil
}

func (s *ExtensionService) ListExecutionsByExtensionID(extensionID string) ([]types.ExtensionExecution, error) {
	execs, err := s.storage.ListExecutionsByExtensionID(extensionID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	return execs, nil
}

func (s *ExtensionService) appendLog(executionID uuid.UUID, stepID *uuid.UUID, level string, message string, data map[string]interface{}) {
	seq, err := s.storage.NextLogSequence(executionID.String())
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("failed to get next log sequence: %v", err), "")
		return
	}
	var payload []byte
	if data != nil {
		payload, _ = json.Marshal(data)
	} else {
		payload = []byte("{}")
	}
	log := &types.ExtensionLog{
		ExecutionID: executionID,
		StepID:      stepID,
		Level:       level,
		Message:     message,
		Data:        payload,
		Sequence:    seq,
	}
	_ = s.storage.CreateExtensionLog(log)
}

func (s *ExtensionService) ListExecutionLogs(executionID string, afterSeq int64, limit int) ([]types.ExtensionLog, *types.ExecutionStatus, error) {
	logs, err := s.storage.ListExtensionLogs(executionID, afterSeq, limit)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, nil, err
	}

	exec, err := s.storage.GetExecutionByID(executionID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, nil, err
	}

	return logs, &exec.Status, nil
}
