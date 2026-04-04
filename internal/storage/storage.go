package storage

import (
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
	FilePath string
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
		FilePath: filepath.Join(dir, "history.json"),
		dir:      dir,
	}, nil
}

func (s *FileStorage) Load() ([]ClipItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.FilePath)
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

func (s *FileStorage) Save(items []ClipItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.FilePath, data, 0644)
}

func (s *FileStorage) Append(item ClipItem) error {
	items, err := s.Load()
	if err != nil {
		return err
	}

	// Check duplicates
	if len(items) > 0 {
		last := items[len(items)-1]
		if last.Type == item.Type {
			if item.Type == Text && last.TextContent == item.TextContent {
				return nil
			}
		}
	}

	items = append(items, item)

	// Limit size (only remove unpinned items)
	if len(items) > 50 {
		items = trimToLimit(items, 50)
	}

	return s.Save(items)
}

// AppendWithLimit appends and trims to the given max size.
func (s *FileStorage) AppendWithLimit(item ClipItem, maxSize int) error {
	items, err := s.Load()
	if err != nil {
		return err
	}

	if len(items) > 0 {
		last := items[len(items)-1]
		if last.Type == item.Type {
			if item.Type == Text && last.TextContent == item.TextContent {
				return nil
			}
		}
	}

	items = append(items, item)

	if len(items) > maxSize {
		items = trimToLimit(items, maxSize)
	}

	return s.Save(items)
}

// trimToLimit removes oldest unpinned items to fit within limit.
func trimToLimit(items []ClipItem, limit int) []ClipItem {
	for len(items) > limit {
		removed := false
		for i := 0; i < len(items); i++ {
			if !items[i].Pinned {
				items = append(items[:i], items[i+1:]...)
				removed = true
				break
			}
		}
		if !removed {
			break // All items are pinned
		}
	}
	return items
}

// Delete removes an item at the given index.
func (s *FileStorage) Delete(index int) error {
	items, err := s.Load()
	if err != nil {
		return err
	}
	if index < 0 || index >= len(items) {
		return nil
	}
	items = append(items[:index], items[index+1:]...)
	return s.Save(items)
}

// TogglePin toggles the pinned state of an item at the given index.
func (s *FileStorage) TogglePin(index int) error {
	items, err := s.Load()
	if err != nil {
		return err
	}
	if index < 0 || index >= len(items) {
		return nil
	}
	items[index].Pinned = !items[index].Pinned
	return s.Save(items)
}

// Clear removes all non-pinned items.
func (s *FileStorage) Clear() error {
	items, err := s.Load()
	if err != nil {
		return err
	}
	var pinned []ClipItem
	for _, item := range items {
		if item.Pinned {
			pinned = append(pinned, item)
		}
	}
	return s.Save(pinned)
}

// ClearAll removes all items including pinned.
func (s *FileStorage) ClearAll() error {
	return s.Save([]ClipItem{})
}

// PurgeExpired removes items past their ExpiresAt time.
func (s *FileStorage) PurgeExpired() error {
	items, err := s.Load()
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
	return s.Save(kept)
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

func (s *FileStorage) SaveSnippets(snippets []Snippet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(snippets, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.snippetPath(), data, 0644)
}

func (s *FileStorage) AddSnippet(name, content string) error {
	snippets, err := s.LoadSnippets()
	if err != nil {
		return err
	}
	// Update if exists
	for i, sn := range snippets {
		if sn.Name == name {
			snippets[i].Content = content
			return s.SaveSnippets(snippets)
		}
	}
	snippets = append(snippets, Snippet{Name: name, Content: content})
	return s.SaveSnippets(snippets)
}

func (s *FileStorage) RemoveSnippet(name string) error {
	snippets, err := s.LoadSnippets()
	if err != nil {
		return err
	}
	for i, sn := range snippets {
		if sn.Name == name {
			snippets = append(snippets[:i], snippets[i+1:]...)
			return s.SaveSnippets(snippets)
		}
	}
	return nil
}
