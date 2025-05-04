package controller

import (
	"bytes"
	"encoding/binary"
	"io"
	"net/http"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *ContainerController) GetContainerLogs(f fuego.ContextWithBody[types.ContainerLogsRequest]) (*shared_types.Response, error) {
	req, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	logsReader, err := c.dockerService.GetContainerLogs(req.ID, container.LogsOptions{
		Follow:     req.Follow,
		Tail:       strconv.Itoa(req.Tail),
		Since:      req.Since,
		Until:      req.Until,
		ShowStdout: req.Stdout,
		ShowStderr: req.Stderr,
	})
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logsReader)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	decodedLogs := decodeDockerLogs(buf.Bytes())

	return &shared_types.Response{
		Status:  "success",
		Message: "Container logs fetched successfully",
		Data:    decodedLogs,
	}, nil
}

func decodeDockerLogs(data []byte) string {
	var result bytes.Buffer
	offset := 0

	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}

		streamType := data[offset]
		length := binary.BigEndian.Uint32(data[offset+4 : offset+8])
		offset += 8

		if offset+int(length) > len(data) {
			break
		}

		if streamType == 1 || streamType == 2 {
			result.Write(data[offset : offset+int(length)])
		}
		offset += int(length)
	}

	return result.String()
}
