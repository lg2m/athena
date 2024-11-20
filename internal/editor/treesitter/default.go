package treesitter

import "github.com/gdamore/tcell/v2"

// Universal styles that work well across languages.
var DefaultStyles = StyleMap{
	"function":      tcell.StyleDefault.Foreground(tcell.ColorGreen),
	"function.name": tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true),
	"variable":      tcell.StyleDefault.Foreground(tcell.ColorYellow),
	"string":        tcell.StyleDefault.Foreground(tcell.ColorRed),
	"keyword":       tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true),
	"comment":       tcell.StyleDefault.Foreground(tcell.ColorGray),
}
