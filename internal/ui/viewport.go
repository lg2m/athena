package ui

// Viewport handles scrolling and visible area management.
type Viewport struct {
	offset  int // lines scrolled from top
	padding int // lines to keep visible above/below cursor
}

func NewViewport(padding int) *Viewport {
	return &Viewport{
		padding: padding,
	}
}

// Update adjusts viewport position to keep cursor visible.
func (v *Viewport) Update(currLine, viewHeight int) {
	if currLine-v.offset < v.padding {
		// cursor too close to top
		v.offset = max(0, currLine-v.padding)
	} else if currLine-v.offset > viewHeight-v.padding {
		// cursor too close to bottom
		v.offset = currLine - (viewHeight - v.padding)
	}
}

// VisibleRange returns the range of visible lines.
func (v *Viewport) VisibleRange(viewHeight, totalLines int) (start, end int) {
	start = v.offset
	end = min(totalLines, v.offset+viewHeight)
	return start, end
}
