package rope

import (
	"github.com/rivo/uniseg"
)

// RopeIterator allows traversal of the Rope's grapheme clusters.
type RopeIterator struct {
	current     *RopeNode
	stack       []*RopeNode
	graphemes   *uniseg.Graphemes
	position    int
	leafStart   int    // Start position of current leaf node
	graphemePos int    // Position within current leaf node
	leafData    string // Current leaf's data for reverse traversal
	// stack     []*RopeNode
	// graphemes *uniseg.Graphemes
	// current   *RopeNode
}

// NewIterator creates a new RopeIterator starting from the beginning.
func (r *Rope) NewIterator() *RopeIterator {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return &RopeIterator{
		current:   r.root,
		stack:     make([]*RopeNode, 0, 32),
		graphemes: nil,
	}
}

// Next advances the iterator and returns the next grapheme cluster.
func (it *RopeIterator) Next() (string, bool) {
	for {
		// Traverse to the leftmost leaf node.
		for it.current != nil {
			it.stack = append(it.stack, it.current)
			it.current = it.current.left
		}

		if len(it.stack) == 0 {
			// Traversal is complete.
			return "", false
		}

		// Pop the node from the stack.
		it.current = it.stack[len(it.stack)-1]
		it.stack = it.stack[:len(it.stack)-1]

		if it.current.left == nil && it.current.right == nil {
			// Leaf node.
			if it.graphemes == nil {
				it.graphemes = uniseg.NewGraphemes(it.current.data)
			}

			if it.graphemes.Next() {
				it.position++
				// Keep the current node and graphemes for the next call.
				return it.graphemes.Str(), true
			} else {
				// Finished with this leaf node, reset graphemes.
				it.graphemes = nil
			}
		}

		// Move to the right subtree.
		it.current = it.current.right
	}
}

func (it *RopeIterator) Prev() (string, bool) {
	if it.position <= 0 {
		return "", false
	}

	// If we're in a leaf node and not at its start, move back within it
	if it.graphemes != nil && it.graphemePos > 0 {
		// Reset graphemes iterator and scan forward to previous position
		it.graphemes = uniseg.NewGraphemes(it.leafData)
		for i := 0; i < it.graphemePos-1; i++ {
			if !it.graphemes.Next() {
				return "", false
			}
		}
		if it.graphemes.Next() {
			it.position--
			it.graphemePos--
			return it.graphemes.Str(), true
		}
	}

	// Find the previous leaf node
	for {
		// If we're at leaf node's start or don't have a current node,
		// we need to traverse to the previous leaf
		if it.current == nil || (it.current.left == nil && it.current.right == nil && it.graphemePos == 0) {
			// Pop nodes until we find one where we came from right
			var lastPopped *RopeNode
			for len(it.stack) > 0 {
				lastPopped = it.current
				it.current = it.stack[len(it.stack)-1]
				it.stack = it.stack[:len(it.stack)-1]

				// If we popped from the right subtree, this node is our predecessor
				if it.current.right == lastPopped {
					// Found our predecessor, if it's a leaf we'll process it
					if it.current.left == nil && it.current.right == nil {
						break
					}
					// If not a leaf, move to rightmost leaf of left subtree
					it.current = it.current.left
					for it.current.right != nil {
						it.stack = append(it.stack, it.current)
						it.current = it.current.right
					}
					break
				}

				// If we popped from the left subtree, keep popping
			}
		}

		// Process the leaf node
		if it.current != nil && it.current.left == nil && it.current.right == nil {
			it.leafData = it.current.data
			it.graphemes = uniseg.NewGraphemes(it.current.data)

			// Count graphemes to find the last one
			it.graphemePos = 0
			for it.graphemes.Next() {
				it.graphemePos++
			}

			// Position at the last grapheme
			if it.graphemePos > 0 {
				it.graphemes = uniseg.NewGraphemes(it.current.data)
				for i := 0; i < it.graphemePos-1; i++ {
					if !it.graphemes.Next() {
						return "", false
					}
				}
				if it.graphemes.Next() {
					it.position--
					it.graphemePos--
					return it.graphemes.Str(), true
				}
			}
		}

		if len(it.stack) == 0 && (it.current == nil || it.graphemePos == 0) {
			// We've traversed everything
			return "", false
		}
	}
}
