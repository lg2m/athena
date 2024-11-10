package buffer

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lg2m/athena/internal/rope"
	"github.com/lg2m/athena/internal/util"
)

var (
	ErrInvalidRange     = errors.New("invalid range")
	ErrInvalidPosition  = errors.New("invalid position")
	ErrInvalidLineCol   = errors.New("invalid line/column")
	ErrInvalidSelection = errors.New("invalid selection")
)

type Selection struct {
	Start int
	End   int
}

// Buffer represents a single text buffer (document).
type Buffer struct {
	document      *rope.Rope
	selection     Selection
	name          string
	filePath      string
	lastSavePoint time.Time
	file          *os.File
	size          int64
	lineCache     []int
	mu            sync.RWMutex
}

// NewBuffer creates a new Buffer with optional initial content.
func NewBuffer(filePath string) (*Buffer, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	document, err := io.ReadAll(file)
	if err != nil {
		file.Close()
		return nil, err
	}

	fp, err := filepath.Abs(filePath)
	if err != nil {
		file.Close()
		return nil, err
	}

	b := &Buffer{
		document:      rope.NewRope(string(document)),
		selection:     Selection{Start: 0, End: 0},
		name:          util.GetFileName(filePath, true),
		filePath:      fp,
		lastSavePoint: time.Now(),
		file:          file,
		size:          int64(len(document)),
	}

	b.updateLineCache()

	return b, nil
}

// Insert inserts text at the cursor's current position.
func (b *Buffer) Insert(s string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// replace selection with new text
	if b.selection.Start != b.selection.End {
		if err := b.document.Delete(b.selection.Start, b.selection.End); err != nil {
			return err
		}
	}

	// insert new text at selection start
	if err := b.document.Insert(b.selection.Start, s); err != nil {
		return err
	}

	// update selection to new position
	newEnd := b.selection.Start + rope.CountGraphemes(s)
	b.selection = Selection{Start: newEnd, End: newEnd}

	// if err := b.document.Insert(b.cursor.Position, s); err != nil {
	// 	return err
	// }

	b.size += int64(len(s))
	// b.cursor.Position += rope.CountGraphemes(s)
	b.updateLineCache()

	return nil
}

// Delete deletes text from the cursor position to position + length.
func (b *Buffer) Delete(start, end int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.document.Delete(start, end); err != nil {
		return err
	}

	if b.selection.Start > start {
		b.selection = Selection{Start: start, End: start}
	}
	// if b.cursor.Position > start {
	// 	b.cursor.SetPosition(start)
	// }

	b.size -= int64(end - start)
	b.updateLineCache()

	return nil
}

// DeleteSelections deletes text in the current selections.
func (b *Buffer) DeleteSelection() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	start, end := b.selection.Start, b.selection.End

	if err := b.document.Delete(start, end); err != nil {
		return err
	}

	b.selection = Selection{Start: start, End: start}
	b.size -= int64(end - start)
	b.updateLineCache()

	return nil
}

// GetText returns text between start and end positions.
func (b *Buffer) GetText(start, end int) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.document.GetTextRange(start, end)
}

// GetSelectedText returns the text within the current selections.
func (b *Buffer) GetSelectedText() (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.document.GetTextRange(b.selection.Start, b.selection.End)
}

// Save writes buffer content to disk.
func (b *Buffer) Save() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.file.Truncate(0); err != nil {
		return err
	}

	if _, err := b.file.Seek(0, 0); err != nil {
		return err
	}

	_, err := b.file.WriteString(b.document.ToString())
	if err != nil {
		return err
	}

	b.lastSavePoint = time.Now()
	return nil
}

// Close properly closes the buffer and its resources
func (b *Buffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Save any remaining dirty chunks
	if err := b.Save(); err != nil {
		return err
	}

	if err := b.file.Close(); err != nil {
		return err
	}

	return nil
}

// CollapseSelectionsToCursor collapses all selections to their end positions.
func (b *Buffer) CollapseSelectionsToCursor() {
	b.mu.Lock()
	defer b.mu.Unlock()

	pos := b.selection.End
	b.selection = Selection{Start: pos, End: pos}
}

// MoveSelections moves the selections by the specified offset.
// If `extend` is true, it extends the selection; otherwise, it moves the cursor.
func (b *Buffer) MoveSelections(offset int, extend bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	newPos := b.selection.End + offset
	newPos = clamp(newPos, 0, b.document.TotalGraphemes())
	if extend {
		// extend the selection end
		b.selection.End = newPos
	} else {
		// move both start and end (cursor movement)
		b.selection = Selection{Start: newPos, End: newPos}
	}

	return nil
}

// MoveSelectionToLineCol moves the selection to a specific line and column.
func (b *Buffer) MoveSelectionToLineCol(line, col int, extend bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if line < 0 || col < 0 || line >= len(b.lineCache) {
		return ErrInvalidLineCol
	}

	lineStart := b.lineCache[line]
	var lineEnd int
	if line+1 < len(b.lineCache) {
		lineEnd = b.lineCache[line+1] - 1 // -1 to exclude newline
	} else {
		lineEnd = b.document.TotalGraphemes()
	}

	actualCol := col
	lineLen := lineEnd - lineStart
	if actualCol > lineLen {
		actualCol = lineLen
	}

	targetPos := lineStart + actualCol

	if extend {
		b.selection.End = targetPos
	} else {
		b.selection = Selection{Start: targetPos, End: targetPos}
	}

	return nil
}

// Selections returns the current selections.
func (b *Buffer) Selection() Selection {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.selection
}

// TotalGraphemes returns the total number of graphemes in the document.
func (b *Buffer) TotalGraphemes() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.document.TotalGraphemes()
}

// PositionToLineCol converts a buffer position to line and column numbers
func (b *Buffer) PositionToLineCol(pos int) (int, int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if pos < 0 || pos > b.document.TotalGraphemes() {
		return 0, 0, ErrInvalidPosition
	}

	// Binary search in lineCache to find the line
	line := 0
	for i := len(b.lineCache) - 1; i >= 0; i-- {
		if pos >= b.lineCache[i] {
			line = i
			break
		}
	}

	column := pos - b.lineCache[line]
	return line, column, nil
}

// GetLine returns the content of a specific line
func (b *Buffer) GetLine(lineNum int) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if lineNum < 0 || lineNum >= len(b.lineCache) {
		return "", ErrInvalidLineCol
	}

	start := b.lineCache[lineNum]
	var end int
	if lineNum+1 < len(b.lineCache) {
		end = b.lineCache[lineNum+1] - 1 // -1 to exclude newline
	} else {
		end = b.document.TotalGraphemes()
	}

	return b.document.GetTextRange(start, end)
}

// LineCount returns the total number of lines in the buffer
func (b *Buffer) LineCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.lineCache)
}

// FileName returns the name of the file related to the buffer.
func (b *Buffer) FileName() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.name
}

// FilePath returns the path of the file related to the buffer.
func (b *Buffer) FilePath() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.filePath
}

// updateLineCache rebuilds the cache of line start positions.
func (b *Buffer) updateLineCache() {
	b.lineCache = make([]int, 0, 1000)
	b.lineCache = append(b.lineCache, 0) // First line always starts at 0

	doc := b.document.ToString()
	for i, r := range doc {
		if r == '\n' {
			b.lineCache = append(b.lineCache, i+1)
		}
	}
}

// clamp clamps a value within a range.
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
