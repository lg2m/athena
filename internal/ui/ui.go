package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/lg2m/athena/internal/editor"
	"github.com/lg2m/athena/internal/editor/buffer"
	"github.com/lg2m/athena/internal/editor/state"
	"github.com/lg2m/athena/internal/util"
	"github.com/rivo/tview"
)

// Option defines a functional option for configuring Athena.
type Option = func(*Athena)

// Athena represents the main UI component of the editor.
// All fields are immutable after initialization unless protected by mu.
type Athena struct {
	app     *tview.Application
	editor  *editor.Editor
	layout  *tview.Flex
	display struct {
		gutters   *tview.TextView
		document  *tview.TextView
		statusBar *tview.TextView
	}
	config struct {
		gutterWidth int
		filePath    string
	}
	mu sync.RWMutex
}

// NewAthena creates and configures a new Athena editor instance.
func NewAthena(opts ...Option) (*Athena, error) {
	a := &Athena{
		app:    tview.NewApplication(),
		editor: editor.NewEditor(),
	}
	a.config.gutterWidth = 6

	a.initializeComponents()

	for _, opt := range opts {
		opt(a)
	}

	if a.config.filePath != "" {
		if err := a.LoadFile(a.config.filePath); err != nil {
			return nil, err
		}
	}

	return a, nil
}

// LoadFile loads a file into the editor and refreshes the display.
func (a *Athena) LoadFile(filePath string) error {
	if err := a.editor.OpenFile(filePath); err != nil {
		return fmt.Errorf("failed to load file: %w", err)
	}
	buffer := a.editor.Manager.GetCurrentBuffer()
	if buffer == nil {
		return fmt.Errorf("unable to get buffer")
	}
	a.refreshContent(buffer)
	return nil
}

// WithFilePath sets the initial file to be loaded into the editor.
func WithFilePath(filePath string) Option {
	return func(a *Athena) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return
		}
		a.config.filePath = absPath
	}
}

// WithGutterWidth sets a custom gutter width.
func WithGutterWidth(width int) Option {
	return func(a *Athena) {
		a.config.gutterWidth = width
	}
}

// initializeComponents sets up all UI components with their configurations.
func (a *Athena) initializeComponents() {
	a.display.gutters = a.newGutters()
	a.display.document = a.newDocument()
	a.display.statusBar = a.newStatusBar()
	a.layout = a.newLayout()
}

func (a *Athena) newGutters() *tview.TextView {
	gutters := tview.NewTextView()
	return gutters.
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)
}

func (a *Athena) newDocument() *tview.TextView {
	doc := tview.NewTextView().
		SetWrap(false).
		SetWordWrap(false).
		SetDynamicColors(true).
		SetRegions(true)
	doc.
		SetTitle("Athena Editor").
		SetInputCapture(a.handleKeyPress)
	return doc
}

func (a *Athena) newStatusBar() *tview.TextView {
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)
	statusBar.SetBackgroundColor(tcell.ColorNavy)
	return statusBar
}

func (a *Athena) newLayout() *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().
				AddItem(a.display.gutters, a.config.gutterWidth, 0, false).
				AddItem(a.display.document, 0, 1, true),
			0, 1, true).
		AddItem(a.display.statusBar, 1, 0, false)
}

// Run starts the Athena application.
func (a *Athena) Run() error {
	if err := a.app.SetRoot(a.layout, true).EnableMouse(true).Run(); err != nil {
		return fmt.Errorf("application error: %w", err)
	}
	return nil
}

// handleKeyPress processes keyboard input events.
// It uses a read lock for operations that don't modify state.
func (a *Athena) handleKeyPress(event *tcell.EventKey) *tcell.EventKey {
	a.mu.RLock()
	buffer := a.editor.Manager.GetCurrentBuffer()
	if buffer == nil {
		a.mu.RUnlock()
		return event
	}
	mode := a.editor.GetMode()
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	switch mode {
	case state.Normal:
		return a.handleNormalMode(event, buffer)
	case state.Insert:
		return a.handleInsertMode(event, buffer)
	default:
		return event
	}
}

func (a *Athena) handleNormalMode(event *tcell.EventKey, b *buffer.Buffer) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlC:
		a.app.Stop()
	case tcell.KeyLeft:
		b.MoveCursor(-1)
		a.refreshContent(b)
	case tcell.KeyRight:
		b.MoveCursor(1)
		a.refreshContent(b)
	case tcell.KeyUp:
		a.moveCursorUp(b)
		a.refreshContent(b)
	case tcell.KeyDown:
		a.moveCursorDown(b)
		a.refreshContent(b)
	case tcell.KeyRune:
		switch event.Rune() {
		case 'i':
			a.editor.SetMode(state.Insert)
			a.refreshContent(b)
		case 'h':
			b.MoveCursor(-1)
			a.refreshContent(b)
		case 'l':
			b.MoveCursor(1)
			a.refreshContent(b)
		case 'k':
			a.moveCursorUp(b)
			a.refreshContent(b)
		case 'j':
			a.moveCursorDown(b)
			a.refreshContent(b)
		}
	}
	return nil
}

func (a *Athena) handleInsertMode(event *tcell.EventKey, b *buffer.Buffer) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyLeft:
		b.MoveCursor(-1)
		a.refreshContent(b)
	case tcell.KeyRight:
		b.MoveCursor(1)
		a.refreshContent(b)
	case tcell.KeyUp:
		a.moveCursorUp(b)
		a.refreshContent(b)
	case tcell.KeyDown:
		a.moveCursorDown(b)
		a.refreshContent(b)
	case tcell.KeyEscape:
		a.editor.SetMode(state.Normal)
		a.refreshContent(b)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		b.MoveCursor(-1)
		_ = b.Delete(1)
		a.refreshContent(b)
	case tcell.KeyEnter:
		_ = b.Insert("\n")
		a.refreshContent(b)
	case tcell.KeyRune:
		if event.Rune() != 0 {
			_ = b.Insert(string(event.Rune()))
			a.refreshContent(b)
		}
	}
	return nil
}

// refreshContent updates the display components with the current buffer state.
// It uses a string builder for efficient string concatenation.
func (a *Athena) refreshContent(b *buffer.Buffer) {
	line, col := b.GetCursorPosition()
	lines := b.GetLines()
	total := len(lines)

	if line >= total {
		line = total - 1
	}

	if col > len([]rune(lines[line])) {
		col = len([]rune(lines[line]))
	}

	// Update gutters (line numbers)
	var gutterBuilder strings.Builder
	gutterBuilder.Grow(total * (a.config.gutterWidth + 1)) // pre-allocate space

	for i := range lines {
		if i == line {
			fmt.Fprintf(&gutterBuilder, "[white]%*d\n", a.config.gutterWidth-1, i+1)
		} else {
			fmt.Fprintf(&gutterBuilder, "[purple]%*d\n", a.config.gutterWidth-1, i+1)
		}
	}

	// Add placeholder lines for empty space
	_, _, _, height := a.display.document.GetInnerRect()
	for i := 0; i < height; i++ {
		fmt.Fprintf(&gutterBuilder, "[purple]%*s\n", a.config.gutterWidth-1, "~")
	}

	a.display.gutters.SetText(gutterBuilder.String())

	// Update status bar
	status := fmt.Sprintf(" %d%% %d:%d %d ",
		util.CalcProgress(total, line+1),
		line+1, col+1, total)
	a.display.statusBar.SetText(status)

	// Update document with cursor
	var documentBuilder strings.Builder
	documentBuilder.Grow(total * 80) // Estimate average line length

	for i, l := range lines {
		if i > 0 {
			documentBuilder.WriteByte('\n')
		}

		if i == line {
			runes := []rune(l)
			if col > len(runes) {
				col = len(runes)
			}
			if col == len(runes) {
				documentBuilder.WriteString(string(runes))
				documentBuilder.WriteString("[::r] [-:-:-]")
			} else {
				documentBuilder.WriteString(string(runes[:col]))
				documentBuilder.WriteString("[::r]")
				documentBuilder.WriteRune(runes[col])
				documentBuilder.WriteString("[-:-:-]")
				documentBuilder.WriteString(string(runes[col+1:]))
			}
		} else {
			documentBuilder.WriteString(l)
		}
	}

	a.display.document.SetText(documentBuilder.String())
}

// moveCursorUp moves the cursor up one line, maintaining column position if possible.
func (a *Athena) moveCursorUp(b *buffer.Buffer) {
	line, col := b.GetCursorPosition()
	if line <= 0 {
		return
	}

	lines := b.GetLines()
	if line-1 >= len(lines) {
		return
	}

	prevLineLen := len([]rune(lines[line-1]))
	newCol := min(col, prevLineLen)

	// calc new cursor index
	curr := b.GetCursorIndex()
	currLineLen := len([]rune(lines[line]))
	newPos := curr - currLineLen - 1 - (prevLineLen - newCol)

	if newPos < 0 {
		newPos = 0 // Clamp to beginning
	}

	b.SetCursor(newPos)
}

// moveCursorDown moves the cursor down one line, maintaining column position if possible.
func (a *Athena) moveCursorDown(b *buffer.Buffer) {
	line, col := b.GetCursorPosition()
	lines := b.GetLines()
	if line >= len(lines)-1 {
		return
	}

	// Move to next line
	nextLineLen := len([]rune(lines[line+1]))
	newCol := min(col, nextLineLen)

	// Calculate new cursor position
	curr := b.GetCursorIndex()
	currLineLen := len([]rune(lines[line]))
	newPos := curr + currLineLen + 1 + newCol

	bufferLen := len([]rune(b.GetText()))
	if newPos > bufferLen {
		newPos = bufferLen // Clamp to end
	}

	b.SetCursor(newPos)
}
