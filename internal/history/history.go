package history

import (
	"bytes"
	"sync"
	"time"
)

type ClipType int

const (
	Text ClipType = iota
	Image
)

type ClipItem struct {
	Type        ClipType
	TextContent string
	ImageData   []byte
	Timestamp   time.Time
}

type Manager struct {
	mu      sync.RWMutex
	items   []ClipItem
	maxSize int
}

func NewManager(maxSize int) *Manager {
	if maxSize <= 0 {
		maxSize = 50 // Default
	}
	return &Manager{
		items:   make([]ClipItem, 0, maxSize),
		maxSize: maxSize,
	}
}

func (m *Manager) Add(item ClipItem) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicates (last item)
	if len(m.items) > 0 {
		last := m.items[len(m.items)-1]
		if last.Type == item.Type {
			if item.Type == Text && last.TextContent == item.TextContent {
				return false
			}
			if item.Type == Image && bytes.Equal(last.ImageData, item.ImageData) {
				return false
			}
		}
	}

	m.items = append(m.items, item)

	// Trim if exceeds maxSize
	// Since we append to the end, we remove from the beginning if needed
	// But wait, user wants Windows style (Win+V). Usually newest is at top visually.
	// My storage is append-only (chronological).
	// Trimming: Keep last N.

	if len(m.items) > m.maxSize {
		// remove oldest (index 0)
		overflow := len(m.items) - m.maxSize
		m.items = m.items[overflow:]
	}

	return true
}

func (m *Manager) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}

func (m *Manager) Get(index int) (ClipItem, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if index < 0 || index >= len(m.items) {
		return ClipItem{}, false
	}
	return m.items[index], true
}

func (m *Manager) GetAll() []ClipItem {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return copy
	dst := make([]ClipItem, len(m.items))
	copy(dst, m.items)
	return dst
}
