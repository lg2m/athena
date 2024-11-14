package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/lg2m/athena/internal/athena/config"
	"github.com/lg2m/athena/internal/editor"
	"github.com/lg2m/athena/internal/editor/state"
	"github.com/lg2m/athena/internal/util"
)

// statusBarMaxLengths holds the maximum lengths for each section.
type statusBarMaxLengths struct {
	left   int
	center int
	right  int
}

// StatusBarView represents the status bar.
type StatusBarView struct {
	BaseView
	editor *editor.Editor
	cfg    *config.EditorConfig

	style      tcell.Style
	left       string
	center     string
	right      string
	truncated  bool
	maxLengths statusBarMaxLengths
}

func NewStatusBarView(e *editor.Editor, cfg *config.EditorConfig) *StatusBarView {
	style := tcell.StyleDefault.Background(tcell.ColorDarkSlateGray).Foreground(tcell.ColorWhite)
	return &StatusBarView{
		editor: e,
		cfg:    cfg,
		style:  style,
	}
}

func (v *StatusBarView) Draw(screen tcell.Screen) {
	v.buildStatusSections()
	v.handleOverflow()
	v.render(screen)
}

// buildStatusSections constructs the left, center, and right sections.
func (v *StatusBarView) buildStatusSections() {
	v.left = v.buildSection(v.cfg.StatusBar.Left)
	v.center = v.buildSection(v.cfg.StatusBar.Center)
	v.right = v.buildSection(v.cfg.StatusBar.Right)
}

// buildSection builds a single section based on the provided options.
func (v *StatusBarView) buildSection(options []config.StatusBarOption) string {
	var builder strings.Builder
	for _, opt := range options {
		builder.WriteString(v.getOptionString(opt))
	}
	return builder.String()
}

// getOptionString returns the string representation for a given status bar option.
func (v *StatusBarView) getOptionString(opt config.StatusBarOption) string {
	switch opt {
	case config.SectionMode:
		switch v.editor.GetMode() {
		case state.Normal:
			return fmt.Sprintf(" %s ", v.cfg.StatusBar.Mode.Normal)
		case state.Insert:
			return fmt.Sprintf(" %s ", v.cfg.StatusBar.Mode.Insert)
		default:
			return " UNK "
		}
	case config.SectionFileName:
		if fileName, err := v.editor.FileName(); err == nil && fileName != "" {
			return fmt.Sprintf(" %s ", fileName)
		}
	case config.SectionFileAbsPath:
		if filePath, err := v.editor.FilePath(); err == nil && filePath != "" {
			return fmt.Sprintf(" %s ", filePath)
		}
	// case config.SectionFileModified:
	// case config.SectionFileEncoding:
	case config.SectionFileType:
		if ext, err := v.editor.FileType(); err == nil && ext != "" {
			return fmt.Sprintf(" %s ", ext)
		}
	// case config.SectionVersionControl:
	case config.SectionCursorPos:
		currLine, currCol, _ := v.editor.GetCurrentPosition()
		return fmt.Sprintf(" %d:%d ", currLine+1, currCol+1)
	case config.SectionLineCount:
		total, _ := v.editor.GetLineCount()
		return fmt.Sprintf(" %d ", total)
	case config.SectionCursorPercentage:
		total, _ := v.editor.GetLineCount()
		currLine, _, _ := v.editor.GetCurrentPosition()
		scrollPercent := util.CalcProgress(total, currLine+1)
		return fmt.Sprintf(" %d%% ", scrollPercent)
	case config.SectionSpacer:
		return " "
	default:
		return ""
	}
	return ""
}

// handleOverflow manages the truncation of sections if the total length exceeds available width.
func (v *StatusBarView) handleOverflow() {
	totalLen := len(v.left) + len(v.center) + len(v.right)
	availableWidth := v.width

	if totalLen <= availableWidth {
		v.maxLengths = statusBarMaxLengths{
			left:   len(v.left),
			center: len(v.center),
			right:  len(v.right),
		}
		return
	}

	overflow := totalLen - availableWidth
	v.truncated = true

	// Prioritize truncating the center section first
	v.center, overflow = truncateString(v.center, overflow)
	if overflow > 0 {
		v.left, overflow = truncateString(v.left, overflow)
	}
	if overflow > 0 {
		v.right, _ = truncateString(v.right, overflow)
	}

	v.maxLengths = statusBarMaxLengths{
		left:   len(v.left),
		center: len(v.center),
		right:  len(v.right),
	}
}

// truncateString truncates the input string by the specified overflow amount.
func truncateString(s string, overflow int) (string, int) {
	if len(s) > overflow {
		return s[:len(s)-overflow], 0
	}
	return "", overflow - len(s)
}

// render outputs the status bar sections to the screen.
func (v *StatusBarView) render(screen tcell.Screen) {
	// Clear the status bar area
	for x := v.x; x < v.x+v.width; x++ {
		screen.SetContent(x, v.y, ' ', nil, v.style)
	}

	// Calculate positions
	leftX := v.x
	rightX := v.x + v.width - v.maxLengths.right
	centerX := v.x + (v.width-v.maxLengths.center)/2

	// Render each section
	v.renderString(screen, v.left, leftX)
	v.renderString(screen, v.center, centerX)
	v.renderString(screen, v.right, rightX)
}

// renderString draws a string on the screen starting at the specified x position.
func (v *StatusBarView) renderString(screen tcell.Screen, s string, startX int) {
	for i, ch := range s {
		xPos := startX + i
		if xPos >= v.x+v.width {
			break
		}
		screen.SetContent(xPos, v.y, ch, nil, v.style)
	}
}
