use crate::{state::Mode, Direction, Granularity};

#[derive(Debug, PartialEq)]
pub enum Command {
    Quit,
    InsertChar(char),
    InsertNewLine,
    DeleteChar,
    MoveCursor(Direction, Granularity),
    MoveCursorLeft,
    MoveCursorRight,
    MoveCursorUp,
    MoveCursorDown,
    SaveFile,
    UpdateMode(Mode),
}
