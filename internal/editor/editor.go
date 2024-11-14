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
	mode          state.EditorMode
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

// FileType returns the file name related to the current active buffer.
func (e *Editor) FileType() (string, error) {
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
func (e *Editor) GetMode() state.EditorMode {
	return e.mode
}

// SetMode sets the current editor mode state.
func (e *Editor) SetMode(mode state.EditorMode) {
	e.mode = mode
}

// InsertText inserts text at the cursor position in the current buffer.
func (e *Editor) InsertText(text string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return ErrNoBuffer
	}

	// TODO: may not be desirable
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

	return e.current.Delete(pos, pos+length)
}

// GetCurrentPosition retrieves the current line and column of the cursor.
func (e *Editor) GetCurrentPosition() (int, int, error) {
	selection := e.current.Selection()
	pos := selection.End
	return e.current.PositionToLineCol(pos)
}

// LineCol retrieves the current line and column of a position.
func (e *Editor) LineCol(pos int) (int, int, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.current == nil {
		return 0, 0, ErrNoBuffer
	}
	return e.current.PositionToLineCol(pos)
}

// Selection retrieves the current selection in the active buffer.
func (e *Editor) Selection() (state.Selection, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.current == nil {
		return state.Selection{}, ErrNoBuffer
	}
	return e.current.Selection(), nil
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
}

// JumpFromCursor moves the cursor a specified number of lines relative to the current cursor position while maintaining the column position.
func (e *Editor) JumpFromCursor(offset int, extend bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return ErrNoBuffer
	}

	// get current pos
	selection := e.current.Selection()
	currLine, currCol, err := e.current.PositionToLineCol(selection.End)
	if err != nil {
		return err
	}

	// calc target line
	targetLine := currLine + offset

	// bounds check
	if targetLine < 0 {
		targetLine = 0
	}
	totalLines := e.current.LineCount()
	if targetLine >= totalLines {
		targetLine = totalLines - 1
	}

	if e.desiredColumn == -1 {
		e.desiredColumn = currCol
	}

	return e.current.MoveSelectionToLineCol(targetLine, e.desiredColumn, extend)
}

// JumpToLine moves the cursor to a specific line number (0-based) and attempts to retain column position (when possible).
func (e *Editor) JumpToLine(lineNum int, extend bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.current == nil {
		return ErrNoBuffer
	}

	// bounds check
	if lineNum < 0 {
		lineNum = 0
	}
	totalLines := e.current.LineCount()
	if lineNum >= totalLines {
		lineNum = totalLines - 1
	}

	// current column for maintaining desired column
	selection := e.current.Selection()
	_, currCol, err := e.current.PositionToLineCol(selection.End)
	if err != nil {
		return err
	}

	if e.desiredColumn == -1 {
		e.desiredColumn = currCol
	}

	return e.current.MoveSelectionToLineCol(lineNum, e.desiredColumn, extend)
}

// JumpToTop moves the cursor to the beginning of the document.
func (e *Editor) JumpToTop(extend bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.current == nil {
		return ErrNoBuffer
	}
	return e.current.MoveSelectionToLineCol(0, 0, extend)
}

// JumpToBottom moves the cursor to the end of the document.
func (e *Editor) JumpToBottom(extend bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.current == nil {
		return ErrNoBuffer
	}
	lastLine := e.current.LineCount() - 1
	return e.current.MoveSelectionToLineCol(lastLine, 0, extend)
}

// MoveToNextWord moves the cursor to the beginning of the next word boundary.
func (e *Editor) MoveToNextWord(extend bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.current == nil {
		return ErrNoBuffer
	}
	return e.current.MoveToNextWord(extend)
}

// MoveToPrevWord moves the cursor to the beginning of the previous word boundary.
func (e *Editor) MoveToPrevWord(extend bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.current == nil {
		return ErrNoBuffer
	}
	return e.current.MoveToPrevWord(extend)
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
