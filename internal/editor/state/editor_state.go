package state

type EditorState struct {
	mode         Mode
	ActiveBuffer string
	Dirty        bool
	Message      string
	Error        string
}

func NewEditorState() *EditorState {
	return &EditorState{
		mode: Normal,
	}
}

// SetMode sets the editor states mode (normal, insert).
func (s *EditorState) SetMode(mode Mode) {
	s.mode = mode
}

// GetMode retrieves the current editor states mode.
func (s *EditorState) GetMode() Mode {
	return s.mode
}
