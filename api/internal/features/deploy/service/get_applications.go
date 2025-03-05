package service

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"strconv"
)

func (s *DeployService) GetApplications(page string, pageSize string) ([]shared_types.Application, int, error) {
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		return nil, 0, err
	}
	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil {
		return nil, 0, err
	}
	applications, totalCount, err := s.storage.GetApplications(pageNum, pageSizeNum)
	if err != nil {
		return nil, 0, err
	}

	return applications, totalCount, nil
}
