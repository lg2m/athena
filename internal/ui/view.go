package ui

import "github.com/gdamore/tcell/v2"

// View represents a drawable UI component.
type View interface {
	Draw(screen tcell.Screen)
	HandleEvent(event tcell.Event) bool
	Resize(x, y, width, height int)
}

// BaseView provides common functionality for views.
type BaseView struct {
	x, y          int
	width, height int
}

// Resize implements view resizing.
func (v *BaseView) Resize(x, y, width, height int) {
	v.x = x
	v.y = y
	v.width = width
	v.height = height
}
