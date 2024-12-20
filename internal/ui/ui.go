package ui

// import (
// 	"fmt"
// 	"path/filepath"
// 	"strings"
// 	"sync"

// 	"github.com/gdamore/tcell/v2"
// 	"github.com/lg2m/athena/internal/editor"
// 	"github.com/lg2m/athena/internal/editor/state"
// 	"github.com/lg2m/athena/internal/util"
// 	"github.com/rivo/tview"
// )

// // Option defines a functional option for configuring Athena.
// type Option = func(*Athena)

// // Athena represents the main UI component of the editor.
// type Athena struct {
// 	app     *tview.Application
// 	editor  *editor.Editor
// 	layout  *tview.Flex
// 	display struct {
// 		gutters   *tview.TextView
// 		document  *tview.TextView
// 		statusBar *tview.TextView
// 	}
// 	config struct {
// 		gutterWidth int
// 		filePath    string
// 	}
// 	viewport struct {
// 		offset  int // number of lines scrolled from top
// 		height  int // visible height of the document view
// 		padding int // number of lines to keep visible above/below cursor
// 	}
// 	mu sync.RWMutex
// }

// // NewAthena creates and configures a new Athena editor instance.
// func NewAthena(opts ...Option) (*Athena, error) {
// 	a := &Athena{
// 		app:    tview.NewApplication(),
// 		editor: editor.NewEditor(),
// 	}
// 	a.config.gutterWidth = 6
// 	a.viewport.padding = 5

// 	a.initializeComponents()

// 	for _, opt := range opts {
// 		opt(a)
// 	}

// 	if a.config.filePath != "" {
// 		if err := a.LoadFile(a.config.filePath); err != nil {
// 			a.showError(err)
// 			return nil, err
// 		}
// 	}

// 	return a, nil
// }

// // LoadFile loads a file into the editor and refreshes the display.
// func (a *Athena) LoadFile(filePath string) error {
// 	if err := a.editor.OpenFile(filePath); err != nil {
// 		return fmt.Errorf("failed to load file: %w", err)
// 	}

// 	a.refreshContent()
// 	return nil
// }

// // WithFilePath sets the initial file to be loaded into the editor.
// func WithFilePath(filePath string) Option {
// 	return func(a *Athena) {
// 		absPath, err := filepath.Abs(filePath)
// 		if err != nil {
// 			return
// 		}
// 		a.config.filePath = absPath
// 	}
// }

// // WithGutterWidth sets a custom gutter width.
// func WithGutterWidth(width int) Option {
// 	return func(a *Athena) {
// 		a.config.gutterWidth = width
// 	}
// }

// // initializeComponents sets up all UI components with their configurations.
// func (a *Athena) initializeComponents() {
// 	a.display.gutters = a.newGutters()
// 	a.display.document = a.newDocument()
// 	a.display.statusBar = a.newStatusBar()
// 	a.layout = a.newLayout()
// }

// func (a *Athena) newGutters() *tview.TextView {
// 	gutters := tview.NewTextView()
// 	return gutters.
// 		SetDynamicColors(true).
// 		SetRegions(true).
// 		SetWrap(false)
// }

// func (a *Athena) newDocument() *tview.TextView {
// 	doc := tview.NewTextView().
// 		SetWrap(false).
// 		SetWordWrap(false).
// 		SetDynamicColors(true).
// 		SetRegions(true)
// 	doc.
// 		SetTitle("Athena Editor").
// 		SetInputCapture(a.handleKeyPress)
// 	return doc
// }

// func (a *Athena) newStatusBar() *tview.TextView {
// 	statusBar := tview.NewTextView().
// 		SetDynamicColors(true).
// 		SetRegions(true).
// 		SetWrap(false)
// 	statusBar.SetBackgroundColor(tcell.ColorNavy)
// 	return statusBar
// }

// func (a *Athena) newLayout() *tview.Flex {
// 	return tview.NewFlex().
// 		SetDirection(tview.FlexRow).
// 		AddItem(
// 			tview.NewFlex().
// 				AddItem(a.display.gutters, a.config.gutterWidth, 0, false).
// 				AddItem(a.display.document, 0, 1, true),
// 			0, 1, true).
// 		AddItem(a.display.statusBar, 1, 0, false)
// }

// // Run starts the Athena application.
// func (a *Athena) Run() error {
// 	if err := a.app.SetRoot(a.layout, true).EnableMouse(true).Run(); err != nil {
// 		return fmt.Errorf("application error: %w", err)
// 	}
// 	return nil
// }

// // handleKeyPress processes keyboard input events.
// // It uses a read lock for operations that don't modify state.
// func (a *Athena) handleKeyPress(event *tcell.EventKey) *tcell.EventKey {
// 	a.mu.Lock()
// 	defer a.mu.Unlock()

// 	switch a.editor.GetMode() {
// 	case state.Normal:
// 		return a.handleNormalMode(event)
// 	case state.Insert:
// 		return a.handleInsertMode(event)
// 	default:
// 		return event
// 	}
// }

// func (a *Athena) handleNormalMode(event *tcell.EventKey) *tcell.EventKey {
// 	switch event.Key() {
// 	case tcell.KeyCtrlC:
// 		if err := a.editor.CloseCurrentBuffer(); err != nil {
// 			a.showError(err)
// 			return nil
// 		}
// 		a.app.Stop()
// 	case tcell.KeyCtrlS:
// 		if err := a.editor.SaveCurrentBuffer(); err != nil {
// 			a.showError(err)
// 		}
// 	case tcell.KeyLeft:
// 		if err := a.editor.MoveCursorHorizontal(-1, false); err != nil {
// 			a.showError(err)
// 		}
// 	case tcell.KeyRight:
// 		if err := a.editor.MoveCursorHorizontal(1, false); err != nil {
// 			a.showError(err)
// 		}
// 	case tcell.KeyUp:
// 		if err := a.editor.MoveCursorVertical(-1, false); err != nil {
// 			a.showError(err)
// 		}
// 	case tcell.KeyDown:
// 		if err := a.editor.MoveCursorVertical(1, false); err != nil {
// 			a.showError(err)
// 		}
// 	case tcell.KeyRune:
// 		switch event.Rune() {
// 		case 'd':
// 			if err := a.editor.DeleteSelection(); err != nil {
// 				a.showError(err)
// 			}
// 		case 'i':
// 			a.editor.SetMode(state.Insert)
// 			a.showError(fmt.Errorf("Testing"))
// 		case 'h':
// 			if err := a.editor.MoveCursorHorizontal(-1, false); err != nil {
// 				a.showError(err)
// 			}
// 		case 'l':
// 			if err := a.editor.MoveCursorHorizontal(1, false); err != nil {
// 				a.showError(err)
// 			}
// 		case 'k':
// 			if err := a.editor.MoveCursorVertical(-1, false); err != nil {
// 				a.showError(err)
// 			}
// 		case 'j':
// 			if err := a.editor.MoveCursorVertical(1, false); err != nil {
// 				a.showError(err)
// 			}
// 		}
// 	}
// 	a.refreshContent()
// 	return nil
// }

// func (a *Athena) handleInsertMode(event *tcell.EventKey) *tcell.EventKey {
// 	switch event.Key() {
// 	case tcell.KeyEscape:
// 		a.editor.SetMode(state.Normal)
// 	case tcell.KeyBackspace, tcell.KeyBackspace2:
// 		if err := a.editor.DeleteText(-1); err != nil {
// 			a.showError(err)
// 		}
// 	case tcell.KeyDelete:
// 		if err := a.editor.DeleteText(1); err != nil {
// 			a.showError(err)
// 		}
// 	case tcell.KeyEnter:
// 		if err := a.editor.InsertText("\n"); err != nil {
// 			a.showError(err)
// 		}
// 	case tcell.KeyLeft:
// 		if err := a.editor.MoveCursorHorizontal(-1, false); err != nil {
// 			a.showError(err)
// 		}
// 	case tcell.KeyRight:
// 		if err := a.editor.MoveCursorHorizontal(1, false); err != nil {
// 			a.showError(err)
// 		}
// 	case tcell.KeyUp:
// 		if err := a.editor.MoveCursorVertical(-1, false); err != nil {
// 			a.showError(err)
// 		}
// 	case tcell.KeyDown:
// 		if err := a.editor.MoveCursorVertical(1, false); err != nil {
// 			a.showError(err)
// 		}
// 	case tcell.KeyRune:
// 		if err := a.editor.InsertText(string(event.Rune())); err != nil {
// 			a.showError(err)
// 		}
// 	}
// 	a.refreshContent()
// 	return nil
// }

// // refreshContent updates the display components with the current buffer state.
// // It uses a string builder for efficient string concatenation.
// func (a *Athena) refreshContent() {
// 	currLine, currCol, err := a.editor.GetCurrentPosition()
// 	if err != nil {
// 		return
// 	}

// 	total, err := a.editor.GetLineCount()
// 	if err != nil {
// 		return
// 	}

// 	a.updateViewport()

// 	// calc visible range
// 	visibleStart := a.viewport.offset
// 	visibleEnd := min(total, a.viewport.offset+a.viewport.height)

// 	// Update gutters (line numbers)
// 	var gutterBuilder strings.Builder
// 	gutterBuilder.Grow(a.viewport.height * (a.config.gutterWidth + 1)) // pre-allocate space

// 	for i := visibleStart; i < visibleEnd; i++ {
// 		if i == currLine {
// 			fmt.Fprintf(&gutterBuilder, "[white]%*d\n", a.config.gutterWidth-1, i+1)
// 		} else if i == total-1 {
// 			fmt.Fprintf(&gutterBuilder, "[purple]%*s\n", a.config.gutterWidth-1, "~")
// 		} else {
// 			fmt.Fprintf(&gutterBuilder, "[purple]%*d\n", a.config.gutterWidth-1, i+1)
// 		}
// 	}

// 	a.display.gutters.SetText(gutterBuilder.String())

// 	// Update status bar with scroll information
// 	mode := a.editor.GetMode().String()
// 	scrollPercent := util.CalcProgress(total, currLine+1)

// 	status := fmt.Sprintf(" %s | %d%% %d:%d %d ",
// 		mode,
// 		scrollPercent,
// 		currLine+1, currCol+1,
// 		total)

// 	a.display.statusBar.SetText(status)

// 	// Update document content with cursor
// 	var documentBuilder strings.Builder
// 	documentBuilder.Grow(a.viewport.height * 80)

// 	for i := visibleStart; i < visibleEnd; i++ {
// 		if i > visibleStart {
// 			documentBuilder.WriteByte('\n')
// 		}

// 		l, err := a.editor.GetLine(i)
// 		if err != nil {
// 			a.showError(err)
// 			continue
// 		}

// 		if i == currLine {
// 			runes := []rune(l)
// 			if currCol > len(runes) {
// 				currCol = len(runes)
// 			}
// 			if currCol == len(runes) {
// 				documentBuilder.WriteString(string(runes))
// 				documentBuilder.WriteString("[::r] [-:-:-]")
// 			} else {
// 				documentBuilder.WriteString(string(runes[:currCol]))
// 				documentBuilder.WriteString("[::r]")
// 				documentBuilder.WriteRune(runes[currCol])
// 				documentBuilder.WriteString("[-:-:-]")
// 				documentBuilder.WriteString(string(runes[currCol+1:]))
// 			}
// 		} else {
// 			documentBuilder.WriteString(l)
// 		}
// 	}

// 	a.display.document.SetText(documentBuilder.String())
// }

// // updateViewport updates the viewport to ensure the cursor is always visible.
// func (a *Athena) updateViewport() {
// 	_, _, _, height := a.display.document.GetInnerRect()
// 	a.viewport.height = height

// 	currLine, _, err := a.editor.GetCurrentPosition()
// 	if err != nil {
// 		return
// 	}

// 	// adjust viewport if cursor moves too close to the edges
// 	if currLine-a.viewport.offset < a.viewport.padding {
// 		// cursor too close to top
// 		a.viewport.offset = max(0, currLine-a.viewport.padding)
// 	} else if currLine-a.viewport.offset > a.viewport.height-a.viewport.padding {
// 		// cursor too close to bttom
// 		a.viewport.offset = currLine - (a.viewport.height - a.viewport.padding)
// 	}
// }

// func (a *Athena) showError(err error) {
// 	if err != nil {
// 		a.display.statusBar.SetText(fmt.Sprintf(" [red]Error: %v[-:-:-]", err))
// 	}
// }
