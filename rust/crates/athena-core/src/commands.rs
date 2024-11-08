use crate::{Direction, Granularity, Mode};

// TODO: determine if we separate this into key commands and prompt commands
// e.g., jkhl, i, a, shift+i, etc. and :save, :quit, etc.

/// Editor commands executed by key presses or explicitly in command prompt.
#[derive(Debug, PartialEq)]
pub enum EditorCommand {
    Quit,
    InsertChar(char),
    Backspace,
    Enter,
    UpdateMode(Mode),
    Append,
    AppendBelow,
    AppendAbove,
    AppendEnd,
    AppendStart,

    InsertNewLine,
    DeleteChar,
    MoveCursor(Direction, Granularity),
    SaveFile,
    // TODO: figure out how to make this nicer
}
