package ui

import (
	"fmt"
	"strconv"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/lg2m/athena/internal/athena/config"
	"github.com/lg2m/athena/internal/editor"
	"github.com/lg2m/athena/internal/editor/state"
)

// DocumentView represents the main document (or file) view.
type DocumentView struct {
	BaseView
	editor   *editor.Editor
	cfg      *config.Config
	viewport *Viewport

	keyBuffer     string
	numericPrefix string

	goToMenu *GoToMenu
}

func NewDocumentView(e *editor.Editor, cfg *config.Config, v *Viewport) *DocumentView {
	return &DocumentView{
		editor:   e,
		cfg:      cfg,
		viewport: v,
		goToMenu: NewGoToMenu(cfg),
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

	mode := v.editor.GetMode()
	cursorShape := v.getCursorShape(mode)

	// Get the current selection range
	// selection, _ := v.editor.Selection()

	for i := 0; i < v.height; i++ {
		lineIdx := start + i
		if lineIdx >= end {
			break
		}

		line, err := v.editor.GetLine(lineIdx)
		if err != nil {
			continue
		}

		runes := []rune(line)

		for x := range runes {
			style := tcell.StyleDefault

			// If we have a selection (start != end) and current position is within selection
			// if selection.Start != selection.End && x >= selection.Start && x < selection.End {
			// 	style = style.Background(tcell.ColorGray)
			// }

			// If this is the cursor position, apply cursor style
			if lineIdx == currLine && x == currCol {
				if mode == state.Normal {
					style = v.getCursorStyle(cursorShape)
				} else {
					style = style.Reverse(true)
				}
			}

			screen.SetContent(v.x+x, v.y+i, runes[x], nil, style)
		}

		// Handle cursor at end of line
		if lineIdx == currLine && currCol >= len(runes) {
			style := tcell.StyleDefault
			if mode == state.Normal {
				style = v.getCursorStyle(cursorShape)
			} else {
				style = style.Reverse(true)
			}
			screen.SetContent(v.x+len(runes), v.y+i, ' ', nil, style)
		}
	}

	v.goToMenu.Draw(screen, v.height)
}

func (v *DocumentView) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		key := getKeyString(ev)
		mode := v.editor.GetMode()
		var keymap map[string]config.KeyAction

		switch mode {
		case state.Normal:
			keymap = v.cfg.Keymap.Normal
		case state.Insert:
			keymap = v.cfg.Keymap.Insert
		}

		// Handle numeric prefixes (digits)
		if isDigit(key) && mode == state.Normal {
			v.numericPrefix += key
			return true
		}

		v.keyBuffer += key

		action, partial, matched := v.matchKeySequence(keymap)
		if matched {
			v.keyBuffer = ""
			return v.executeAction(action)
		} else if partial {
			if v.keyBuffer[0] == 'g' && !v.goToMenu.visible {
				v.goToMenu.Show()
			}

			if key == "<esc>" {
				v.goToMenu.Hide()
				v.numericPrefix = ""
				v.keyBuffer = ""
				return false
			}

			return true
		} else {
			v.keyBuffer = ""
			if ev.Key() == tcell.KeyRune && mode == state.Insert {
				_ = v.editor.InsertText(string(ev.Rune()))
				return true
			}
		}
	}
	return false
}

func (v *DocumentView) matchKeySequence(keymap config.KeyMap) (string, bool, bool) {
	if len(v.keyBuffer) == 0 || keymap == nil {
		return "", false, false
	}

	if actionVal, exists := keymap[v.keyBuffer]; exists {
		if actionStr, ok := actionVal.(string); ok {
			return actionStr, true, true
		}
	}

	firstKey := string(v.keyBuffer[0])
	actionVal, exists := keymap[firstKey]
	if !exists {
		// First key does not exist in keymap.
		return "", false, false
	}

	switch val := actionVal.(type) {
	case map[string]interface{}:

		if len(v.keyBuffer) == 1 {
			// Only the first key is present; it's a partial match.
			return "", true, false
		}

		secondKey := string(v.keyBuffer[1])
		if secondAction, exists := val[secondKey]; exists {
			if actionStr, ok := secondAction.(string); ok {
				return actionStr, true, true
			}
			// If the secondAction exists but is not a string, it's an unexpected type.
			return "", false, false
		}

		return "", true, false

	default:
		// Unsupported type encountered in keymap.
		return "", false, false
	}

	// if action, ok := keymap[v.keyBuffer]; ok {
	// 	if s, isStr := action.(string); isStr {
	// 		return s, true, true
	// 	}
	// }

	// key := string(v.keyBuffer[0])

	// action, exists := keymap[key]
	// if !exists {
	// 	return "", false, false
	// }

	// switch val := action.(type) {
	// case map[string]interface{}:
	// 	if len(v.keyBuffer) == 1 {
	// 		// partial, not full
	// 		return "", true, false
	// 	}

	// 	// We have a second key
	// 	secondKey := string(v.keyBuffer[1])
	// 	if secondAction, exists := val[secondKey]; exists {
	// 		return secondAction.(string), true, true
	// 	}

	// 	return "", true, false

	// case string:
	// 	if len(v.keyBuffer) == 1 {
	// 		_ = v.editor.InsertText(fmt.Sprintf("Matching key: %+v %+v\n", key, keymap))
	// 		return val, true, true
	// 	}
	// }

	// return "", false, false
}

func (v *DocumentView) getNumericPrefixOrDefault(defaultValue int) int {
	if v.numericPrefix != "" {
		if n, err := strconv.Atoi(v.numericPrefix); err == nil {
			v.numericPrefix = ""
			return n
		}
		v.numericPrefix = ""
	}
	return defaultValue
}

func (v *DocumentView) executeAction(action string) bool {
	switch action {
	case "enter_insert_mode":
		v.editor.SetMode(state.Insert)
	case "enter_normal_mode":
		v.editor.SetMode(state.Normal)
	case "move_left":
		_ = v.editor.MoveCursorHorizontal(-1, false)
	case "move_right":
		_ = v.editor.MoveCursorHorizontal(1, false)
	case "move_down":
		mult := v.getNumericPrefixOrDefault(1)
		_ = v.editor.JumpFromCursor(mult, false)
		v.centerCursor()
	case "move_up":
		mult := v.getNumericPrefixOrDefault(1)
		_ = v.editor.JumpFromCursor(-mult, false)
		v.centerCursor()
	case "move_next_word":
		_ = v.editor.MoveToNextWord(false)
		v.centerCursor()
	case "move_prev_word":
		_ = v.editor.MoveToPrevWord(false)
		v.centerCursor()
	case "delete_backwards":
		_ = v.editor.DeleteText(-1)
	case "delete_forward":
		_ = v.editor.DeleteText(1)
	case "new_line":
		_ = v.editor.InsertText("\n")
	case "show_goto_menu":
		v.goToMenu.Show()
	case "go_to_top":
		lineNum := v.getNumericPrefixOrDefault(1) - 1
		if lineNum < 0 {
			lineNum = 0
		}
		_ = v.editor.JumpToLine(lineNum, false)
		v.centerCursor()
		v.goToMenu.Hide()
	case "go_to_bottom":
		_ = v.editor.JumpToBottom(false)
		v.centerCursor()
		v.goToMenu.Hide()
	default:
		return false
	}
	v.numericPrefix = ""
	return true
}

func (v *DocumentView) centerCursor() {
	// Get current cursor position
	if line, _, err := v.editor.GetCurrentPosition(); err == nil {
		// Adjust view offset to center the cursor
		viewHeight := v.height
		newOffset := line - (viewHeight / 2)
		if newOffset < 0 {
			newOffset = 0
		}
		v.viewport.offset = newOffset
	}
}

func (v *DocumentView) getCursorShape(mode state.EditorMode) config.CursorShape {
	switch mode {
	case state.Insert:
		return v.cfg.Editor.CursorShape.Insert
	default:
		return v.cfg.Editor.CursorShape.Normal
	}
}

func (v *DocumentView) getCursorStyle(shape config.CursorShape) tcell.Style {
	style := tcell.StyleDefault
	switch shape {
	case config.CursorBlock:
		return style.Reverse(true)
	case config.CursorLine:
		return style.Background(tcell.ColorWhiteSmoke)
	case config.CursorUnder:
		return style.Underline(true)
	default:
		return style.Reverse(true)
	}
}

func getKeyString(ev *tcell.EventKey) string {
	if ev.Modifiers()&tcell.ModCtrl != 0 && ev.Key() == tcell.KeyRune {
		return fmt.Sprintf("<c-%c>", ev.Rune())
	}

	switch ev.Key() {
	case tcell.KeyEscape:
		return "<esc>"
	case tcell.KeyEnter:
		return "<cr>"
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return "<bs>"
	case tcell.KeyDelete:
		return "<del>"
	case tcell.KeyTab:
		return "<tab>"
	case tcell.KeyLeft:
		return "<left>"
	case tcell.KeyRight:
		return "<right>"
	case tcell.KeyUp:
		return "<up>"
	case tcell.KeyDown:
		return "<down>"
	case tcell.KeyRune:
		return string(ev.Rune())
	default:
		return ev.Name()
	}
}

func isDigit(key string) bool {
	return len(key) == 1 && unicode.IsDigit(rune(key[0]))
}

// GoToMenu represents the menu overlay for goto commands
type GoToMenu struct {
	visible bool
	x, y    int // Position of the menu
	width   int // Width of the menu
	options []string
}

func NewGoToMenu(cfg *config.Config) *GoToMenu {
	options := []string{"[g] → goto commands"}
	if gotoMappings, ok := cfg.Keymap.Normal["g"].(map[string]interface{}); ok {
		for key, action := range gotoMappings {
			if key != "default" {
				option := fmt.Sprintf("  g%s → %s", key, action)
				options = append(options, option)
			}
		}
	}
	return &GoToMenu{
		options: options,
		width:   25,
	}
}

// Show makes the menu visible
func (m *GoToMenu) Show() {
	m.visible = true
}

// Hide makes the menu invisible
func (m *GoToMenu) Hide() {
	m.visible = false
}

func (m *GoToMenu) Visible() bool {
	return m.visible
}

func (m *GoToMenu) Draw(screen tcell.Screen, viewHeight int) {
	if !m.visible {
		return
	}

	// Position menu at bottom right
	startY := viewHeight - len(m.options) - 1
	w, _ := screen.Size()
	startX := w - m.width - 2

	// Store position for potential future use
	m.x = startX
	m.y = startY

	// Styles
	style := tcell.StyleDefault.Background(tcell.ColorGray).Foreground(tcell.ColorWhite)
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	// Draw top border
	screen.SetContent(startX, startY-1, '╭', nil, borderStyle)
	screen.SetContent(startX+m.width+1, startY-1, '╮', nil, borderStyle)
	for x := startX + 1; x < startX+m.width+1; x++ {
		screen.SetContent(x, startY-1, '─', nil, borderStyle)
	}

	// Draw options with background
	for i, opt := range m.options {
		y := startY + i
		// Draw left border
		screen.SetContent(startX, y, '│', nil, borderStyle)

		// Draw option text
		runes := []rune(opt)
		for x := 0; x < m.width; x++ {
			ch := ' '
			if x < len(runes) {
				ch = runes[x]
			}
			screen.SetContent(startX+x+1, y, ch, nil, style)
		}

		// Draw right border
		screen.SetContent(startX+m.width+1, y, '│', nil, borderStyle)
	}

	// Draw bottom border
	screen.SetContent(startX, startY+len(m.options), '╰', nil, borderStyle)
	screen.SetContent(startX+m.width+1, startY+len(m.options), '╯', nil, borderStyle)
	for x := startX + 1; x < startX+m.width+1; x++ {
		screen.SetContent(x, startY+len(m.options), '─', nil, borderStyle)
	}
}
