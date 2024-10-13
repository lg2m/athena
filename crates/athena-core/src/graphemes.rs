use ropey::{iter::Chunks, RopeSlice};
use unicode_segmentation::{GraphemeCursor, GraphemeIncomplete, UnicodeSegmentation};
use unicode_width::UnicodeWidthStr;

/// A trait to handle grapheme boundaries and width. This makes the logic extendable for different
/// types of underlying text storage.
pub trait GraphemeOperations {
    fn grapheme_width(&self) -> usize;
    fn prev_grapheme_boundary(&self, index: usize) -> usize;
    fn next_grapheme_boundary(&self, index: usize) -> usize;
    fn is_grapheme_boundary(&self, index: usize) -> bool;
}

/// Implementation of `GraphemeOperations` for `RopeSlice`.
impl<'a> GraphemeOperations for RopeSlice<'a> {
    fn grapheme_width(&self) -> usize {
        if self.len_chars() == 0 {
            return 0;
        }

        // Get the first chunk of the RopeSlice
        let (chunk, _, _, _) = self.chunk_at_char(0);

        // Check if the first byte is ASCII
        if chunk.as_bytes()[0] <= 127 {
            1 // Fast path for ASCII
        } else {
            // Calc the width of the first grapheme cluster
            let grapheme = chunk.graphemes(true).next().unwrap_or("");
            UnicodeWidthStr::width(grapheme).max(1)
        }
    }

    fn prev_grapheme_boundary(&self, char_idx: usize) -> usize {
        let byte_idx = self.char_to_byte(char_idx);
        let mut gc = GraphemeCursor::new(byte_idx, self.len_bytes(), true);

        let (mut chunk, mut chunk_byte_idx, mut chunk_char_idx, _) = self.chunk_at_byte(byte_idx);

        loop {
            match gc.prev_boundary(chunk, chunk_byte_idx) {
                Ok(None) => return 0,
                Ok(Some(n)) => {
                    let tmp = byte_to_char_idx(chunk, n - chunk_byte_idx);
                    return chunk_char_idx + tmp;
                }
                Err(GraphemeIncomplete::PrevChunk) => {
                    let (a, b, c, _) = self.chunk_at_byte(chunk_byte_idx - 1);
                    chunk = a;
                    chunk_byte_idx = b;
                    chunk_char_idx = c;
                }
                Err(GraphemeIncomplete::PreContext(n)) => {
                    let ctx_chunk = self.chunk_at_byte(n - 1).0;
                    gc.provide_context(ctx_chunk, n - ctx_chunk.len());
                }
                _ => unreachable!(),
            }
        }
    }

    fn next_grapheme_boundary(&self, char_idx: usize) -> usize {
        let byte_idx = self.char_to_byte(char_idx);
        let mut gc = GraphemeCursor::new(byte_idx, self.len_bytes(), true);

        let (mut chunk, mut chunk_byte_idx, mut chunk_char_idx, _) = self.chunk_at_byte(byte_idx);

        loop {
            match gc.next_boundary(chunk, chunk_byte_idx) {
                Ok(None) => return self.len_chars(),
                Ok(Some(n)) => {
                    let tmp = byte_to_char_idx(chunk, n - chunk_byte_idx);
                    return chunk_char_idx + tmp;
                }
                Err(GraphemeIncomplete::NextChunk) => {
                    chunk_byte_idx += chunk.len();
                    let (a, _, c, _) = self.chunk_at_byte(chunk_byte_idx);
                    chunk = a;
                    chunk_char_idx = c;
                }
                Err(GraphemeIncomplete::PreContext(n)) => {
                    let ctx_chunk = self.chunk_at_byte(n - 1).0;
                    gc.provide_context(ctx_chunk, n - ctx_chunk.len());
                }
                _ => unreachable!(),
            }
        }
    }

    fn is_grapheme_boundary(&self, char_idx: usize) -> bool {
        let byte_idx = self.char_to_byte(char_idx);
        let (chunk, chunk_byte_idx, _, _) = self.chunk_at_byte(byte_idx);
        let mut gc = GraphemeCursor::new(byte_idx, self.len_bytes(), true);

        loop {
            match gc.is_boundary(chunk, chunk_byte_idx) {
                Ok(n) => return n,
                Err(GraphemeIncomplete::PreContext(n)) => {
                    let (ctx_chunk, ctx_byte_start, _, _) = self.chunk_at_byte(n - 1);
                    gc.provide_context(ctx_chunk, ctx_byte_start);
                }
                _ => unreachable!(),
            }
        }
    }
}

/// Convert byte index to character index (necessary for UTF-8 processing).
fn byte_to_char_idx(text: &str, index: usize) -> usize {
    text[..index].chars().count()
}

/// Iterator for grapheme clusters in a TextSlice.
pub struct GraphemeIter<'a> {
    text: RopeSlice<'a>,
    cursor: GraphemeCursor,
    chunks: Chunks<'a>,
    cur_chunk: &'a str,
    cur_chunk_start: usize,
}

impl<'a> GraphemeIter<'a> {
    pub fn new(slice: RopeSlice<'a>) -> GraphemeIter<'a> {
        let mut chunks = slice.chunks();
        let first_chunk = chunks.next().unwrap_or("");
        GraphemeIter {
            text: slice,
            cursor: GraphemeCursor::new(0, slice.len_bytes(), true),
            chunks,
            cur_chunk: first_chunk,
            cur_chunk_start: 0,
        }
    }
}

impl<'a> Iterator for GraphemeIter<'a> {
    type Item = RopeSlice<'a>;

    fn next(&mut self) -> Option<RopeSlice<'a>> {
        let a = self.cursor.cur_cursor();
        let b;

        loop {
            match self
                .cursor
                .next_boundary(self.cur_chunk, self.cur_chunk_start)
            {
                Ok(None) => {
                    return None;
                }
                Ok(Some(n)) => {
                    b = n;
                    break;
                }
                Err(GraphemeIncomplete::NextChunk) => {
                    self.cur_chunk_start += self.cur_chunk.len();
                    self.cur_chunk = self.chunks.next().unwrap_or("");
                }
                _ => unreachable!(),
            }
        }

        if a < self.cur_chunk_start {
            let a_char = self.text.byte_to_char(a);
            let b_char = self.text.byte_to_char(b);

            Some(self.text.slice(a_char..b_char))
        } else {
            let a2 = a - self.cur_chunk_start;
            let b2 = b - self.cur_chunk_start;
            Some((&self.cur_chunk[a2..b2]).into())
        }
    }
}
