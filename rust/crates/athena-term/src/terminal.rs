use anyhow::Result;
use crossterm::{cursor, execute, terminal, QueueableCommand};
use std::io::{self, Stdout, Write};

#[macro_export]
macro_rules! display {
    ( $self:expr, $( $x:expr ),* ) => {
        queue!($self.terminal.stdout, SetAttribute(Attribute::NormalIntensity))?;
        $(
            queue!($self.terminal.stdout, Print($x))?;
        )*
    };
}

pub struct Terminal {
    pub stdout: Stdout,
}

impl Terminal {
    #[must_use]
    pub fn new() -> Self {
        Self {
            stdout: io::stdout(),
        }
    }

    /// Setup terminal.
    pub fn start(&mut self) -> Result<()> {
        std::panic::set_hook(Box::new(|e| {
            terminal::disable_raw_mode().unwrap();
            execute!(io::stdout(), terminal::LeaveAlternateScreen, cursor::Show).unwrap();
            eprintln!("{e}");
        }));

        terminal::enable_raw_mode()?;

        execute!(
            self.stdout,
            terminal::EnterAlternateScreen,
            terminal::Clear(terminal::ClearType::All)
        )?;

        Ok(())
    }

    /// Brings terminal back to it's original state.
    pub fn stop(&mut self) -> Result<()> {
        Ok(())
    }

    /// Show the cursor on screen.
    pub fn show_cursor(&mut self) -> Result<()> {
        self.stdout.queue(cursor::Show)?;
        Ok(())
    }

    /// Hide the cursor on screen.
    pub fn hide_cursor(&mut self) -> Result<()> {
        self.stdout.queue(cursor::Hide)?;
        Ok(())
    }

    /// Clear the current line where cursor is at.
    pub fn clear_current_line(&mut self) -> Result<()> {
        self.stdout
            .queue(terminal::Clear(terminal::ClearType::CurrentLine))?;
        Ok(())
    }

    /// Flush the stdout.
    pub fn flush(&mut self) -> Result<()> {
        self.stdout.flush()?;
        Ok(())
    }

    /// Move cursor to x, y pos on screen.
    pub fn goto<T: Into<usize>>(&mut self, x: T, y: T) -> Result<()> {
        self.stdout.queue(cursor::MoveTo(
            u16::try_from(x.into()).unwrap_or(u16::MAX),
            u16::try_from(y.into()).unwrap_or(u16::MAX),
        ))?;
        Ok(())
    }

    /// Moves to a line and ensures that it is cleared.
    pub fn prepare_line(&mut self, y: usize) -> Result<()> {
        self.goto(0, y)?;
        self.clear_current_line()
    }

    pub fn size(&self) -> Result<(usize, usize)> {
        let (width, height) = terminal::size()?;
        Ok((width as usize, (height as usize).saturating_sub(1)))
    }
}
