use anyhow::Result;

use athena_core::{state::EditorEvent, EditorState};

use crate::terminal::Terminal;

pub mod document;
pub mod status_bar;

pub trait View: Send {
    fn render(&mut self, terminal: &mut Terminal, state: &EditorState) -> Result<()>;
    fn handle_event(&mut self, event: &EditorEvent, state: &EditorState) -> Result<()>;
    fn is_dirty(&self) -> bool {
        true
    }
    fn mark_clean(&mut self);
}
