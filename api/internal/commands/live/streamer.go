package live

import "strings"

// Streamer buffers incoming agent response chunks and flushes complete lines
// to the writer. This gives a progressive streaming feel — text appears
// line by line as the agent generates it.
type Streamer struct {
	writer *Writer
	buf    strings.Builder
}

// NewStreamer creates a streamer that outputs through the given writer.
func NewStreamer(writer *Writer) *Streamer {
	return &Streamer{writer: writer}
}

// WriteChunk appends a text chunk from the agent SSE stream.
// Complete lines (ending with \n) are flushed immediately to writer.Detail().
// Partial lines stay in the buffer until the next chunk completes them.
func (s *Streamer) WriteChunk(chunk string) {
	s.buf.WriteString(chunk)

	for {
		content := s.buf.String()
		idx := strings.Index(content, "\n")
		if idx < 0 {
			break
		}
		line := content[:idx]
		s.buf.Reset()
		s.buf.WriteString(content[idx+1:])

		if line == "" {
			s.writer.Blank()
		} else {
			s.writer.Detail(line)
		}
	}
}

// Flush outputs any remaining buffered text. Call this when the agent
// response stream ends to ensure nothing is left in the buffer.
func (s *Streamer) Flush() {
	remaining := s.buf.String()
	s.buf.Reset()
	if remaining != "" {
		s.writer.Detail(remaining)
	}
}
