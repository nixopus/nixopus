package live

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	chunkSmallFile = 60
	chunkMin       = 4
	chunkMax       = 200
)

// fileChunk represents a semantic chunk of file content.
type fileChunk struct {
	hash      string
	filePath  string
	startLine int
	endLine   int
	content   string
}

var chunkExtLang = map[string]string{
	".ts": "typescript", ".tsx": "typescript", ".mts": "typescript", ".cts": "typescript",
	".js": "javascript", ".jsx": "javascript", ".mjs": "javascript", ".cjs": "javascript",
	".py":   "python",
	".go":   "go",
	".rs":   "rust",
	".java": "java", ".kt": "kotlin", ".scala": "scala",
	".c": "c", ".h": "c", ".cpp": "cpp", ".cc": "cpp", ".hpp": "cpp", ".cxx": "cpp",
	".cs":    "csharp",
	".rb":    "ruby",
	".php":   "php",
	".swift": "swift",
	".lua":   "lua",
	".r":     "r",
	".ex":    "elixir", ".exs": "elixir",
	".erl": "erlang",
	".hs":  "haskell",
	".ml":  "ocaml", ".mli": "ocaml",
	".zig":  "zig",
	".dart": "dart",
	".elm":  "elm",
	".clj":  "clojure", ".cljs": "clojure", ".cljc": "clojure",
	".vue": "vue", ".svelte": "svelte",
	".sh": "shell", ".bash": "shell", ".zsh": "shell", ".fish": "shell",
	".sql":  "sql",
	".yaml": "yaml", ".yml": "yaml",
	".toml": "toml",
	".json": "json",
	".xml":  "xml", ".html": "html", ".htm": "html",
	".css": "css", ".scss": "scss", ".less": "less",
	".md": "markdown", ".mdx": "markdown", ".rst": "markdown",
	".dockerfile": "dockerfile",
	".tf":         "terraform", ".hcl": "terraform",
	".proto":   "protobuf",
	".graphql": "graphql", ".gql": "graphql",
	".sol": "solidity",
	".v":   "v",
	".nim": "nim",
	".cr":  "crystal",
	".pl":  "perl", ".pm": "perl",
}

var chunkBoundaryPatterns = map[string]*regexp.Regexp{
	"typescript": regexp.MustCompile(`^(?:export\s+)?(?:async\s+)?(?:function|class|interface|type|enum|const|let|abstract\s+class|namespace)\s`),
	"javascript": regexp.MustCompile(`^(?:export\s+)?(?:async\s+)?(?:function|class|const|let|var|module\.exports)\s`),
	"python":     regexp.MustCompile(`^(?:async\s+)?(?:def|class)\s|^@\w|^(?:import|from)\s`),
	"go":         regexp.MustCompile(`^(?:func|type|var|const|import|package)\s`),
	"rust":       regexp.MustCompile(`^(?:pub(?:\([^)]*\))?\s+)?(?:async\s+)?(?:fn|struct|enum|trait|impl|mod|type|use|const|static)\s|^#\[`),
	"java":       regexp.MustCompile(`^(?:public|private|protected|static|final|abstract|class|interface|enum|import|package|@\w)\s`),
	"kotlin":     regexp.MustCompile(`^(?:fun|class|interface|object|enum|import|package|val|var|data\s+class|sealed\s+class|annotation)\s`),
	"scala":      regexp.MustCompile(`^(?:def|class|trait|object|val|var|import|package|case\s+class|sealed\s+trait)\s`),
	"csharp":     regexp.MustCompile(`^(?:public|private|protected|internal|static|class|interface|enum|struct|namespace|using|record)\s`),
	"ruby":       regexp.MustCompile(`^(?:def|class|module|require|include|attr_|private|protected|public)\s`),
	"php":        regexp.MustCompile(`^(?:function|class|interface|trait|namespace|use|public|private|protected|static|abstract|final)\s`),
	"swift":      regexp.MustCompile(`^(?:func|class|struct|enum|protocol|import|extension|var|let|typealias|actor)\s`),
	"c":          regexp.MustCompile(`^(?:void|int|char|float|double|long|short|unsigned|signed|static|extern|struct|enum|union|typedef)\s`),
	"cpp":        regexp.MustCompile(`^(?:void|int|char|float|double|long|short|unsigned|signed|static|extern|class|struct|enum|union|typedef|namespace|template|virtual|inline|constexpr)\s`),
	"lua":        regexp.MustCompile(`^(?:function|local\s+function)\s`),
	"elixir":     regexp.MustCompile(`^(?:def|defp|defmodule|defmacro|defstruct|defguard|defprotocol|defimpl|import|use|alias)\s`),
	"dart":       regexp.MustCompile(`^(?:class|void|Future|Stream|int|double|String|bool|List|Map|Set|abstract|mixin|extension|enum)\s`),
	"haskell":    regexp.MustCompile(`^(?:module|import|data|type|class|instance|newtype|deriving)\s|^[a-z]\w*\s*::`),
	"ocaml":      regexp.MustCompile(`^(?:let|type|module|open|val|external|class)\s`),
	"solidity":   regexp.MustCompile(`^(?:contract|function|event|modifier|struct|enum|mapping|library|interface|pragma)\s`),
}

func chunkDetectLanguage(filename string) string {
	base := strings.ToLower(filename)
	if strings.HasSuffix(base, "dockerfile") || strings.HasSuffix(base, ".dockerfile") {
		return "dockerfile"
	}
	if strings.HasSuffix(base, "makefile") || strings.HasSuffix(base, ".mk") {
		return "makefile"
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if lang, ok := chunkExtLang[ext]; ok {
		return lang
	}
	return "unknown"
}

func chunkFile(filePath string, content []byte) []fileChunk {
	s := string(content)
	lines := strings.Split(s, "\n")

	if len(lines) <= chunkSmallFile {
		text := strings.TrimSpace(s)
		if text == "" {
			return nil
		}
		return []fileChunk{makeFileChunk(filePath, s, 1, len(lines))}
	}

	language := chunkDetectLanguage(filePath)
	pattern := chunkBoundaryPatterns[language]

	if pattern == nil {
		return chunkAtBlankLines(filePath, lines)
	}
	return chunkAtBoundaries(filePath, lines, pattern)
}

func chunkAtBoundaries(filePath string, lines []string, pattern *regexp.Regexp) []fileChunk {
	breaks := []int{0}

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if len(line) == 0 {
			continue
		}
		trimmed := strings.TrimLeft(line, " \t")
		if len(trimmed) == 0 {
			continue
		}
		if len(line) != len(trimmed) {
			continue
		}
		if pattern.MatchString(trimmed) && i-breaks[len(breaks)-1] >= chunkMin {
			breaks = append(breaks, i)
		}
	}

	return chunkAssemble(filePath, lines, breaks)
}

func chunkAtBlankLines(filePath string, lines []string) []fileChunk {
	breaks := []int{0}
	blanks := 0

	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			blanks++
		} else {
			if blanks >= 2 && i-breaks[len(breaks)-1] >= chunkMin {
				breaks = append(breaks, i)
			}
			blanks = 0
		}
	}

	return chunkAssemble(filePath, lines, breaks)
}

func chunkAssemble(filePath string, lines []string, breaks []int) []fileChunk {
	var chunks []fileChunk

	for i := 0; i < len(breaks); i++ {
		start := breaks[i]
		end := len(lines)
		if i+1 < len(breaks) {
			end = breaks[i+1]
		}

		if end-start > chunkMax {
			for j := start; j < end; j += chunkMax {
				blockEnd := j + chunkMax
				if blockEnd > end {
					blockEnd = end
				}
				text := strings.Join(lines[j:blockEnd], "\n")
				if strings.TrimSpace(text) != "" {
					chunks = append(chunks, makeFileChunk(filePath, text, j+1, blockEnd))
				}
			}
		} else {
			text := strings.Join(lines[start:end], "\n")
			if strings.TrimSpace(text) != "" {
				chunks = append(chunks, makeFileChunk(filePath, text, start+1, end))
			}
		}
	}

	if len(chunks) == 0 {
		text := strings.Join(lines, "\n")
		if strings.TrimSpace(text) != "" {
			chunks = append(chunks, makeFileChunk(filePath, text, 1, len(lines)))
		}
	}
	return chunks
}

func makeFileChunk(filePath string, content string, startLine, endLine int) fileChunk {
	h := sha256.Sum256([]byte(content))
	return fileChunk{
		hash:      hex.EncodeToString(h[:]),
		filePath:  filePath,
		startLine: startLine,
		endLine:   endLine,
		content:   content,
	}
}

// IndexFileChunks chunks file content and persists to application_file_chunks.
// Skips binary/large files. Call when file is written (has content).
func IndexFileChunks(ctx context.Context, store *shared_storage.Store, applicationID uuid.UUID, path string, content []byte) error {
	if store == nil || store.DB == nil {
		return nil
	}
	if len(content) > maxIndexableSize() {
		return nil
	}
	// Skip binary-ish content
	if !isTextContent(content) {
		return nil
	}

	chunks := chunkFile(path, content)
	if len(chunks) == 0 {
		return nil
	}

	language := chunkDetectLanguage(path)

	// Delete existing chunks for this path, then insert new ones
	_, err := store.DB.NewDelete().
		Model((*types.ApplicationFileChunk)(nil)).
		Where("application_id = ? AND path = ?", applicationID, path).
		Exec(ctx)
	if err != nil {
		return err
	}

	for _, c := range chunks {
		ac := &types.ApplicationFileChunk{
			ID:            uuid.New(),
			ApplicationID: applicationID,
			Path:          path,
			StartLine:     c.startLine,
			EndLine:       c.endLine,
			Content:       c.content,
			ChunkHash:     c.hash,
			Language:      language,
			CreatedAt:     time.Now().UTC(),
		}
		_, err := store.DB.NewInsert().Model(ac).Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteFileChunks removes all chunks for a path. Call when file is deleted.
func DeleteFileChunks(ctx context.Context, store *shared_storage.Store, applicationID uuid.UUID, path string) error {
	if store == nil || store.DB == nil {
		return nil
	}
	_, err := store.DB.NewDelete().
		Model((*types.ApplicationFileChunk)(nil)).
		Where("application_id = ? AND path = ?", applicationID, path).
		Exec(ctx)
	return err
}

// isTextContent returns true if content looks like text (not binary).
func isTextContent(b []byte) bool {
	const maxCheck = 1024
	n := len(b)
	if n > maxCheck {
		n = maxCheck
	}
	for i := 0; i < n; i++ {
		c := b[i]
		if c == 0 {
			return false
		}
		if c < 32 && c != '\t' && c != '\n' && c != '\r' {
			return false
		}
	}
	return true
}
