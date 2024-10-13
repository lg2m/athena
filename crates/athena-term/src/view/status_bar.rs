use anyhow::Result;
use crossterm::{
    cursor,
    style::{Color, Print, SetBackgroundColor, SetForegroundColor},
    terminal::{self, Clear, ClearType},
    QueueableCommand,
};
use std::{
    collections::HashMap,
    fmt::Display,
    io::{Stdout, Write},
};

use athena_core::{
    state::{coords_at_pos, AppEvent},
    Mode, State,
};

use super::View;

pub type StatusBarConfig = HashMap<Section, Vec<String>>;

#[derive(Debug, PartialEq, Eq, Hash)]
pub enum Section {
    Left,
    Center,
    Right,
}

#[derive(Debug, Clone, PartialEq)]
pub enum StatusItemKind {
    Mode(Mode),                   // e.g., insert, normal
    CursorPosition(usize, usize), // row, col
    LineCount(usize),             // lines in file
    FileName(String),
    FileEncoding(String),
    Language(String), // programming language or file type
}

impl Display for StatusItemKind {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            StatusItemKind::Mode(mode) => write!(
                f,
                "{}",
                match mode {
                    Mode::Normal => "NOR".to_string(),
                    Mode::Insert => "INS".to_string(),
                }
            ),
            StatusItemKind::CursorPosition(row, col) => write!(f, "{}:{}", row, col),
            StatusItemKind::LineCount(count) => write!(f, "{}", count),
            StatusItemKind::FileName(name) => write!(f, "{}", name),
            StatusItemKind::FileEncoding(encoding) => write!(f, "{}", encoding),
            StatusItemKind::Language(lang) => write!(f, "{}", lang),
        }
    }
}

#[derive(Debug, Clone)]
pub struct StatusItem {
    kind: StatusItemKind,
}

impl StatusItem {
    pub fn new(kind: StatusItemKind) -> Self {
        Self { kind }
    }
}

pub struct StatusBar {
    config: StatusBarConfig, // left, center, right - and their item names
    items: HashMap<String, StatusItem>,
    dirty: bool,
}

impl StatusBar {
    pub fn new(config: StatusBarConfig) -> Self {
        Self {
            config,
            items: HashMap::new(),
            dirty: false,
        }
    }

    pub fn with_default(mut self) -> Self {
        self.update_item("mode", StatusItemKind::Mode(Mode::Normal));
        self.update_item("position", StatusItemKind::CursorPosition(1, 1));
        self.update_item("total-line-count", StatusItemKind::LineCount(1));
        self.dirty = true;
        self
    }

    pub fn update_item(&mut self, key: &str, item: StatusItemKind) {
        self.items.insert(key.to_string(), StatusItem::new(item));
        self.dirty = true;
    }

    fn build_section(&self, section: Section) -> String {
        if let Some(items) = self.config.get(&section) {
            items
                .iter()
                .filter_map(|k| self.items.get(k))
                .map(|i| i.kind.to_string())
                .collect::<Vec<_>>()
                .join(" | ")
        } else {
            "".to_string()
        }
    }

    fn build(&self, width: usize) -> String {
        let left = self.build_section(Section::Left);
        let center = self.build_section(Section::Center);
        let right = self.build_section(Section::Right);

        let avail_width = width.saturating_sub(left.len() + right.len());
        let left_padding = avail_width / 2;
        let right_padding = avail_width - left_padding;

        format!(
            "{:<width_left$}{:^width_center$}{:>width_right$}",
            left,
            center,
            right,
            width_left = left.len() + left_padding,
            width_center = center.len(),
            width_right = right.len() + right_padding,
        )
    }
}

impl View for StatusBar {
    fn render(&mut self, stdout: &mut Stdout, _state: &State) -> Result<()> {
        let size = terminal::size()?;
        let content = self.build(size.0 as usize);

        stdout
            .queue(cursor::Hide)?
            .queue(cursor::MoveTo(0, size.1))? // Move to the start of the status bar position
            .queue(Clear(ClearType::CurrentLine))? // Clear current line
            .queue(SetForegroundColor(Color::White))?
            .queue(SetBackgroundColor(Color::Black))?
            .queue(Print(content))?
            .queue(SetForegroundColor(Color::Reset))?
            .queue(SetBackgroundColor(Color::Reset))?
            .queue(cursor::Show)?
            .flush()?;

        Ok(())
    }

    fn handle_event(&mut self, event: &AppEvent, state: &State) -> Result<()> {
        match event {
            AppEvent::CursorMoved(row, col) => {
                self.update_item(
                    "position",
                    StatusItemKind::CursorPosition(*row + 1, *col + 1),
                );
            }
            AppEvent::ModeChanged(mode) => {
                self.update_item("mode", StatusItemKind::Mode(*mode));
            }
            AppEvent::BufferChanged => {
                let line_count = state.buffer.len_lines();
                self.update_item("total-line-count", StatusItemKind::LineCount(line_count));

                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.update_item(
                    "position",
                    StatusItemKind::CursorPosition(coords.0 + 1, coords.1 + 1),
                );
            }
            _ => {}
        }
        Ok(())
    }

    fn is_dirty(&self) -> bool {
        self.dirty
    }

    fn mark_clean(&mut self) {
        self.dirty = false;
    }
}

pub fn default_status_bar_config() -> StatusBarConfig {
    let mut config = HashMap::new();
    config.insert(Section::Left, vec!["mode".to_string()]);
    config.insert(
        Section::Right,
        vec!["position".to_string(), "total-line-count".to_string()],
    );
    config
}
