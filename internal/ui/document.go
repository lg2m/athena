package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/lg2m/athena/internal/athena/config"
	"github.com/lg2m/athena/internal/editor"
)

// DocumentView represents the main document (or file) view.
type DocumentView struct {
	BaseView
	editor   *editor.Editor
	cfg      *config.Config
	viewport *Viewport
}

func NewDocumentView(e *editor.Editor, cfg *config.Config, v *Viewport) *DocumentView {
	return &DocumentView{
		editor:   e,
		cfg:      cfg,
		viewport: v,
	}
}

// Draw implements the document view.
func (v *DocumentView) Draw(screen tcell.Screen) {
	currLine, currCol, _ := v.editor.GetCurrentPosition()
	total, _ := v.editor.GetLineCount()

	// Update viewport to ensure cursor visibility
	v.viewport.Update(currLine, v.height)

	// Get visible range from viewport
	start, end := v.viewport.VisibleRange(v.height, total)

	for i := 0; i < v.height; i++ {
		lineIdx := start + i
		if lineIdx >= end {
			break
		}

		line, err := v.editor.GetLine(lineIdx)
		if err != nil {
			continue
		}

		style := tcell.StyleDefault
		runes := []rune(line)

		for x, ch := range runes {
			if lineIdx == currLine && x == currCol {
				style = style.Reverse(true)
			} else {
				style = style.Reverse(false)
			}
			screen.SetContent(v.x+x, v.y+i, ch, nil, style)
		}

		// Draw cursor at end of line if needed
		if lineIdx == currLine && currCol >= len(runes) {
			screen.SetContent(v.x+len(runes), v.y+i, ' ', nil, style.Reverse(true))
		}
	}
}
