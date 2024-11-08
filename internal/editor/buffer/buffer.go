package buffer

import (
	"errors"
	"io"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidRange    = errors.New("invalid range")
	ErrOutOfBounds     = errors.New("position out of bounds")
	ErrBufferEmpty     = errors.New("buffer is empty")
	ErrInvalidLineCol  = errors.New("invalid line or column")
	ErrInvalidPosition = errors.New("invalid position")
)

// Buffer represents a single text buffer (document).
type Buffer struct {
	manager       *ChunkManager
	cursor        *Cursor
	name          string
	lastSavePoint time.Time
	mu            sync.RWMutex
}

// NewBuffer creates a new Buffer with optional initial content.
func NewBuffer(filePath string) (*Buffer, error) {
	manager, err := NewChunkManager(filePath)
	if err != nil {
		return nil, err
	}

	return &Buffer{
		manager:       manager,
		cursor:        &Cursor{Position: 0},
		name:          filePath,
		lastSavePoint: time.Now(),
	}, nil
}

// Insert inserts text at the cursor's current position.
func (b *Buffer) Insert(s string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	chunkID := b.cursor.Position / b.manager.chunkSize
	chunk, err := b.manager.GetChunk(chunkID)
	if err != nil {
		return err
	}

	// Perform the insertion
	relativePos := b.cursor.Position % b.manager.chunkSize
	chunkData := chunk.Read()
	newData := chunkData[:relativePos] + s + chunkData[relativePos:]
	chunk.Write(newData)

	// Update cursor
	b.cursor.Position += len(s)

	return nil
}

// Delete deletes text from the cursor position to position + length.
func (b *Buffer) Delete(start, end int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if start < 0 || end > b.Size() || start > end {
		return ErrInvalidRange
	}

	// Get affected chunks
	startChunk := start / b.manager.chunkSize
	endChunk := end / b.manager.chunkSize

	// Handle single chunk case
	if startChunk == endChunk {
		chunk, err := b.manager.GetChunk(startChunk)
		if err != nil {
			return err
		}

		chunkData := chunk.Read()
		relStart := start % b.manager.chunkSize
		relEnd := end % b.manager.chunkSize
		newData := chunkData[:relStart] + chunkData[relEnd:]
		chunk.Write(newData)

		// Update cursor if needed
		if b.cursor.Position > start {
			b.cursor.Position = start
		}

		return nil
	}

	// Handle multi-chunk deletion
	var newContent strings.Builder

	// First chunk
	firstChunk, err := b.manager.GetChunk(startChunk)
	if err != nil {
		return err
	}
	firstChunkData := firstChunk.Read()
	relStart := start % b.manager.chunkSize
	newContent.WriteString(firstChunkData[:relStart])

	// Last chunk
	lastChunk, err := b.manager.GetChunk(endChunk)
	if err != nil && err != io.EOF {
		return err
	}
	if err != io.EOF {
		lastChunkData := lastChunk.Read()
		relEnd := end % b.manager.chunkSize
		newContent.WriteString(lastChunkData[relEnd:])
	}

	// Write the combined content to the first chunk
	firstChunk.Write(newContent.String())

	// Mark intermediate chunks for deletion
	for chunkID := startChunk + 1; chunkID <= endChunk; chunkID++ {
		delete(b.manager.chunks, chunkID)
		b.manager.loadedChunks--
	}

	// Update cursor if needed
	if b.cursor.Position > start {
		b.cursor.Position = start
	}

	return nil
}

// MoveCursor moves the cursor by the specified offset.
func (b *Buffer) MoveCursor(offset int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	newPos := b.cursor.Position + offset
	return b.SetCursor(newPos)
}

// MoveCursorToLineCol moves the cursor to a specific line and column
// It clamps the column to the end of the line if necessary but preserves the desired column.
func (b *Buffer) MoveCursorToLineCol(line, col int) (int, int, error) {
	if line < 0 || col < 0 {
		return 0, 0, ErrInvalidLineCol
	}

	currentLine := 0
	currentCol := 0
	position := 0

	targetPosition := -1
	targetCol := col

	for chunkID := 0; ; chunkID++ {
		chunk, err := b.manager.GetChunk(chunkID)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, 0, err
		}
		chunkData := chunk.Read()
		for _, r := range chunkData {
			if currentLine == line {
				if currentCol == col {
					targetPosition = position
					targetCol = currentCol
					break
				}
				if r == '\n' {
					// EOL reached - clamp to the end of column
					targetPosition = position
					targetCol = currentCol
					break
				}
				currentCol++
			}
			if r == '\n' {
				currentLine++
				currentCol = 0
			}
			position++
		}
		if targetPosition != -1 {
			break
		}
	}
	if targetPosition == -1 && currentLine == line {
		// line exists but might be empty
		targetPosition = position
		targetCol = currentCol
	}
	if targetPosition != -1 {
		if err := b.SetCursor(targetPosition); err != nil {
			return 0, 0, err
		}
		return currentLine, targetCol, nil
	}

	return 0, 0, ErrInvalidLineCol
}

// PositionToLineCol converts a buffer position to line and column numbers.
func (b *Buffer) PositionToLineCol(pos int) (int, int, error) {
	if pos < 0 || pos > b.Size() {
		return 0, 0, ErrInvalidPosition
	}
	line := 0
	column := 0
	position := 0
	for chunkID := 0; ; chunkID++ {
		chunk, err := b.manager.GetChunk(chunkID)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, 0, err
		}
		chunkData := chunk.Read()
		for _, r := range chunkData {
			if position == pos {
				return line, column, nil
			}
			if r == '\n' {
				line++
				column = 0
			} else {
				column++
			}
			position++
		}
	}
	// If pos is at the end of the buffer
	if position == pos {
		return line, column, nil
	}
	return 0, 0, ErrInvalidPosition
}

// SetCursor sets the cursor to the specified position.
func (b *Buffer) SetCursor(pos int) error {
	if pos < 0 || pos > b.Size() {
		return ErrOutOfBounds
	}
	b.cursor.Position = pos
	return nil
}

// GetLine returns the content of a specific line
func (b *Buffer) GetLine(lineNum int) (string, error) {
	if lineNum < 0 {
		return "", ErrInvalidLineCol
	}

	var line strings.Builder
	currentLine := 0
	foundLine := false

	// Iterate through chunks to find the line
	for chunkID := 0; ; chunkID++ {
		chunk, err := b.manager.GetChunk(chunkID)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		chunkData := chunk.Read()
		for _, r := range chunkData {
			if currentLine == lineNum {
				foundLine = true
				if r == '\n' {
					return line.String(), nil
				}
				line.WriteRune(r)
			} else if currentLine > lineNum {
				return line.String(), nil
			}

			if r == '\n' {
				currentLine++
			}
		}
	}

	if !foundLine {
		return "", ErrInvalidLineCol
	}
	return line.String(), nil
}

// GetLines returns a range of lines
func (b *Buffer) GetLines(start, end int) ([]string, error) {
	if start < 0 || end < start {
		return nil, ErrInvalidRange
	}

	lines := make([]string, 0, end-start+1)
	var currentLine strings.Builder
	lineNum := 0

	// Iterate through chunks to collect lines
	for chunkID := 0; ; chunkID++ {
		chunk, err := b.manager.GetChunk(chunkID)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		chunkData := chunk.Read()
		for _, r := range chunkData {
			if lineNum >= start && lineNum <= end {
				if r == '\n' {
					lines = append(lines, currentLine.String())
					currentLine.Reset()
					lineNum++
					if lineNum > end {
						return lines, nil
					}
				} else {
					currentLine.WriteRune(r)
				}
			} else if lineNum > end {
				return lines, nil
			} else if r == '\n' {
				lineNum++
			}
		}
	}

	if currentLine.Len() > 0 && lineNum <= end {
		lines = append(lines, currentLine.String())
	}

	return lines, nil
}

// GetTextRange retrieves text from start to end graphemes indices (exclusive).
func (b *Buffer) GetTextRange(start, end int) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	startChunk := start / b.manager.chunkSize
	endChunk := end / b.manager.chunkSize

	var result strings.Builder

	for chunkID := startChunk; chunkID <= endChunk; chunkID++ {
		chunk, err := b.manager.GetChunk(chunkID)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		chunkStart := chunkID * b.manager.chunkSize
		chunkData := chunk.Read()

		// Calculate relative positions within chunk
		relativeStart := max(0, start-chunkStart)
		relativeEnd := min(len(chunkData), end-chunkStart)

		result.WriteString(chunkData[relativeStart:relativeEnd])
	}

	return result.String(), nil
}

// GetCursorPosition returns the cursor's line and column numbers.
func (b *Buffer) GetCursorPosition() (line, column int) {
	textBeforeCursor, _ := b.GetTextRange(0, b.cursor.Position)
	lines := strings.Split(textBeforeCursor, "\n")
	line = len(lines) - 1
	column = len([]rune(lines[line]))
	return
}

// GetCursor returns the cursor's position in the graphemes cluster.
func (b *Buffer) GetCursor() int {
	return b.cursor.Position
}

// Save writes the buffer content to the associated file.
func (b *Buffer) Save() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Save all dirty chunks
	for _, chunk := range b.manager.chunks {
		if chunk.Dirty {
			if err := b.manager.saveChunk(chunk); err != nil {
				return err
			}
		}
	}

	b.lastSavePoint = time.Now()
	return nil
}

// Size returns the total size of the buffer in bytes
func (b *Buffer) Size() int {
	return int(b.manager.totalSize)
}

// Close properly closes the buffer and its resources
func (b *Buffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Save any remaining dirty chunks
	if err := b.Save(); err != nil {
		return err
	}

	// Close the file handle
	if err := b.manager.file.Close(); err != nil {
		return err
	}

	// Clear the chunks map
	b.manager.chunks = nil

	return nil
}

// LineCount returns the total number of lines in the buffer
func (b *Buffer) LineCount() (int, error) {
	count := 0
	for chunkID := 0; ; chunkID++ {
		chunk, err := b.manager.GetChunk(chunkID)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

		chunkData := chunk.Read()
		count += strings.Count(chunkData, "\n")
	}

	// Add 1 if buffer doesn't end with newline
	lastChunk, err := b.manager.GetChunk(int(b.manager.totalSize / int64(b.manager.chunkSize)))
	if err == nil && !strings.HasSuffix(lastChunk.Read(), "\n") {
		count++
	}

	return count, nil
}
