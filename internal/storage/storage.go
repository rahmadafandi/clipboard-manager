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
	Type        ClipType  `json:"type"`
	TextContent string    `json:"text_content,omitempty"`
	ImageData   []byte    `json:"image_data,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

type FileStorage struct {
	mu       sync.Mutex
	FilePath string
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
	// Not efficient for large history, but fine for <100 items. Need to lock properly.
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
			// Image comparison logic (omitted for brevity in append, relying on watcher usually)
		}
	}

	items = append(items, item)

	// Limit size
	if len(items) > 50 {
		items = items[len(items)-50:]
	}

	return s.Save(items)
}
