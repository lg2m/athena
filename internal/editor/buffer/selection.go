package buffer

import (
	"errors"

	"github.com/lg2m/athena/internal/rope"
)

// Selection represents a range of text within the buffer.
type Selection struct {
	Start int // Start pos in grapheme indices
	End   int // End pos in grapheme indices
}

// NewSelection creates a new Selection instance.
// It initializes both Start and End to the cursor position.
func NewSelection(cursorPos int) *Selection {
	return &Selection{
		Start: cursorPos,
		End:   cursorPos,
	}
}

// SetStart sets the start position of the selection.
func (s *Selection) SetStart(pos int) {
	s.Start = pos
}

// SetEnd sets the end position of the selection.
func (s *Selection) SetEnd(pos int) {
	s.End = pos
}

// Clear resets the selection to an empty state.
func (s *Selection) Clear() {
	s.Start = 0
	s.End = 0
}

// IsEmpty checks if the selection is empty.
func (s *Selection) IsEmpty() bool {
	return s.Start == s.End
}

// Validate ensures that the selection indices are within bounds and Start <= End.
func (s *Selection) Validate(totalGraphemes int) error {
	if s.Start < 0 || s.End < 0 {
		return errors.New("selection indices cannot be negative")
	}
	if s.Start > totalGraphemes || s.End > totalGraphemes {
		return errors.New("selection indices exceed buffer length")
	}
	if s.Start > s.End {
		return errors.New("selection start cannot be after end")
	}
	return nil
}

// GetSelectedText retrieves the selected text from the rope.
func (s *Selection) GetSelectedText(r *rope.Rope) (string, error) {
	if s.IsEmpty() {
		return "", nil
	}
	if err := s.Validate(r.TotalGraphemes()); err != nil {
		return "", err
	}
	selectedText, err := r.GetTextRange(s.Start, s.End)
	if err != nil {
		return "", err
	}
	return selectedText, nil
}
