use anyhow::Result;
use std::io::Stdout;

use athena_core::{state::AppEvent, State};

pub mod editor;
pub mod status_bar;

pub trait View: Send {
    fn render(&mut self, stdout: &mut Stdout, state: &State) -> Result<()>;
    fn handle_event(&mut self, event: &AppEvent, state: &State) -> Result<()>;
    fn is_dirty(&self) -> bool {
        true
    }
    fn mark_clean(&mut self);
}
