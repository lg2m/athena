package treesitter

import "github.com/gdamore/tcell/v2"

var (
	ColorRed        = tcell.NewHexColor(0xf7768e)
	ColorOrange     = tcell.NewHexColor(0xff9e64)
	ColorYellow     = tcell.NewHexColor(0xe0af68)
	ColorLightGreen = tcell.NewHexColor(0x9ece6a)
	ColorGreen      = tcell.NewHexColor(0x73daca)
	ColorAqua       = tcell.NewHexColor(0x2ac3de)
	ColorTeal       = tcell.NewHexColor(0x1abc9c)
	ColorTurquoise  = tcell.NewHexColor(0x89ddff)
	ColorLightCyan  = tcell.NewHexColor(0xb4f9f8)
	ColorCyan       = tcell.NewHexColor(0x7dcfff)
	ColorBlue       = tcell.NewHexColor(0x7aa2f7)
	ColorPurple     = tcell.NewHexColor(0x9d7cd8)
	ColorMagenta    = tcell.NewHexColor(0xbb9af7)
	ColorComment    = tcell.NewHexColor(0x565f89)
	ColorBlack      = tcell.NewHexColor(0x414868)

	ColorFg          = tcell.NewHexColor(0xc0caf5)
	ColorFgDark      = tcell.NewHexColor(0xa9b1d6)
	ColorFgGutter    = tcell.NewHexColor(0x3b4261)
	ColorBg          = tcell.NewHexColor(0x1a1b26)
	ColorBgSelection = tcell.NewHexColor(0x283457)
)

// Universal styles that work well across languages.
var DefaultStyles = StyleMap{
	// Keywords
	"keyword":          tcell.StyleDefault.Foreground(ColorPurple).Italic(true),
	"keyword.control":  tcell.StyleDefault.Foreground(ColorMagenta),
	"keyword.operator": tcell.StyleDefault.Foreground(ColorMagenta),

	// Types
	"type":              tcell.StyleDefault.Foreground(ColorAqua),
	"type.builtin":      tcell.StyleDefault.Foreground(ColorAqua).Bold(true),
	"type.enum.variant": tcell.StyleDefault.Foreground(ColorOrange),

	// Functions
	"function": tcell.StyleDefault.Foreground(ColorBlue).Italic(true),
	// "function.name":    tcell.StyleDefault.Foreground(ColorBlue).Bold(true),
	"function.builtin": tcell.StyleDefault.Foreground(ColorAqua),
	"function.macro":   tcell.StyleDefault.Foreground(ColorCyan),
	"function.special": tcell.StyleDefault.Foreground(ColorCyan),

	// Operators and Punctuation
	"operator":    tcell.StyleDefault.Foreground(ColorTurquoise),
	"punctuation": tcell.StyleDefault.Foreground(ColorTurquoise),

	// Variables and Constants
	"variable":                  tcell.StyleDefault.Foreground(ColorFg),
	"variable.builtin":          tcell.StyleDefault.Foreground(ColorRed),
	"constant":                  tcell.StyleDefault.Foreground(ColorOrange),
	"constant.builtin":          tcell.StyleDefault.Foreground(ColorAqua),
	"constant.character.escape": tcell.StyleDefault.Foreground(ColorMagenta),

	// Literals
	"string":         tcell.StyleDefault.Foreground(ColorLightGreen),
	"string.special": tcell.StyleDefault.Foreground(ColorAqua),
	"string.regexp":  tcell.StyleDefault.Foreground(ColorLightCyan),

	// Comments and Documentation
	"comment":     tcell.StyleDefault.Foreground(ColorComment).Italic(true),
	"comment.doc": tcell.StyleDefault.Foreground(ColorYellow),

	// Special
	"label":     tcell.StyleDefault.Foreground(ColorBlue),
	"tag":       tcell.StyleDefault.Foreground(ColorMagenta),
	"attribute": tcell.StyleDefault.Foreground(ColorCyan),
	"namespace": tcell.StyleDefault.Foreground(ColorCyan).Bold(true),

	// Diagnostics
	"error":   tcell.StyleDefault.Foreground(ColorRed).Bold(true),
	"warning": tcell.StyleDefault.Foreground(ColorYellow),
	"info":    tcell.StyleDefault.Foreground(ColorCyan),
	"hint":    tcell.StyleDefault.Foreground(ColorTeal),

	// "function.name":   tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true),
	// "function.method": tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true),

	// "number":  tcell.StyleDefault.Foreground(tcell.ColorRed),
	// "boolean": tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true),
}
