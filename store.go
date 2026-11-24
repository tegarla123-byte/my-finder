package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Shortcut struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type storePayload struct {
	Shortcuts []Shortcut `json:"shortcuts"`
}

type Store struct {
	filePath string
	Items    []Shortcut
}

func defaultStorePath() string {
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		tempDir := strings.ToLower(filepath.Clean(os.TempDir()))
		cleanExeDir := strings.ToLower(filepath.Clean(exeDir))
		// Use portable file next to .exe unless running from go-build temp dir
		if !strings.HasPrefix(cleanExeDir, tempDir) {
			candidate := filepath.Join(exeDir, "shortcuts.json")
			if isDirWritable(exeDir) {
				return candidate
			}
		}
	}
	cfgDir, err := os.UserConfigDir()
	if err == nil && cfgDir != "" {
		appDir := filepath.Join(cfgDir, "MyFinder")
		_ = os.MkdirAll(appDir, 0o755)
		return filepath.Join(appDir, "shortcuts.json")
	}
	return "shortcuts.json"
}

func isDirWritable(dir string) bool {
	f, err := os.CreateTemp(dir, ".myfinder-write-check-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}

func newStore(path string) *Store {
	if path == "" {
		path = defaultStorePath()
	}
	return &Store{
		filePath: path,
		Items:    []Shortcut{},
	}
}

func (s *Store) FilePath() string {
	return s.filePath
}

func (s *Store) Load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.Items = []Shortcut{}
			return nil
		}
		return err
	}
	var payload storePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	s.Items = make([]Shortcut, 0, len(payload.Shortcuts))
	for _, item := range payload.Shortcuts {
		item.Path = normalizePath(item.Path)
		item.Name = strings.TrimSpace(item.Name)
		if item.Path == "" {
			continue
		}
		if item.Name == "" {
			item.Name = defaultNameForPath(item.Path)
		}
		if item.ID == "" {
			item.ID = generateID()
		}
		s.Items = append(s.Items, item)
	}
	return nil
}

func (s *Store) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.filePath), 0o755); err != nil {
		return err
	}
	payload := storePayload{Shortcuts: s.Items}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := s.filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.filePath); err != nil {
		// Fallback on Windows if rename over existing fails due to antivirus lock
		if writeErr := os.WriteFile(s.filePath, data, 0o644); writeErr != nil {
			_ = os.Remove(tmp)
			return writeErr
		}
		_ = os.Remove(tmp)
	}
	return nil
}

func (s *Store) Add(name, rawPath string) (Shortcut, error) {
	cleanPath := normalizePath(rawPath)
	if cleanPath == "" {
		return Shortcut{}, errors.New("path tidak boleh kosong")
	}
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		cleanName = defaultNameForPath(cleanPath)
	}
	item := Shortcut{
		ID:   generateID(),
		Name: cleanName,
		Path: cleanPath,
	}
	s.Items = append(s.Items, item)
	if err := s.Save(); err != nil {
		return item, err
	}
	return item, nil
}

func (s *Store) Update(id, name, rawPath string) error {
	cleanPath := normalizePath(rawPath)
	if cleanPath == "" {
		return errors.New("path tidak boleh kosong")
	}
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		cleanName = defaultNameForPath(cleanPath)
	}
	for i := range s.Items {
		if s.Items[i].ID == id {
			s.Items[i].Name = cleanName
			s.Items[i].Path = cleanPath
			return s.Save()
		}
	}
	return errors.New("shortcut tidak ditemukan")
}

func (s *Store) Delete(id string) error {
	idx := s.indexByID(id)
	if idx < 0 {
		return nil
	}
	s.Items = append(s.Items[:idx], s.Items[idx+1:]...)
	return s.Save()
}

func (s *Store) Move(id string, delta int) bool {
	idx := s.indexByID(id)
	if idx < 0 {
		return false
	}
	target := idx + delta
	if target < 0 || target >= len(s.Items) {
		return false
	}
	s.Items[idx], s.Items[target] = s.Items[target], s.Items[idx]
	_ = s.Save()
	return true
}

func (s *Store) indexByID(id string) int {
	for i := range s.Items {
		if s.Items[i].ID == id {
			return i
		}
	}
	return -1
}

// Filter returns indices into s.Items matching all whitespace-separated query tokens.
func (s *Store) Filter(query string) []int {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		res := make([]int, len(s.Items))
		for i := range s.Items {
			res[i] = i
		}
		return res
	}
	tokens := strings.Fields(q)
	res := make([]int, 0, len(s.Items))
	for i, item := range s.Items {
		haystack := strings.ToLower(item.Name + " " + item.Path)
		matched := true
		for _, tok := range tokens {
			if !strings.Contains(haystack, tok) {
				matched = false
				break
			}
		}
		if matched {
			res = append(res, i)
		}
	}
	return res
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.Trim(p, `"`)
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	p = strings.ReplaceAll(p, "/", `\`)
	// Drive letter like "D:" -> "D:\"
	if len(p) == 2 && p[1] == ':' && isAlpha(p[0]) {
		return strings.ToUpper(p[:1]) + `:\`
	}
	// Preserve UNC prefix
	if strings.HasPrefix(p, `\\`) {
		cleaned := filepath.Clean(p)
		return cleaned
	}
	cleaned := filepath.Clean(p)
	if len(cleaned) == 2 && cleaned[1] == ':' && isAlpha(cleaned[0]) {
		return strings.ToUpper(cleaned[:1]) + `:\`
	}
	if len(cleaned) >= 3 && cleaned[1] == ':' && cleaned[2] == '\\' && isAlpha(cleaned[0]) {
		cleaned = strings.ToUpper(cleaned[:1]) + cleaned[1:]
	}
	return cleaned
}

func defaultNameForPath(p string) string {
	p = normalizePath(p)
	if len(p) == 3 && p[1] == ':' && p[2] == '\\' {
		return "Drive " + strings.ToUpper(p[:1])
	}
	base := filepath.Base(p)
	if base == "" || base == "." || base == `\` {
		return p
	}
	return base
}

func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func generateID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
