package history

import (
	"testing"
	"time"
)

func TestManager_Add(t *testing.T) {
	m := NewManager(5)

	// Test adding items
	item1 := ClipItem{Type: Text, TextContent: "test1", Timestamp: time.Now()}
	if !m.Add(item1) {
		t.Error("Failed to add item1")
	}

	if m.Len() != 1 {
		t.Errorf("Expected length 1, got %d", m.Len())
	}

	// Test duplicate text
	if m.Add(item1) {
		t.Error("Should not add duplicate item1")
	}
	if m.Len() != 1 {
		t.Errorf("Expected length 1 after duplicate add, got %d", m.Len())
	}

	// Test adding more items
	item2 := ClipItem{Type: Text, TextContent: "test2", Timestamp: time.Now()}
	m.Add(item2)
	if m.Len() != 2 {
		t.Errorf("Expected length 2, got %d", m.Len())
	}

	// Test max size
	for i := 0; i < 10; i++ {
		m.Add(ClipItem{Type: Text, TextContent: string(rune(i)), Timestamp: time.Now()})
	}

	if m.Len() != 5 {
		t.Errorf("Expected length to be capped at 5, got %d", m.Len())
	}
}

func TestManager_ImageDuplicate(t *testing.T) {
	m := NewManager(5)

	imgData := []byte{0x00, 0x01, 0x02}
	item1 := ClipItem{Type: Image, ImageData: imgData, Timestamp: time.Now()}

	m.Add(item1)

	// Try adding same image
	item2 := ClipItem{Type: Image, ImageData: []byte{0x00, 0x01, 0x02}, Timestamp: time.Now()}
	if m.Add(item2) {
		t.Error("Should not add duplicate image")
	}

	// Different image
	item3 := ClipItem{Type: Image, ImageData: []byte{0x00, 0x01, 0x03}, Timestamp: time.Now()}
	if !m.Add(item3) {
		t.Error("Should add different image")
	}
}

func TestManager_Get(t *testing.T) {
	m := NewManager(5)
	m.Add(ClipItem{Type: Text, TextContent: "1"})
	m.Add(ClipItem{Type: Text, TextContent: "2"})

	// Get valid
	item, ok := m.Get(0)
	if !ok || item.TextContent != "1" {
		t.Error("Failed to get item 0")
	}

	item, ok = m.Get(1)
	if !ok || item.TextContent != "2" {
		t.Error("Failed to get item 1")
	}

	// Get invalid
	_, ok = m.Get(2)
	if ok {
		t.Error("Should not get item 2")
	}

	_, ok = m.Get(-1)
	if ok {
		t.Error("Should not get item -1")
	}
}
