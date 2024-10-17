use ropey::{
    iter::{Chars, Chunks},
    RopeSlice,
};
use unicode_segmentation::{GraphemeCursor, GraphemeIncomplete, UnicodeSegmentation};
use unicode_width::UnicodeWidthStr;

/// A trait to handle grapheme boundaries and width. This makes the logic extendable for different
/// types of underlying text storage.
pub trait GraphemeOperations {
    fn grapheme_width(&self) -> usize;
    fn prev_grapheme_boundary(&self, index: usize) -> usize;
    fn prev_word_boundary(&self, char_idx: usize) -> usize;
    fn next_grapheme_boundary(&self, index: usize) -> usize;
    fn next_word_boundary(&self, char_idx: usize) -> usize;
    fn is_grapheme_boundary(&self, index: usize) -> bool;
}

/// Implementation of `GraphemeOperations` for `RopeSlice`.
impl<'a> GraphemeOperations for RopeSlice<'a> {
    fn grapheme_width(&self) -> usize {
        if self.len_chars() == 0 {
            return 0;
        }

        let (chunk, chunk_start, _, _) = self.chunk_at_char(0);
        let end_char_idx = chunk.char_indices().nth(1).map_or(chunk.len(), |(i, _)| i);

        let mut graphemes = chunk[..end_char_idx].graphemes(true);
        if let Some(grapheme) = graphemes.next() {
            UnicodeWidthStr::width(grapheme).max(1)
        } else {
            1
        }

        ///// OLD
        // // Get the first chunk of the RopeSlice
        // let (chunk, _, _, _) = self.chunk_at_char(0);

        // // Check if the first byte is ASCII
        // if chunk.as_bytes()[0] <= 127 {
        //     1 // Fast path for ASCII
        // } else {
        //     // Calc the width of the first grapheme cluster
        //     let grapheme = chunk.graphemes(true).next().unwrap_or("");
        //     UnicodeWidthStr::width(grapheme).max(1)
        // }

        ////// NEW v1
        // let mut iter = self.chars();
        // let first_grapheme = iter.next().unwrap_or('\0').to_string();
        // let mut grapheme_iter = first_grapheme.graphemes(true);
        // let grapheme = grapheme_iter.next().unwrap_or("");
        // UnicodeWidthStr::width(grapheme).max(1)
    }

    fn prev_grapheme_boundary(&self, char_idx: usize) -> usize {
        if char_idx == 0 {
            return 0;
        }

        let byte_idx = self.char_to_byte(char_idx);
        let max_context = 128;

        let start_byte = byte_idx.saturating_sub(max_context);
        let (start_chunk, chunk_byte_start, chunk_char_start, _) = self.chunk_at_byte(start_byte);

        let context_len = byte_idx - chunk_byte_start;
        let context = &start_chunk[..context_len];

        let mut prev_boundary = 0;
        for (i, _) in context.grapheme_indices(true) {
            let char_count = context[..i].chars().count();
            if char_count >= char_idx - chunk_char_start {
                break;
            }
            prev_boundary = char_count;
        }

        chunk_char_start + prev_boundary

        ////// NEW v1
        // const MAX_CONTEXT: usize = 64;
        // let char_idx = char_idx.min(self.len_chars());
        // let start_char = char_idx.saturating_sub(MAX_CONTEXT);
        // let slice = self.slice(start_char..char_idx);
        // let context_str = slice.chars().collect::<String>();

        // let mut last_boundary = start_char;
        // for (i, _) in context_str.grapheme_indices(true) {
        //     let grapheme_char_idx = start_char + context_str[..i].chars().count();
        //     if grapheme_char_idx >= char_idx {
        //         break;
        //     }
        //     last_boundary = grapheme_char_idx;
        // }
        // last_boundary

        ///// OLD
        // let byte_idx = self.char_to_byte(char_idx);
        // let mut gc = GraphemeCursor::new(byte_idx, self.len_bytes(), true);

        // let (mut chunk, mut chunk_byte_idx, mut chunk_char_idx, _) = self.chunk_at_byte(byte_idx);

        // loop {
        //     match gc.prev_boundary(chunk, chunk_byte_idx) {
        //         Ok(None) => return 0,
        //         Ok(Some(n)) => {
        //             let tmp = byte_to_char_idx(chunk, n - chunk_byte_idx);
        //             return chunk_char_idx + tmp;
        //         }
        //         Err(GraphemeIncomplete::PrevChunk) => {
        //             let (a, b, c, _) = self.chunk_at_byte(chunk_byte_idx - 1);
        //             chunk = a;
        //             chunk_byte_idx = b;
        //             chunk_char_idx = c;
        //         }
        //         Err(GraphemeIncomplete::PreContext(n)) => {
        //             let ctx_chunk = self.chunk_at_byte(n - 1).0;
        //             gc.provide_context(ctx_chunk, n - ctx_chunk.len());
        //         }
        //         _ => unreachable!(),
        //     }
        // }
    }

    fn prev_word_boundary(&self, char_idx: usize) -> usize {
        if char_idx == 0 {
            return 0;
        }

        let max_context = 128;
        let start_idx = char_idx.saturating_sub(max_context);
        let slice = self.slice(start_idx..char_idx);

        let context_str = slice.chars().collect::<String>();
        let mut last_boundary = start_idx;

        for (i, _) in context_str.split_word_bound_indices() {
            let word_char_idx = start_idx + context_str[..i].chars().count();
            if word_char_idx >= char_idx {
                break;
            }
            last_boundary = word_char_idx;
        }

        last_boundary
    }

    fn next_grapheme_boundary(&self, char_idx: usize) -> usize {
        if char_idx >= self.len_chars() {
            return self.len_chars();
        }

        let byte_idx = self.char_to_byte(char_idx);
        let max_context = 128;

        // Determine the end of the context
        let (chunk, chunk_byte_start, _, _) = self.chunk_at_byte(byte_idx);
        let end_byte = (byte_idx - chunk_byte_start)
            + max_context.min(chunk.len() - (byte_idx - chunk_byte_start));

        // Get the context slice
        let context = &chunk[(byte_idx - chunk_byte_start)..end_byte];

        // Find the next grapheme boundary
        for (i, _) in context.grapheme_indices(true) {
            let char_count = context[..i].chars().count();
            if char_count > 0 {
                return char_idx + char_count;
            }
        }

        // If no boundary found, return the end of the text
        self.len_chars()

        ////// NEW v1
        // const MAX_CONTEXT: usize = 64;
        // let end_char = (char_idx + MAX_CONTEXT).min(self.len_chars());
        // let slice = self.slice(char_idx..end_char);
        // let context_str = slice.chars().collect::<String>();

        // for (i, _) in context_str.grapheme_indices(true) {
        //     let grapheme_char_idx = char_idx + context_str[..i].chars().count();
        //     if grapheme_char_idx > char_idx {
        //         return grapheme_char_idx;
        //     }
        // }

        // self.len_chars()

        ////// OLD
        // let byte_idx = self.char_to_byte(char_idx);
        // let mut gc = GraphemeCursor::new(byte_idx, self.len_bytes(), true);

        // let (mut chunk, mut chunk_byte_idx, mut chunk_char_idx, _) = self.chunk_at_byte(byte_idx);

        // loop {
        //     match gc.next_boundary(chunk, chunk_byte_idx) {
        //         Ok(None) => return self.len_chars(),
        //         Ok(Some(n)) => {
        //             let tmp = byte_to_char_idx(chunk, n - chunk_byte_idx);
        //             return chunk_char_idx + tmp;
        //         }
        //         Err(GraphemeIncomplete::NextChunk) => {
        //             chunk_byte_idx += chunk.len();
        //             let (a, _, c, _) = self.chunk_at_byte(chunk_byte_idx);
        //             chunk = a;
        //             chunk_char_idx = c;
        //         }
        //         Err(GraphemeIncomplete::PreContext(n)) => {
        //             let ctx_chunk = self.chunk_at_byte(n - 1).0;
        //             gc.provide_context(ctx_chunk, n - ctx_chunk.len());
        //         }
        //         _ => unreachable!(),
        //     }
        // }
    }

    fn next_word_boundary(&self, char_idx: usize) -> usize {
        if char_idx >= self.len_chars() {
            return self.len_chars();
        }

        let max_context = 128;
        let end_idx = (char_idx + max_context).min(self.len_chars());
        let slice = self.slice(char_idx..end_idx);

        let context_str: String = slice.chars().collect();

        for (i, _) in context_str.split_word_bound_indices() {
            let word_char_idx = char_idx + context_str[..i].chars().count();
            if word_char_idx > char_idx {
                return word_char_idx;
            }
        }

        self.len_chars()
    }

    fn is_grapheme_boundary(&self, char_idx: usize) -> bool {
        if char_idx == 0 || char_idx == self.len_chars() {
            return true;
        }

        let byte_idx = self.char_to_byte(char_idx);
        let max_context = 128;

        // Determine the start and end of the context
        let start_byte = byte_idx.saturating_sub(max_context);
        let end_byte = (byte_idx + max_context).min(self.len_bytes());

        // Collect context across chunks if necessary
        let mut context = String::new();
        let mut cur_byte = start_byte;
        while cur_byte < end_byte {
            let (chunk, chunk_byte_start, _, _) = self.chunk_at_byte(cur_byte);
            let remaining = end_byte - cur_byte;
            let chunk_offset = cur_byte - chunk_byte_start;
            let len = remaining.min(chunk.len() - chunk_offset);
            context.push_str(&chunk[chunk_offset..chunk_offset + len]);
            cur_byte += len;
        }

        // Find if char_idx is a grapheme boundary
        let mut char_count = 0;
        for (i, _) in context.grapheme_indices(true) {
            let gc = context[..i].chars().count();
            if char_count + gc == char_idx {
                return true;
            }
            if char_count + gc > char_idx {
                break;
            }
            char_count += gc;
        }

        false

        ////// NEW v1
        // const MAX_CONTEXT: usize = 64;
        // let start_char = char_idx.saturating_sub(MAX_CONTEXT);
        // let end_char = (char_idx + MAX_CONTEXT).min(self.len_chars());
        // let slice = self.slice(start_char..end_char);
        // let context_str = slice.chars().collect::<String>();

        // // let target_offset = char_idx - start_char;
        // for (i, _) in context_str.grapheme_indices(true) {
        //     let grapheme_char_idx = start_char + context_str[..i].chars().count();
        //     if grapheme_char_idx == char_idx {
        //         return true;
        //     }
        //     if grapheme_char_idx > char_idx {
        //         break;
        //     }
        // }

        // false

        ////// OLD
        // let byte_idx = self.char_to_byte(char_idx);
        // let (chunk, chunk_byte_idx, _, _) = self.chunk_at_byte(byte_idx);
        // let mut gc = GraphemeCursor::new(byte_idx, self.len_bytes(), true);

        // loop {
        //     match gc.is_boundary(chunk, chunk_byte_idx) {
        //         Ok(n) => return n,
        //         Err(GraphemeIncomplete::PreContext(n)) => {
        //             let (ctx_chunk, ctx_byte_start, _, _) = self.chunk_at_byte(n - 1);
        //             gc.provide_context(ctx_chunk, ctx_byte_start);
        //         }
        //         _ => unreachable!(),
        //     }
        // }
    }
}

/// Convert byte index to character index (necessary for UTF-8 processing).
fn byte_to_char_idx(text: &str, index: usize) -> usize {
    text[..index].chars().count()
}

/// Iterator for grapheme clusters in a TextSlice.
pub struct GraphemeIter<'a> {
    rope_slice: RopeSlice<'a>,
    char_idx: usize,
    len_chars: usize,
    // text: RopeSlice<'a>,
    // cursor: GraphemeCursor,
    // chunks: Chunks<'a>,
    // cur_chunk: &'a str,
    // cur_chunk_start: usize,
}

impl<'a> GraphemeIter<'a> {
    pub fn new(slice: RopeSlice<'a>) -> GraphemeIter<'a> {
        let len_chars = slice.len_chars();
        // let mut chunks = slice.chunks();
        // let first_chunk = chunks.next().unwrap_or("");
        GraphemeIter {
            rope_slice: slice,
            char_idx: 0,
            len_chars,
            // text: slice,
            // cursor: GraphemeCursor::new(0, slice.len_bytes(), true),
            // chunks,
            // cur_chunk: first_chunk,
            // cur_chunk_start: 0,
        }
    }
}

impl<'a> Iterator for GraphemeIter<'a> {
    type Item = RopeSlice<'a>;

    fn next(&mut self) -> Option<RopeSlice<'a>> {
        if self.char_idx >= self.len_chars {
            return None;
        }

        let next_boundary = self.rope_slice.next_grapheme_boundary(self.char_idx);
        let grapheme = self.rope_slice.slice(self.char_idx..next_boundary);
        self.char_idx = next_boundary;
        Some(grapheme)

        ////// OLD
        // let a = self.cursor.cur_cursor();
        // let b;

        // loop {
        //     match self
        //         .cursor
        //         .next_boundary(self.cur_chunk, self.cur_chunk_start)
        //     {
        //         Ok(None) => {
        //             return None;
        //         }
        //         Ok(Some(n)) => {
        //             b = n;
        //             break;
        //         }
        //         Err(GraphemeIncomplete::NextChunk) => {
        //             self.cur_chunk_start += self.cur_chunk.len();
        //             self.cur_chunk = self.chunks.next().unwrap_or("");
        //         }
        //         _ => unreachable!(),
        //     }
        // }

        // if a < self.cur_chunk_start {
        //     let a_char = self.text.byte_to_char(a);
        //     let b_char = self.text.byte_to_char(b);

        //     Some(self.text.slice(a_char..b_char))
        // } else {
        //     let a2 = a - self.cur_chunk_start;
        //     let b2 = b - self.cur_chunk_start;
        //     Some((&self.cur_chunk[a2..b2]).into())
        // }
    }
}
