package buffer

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lg2m/athena/internal/editor/state"
	"github.com/lg2m/athena/internal/editor/treesitter"
	"github.com/lg2m/athena/internal/rope"
	"github.com/lg2m/athena/internal/util"
	"github.com/rivo/uniseg"
)

var (
	ErrInvalidRange     = errors.New("buffer: poition range exeeds document boundaries")
	ErrInvalidPosition  = errors.New("buffer: position exceeds document boundaries")
	ErrInvalidLineCol   = errors.New("buffer: line/column position out of bounds")
	ErrInvalidSelection = errors.New("buffer: selection boundaries are invalid")
)

// Buffer represents a text buffer with support for syntax highlighting and concurrent access.
type Buffer struct {
	document      *rope.Rope
	selection     state.Selection
	filePath      string
	lastSavePoint time.Time
	file          *os.File
	size          int64
	lineCache     []int
	highlighter   *treesitter.Highlighter
	dirty         bool

	FileUtil *util.FileUtil

	lineCacheMu sync.RWMutex
	mu          sync.RWMutex
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

	// Setup registry
	registry := treesitter.NewRegistry(treesitter.DefaultStyles)

	// Register langauges
	registry.RegisterLanguage(&treesitter.RustLanguage{})

	// Create highlighter
	highlighter, err := treesitter.NewHighlighter(registry, "rust")
	if err != nil {
		file.Close()
		return nil, err
	}

	b := &Buffer{
		document:      rope.NewRope(string(document)),
		selection:     state.Selection{Start: 0, End: 0},
		filePath:      fp,
		lastSavePoint: time.Now(),
		file:          file,
		size:          int64(len(document)),
		highlighter:   highlighter,
		FileUtil:      util.NewFileUtil(nil),
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
		b.size -= int64(b.selection.End - b.selection.Start)
	}

	// insert new text at selection start
	if err := b.document.Insert(b.selection.Start, s); err != nil {
		return err
	}

	// update selection to new position
	graphemeCount := countGraphemes(s)
	newEnd := b.selection.Start + graphemeCount
	b.selection = state.Selection{Start: newEnd, End: newEnd}

	b.size += int64(len(s))
	b.dirty = true
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
		b.selection = state.Selection{Start: start, End: start}
	}

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

	b.selection = state.Selection{Start: start, End: start}
	b.size -= int64(end - start)
	b.updateLineCache()
	return nil
}

// GetSelectedText returns the text within the current selections.
func (b *Buffer) GetSelectedText() (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.document.Substring(b.selection.Start, b.selection.End)
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

	_, err := b.file.WriteString(b.document.String())
	if err != nil {
		return err
	}

	b.lastSavePoint = time.Now()
	b.dirty = false
	return nil
}

// Close properly closes the buffer and its resources
func (b *Buffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Save remaining dirty content
	if b.dirty {
		if err := b.Save(); err != nil {
			return err
		}
	}
	return b.file.Close()
}

// CollapseSelectionsToCursor collapses all selections to their end positions.
func (b *Buffer) CollapseSelectionsToCursor() {
	b.mu.Lock()
	defer b.mu.Unlock()

	pos := b.selection.End
	b.selection = state.Selection{Start: pos, End: pos}
}

// Selections returns the current selections.
func (b *Buffer) Selection() state.Selection {
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

	b.lineCacheMu.RLock()
	defer b.lineCacheMu.RUnlock()

	// search lineCache to find the line
	left, right := 0, len(b.lineCache)-1
	var line int
	for left <= right {
		mid := (left + right) / 2
		if b.lineCache[mid] <= pos {
			line = mid
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	column := pos - b.lineCache[line]
	return line, column, nil
}

// GetLine returns the content of a specific line
func (b *Buffer) GetLine(lineNum int) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	b.lineCacheMu.RLock()
	defer b.lineCacheMu.RUnlock()

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

	return b.document.Substring(start, end)
}

func (b *Buffer) GetHighlights() ([]treesitter.Highlight, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.highlighter.GetHighlights([]byte(b.document.String()))
}

// LineCount returns the total number of lines in the buffer
func (b *Buffer) LineCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	b.lineCacheMu.RLock()
	defer b.lineCacheMu.RUnlock()

	return len(b.lineCache)
}

// FileName returns the name of the file related to the buffer.
func (b *Buffer) FileName() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.FileUtil.GetFileName(b.filePath, true)
}

// FileType returns the type of file in the buffer.
func (b *Buffer) FileType() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.FileUtil.GetFileExt(b.filePath)
}

// FilePath returns the path of the file related to the buffer.
func (b *Buffer) FilePath() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.filePath
}

// updateLineCache rebuilds the cache of line start positions.
func (b *Buffer) updateLineCache() {
	b.lineCacheMu.Lock()
	defer b.lineCacheMu.Unlock()

	b.lineCache = []int{0}
	iter := b.document.NewIterator()
	var pos int
	for grapheme, ok := iter.Next(); ok; grapheme, ok = iter.Next() {
		if grapheme == "\n" {
			b.lineCache = append(b.lineCache, pos+1)
		}
		pos++
	}
}

// countGraphemes counts the grapheme clusters in a string.
func countGraphemes(s string) int {
	gr := uniseg.NewGraphemes(s)
	count := 0
	for gr.Next() {
		count++
	}
	return count
}
