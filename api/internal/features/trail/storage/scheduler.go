package storage

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type serverCandidate struct {
	bun.BaseModel `bun:"table:infra_servers,alias:isr"`

	ID         uuid.UUID `bun:"id"`
	MaxVCPUs   int       `bun:"max_vcpus"`
	MaxMemMB   int       `bun:"max_memory_mb"`
	MaxDiskGB  int       `bun:"max_disk_gb"`
	UsedVCPUs  int       `bun:"used_vcpus"`
	UsedMemMB  int       `bun:"used_mem"`
	UsedDiskGB int       `bun:"used_disk"`
}

// activeProvisionStatuses lists provision_status values where a VM is
// consuming resources on the host (in-progress or running).
var activeProvisionStatuses = []string{
	"initializing", "creating_container",
	"configuring_ssh", "setting_up_subdomain", "completed",
}

// SelectBestServer picks the active infra server with the most remaining
// capacity that can fit the requested resources. Returns an empty string
// (not an error) when no infra_servers are registered, so the caller can
// fall back to legacy queue names for backward compatibility.
func (s *TrailStorage) SelectBestServer(vcpus, memMB, diskGB int) (string, error) {
	var candidates []serverCandidate

	err := s.DB.NewSelect().
		TableExpr("infra_servers AS isr").
		ColumnExpr("isr.id").
		ColumnExpr("isr.max_vcpus").
		ColumnExpr("isr.max_memory_mb").
		ColumnExpr("isr.max_disk_gb").
		ColumnExpr("COALESCE(SUM(upd.vcpu_count), 0) + COALESCE(pv.pool_vcpus, 0) AS used_vcpus").
		ColumnExpr("COALESCE(SUM(upd.memory_mb), 0) + COALESCE(pv.pool_mem, 0) AS used_mem").
		ColumnExpr("COALESCE(SUM(upd.disk_size_gb), 0) + COALESCE(pv.pool_disk, 0) AS used_disk").
		Join(`LEFT JOIN user_provision_details AS upd ON upd.server_id = isr.id`+
			` AND EXISTS (SELECT 1 FROM "user" u WHERE u.id = upd.user_id AND u.provision_status IN (?))`,
			bun.In(activeProvisionStatuses)).
		Join(`LEFT JOIN (`+
			`SELECT vp.server_id, `+
			`SUM(vp.vcpu_count) AS pool_vcpus, `+
			`SUM(vp.memory_mb) AS pool_mem, `+
			`SUM(vp.disk_size_gb) AS pool_disk `+
			`FROM vm_pool AS vp `+
			`WHERE vp.status IN ('warm', 'claiming') `+
			`GROUP BY vp.server_id`+
			`) AS pv ON pv.server_id = isr.id`).
		Where("isr.status = ?", "active").
		GroupExpr("isr.id, pv.pool_vcpus, pv.pool_mem, pv.pool_disk").
		Scan(s.Ctx, &candidates)

	if err != nil {
		return "", fmt.Errorf("scheduler query failed: %w", err)
	}

	if len(candidates) == 0 {
		return "", nil
	}

	return pickBestServer(candidates, vcpus, memMB, diskGB)
}

// pickBestServer selects the server with the most remaining headroom on its
// tightest resource dimension, among servers that can fit the request.
func pickBestServer(candidates []serverCandidate, vcpus, memMB, diskGB int) (string, error) {
	bestIdx := -1
	bestMinHeadroom := -1.0

	for i, c := range candidates {
		freeVCPU := c.MaxVCPUs - c.UsedVCPUs
		freeMem := c.MaxMemMB - c.UsedMemMB
		freeDisk := c.MaxDiskGB - c.UsedDiskGB

		if freeVCPU < vcpus || freeMem < memMB || freeDisk < diskGB {
			continue
		}

		hCPU := float64(freeVCPU-vcpus) / float64(max(c.MaxVCPUs, 1))
		hMem := float64(freeMem-memMB) / float64(max(c.MaxMemMB, 1))
		hDisk := float64(freeDisk-diskGB) / float64(max(c.MaxDiskGB, 1))

		minH := hCPU
		if hMem < minH {
			minH = hMem
		}
		if hDisk < minH {
			minH = hDisk
		}

		if minH > bestMinHeadroom {
			bestMinHeadroom = minH
			bestIdx = i
		}
	}

	if bestIdx < 0 {
		return "", fmt.Errorf("no server has enough capacity for the requested resources (vcpus=%d mem=%dMB disk=%dGB)", vcpus, memMB, diskGB)
	}

	return candidates[bestIdx].ID.String(), nil
}
