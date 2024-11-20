package buffer

import (
	"unicode"
	"unicode/utf8"

	"github.com/lg2m/athena/internal/editor/state"
	"github.com/lg2m/athena/internal/util"
)

// MoveSelections moves the selections by the specified offset.
// If `extend` is true, it extends the selection; otherwise, it moves the cursor.
func (b *Buffer) MoveSelections(offset int, extend bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	newPos := b.selection.End + offset
	newPos = util.Clamp(newPos, 0, b.document.TotalGraphemes())
	if extend {
		// extend the selection end
		b.selection.End = newPos
	} else {
		// move both start and end (cursor movement)
		b.selection = state.Selection{Start: newPos, End: newPos}
	}

	return nil
}

// MoveSelectionToLineCol moves the selection to a specific line and column.
func (b *Buffer) MoveSelectionToLineCol(line, col int, extend bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lineCacheMu.RLock()
	defer b.lineCacheMu.RUnlock()

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
		b.selection = state.Selection{Start: targetPos, End: targetPos}
	}

	return nil
}

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
	curr, err := b.document.Substring(pos, pos+1)
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

		nextGrapheme, err := b.document.Substring(nextPos, nextPos+1)
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
