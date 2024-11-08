package buffer

import (
	"errors"
	"os"
	"strings"

	"github.com/lg2m/athena/internal/rope"
)

// Buffer represents a single text buffer (document).
type Buffer struct {
	rope   *rope.Rope
	cursor *Cursor
	name   string
}

// NewBuffer creates a new Buffer with optional initial content.
func NewBuffer(name, initial string) *Buffer {
	return &Buffer{
		rope:   rope.NewRope(initial),
		cursor: &Cursor{Position: 0},
		name:   name,
	}
}

// NewBufferFromFile creates a new Buffer by loading content from a file.
func NewBufferFromFile(filePath string) (*Buffer, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return NewBuffer(filePath, string(data)), nil
}

// Insert inserts text at the cursor's current position.
func (b *Buffer) Insert(s string) error {
	if err := b.rope.Insert(b.cursor.Position, s); err != nil {
		return err
	}
	b.cursor.Position += rope.CountGraphemes(s)
	return nil
}

// Delete deletes text from the cursor position to position + length.
func (b *Buffer) Delete(length int) error {
	if length < 0 {
		return errors.New("length must be non-negative")
	}
	return b.rope.Delete(b.cursor.Position, b.cursor.Position+length)
}

// MoveCursor moves the cursor by the specified offset.
func (b *Buffer) MoveCursor(offset int) {
	newPos := b.cursor.Position + offset
	totalGraphemes := b.rope.TotalGraphemes()
	if newPos < 0 {
		newPos = 0
	}
	if newPos > totalGraphemes {
		newPos = totalGraphemes
	}
	b.cursor.SetPosition(newPos)
}

// SetCursor sets the cursor to the specified position.
func (b *Buffer) SetCursor(pos int) {
	totalGraphemes := b.rope.TotalGraphemes()
	if pos < 0 {
		pos = 0
	}
	if pos > totalGraphemes {
		pos = totalGraphemes
	}
	b.cursor.SetPosition(pos)
}

// GetText returns the entire content of the buffer.
func (b *Buffer) GetText() string {
	return b.rope.ToString()
}

// GetTextRange retrieves text from start to end graphemes indices (exclusive).
func (b *Buffer) GetTextRange(start, end int) (string, error) {
	return b.rope.GetTextRange(start, end)
}

// GetCursorPosition returns the cursor's line and column numbers.
func (b *Buffer) GetCursorPosition() (line, column int) {
	textBeforeCursor, _ := b.GetTextRange(0, b.cursor.Position)
	lines := strings.Split(textBeforeCursor, "\n")
	line = len(lines) - 1
	column = len([]rune(lines[line]))
	return
}

// GetCursorIndex returns the cursor's position in the graphemes cluster.
func (b *Buffer) GetCursorIndex() int {
	return b.cursor.Position
}

// GetLines returns the buffer content split into lines.
func (b *Buffer) GetLines() []string {
	text := b.GetText()
	return strings.Split(text, "\n")
}

// TotalGraphemes returns the total number of grapheme clusters in the buffer.
func (b *Buffer) TotalGraphemes() int {
	return b.rope.TotalGraphemes()
}

// Save writes the buffer content to the associated file.
func (b *Buffer) Save() error {
	if b.name == "" {
		return errors.New("no file associated with this buffer")
	}
	content := b.GetText()
	return os.WriteFile(b.name, []byte(content), 0644)
}
