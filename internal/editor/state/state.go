package state

// EditorMode represents the editor mode.
//
//	e.g., Insert and Normal.
type EditorMode uint8

const (
	Normal EditorMode = iota
	Insert
)

// Selection represents the cursor and the text being selected.
type Selection struct {
	Start int
	End   int
}
