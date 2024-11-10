package athena

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lg2m/athena/internal/athena/config"
	"github.com/lg2m/athena/internal/editor"
	"github.com/lg2m/athena/internal/ui"
)

// Athena represents the main application.
type Athena struct {
	screen tcell.Screen
	cfg    *config.Config
	editor *editor.Editor
	views  struct {
		gutters   *ui.GuttersView
		document  *ui.DocumentView
		statusBar *ui.StatusBarView
	}
	viewport *ui.Viewport // Shared viewport for synchronized scrolling
}

// NewAthena creates an instance of the athena text-editor.
func NewAthena(cfg *config.Config, filePath string) (*Athena, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize screen: %w", err)
	}

	a := &Athena{
		screen:   screen,
		cfg:      cfg,
		editor:   editor.NewEditor(),
		viewport: ui.NewViewport(cfg.Editor.ScrollPadding),
	}

	if err := a.editor.OpenFile(filePath); err != nil {
		return nil, fmt.Errorf("failed to load file: %w", err)
	}

	a.initializeViews()

	return a, nil
}

// Run starts the Athena application.
func (a *Athena) Run() error {
	defer a.screen.Fini()

	for {
		a.draw()
		a.screen.Show()

		ev := a.screen.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC {
				return nil
			}
		case *tcell.EventResize:
			a.screen.Sync()
			a.resizeViews()
		}
	}
}

func (a *Athena) initializeViews() {
	a.views.gutters = ui.NewGuttersView(a.editor, a.cfg, a.viewport)
	a.views.document = ui.NewDocumentView(a.editor, a.cfg, a.viewport)
	a.views.statusBar = ui.NewStatusBarView(a.editor, &a.cfg.Editor)
	a.resizeViews()
}

func (a *Athena) draw() {
	a.screen.Clear()

	a.views.gutters.Draw(a.screen)
	a.views.document.Draw(a.screen)
	a.views.statusBar.Draw(a.screen)
}

func (a *Athena) resizeViews() {
	width, height := a.screen.Size()
	a.views.gutters.Resize(0, 0, 6, height-1)
	a.views.document.Resize(6, 0, width-6, height-1)
	a.views.statusBar.Resize(0, height-1, width, 1)
}
