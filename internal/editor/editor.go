package editor

import (
	"errors"
	"path/filepath"
	"sync"

	"github.com/lg2m/athena/internal/editor/buffer"
	"github.com/lg2m/athena/internal/editor/state"
)

var (
	ErrNoBuffer         = errors.New("no current buffer")
	ErrBufferNotFound   = errors.New("buffer not found")
	ErrInvalidOperation = errors.New("invalid operation for current mode")
	ErrUnsavedChanges   = errors.New("unsaved changes exist")
)

// Editor represents the main editor application.
type Editor struct {
	Manager       *buffer.BufferManager
	mode          state.Mode
	DesiredColumn int // track movement
	mu            sync.RWMutex
}

// NewEditor initializes a new Editor instance.
func NewEditor() *Editor {
	return &Editor{
		Manager: buffer.NewBufferManager(),
		mode:    state.Normal,
	}
}

// OpenFile opens a file and adds it to the buffer manager.
func (e *Editor) OpenFile(filePath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	// Check if buffer already exists
	if _, exists := e.Manager.GetBuffer(absPath); exists {
		return e.Manager.SetCurrentBuffer(absPath)
	}

	// Create new buffer
	buf, err := buffer.NewBuffer(absPath)
	if err != nil {
		return err
	}

	e.Manager.AddBuffer(buf)
	return nil
}

// GetMode returns the current mode state.
func (e *Editor) GetMode() state.Mode {
	return e.mode
}

// SetMode sets the current editor mode state.
func (e *Editor) SetMode(mode state.Mode) {
	e.mode = mode
}

// InsertText inserts text at the cursor position in the current buffer.
func (e *Editor) InsertText(text string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	buf := e.Manager.GetCurrentBuffer()
	if buf == nil {
		return ErrNoBuffer
	}

	if e.mode != state.Insert {
		return ErrInvalidOperation
	}

	return buf.Insert(text)
}

// DeleteText deletes text of specified length from the cursor position.
func (e *Editor) DeleteText(length int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	buf := e.Manager.GetCurrentBuffer()
	if buf == nil {
		return ErrNoBuffer
	}

	pos := buf.GetCursor()
	if length < 0 {
		// Handle backward delete
		pos += length
		length = -length
	}

	return buf.Delete(pos, pos+length)
}

// GetCurrentPosition retrieves the current line and column of the cursor.
func (e *Editor) GetCurrentPosition() (int, int, error) {
	b := e.Manager.GetCurrentBuffer()
	if b == nil {
		return 0, 0, ErrNoBuffer
	}
	pos := b.GetCursor()
	line, col, err := b.PositionToLineCol(pos)
	if err != nil {
		return 0, 0, err
	}
	// e.DesiredColumn = col
	return line, col, nil
}

// MoveCursorHorizontal moves the cursor in the current buffer.
func (e *Editor) MoveCursorHorizontal(offset int) error {
	buf := e.Manager.GetCurrentBuffer()
	if buf == nil {
		return ErrNoBuffer
	}
	if err := buf.MoveCursor(offset); err != nil {
		return err
	}
	_, col, err := buf.PositionToLineCol(buf.GetCursor())
	if err != nil {
		return err
	}
	e.DesiredColumn = col
	return nil
}

// MoveVertical moves the cursor vertically while maintaining the desired column position.
func (e *Editor) MoveCursorVertical(offset int) error {
	buf := e.Manager.GetCurrentBuffer()
	if buf == nil {
		return ErrNoBuffer
	}
	currLine, currCol, err := buf.PositionToLineCol(buf.GetCursor())
	if err != nil {
		return err
	}
	if e.DesiredColumn == 0 {
		e.DesiredColumn = currCol
	}
	targetLine := currLine + offset
	if targetLine < 0 {
		targetLine = 0
	}
	totalLines, err := buf.LineCount()
	if err != nil {
		return err
	}
	if targetLine >= totalLines {
		targetLine = totalLines - 1
	}
	_, _, err = buf.MoveCursorToLineCol(targetLine, e.DesiredColumn)
	if err != nil {
		return err
	}
	return nil
}

// MoveCursorToLineCol moves cursor to specific line and column.
func (e *Editor) MoveCursorToLineCol(line int) error {
	buf := e.Manager.GetCurrentBuffer()
	if buf == nil {
		return ErrNoBuffer
	}
	desiredLine := line
	desiredCol := e.DesiredColumn
	_, _, err := buf.MoveCursorToLineCol(desiredLine, desiredCol)
	if err != nil {
		return err
	}

	// If the cursor was clamped, desiredColumn remains unchanged
	// The cursor will attempt to move to desiredColumn on subsequent movements
	// e.DesiredColumn = desiredCol
	return nil
}

// SaveCurrentBuffer saves the current buffer.
func (e *Editor) SaveCurrentBuffer() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	buf := e.Manager.GetCurrentBuffer()
	if buf == nil {
		return ErrNoBuffer
	}
	return buf.Save()
}

// CloseCurrentBuffer closes the current buffer.
func (e *Editor) CloseCurrentBuffer() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	buf := e.Manager.GetCurrentBuffer()
	if buf == nil {
		return ErrNoBuffer
	}

	// TODO: check for unsaved changes

	return buf.Close()
}

// GetLineCount returns the total number of lines in the buffer.
func (e *Editor) GetLineCount() (int, error) {
	b := e.Manager.GetCurrentBuffer()
	if b == nil {
		return 0, ErrNoBuffer
	}
	return b.LineCount()
}
