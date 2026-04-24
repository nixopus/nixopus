package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseResticSummary(t *testing.T) {
	id, n := parseResticSummary("")
	assert.Empty(t, id)
	assert.Zero(t, n)

	out := `ignored
{"message_type":"summary","snapshot_id":"snap-abc","total_bytes_processed":99}
`
	id, n = parseResticSummary(out)
	assert.Equal(t, "snap-abc", id)
	assert.Equal(t, int64(99), n)

	// summary line without snapshot_id; non-numeric total_bytes to skip Sscanf
	line := `{"message_type":"summary","total_bytes_processed":"x"}`
	id, n = parseResticSummary(line)
	assert.Empty(t, id)
	assert.Zero(t, n)

	badNum := `{"message_type":"summary","snapshot_id":"x","total_bytes_processed":"bad"}`
	id, n = parseResticSummary(badNum)
	assert.Equal(t, "x", id)
	assert.Zero(t, n)
}

func Test_extractJSONField(t *testing.T) {
	assert.Equal(t, "42", extractJSONField(`{"total_bytes_processed":42,"x":1}`, "total_bytes_processed"))
	assert.Empty(t, extractJSONField(`{"x":1}`, "missing"))
	assert.Equal(t, "tail", extractJSONField(`{"k":tail}`, "k"))
	assert.Equal(t, " rawvalue", extractJSONField(`{"total_bytes_processed": rawvalue`, "total_bytes_processed"))
}

func Test_regionOrDefault(t *testing.T) {
	assert.Equal(t, "us-east-1", regionOrDefault(""))
	assert.Equal(t, "eu-west-1", regionOrDefault("eu-west-1"))
}

func Test_extractIDFromChannel(t *testing.T) {
	assert.Equal(t, "rid", extractIDFromChannel("pfx:rid", "pfx:"))
	assert.Empty(t, extractIDFromChannel("other", "pfx:"))
}
