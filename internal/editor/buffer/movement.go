package buffer

import (
	"unicode"
	"unicode/utf8"

	"github.com/lg2m/athena/internal/editor/state"
	"github.com/lg2m/athena/internal/util"
)

// MoveToNextWord moves the cursor to the next word boundary.
func (b *Buffer) MoveToNextWord(extend bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	newPos := b.findNextWordBoundary(b.selection.End, 1)

	if extend {
		// Extend selection to include the word
		b.selection.End = newPos
	} else {
		// Move cursor to new position (collapse selection)
		b.selection = state.Selection{Start: newPos, End: newPos}
	}

	return nil
}

// MoveToPrevWord moves the cursor to the previous word boundary
func (b *Buffer) MoveToPrevWord(extend bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	newPos := b.findNextWordBoundary(b.selection.Start-1, -1)

	if extend {
		if b.selection.End == b.selection.Start {
			b.selection.End = b.selection.Start
			b.selection.Start = newPos
		} else {
			b.selection.Start = newPos
		}
	} else {
		b.selection = state.Selection{Start: newPos, End: newPos}
	}

	return nil
}

// findNextWordBoundary finds the next word boundary position from the given position.
// direction: 1 for forward, -1 for backward TODO make constants
func (b *Buffer) findNextWordBoundary(pos int, direction int) int {
	totalLen := b.document.TotalGraphemes()
	if pos >= totalLen {
		return totalLen
	}
	if pos < 0 {
		return 0
	}

	// Get current grapheme to determine if we're in a word
	curr, err := b.document.GetTextRange(pos, pos+1)
	if err != nil {
		return pos
	}
	currType := getWordType(curr)

	nextPos := pos
	for {
		nextPos += direction
		if nextPos >= totalLen || nextPos < 0 {
			return util.Clamp(nextPos, 0, totalLen)
		}

		nextGrapheme, err := b.document.GetTextRange(nextPos, nextPos+1)
		if err != nil {
			return nextPos
		}
		nextType := getWordType(nextGrapheme)

		if nextType != currType {
			if direction > 0 {
				return nextPos
			} else {
				return nextPos + 1
			}
		}
	}
}

type WordType uint8

const (
	None       WordType = iota // none
	Letter                     // letters, numbers, underscores
	Whitespace                 // spaces, tabs, newlines
	Symbol                     // symbols, operators, punctuation
)

// getWordType returns the type of the grapheme cluster.
func getWordType(s string) WordType {
	if s == "" {
		return None
	}

	r, _ := utf8.DecodeRuneInString(s)
	switch {
	case unicode.IsSpace(r):
		return Whitespace
	case unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_':
		return Letter
	default:
		return Symbol
	}
}
