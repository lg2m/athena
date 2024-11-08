package editor

import (
	"errors"

	"github.com/lg2m/athena/internal/editor/buffer"
	"github.com/lg2m/athena/internal/editor/state"
)

// NormalOp represents a normal mode operation.
type NormalOp string

const (
	// Movements
	MoveLeft  NormalOp = "h"
	MoveRight NormalOp = "l"
	MoveUp    NormalOp = "k"
	MoveDown  NormalOp = "j"

	// Selections
	SelectAll  NormalOp = "%"
	SelectLine NormalOp = "x"
	SelectWord NormalOp = "w"

	// Edit
	DeleteSelection NormalOp = "d"

	// Mode switching
	EnterInsert     NormalOp = "i"
	InsertBeginLine NormalOp = "I"
	AppendAfter     NormalOp = "a"
	AppendEndLine   NormalOp = "A"
	AppendLineBelow NormalOp = "o"
	InsertLineAbove NormalOp = "O"
)

// IsMovement returns true if the operation is a movement operation
func (op NormalOp) IsMovement() bool {
	switch op {
	case MoveLeft, MoveRight, MoveUp, MoveDown:
		return true
	default:
		return false
	}
}

// IsSelection returns true if the operation is a selection operation
func (op NormalOp) IsSelection() bool {
	switch op {
	case SelectAll, SelectLine, SelectWord:
		return true
	default:
		return false
	}
}

// String returns the string representation of the operation
func (op NormalOp) String() string {
	return string(op)
}

// ParseNormalOp converts a string to a NormalOp if valid
func ParseNormalOp(s string) (NormalOp, bool) {
	switch s {
	case string(MoveLeft), string(MoveRight), string(MoveUp), string(MoveDown),
		string(SelectAll), string(SelectLine), string(SelectWord),
		string(DeleteSelection), string(EnterInsert), string(AppendAfter):
		return NormalOp(s), true
	default:
		return "", false
	}
}

// Editor represents the main editor application.
type Editor struct {
	Manager *buffer.BufferManager
	mode    state.Mode
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
	buffer, err := buffer.NewBufferFromFile(filePath)
	if err != nil {
		return err
	}
	e.Manager.AddBuffer(buffer)
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
	buffer := e.Manager.GetCurrentBuffer()
	if buffer == nil {
		return errors.New("no current buffer")
	}
	return buffer.Insert(text)
}

// DeleteText deletes text of specified length from the cursor position.
func (e *Editor) DeleteText(length int) error {
	buffer := e.Manager.GetCurrentBuffer()
	if buffer == nil {
		return errors.New("no current buffer")
	}
	return buffer.Delete(length)
}

// MoveCursor moves the cursor in the current buffer.
func (e *Editor) MoveCursor(offset int) error {
	buffer := e.Manager.GetCurrentBuffer()
	if buffer == nil {
		return errors.New("no current buffer")
	}
	buffer.MoveCursor(offset)
	return nil
}

// SaveCurrentBuffer saves the current buffer to its file.
func (e *Editor) SaveCurrentBuffer() error {
	buffer := e.Manager.GetCurrentBuffer()
	if buffer == nil {
		return errors.New("no current buffer")
	}
	return buffer.Save()
}
