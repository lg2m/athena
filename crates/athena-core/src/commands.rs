use crate::{state::Mode, Direction, Granularity};

#[derive(Debug, PartialEq)]
pub enum Command {
    Quit,
    InsertChar(char),
    InsertNewLine,
    DeleteChar,
    MoveCursor(Direction, Granularity),
    SaveFile,
    UpdateMode(Mode),
    // TODO: figure out how to make this nicer
    Append,
    AppendBelow,
    AppendAbove,
    AppendEnd,
    AppendStart,
}
