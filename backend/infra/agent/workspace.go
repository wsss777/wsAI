package agent

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const maxFileBytes = 128 * 1024

var blockedNames = map[string]bool{".git": true, ".env": true, "node_modules": true, ".idea": true, ".gocache": true}

type Workspace struct{ root string }
type FileEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
	Size int64  `json:"size"`
}
type SearchMatch struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Text string `json:"text"`
}
type FileContent struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

func NewWorkspace(root string) (*Workspace, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("agent workspace root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return nil, errors.New("agent workspace root is not a directory")
	}
	return &Workspace{root: resolved}, nil
}

func (w *Workspace) Resolve(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = "."
	}
	if filepath.IsAbs(raw) || strings.Contains(raw, "..") {
		return "", "", errors.New("path must be a workspace-relative path")
	}
	clean := filepath.Clean(raw)
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		if blockedNames[strings.ToLower(part)] {
			return "", "", errors.New("path is not available to agent")
		}
	}
	target := filepath.Join(w.root, clean)
	resolved, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", "", err
	}
	rel, err := filepath.Rel(w.root, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", errors.New("path escapes workspace")
	}
	return resolved, filepath.ToSlash(rel), nil
}

func (w *Workspace) List(path string, depth int) ([]FileEntry, error) {
	if depth < 0 || depth > 4 {
		depth = 2
	}
	target, _, err := w.Resolve(path)
	if err != nil {
		return nil, err
	}
	entries := make([]FileEntry, 0)
	err = filepath.WalkDir(target, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		rel, _ := filepath.Rel(target, p)
		if rel != "." && len(strings.Split(rel, string(filepath.Separator))) > depth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() && blockedNames[strings.ToLower(d.Name())] {
			return filepath.SkipDir
		}
		if rel == "." {
			return nil
		}
		info, _ := d.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		kind := "file"
		if d.IsDir() {
			kind = "dir"
		}
		workspaceRel, relErr := filepath.Rel(w.root, p)
		if relErr == nil && len(entries) < 300 {
			entries = append(entries, FileEntry{Path: filepath.ToSlash(workspaceRel), Type: kind, Size: size})
		}
		return nil
	})
	return entries, err
}

func (w *Workspace) Read(path string, start, end int) (FileContent, error) {
	target, rel, err := w.Resolve(path)
	if err != nil {
		return FileContent{}, err
	}
	info, err := os.Stat(target)
	if err != nil || info.IsDir() || info.Size() > maxFileBytes {
		return FileContent{}, errors.New("file cannot be read")
	}
	b, err := os.ReadFile(target)
	if err != nil || !utf8.Valid(b) {
		return FileContent{}, errors.New("file is not readable UTF-8 text")
	}
	lines := strings.Split(string(b), "\n")
	if start < 1 {
		start = 1
	}
	if end < start || end > len(lines) {
		end = len(lines)
	}
	var out strings.Builder
	for i := start - 1; i < end; i++ {
		fmt.Fprintf(&out, "%d: %s\n", i+1, lines[i])
	}
	return FileContent{Path: rel, Content: out.String(), StartLine: start, EndLine: end}, nil
}

func (w *Workspace) Search(query, path string) ([]SearchMatch, error) {
	if query = strings.TrimSpace(query); query == "" {
		return nil, errors.New("query is required")
	}
	target, _, err := w.Resolve(path)
	if err != nil {
		return nil, err
	}
	result := make([]SearchMatch, 0)
	err = filepath.WalkDir(target, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || len(result) >= 100 {
			return nil
		}
		if d.IsDir() {
			if blockedNames[strings.ToLower(d.Name())] {
				return filepath.SkipDir
			}
			return nil
		}
		info, e := d.Info()
		if e != nil || info.Size() > maxFileBytes {
			return nil
		}
		f, e := os.Open(p)
		if e != nil {
			return nil
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 4096), maxFileBytes)
		line := 0
		for scanner.Scan() && len(result) < 100 {
			line++
			text := scanner.Text()
			if strings.Contains(strings.ToLower(text), strings.ToLower(query)) {
				rel, _ := filepath.Rel(w.root, p)
				result = append(result, SearchMatch{Path: filepath.ToSlash(rel), Line: line, Text: strings.TrimSpace(text)})
			}
		}
		return nil
	})
	return result, err
}
