package buffer

// Cursor represents the cursor in the buffer.
type Cursor struct {
	Position int // Position in terms of grapheme clusters
}

// SetPosition sets the cursor position index.
func (c *Cursor) SetPosition(pos int) {
	c.Position = pos
}
