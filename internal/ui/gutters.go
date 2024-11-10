package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lg2m/athena/internal/athena/config"
	"github.com/lg2m/athena/internal/editor"
)

// GuttersView represents the line numbers view.
type GuttersView struct {
	BaseView
	editor   *editor.Editor
	cfg      *config.Config
	viewport *Viewport
}

func NewGuttersView(e *editor.Editor, cfg *config.Config, v *Viewport) *GuttersView {
	return &GuttersView{editor: e, cfg: cfg, viewport: v}
}

// Draw implements the gutter view.
func (v *GuttersView) Draw(screen tcell.Screen) {
	currLine, _, _ := v.editor.GetCurrentPosition()
	total, _ := v.editor.GetLineCount()

	start, _ := v.viewport.VisibleRange(v.height, total)

	style := tcell.StyleDefault.Foreground(tcell.ColorPurple)
	currStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	for i := 0; i < v.height; i++ {
		lineNum := start + i + 1
		y := i

		var numStr string
		lineStyle := style

		if lineNum > total {
			// Draw '~' for lines beyond the end of the file (EOF).
			numStr = fmt.Sprintf("%*s", v.width-1, "~")
		} else {
			switch v.cfg.Editor.LineNumber {
			case config.LineNumberAbsolute:
				// Absolute numbering: display the actual line number.
				numStr = fmt.Sprintf("%*d", v.width-1, lineNum)
				if lineNum == currLine+1 {
					// Highlight the current line number.
					lineStyle = currStyle
				}
			case config.LineNumberRelative:
				if lineNum == currLine+1 {
					// Current line: display absolute number with a distinct style.
					numStr = fmt.Sprintf("%*d", v.width-1, lineNum)
					lineStyle = currStyle
				} else {
					// Relative numbering: display the distance from the current line.
					distance := lineNum - (currLine + 1)
					if distance < 0 {
						distance = -distance
					}
					numStr = fmt.Sprintf("%*d", v.width-1, distance)
				}
			default:
				numStr = ""
			}
		}

		// Render the line number string on the screen.
		for x, ch := range numStr {
			screen.SetContent(v.x+x, v.y+y, ch, nil, lineStyle)
		}
	}
}
