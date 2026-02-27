package syncproto

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"sort"
)

const merkleFanOut = 16

// Tree represents a Merkle tree over path→checksum leaves.
// Used for efficient diff: root comparison can skip full enumeration.
type Tree struct {
	RootHash string
	leaves   map[string]string // path -> checksum (normalized paths)
}

// BuildFromPaths builds a Merkle tree from path→checksum leaves.
// Paths are sorted deterministically for identical tree structure.
func BuildFromPaths(leaves map[string]string) *Tree {
	if len(leaves) == 0 {
		return &Tree{RootHash: "", leaves: make(map[string]string)}
	}

	norm := make(map[string]string, len(leaves))
	for p, c := range leaves {
		norm[normalizeMerklePath(p)] = c
	}

	paths := make([]string, 0, len(norm))
	for p := range norm {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	hashes := make([]string, len(paths))
	for i, p := range paths {
		hashes[i] = norm[p]
	}

	root := buildInternal(hashes)
	return &Tree{RootHash: root, leaves: norm}
}

// buildInternal builds internal nodes bottom-up. Each level hashes
// sorted child hashes. Fan-out of merkleFanOut for balanced height.
func buildInternal(hashes []string) string {
	if len(hashes) == 0 {
		return ""
	}
	if len(hashes) == 1 {
		return hashes[0]
	}

	// Group into chunks of fanOut, hash each chunk
	var next []string
	for i := 0; i < len(hashes); i += merkleFanOut {
		end := i + merkleFanOut
		if end > len(hashes) {
			end = len(hashes)
		}
		chunk := hashes[i:end]
		h := hashConcat(chunk)
		next = append(next, h)
	}

	return buildInternal(next)
}

func hashConcat(hashes []string) string {
	h := sha256.New()
	for _, s := range hashes {
		decoded, _ := hex.DecodeString(s)
		h.Write(decoded)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Leaves returns a copy of path→checksum for the tree.
func (t *Tree) Leaves() map[string]string {
	if t.leaves == nil {
		return nil
	}
	out := make(map[string]string, len(t.leaves))
	for p, c := range t.leaves {
		out[p] = c
	}
	return out
}

// DiffAgainst returns paths that differ between local tree and server leaves.
// Includes: paths to add/modify (local different or missing on server),
// and paths to delete (server has, local doesn't - caller handles as delete).
func DiffAgainst(local *Tree, serverLeaves map[string]string) (toSync []string, toDelete []string) {
	serverNorm := make(map[string]string)
	for p, c := range serverLeaves {
		serverNorm[normalizeMerklePath(p)] = c
	}

	localLeaves := local.Leaves()
	if localLeaves == nil {
		localLeaves = make(map[string]string)
	}

	for path, localHash := range localLeaves {
		serverHash, ok := serverNorm[path]
		if !ok || serverHash != localHash {
			toSync = append(toSync, path)
		}
	}

	for path := range serverNorm {
		if _, ok := localLeaves[path]; !ok {
			toDelete = append(toDelete, path)
		}
	}

	return toSync, toDelete
}

func normalizeMerklePath(p string) string {
	return filepath.ToSlash(filepath.Clean(p))
}

// ComputeSimhash produces a 256-bit simhash from path→checksum leaves.
// Octodel parity: hashes each file checksum into a 256-dim vector, thresholds, outputs hex.
// Used for codebase similarity fingerprint.
func ComputeSimhash(leaves map[string]string) string {
	if len(leaves) == 0 {
		return ""
	}
	v := make([]int32, 256)

	for _, hash := range leaves {
		if len(hash) != 64 {
			continue
		}
		decoded, err := hex.DecodeString(hash)
		if err != nil || len(decoded) != 32 {
			continue
		}
		for i := 0; i < 256; i++ {
			byteIdx := i >> 3
			bitIdx := i & 7
			if (decoded[byteIdx]>>bitIdx)&1 == 1 {
				v[i]++
			} else {
				v[i]--
			}
		}
	}

	out := make([]byte, 32)
	for i := 0; i < 256; i++ {
		if v[i] > 0 {
			byteIdx := i >> 3
			bitIdx := i & 7
			out[byteIdx] |= 1 << bitIdx
		}
	}
	return hex.EncodeToString(out)
}
