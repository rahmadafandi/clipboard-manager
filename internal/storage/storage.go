package storage

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ClipType int

const (
	Text ClipType = iota
	Image
)

type ClipItem struct {
	Type        ClipType   `json:"type"`
	TextContent string     `json:"text_content,omitempty"`
	ImageData   []byte     `json:"image_data,omitempty"`
	Timestamp   time.Time  `json:"timestamp"`
	Pinned      bool       `json:"pinned,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type Snippet struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type FileStorage struct {
	mu       sync.Mutex
	filePath string
	dir      string
}

func NewFileStorage() (*FileStorage, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(cacheDir, "clipboard-manager")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &FileStorage{
		filePath: filepath.Join(dir, "history.json"),
		dir:      dir,
	}, nil
}

func (s *FileStorage) Path() string {
	return s.filePath
}

// load reads items without locking (caller must hold mu).
func (s *FileStorage) load() ([]ClipItem, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ClipItem{}, nil
		}
		return nil, err
	}
	var items []ClipItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// save writes items without locking (caller must hold mu).
func (s *FileStorage) save(items []ClipItem) error {
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *FileStorage) Load() ([]ClipItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load()
}

func (s *FileStorage) Save(items []ClipItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.save(items)
}

func (s *FileStorage) AppendWithLimit(item ClipItem, maxSize int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.load()
	if err != nil {
		return err
	}

	if len(items) > 0 {
		last := items[len(items)-1]
		if last.Type == item.Type {
			if item.Type == Text && last.TextContent == item.TextContent {
				return nil
			}
			if item.Type == Image && bytes.Equal(last.ImageData, item.ImageData) {
				return nil
			}
		}
	}

	items = append(items, item)

	if len(items) > maxSize {
		items = trimToLimit(items, maxSize)
	}

	return s.save(items)
}

// trimToLimit removes oldest unpinned items to fit within limit.
func trimToLimit(items []ClipItem, limit int) []ClipItem {
	var pinned, unpinned []ClipItem
	for _, item := range items {
		if item.Pinned {
			pinned = append(pinned, item)
		} else {
			unpinned = append(unpinned, item)
		}
	}
	keep := limit - len(pinned)
	if keep < 0 {
		keep = 0
	}
	if len(unpinned) > keep {
		unpinned = unpinned[len(unpinned)-keep:]
	}
	result := make([]ClipItem, 0, len(pinned)+len(unpinned))
	result = append(result, pinned...)
	result = append(result, unpinned...)
	return result
}

func (s *FileStorage) Delete(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.load()
	if err != nil {
		return err
	}
	if index < 0 || index >= len(items) {
		return nil
	}
	items = append(items[:index], items[index+1:]...)
	return s.save(items)
}

func (s *FileStorage) TogglePin(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.load()
	if err != nil {
		return err
	}
	if index < 0 || index >= len(items) {
		return nil
	}
	items[index].Pinned = !items[index].Pinned
	return s.save(items)
}

func (s *FileStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.load()
	if err != nil {
		return err
	}
	var pinned []ClipItem
	for _, item := range items {
		if item.Pinned {
			pinned = append(pinned, item)
		}
	}
	return s.save(pinned)
}

func (s *FileStorage) ClearAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.save([]ClipItem{})
}

func (s *FileStorage) PurgeExpired() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.load()
	if err != nil {
		return err
	}
	now := time.Now()
	var kept []ClipItem
	for _, item := range items {
		if item.ExpiresAt != nil && now.After(*item.ExpiresAt) {
			continue
		}
		kept = append(kept, item)
	}
	if len(kept) == len(items) {
		return nil
	}
	return s.save(kept)
}

// --- Snippet Storage ---

func (s *FileStorage) snippetPath() string {
	return filepath.Join(s.dir, "snippets.json")
}

func (s *FileStorage) LoadSnippets() ([]Snippet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.snippetPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []Snippet{}, nil
		}
		return nil, err
	}
	var snippets []Snippet
	if err := json.Unmarshal(data, &snippets); err != nil {
		return nil, err
	}
	return snippets, nil
}

func (s *FileStorage) saveSnippets(snippets []Snippet) error {
	data, err := json.Marshal(snippets)
	if err != nil {
		return err
	}
	return os.WriteFile(s.snippetPath(), data, 0644)
}

func (s *FileStorage) AddSnippet(name, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.snippetPath())
	var snippets []Snippet
	if err == nil {
		json.Unmarshal(data, &snippets)
	}

	for i, sn := range snippets {
		if sn.Name == name {
			snippets[i].Content = content
			return s.saveSnippets(snippets)
		}
	}
	snippets = append(snippets, Snippet{Name: name, Content: content})
	return s.saveSnippets(snippets)
}

func (s *FileStorage) RemoveSnippet(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.snippetPath())
	if err != nil {
		return nil
	}
	var snippets []Snippet
	if err := json.Unmarshal(data, &snippets); err != nil {
		return nil
	}

	for i, sn := range snippets {
		if sn.Name == name {
			snippets = append(snippets[:i], snippets[i+1:]...)
			return s.saveSnippets(snippets)
		}
	}
	return nil
}
