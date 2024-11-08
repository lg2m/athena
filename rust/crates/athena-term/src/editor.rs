use std::{collections::HashMap, io::Write, sync::Arc};

use anyhow::Result;
use crossterm::event::{self, Event, KeyCode, KeyModifiers};
use tokio::sync::{
    mpsc::{self, Receiver, Sender},
    RwLock,
};

use athena_core::{
    get_config_or_default,
    state::{coords_at_pos, EditorEvent},
    Config, Direction, EditorCommand, EditorState, Granularity, Mode,
};

use crate::{
    terminal::Terminal,
    view::{document::Document, status_bar::StatusBar, View},
};

pub struct Editor {
    pub terminal: Terminal,
    config: Config,
    state: Arc<RwLock<EditorState>>,
    views: HashMap<String, Box<dyn View>>,
    event_sender: Sender<EditorEvent>,
    event_receiver: Receiver<EditorEvent>,
    command_sender: Sender<EditorCommand>,
    command_receiver: Receiver<EditorCommand>,
    dirty: bool,
}

impl Editor {
    #[must_use]
    pub fn new() -> Self {
        let config = get_config_or_default();

        let (event_sender, event_receiver) = mpsc::channel(100);
        let (command_sender, command_receiver) = mpsc::channel(100);

        Self {
            terminal: Terminal::new(),
            config,
            state: Arc::new(RwLock::new(EditorState::new())),
            views: HashMap::new(),
            event_sender,
            event_receiver,
            command_sender,
            command_receiver,
            dirty: true,
        }
    }

    /// we need a config and to setup views by default.
    pub fn with_default(mut self) -> Self {
        self.add_view(
            "text_editor",
            Box::new(Document::new(&self.config.editor.gutters)),
        );
        self.add_view(
            "status_bar",
            Box::new(StatusBar::new(&self.config.editor.status_bar)),
        );
        self
    }

    /// Runs the editor
    pub async fn run(&mut self) -> Result<()> {
        // setup the terminal
        self.terminal.start()?;

        // initial render
        self.render().await?;

        loop {
            tokio::select! {
                Some(command) = self.command_receiver.recv() => {
                    if command == EditorCommand::Quit {
                        break;
                    }
                    self.handle_command(command).await?;
                    self.render().await?;
                }
                Some(event) = self.event_receiver.recv() => {
                    self.handle_event(event).await?;
                    self.render().await?;
                }
            }
        }

        self.shutdown()
    }

    /// Shuts down terminal and restores original state along with handling any necessary editor state.
    fn shutdown(&mut self) -> Result<()> {
        self.terminal.stop()?;
        Ok(())
    }

    /// Renders a single frame of the editor.
    async fn render(&mut self) -> Result<()> {
        if !self.is_dirty() {
            return Ok(());
        }

        self.mark_clean();

        self.terminal.hide_cursor()?;

        let state = self.state.read().await;

        for view in self.views.values_mut() {
            if view.is_dirty() {
                view.render(&mut self.terminal, &state)?;
                view.mark_clean();
            }
        }

        let cursor_shape = match state.mode {
            Mode::Insert => "\x1B[6 q", // Block cursor for insert mode
            Mode::Normal => "\x1B[2 q", // Line cursor for normal mode
        };
        self.terminal.stdout.write(cursor_shape.as_bytes())?;

        let pos = state.cursor.index;
        let coords = coords_at_pos(&state.buffer.slice(..), pos);
        self.terminal.goto(coords.1 + 5, coords.0)?;

        // TODO: move cursor to proper location and show cursor
        self.terminal.show_cursor()?;

        self.terminal.flush()?;

        Ok(())
    }

    /// Handles incoming commands from the `event_handler`
    async fn handle_command(&mut self, command: EditorCommand) -> Result<()> {
        // editor will handle receiving all commands from the terminal,
        // it will then fire off events to views or other areas of the app.
        // state should be updated by the editor?? (undecided)
        // if we send an event we need to mark state as dirty, else keep as is.
        let mut state = self.state.write().await;
        match command {
            EditorCommand::InsertChar(ch) => {
                state.insert_char(ch);
                self.event_sender.send(EditorEvent::BufferChanged).await?;
            }
            EditorCommand::Backspace => {
                state.backspace();
                self.event_sender.send(EditorEvent::BufferChanged).await?;
            }
            EditorCommand::Enter => match state.mode {
                Mode::Insert => {
                    state.insert_newline();
                    self.event_sender.send(EditorEvent::BufferChanged).await?;
                }
                _ => (),
            },
            EditorCommand::UpdateMode(mode) => {
                state.update_mode(mode);
                self.event_sender
                    .send(EditorEvent::ModeChanged(mode))
                    .await?;

                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.1, coords.0))
                    .await?;
            }
            EditorCommand::Append => {
                state.append();
                self.event_sender
                    .send(EditorEvent::ModeChanged(Mode::Insert))
                    .await?;

                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.1, coords.0))
                    .await?;
            }
            EditorCommand::AppendBelow => {
                state.insert_newline_below();
                self.event_sender
                    .send(EditorEvent::ModeChanged(Mode::Insert))
                    .await?;

                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.1, coords.0))
                    .await?;
            }
            EditorCommand::AppendAbove => {
                state.insert_newline_above();
                self.event_sender
                    .send(EditorEvent::ModeChanged(Mode::Insert))
                    .await?;

                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.1, coords.0))
                    .await?;
            }
            EditorCommand::AppendEnd => {
                state.append_end_of_line();
                self.event_sender
                    .send(EditorEvent::ModeChanged(Mode::Insert))
                    .await?;

                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.1, coords.0))
                    .await?;
            }
            EditorCommand::AppendStart => {
                state.insert_start_of_line();
                self.event_sender
                    .send(EditorEvent::ModeChanged(Mode::Insert))
                    .await?;

                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.1, coords.0))
                    .await?;
            }
            EditorCommand::MoveCursor(direction, granularity) => {
                state.move_cursor(direction, granularity);
                let pos = state.cursor.index;
                let coords = coords_at_pos(&state.buffer.slice(..), pos);
                self.event_sender
                    .send(EditorEvent::CursorMoved(coords.1, coords.0))
                    .await?;
            }
            _ => (),
        }

        self.dirty = true;

        Ok(())
    }

    /// Handles incoming events sent by `handle_command`
    async fn handle_event(&mut self, event: EditorEvent) -> Result<()> {
        let state = self.state.read().await;
        for view in self.views.values_mut() {
            view.handle_event(&event, &state)?;
        }
        Ok(())
    }

    fn add_view(&mut self, name: &str, view: Box<dyn View>) {
        self.views.insert(name.to_string(), view);
    }

    #[inline]
    /// Checks if state has changed and needs a re-render.
    fn is_dirty(&self) -> bool {
        self.dirty
    }

    #[inline]
    /// Used after rendering changed state.
    fn mark_clean(&mut self) {
        self.dirty = false;
    }
}

/// Run the editor
pub async fn run_editor() -> Result<()> {
    let mut editor = Editor::new().with_default();
    let command_sender = editor.command_sender.clone();
    let state = editor.state.clone();

    tokio::spawn(async move {
        event_handler(command_sender, state).await;
    });

    editor.run().await
}

/// Takes terminal events and sends commands that our editor handles.
async fn event_handler(sender: Sender<EditorCommand>, state: Arc<RwLock<EditorState>>) {
    loop {
        if let Ok(event) = event::read() {
            let state = state.read().await;
            let command = match event {
                Event::Key(key_event) => {
                    key_event_handler((key_event.modifiers, key_event.code), &state.mode)
                }
                _ => None,
            };

            if let Some(command) = command {
                sender.send(command).await.unwrap();
            }
        }
    }
}

/// Handles key-specific events
fn key_event_handler(key: (KeyModifiers, KeyCode), mode: &Mode) -> Option<EditorCommand> {
    match (mode, key) {
        // NORMAL MODE
        (Mode::Normal, (KeyModifiers::NONE, KeyCode::Char('q'))) => Some(EditorCommand::Quit),
        // insertions
        (Mode::Normal, (KeyModifiers::NONE, KeyCode::Char('i'))) => {
            Some(EditorCommand::UpdateMode(Mode::Insert))
        }
        (Mode::Normal, (KeyModifiers::SHIFT, KeyCode::Char('I'))) => {
            Some(EditorCommand::AppendStart)
        }
        (Mode::Normal, (KeyModifiers::NONE, KeyCode::Char('a'))) => Some(EditorCommand::Append),
        (Mode::Normal, (KeyModifiers::SHIFT, KeyCode::Char('A'))) => Some(EditorCommand::AppendEnd),
        (Mode::Normal, (KeyModifiers::NONE, KeyCode::Char('o'))) => {
            Some(EditorCommand::AppendBelow)
        }
        (Mode::Normal, (KeyModifiers::SHIFT, KeyCode::Char('O'))) => {
            Some(EditorCommand::AppendAbove)
        }

        // movements
        (Mode::Normal, (KeyModifiers::NONE, KeyCode::Char('h'))) => Some(
            EditorCommand::MoveCursor(Direction::Backward, Granularity::Character),
        ),
        (Mode::Normal, (KeyModifiers::NONE, KeyCode::Char('l'))) => Some(
            EditorCommand::MoveCursor(Direction::Forward, Granularity::Character),
        ),
        (Mode::Normal, (KeyModifiers::NONE, KeyCode::Char('j'))) => Some(
            EditorCommand::MoveCursor(Direction::Forward, Granularity::Line),
        ),
        (Mode::Normal, (KeyModifiers::NONE, KeyCode::Char('k'))) => Some(
            EditorCommand::MoveCursor(Direction::Backward, Granularity::Line),
        ),

        // INSERT MODE
        (Mode::Insert, (KeyModifiers::NONE, KeyCode::Esc)) => {
            Some(EditorCommand::UpdateMode(Mode::Normal))
        }
        (Mode::Insert, (KeyModifiers::NONE, KeyCode::Char(ch))) => {
            Some(EditorCommand::InsertChar(ch))
        }
        (Mode::Insert, (KeyModifiers::NONE, KeyCode::Backspace)) => Some(EditorCommand::Backspace),
        (Mode::Insert, (KeyModifiers::NONE, KeyCode::Enter)) => Some(EditorCommand::Enter),

        // NONE
        _ => None,
    }
}
