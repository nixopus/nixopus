package service

import (
	"errors"
	"testing"

	"github.com/docker/docker/api/types/container"
	dockerfilters "github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/stretchr/testify/require"
)

func TestListImages_containerIDRejected(t *testing.T) {
	t.Parallel()
	stub := &stubDockerRepository{}
	stub.getContainerByID = func(string) (container.InspectResponse, error) {
		return container.InspectResponse{}, errors.New("missing")
	}
	_, err := ListImages(stub, logger.NewLogger(), ListImagesOptions{ContainerID: "nope"})
	require.Error(t, err)
}

func TestListImages_emptyResult(t *testing.T) {
	t.Parallel()
	stub := &stubDockerRepository{}
	stub.getContainerByID = func(string) (container.InspectResponse, error) {
		return container.InspectResponse{}, nil
	}
	stub.listAllImages = func(image.ListOptions) []image.Summary { return nil }
	resp, err := ListImages(stub, logger.NewLogger(), ListImagesOptions{ContainerID: "ok"})
	require.NoError(t, err)
	require.Equal(t, "No images found", resp.Message)
	require.Empty(t, resp.Data)
}

func TestListImages_referenceFilter_appendsWildcard(t *testing.T) {
	t.Parallel()
	want := dockerfilters.NewArgs()
	want.Add("reference", "myrepo*")
	stub := &stubDockerRepository{}
	stub.listAllImages = func(opts image.ListOptions) []image.Summary {
		require.Equal(t, want.Get("reference"), opts.Filters.Get("reference"))
		return []image.Summary{}
	}
	_, err := ListImages(stub, logger.NewLogger(), ListImagesOptions{ImagePrefix: "myrepo"})
	require.NoError(t, err)

	stub2 := &stubDockerRepository{}
	stub2.listAllImages = func(opts image.ListOptions) []image.Summary {
		require.Equal(t, []string{"pinned*"}, opts.Filters.Get("reference"))
		return []image.Summary{}
	}
	_, err = ListImages(stub2, logger.NewLogger(), ListImagesOptions{ImagePrefix: "pinned*"})
	require.NoError(t, err)
}

func TestListImages_mapsNilSlicesToConcrete(t *testing.T) {
	t.Parallel()
	stub := &stubDockerRepository{}
	stub.listAllImages = func(image.ListOptions) []image.Summary {
		return []image.Summary{
			{
				ID:          "id1",
				Created:     55,
				Size:        99,
				SharedSize:  1,
				VirtualSize: 100,
			},
		}
	}
	resp, err := ListImages(stub, logger.NewLogger(), ListImagesOptions{})
	require.NoError(t, err)
	require.Equal(t, "Images listed successfully", resp.Message)
	require.Len(t, resp.Data, 1)
	im := resp.Data[0]
	require.Empty(t, im.RepoDigests)
	require.Empty(t, im.RepoTags)
	require.NotNil(t, im.Labels)
}

func TestPruneImages_paths(t *testing.T) {
	t.Parallel()
	stub := &stubDockerRepository{}
	stub.pruneImages = func(f dockerfilters.Args) (image.PruneReport, error) {
		require.Equal(t, []string{"t"}, f.Get("until"))
		require.Equal(t, []string{"k=v"}, f.Get("label"))
		require.Equal(t, []string{"true"}, f.Get("dangling"))
		return image.PruneReport{}, errors.New("prune boom")
	}
	_, err := PruneImages(stub, logger.NewLogger(), PruneImagesOptions{
		Until: "t", Label: "k=v", Dangling: true,
	})
	require.Error(t, err)

	stub.pruneImages = func(dockerfilters.Args) (image.PruneReport, error) {
		return image.PruneReport{
			SpaceReclaimed: 42,
			ImagesDeleted: []image.DeleteResponse{
				{Untagged: "a", Deleted: "b"},
			},
		}, nil
	}
	resp, err := PruneImages(stub, logger.NewLogger(), PruneImagesOptions{})
	require.NoError(t, err)
	require.Len(t, resp.Data.ImagesDeleted, 1)
	require.Equal(t, "a", resp.Data.ImagesDeleted[0].Untagged)
	require.Equal(t, "b", resp.Data.ImagesDeleted[0].Deleted)
	require.Equal(t, uint64(42), resp.Data.SpaceReclaimed)
}

func TestUpdateContainerResources_validationAndSuccess(t *testing.T) {
	t.Parallel()
	log := logger.NewLogger()
	stub := &stubDockerRepository{}

	_, err := UpdateContainerResources(stub, log, UpdateContainerResourcesOptions{
		ContainerID: "c",
		Memory:      5 * 1024 * 1024,
	})
	require.Error(t, err)

	_, err = UpdateContainerResources(stub, log, UpdateContainerResourcesOptions{
		ContainerID: "c",
		Memory:      10 * 1024 * 1024,
		MemorySwap:  5 * 1024 * 1024,
	})
	require.Error(t, err)

	_, err = UpdateContainerResources(stub, log, UpdateContainerResourcesOptions{
		ContainerID: "c",
		CPUShares:   1,
	})
	require.Error(t, err)

	stub.getContainerByID = func(string) (container.InspectResponse, error) {
		return container.InspectResponse{}, errors.New("inspect fail")
	}
	_, err = UpdateContainerResources(stub, log, UpdateContainerResourcesOptions{
		ContainerID: "c",
		MemorySwap:  -1,
	})
	require.Error(t, err)

	stub.getContainerByID = func(string) (container.InspectResponse, error) {
		return container.InspectResponse{
			ContainerJSONBase: &container.ContainerJSONBase{
				State: &container.State{},
			},
		}, nil
	}
	_, err = UpdateContainerResources(stub, log, UpdateContainerResourcesOptions{
		ContainerID: "c",
	})
	require.Error(t, err)

	stub.getContainerByID = func(string) (container.InspectResponse, error) {
		return container.InspectResponse{
			ContainerJSONBase: &container.ContainerJSONBase{
				State: &container.State{Running: true},
			},
		}, nil
	}
	stub.updateResources = func(string, container.UpdateConfig) (container.ContainerUpdateOKBody, error) {
		return container.ContainerUpdateOKBody{}, errors.New("update api")
	}
	_, err = UpdateContainerResources(stub, log, UpdateContainerResourcesOptions{ContainerID: "c"})
	require.Error(t, err)

	const memOk = int64(16 * 1024 * 1024)
	stub.updateResources = func(id string, cfg container.UpdateConfig) (container.ContainerUpdateOKBody, error) {
		require.Equal(t, "ok", id)
		require.Equal(t, memOk, cfg.Resources.Memory)
		require.Equal(t, int64(-1), cfg.Resources.MemorySwap)
		require.Equal(t, int64(2), cfg.Resources.CPUShares)
		return container.ContainerUpdateOKBody{Warnings: []string{"w1"}}, nil
	}
	respOK, err := UpdateContainerResources(stub, log, UpdateContainerResourcesOptions{
		ContainerID: "ok",
		Memory:      memOk,
		MemorySwap:  -1,
		CPUShares:   2,
	})
	require.NoError(t, err)
	require.Equal(t, "success", respOK.Status)
	require.Equal(t, []string{"w1"}, respOK.Data.Warnings)
}
