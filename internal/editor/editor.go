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
	ErrNoSelections     = errors.New("no selections")
	ErrBufferNotFound   = errors.New("buffer not found")
	ErrInvalidOperation = errors.New("invalid operation for current mode")
	ErrUnsavedChanges   = errors.New("unsaved changes exist")
)

// Editor represents the main editor application.
type Editor struct {
	buffers       map[string]*buffer.Buffer // keys by absolute file path
	current       *buffer.Buffer
	mode          state.Mode
	desiredColumn int // track movement
	mu            sync.RWMutex
}

// NewEditor initializes a new Editor instance.
func NewEditor() *Editor {
	return &Editor{
		buffers:       make(map[string]*buffer.Buffer),
		mode:          state.Normal,
		desiredColumn: -1,
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

	// check if buffer exists
	if b, exists := e.buffers[absPath]; exists {
		e.current = b
		return nil
	}

	// create new buffer
	b, err := buffer.NewBuffer(absPath)
	if err != nil {
		return err
	}

	e.buffers[absPath] = b
	e.current = b
	return nil
}

// FileName returns the file name related to the current active buffer.
func (e *Editor) FileName() (string, error) {
	if e.current == nil {
		return "", ErrNoBuffer
	}
	return e.current.FileName(), nil
}

// FilePath returns the path of the file related to the current active buffer.
func (e *Editor) FilePath() (string, error) {
	if e.current == nil {
		return "", ErrNoBuffer
	}
	return e.current.FilePath(), nil
}

// SwitchBuffer switches to a buffer by file path.
func (e *Editor) SwitchBuffer(filePath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	b, err := e.getBuffer(filePath)
	if err != nil {
		return err
	}

	e.current = b
	return nil
}

// GetBufferList returns a list of all open buffer file paths
func (e *Editor) GetBufferList() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	paths := make([]string, 0, len(e.buffers))
	for path := range e.buffers {
		paths = append(paths, path)
	}

	return paths
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

	if e.current == nil {
		return ErrNoBuffer
	}

	if e.mode != state.Insert {
		return ErrInvalidOperation
	}

	e.current.CollapseSelectionsToCursor()

	return e.current.Insert(text)
}

func (e *Editor) DeleteSelection() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return ErrNoBuffer
	}

	return e.current.DeleteSelection()
}

// DeleteText deletes text of specified length from the cursor position.
func (e *Editor) DeleteText(length int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return ErrNoBuffer
	}

	selection := e.current.Selection()
	pos := selection.End

	if length < 0 {
		// Handle backward delete
		pos += length
		length = -length
	}

	// pos := e.current.Cursor()
	// if length < 0 {
	// 	// Handle backward delete
	// 	pos += length
	// 	length = -length
	// }

	return e.current.Delete(pos, pos+length)
}

// GetCurrentPosition retrieves the current line and column of the cursor.
func (e *Editor) GetCurrentPosition() (int, int, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.current == nil {
		return 0, 0, ErrNoBuffer
	}

	selection := e.current.Selection()
	pos := selection.End
	return e.current.PositionToLineCol(pos)

	// pos := e.current.Cursor()
	// return e.current.PositionToLineCol(pos)
}

// MoveCursorHorizontal moves the cursor horizontally in the current buffer.
func (e *Editor) MoveCursorHorizontal(offset int, extend bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return ErrNoBuffer
	}

	if err := e.current.MoveSelections(offset, extend); err != nil {
		return err
	}

	// Update desiredColumn based on the selection's end position
	selection := e.current.Selection()

	pos := selection.End
	_, col, err := e.current.PositionToLineCol(pos)
	if err != nil {
		return err
	}

	e.desiredColumn = col
	return nil

	// if err := e.current.MoveCursor(offset); err != nil {
	// 	return err
	// }

	// _, col, err := e.current.PositionToLineCol(e.current.Cursor())
	// if err != nil {
	// 	return err
	// }

	// e.desiredColumn = col
	// return nil
}

// MoveCursorVertical moves the cursor vertically while maintaining the desired column position in the current buffer.
func (e *Editor) MoveCursorVertical(offset int, extend bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return ErrNoBuffer
	}

	// get the pos of the first selection's end
	selection := e.current.Selection()
	pos := selection.End

	currLine, currCol, err := e.current.PositionToLineCol(pos)
	if err != nil {
		return err
	}

	if e.desiredColumn == -1 {
		e.desiredColumn = currCol
	}

	targetLine := currLine + offset
	if targetLine < 0 {
		targetLine = 0
	}

	totalLines := e.current.LineCount()
	if targetLine >= totalLines {
		targetLine = totalLines - 1
	}

	return e.current.MoveSelectionToLineCol(targetLine, e.desiredColumn, extend)

	// currLine, currCol, err := e.current.PositionToLineCol(e.current.Cursor())
	// if err != nil {
	// 	return err
	// }

	// if e.desiredColumn == 0 {
	// 	e.desiredColumn = currCol
	// }

	// targetLine := currLine + offset
	// if targetLine < 0 {
	// 	targetLine = 0
	// }

	// totalLines := e.current.LineCount()
	// if targetLine >= totalLines {
	// 	targetLine = totalLines - 1
	// }

	// _, _, err = e.current.MoveCursorToLineCol(targetLine, e.desiredColumn)
	// return err
}

// SaveCurrentBuffer saves the current buffer.
func (e *Editor) SaveCurrentBuffer() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return ErrNoBuffer
	}
	return e.current.Save()
}

// CloseCurrentBuffer closes the current buffer.
func (e *Editor) CloseCurrentBuffer() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return ErrNoBuffer
	}

	if err := e.current.Close(); err != nil {
		return err
	}

	for path, b := range e.buffers {
		if b == e.current {
			delete(e.buffers, path)
			break
		}
	}

	for _, buf := range e.buffers {
		if buf != nil {
			e.current = buf
			return nil
		}
	}

	e.current = nil
	return nil
}

// GetLine returns a line as a string from the document.
func (e *Editor) GetLine(lineNum int) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.current == nil {
		return "", ErrNoBuffer
	}
	return e.current.GetLine(lineNum)
}

// GetLineCount returns the total number of lines in the buffer.
func (e *Editor) GetLineCount() (int, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.current == nil {
		return 0, ErrNoBuffer
	}
	return e.current.LineCount(), nil
}

// getBuffer returns a buffer by file path
func (e *Editor) getBuffer(filePath string) (*buffer.Buffer, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	buf, exists := e.buffers[absPath]
	if !exists {
		return nil, ErrBufferNotFound
	}

	return buf, nil
}
