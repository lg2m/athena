package rope

import (
	"errors"
	"strings"
	"sync"

	"github.com/rivo/uniseg"
)

// MaxLeafSize defines the maximum number of grapheme clusters in a leaf node.
const MaxLeafSize = 256

var (
	ErrInvalidRange = errors.New("invalid range")
	ErrOutOfBounds  = errors.New("index out of bounds")
)

// RopeNode represents a node in the Rope data structure.
type RopeNode struct {
	left   *RopeNode
	right  *RopeNode
	weight int    // Number of grapheme clusters in the left subtree
	data   string // Only for leaf nodes
}

// Rope represents the Rope data structure.
type Rope struct {
	root *RopeNode
	mu   sync.RWMutex
}

// NewRope creates a new Rope from a string.
func NewRope(s string) *Rope {
	leaves := splitIntoLeaves(s, MaxLeafSize)
	root := buildBalancedTree(leaves)
	return &Rope{root: root}
}

// Insert inserts text at a given grapheme index in the Rope.
func (r *Rope) Insert(index int, s string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if index < 0 || index > r.root.totalGraphemes() {
		return ErrOutOfBounds
	}

	// Split the rope at the position
	left, right := r.root.Split(index)
	insertRope := NewRope(s)
	newLeft := concatenateNodes(left, insertRope.root)
	r.root = concatenateNodes(newLeft, right)
	return nil
}

// Delete removes grapheme clusters from start to end (exclusive).
func (r *Rope) Delete(start, end int) error {
	if start < 0 || end > r.root.totalGraphemes() || start > end {
		return ErrInvalidRange
	}
	left, temp := r.root.Split(start)
	_, right := temp.Split(end - start)
	r.root = concatenateNodes(left, right)
	return nil
}

// ToString returns the string representation of the Rope.
func (r *Rope) ToString() string {
	var sb strings.Builder
	r.root.writeToString(&sb)
	return sb.String()
}

// GetTextRange retrieves text from start to end grapheme indices (exclusive).
func (r *Rope) GetTextRange(start, end int) (string, error) {
	if start < 0 || end > r.root.totalGraphemes() || start > end {
		return "", errors.New("invalid text range")
	}
	var sb strings.Builder
	r.root.appendTextRange(&sb, start, end)
	return sb.String(), nil
}

func (r *Rope) TotalGraphemes() int {
	return r.root.totalGraphemes()
}

// Split splits the RopeNode at the given grapheme index.
func (n *RopeNode) Split(index int) (*RopeNode, *RopeNode) {
	if n == nil {
		return nil, nil
	}
	if n.left == nil && n.right == nil {
		// Leaf node: split the data string
		gr := uniseg.NewGraphemes(n.data)
		var leftData, rightData strings.Builder
		count := 0
		for gr.Next() {
			if count < index {
				leftData.WriteString(gr.Str())
			} else {
				rightData.WriteString(gr.Str())
			}
			count++
		}
		leftNode := &RopeNode{
			data:   leftData.String(),
			weight: CountGraphemes(leftData.String()),
		}
		rightNode := &RopeNode{
			data:   rightData.String(),
			weight: CountGraphemes(rightData.String()),
		}
		return leftNode, rightNode
	}
	if index < n.weight {
		// split in left subtree
		leftLeft, leftRight := n.left.Split(index)
		newRight := concatenateNodes(leftRight, n.right)
		return leftLeft, newRight
	}
	// split in right subtree
	rightLeft, rightRight := n.right.Split(index - n.weight)
	newLeft := concatenateNodes(n.left, rightLeft)
	return newLeft, rightRight
}

// writeToString
func (n *RopeNode) writeToString(sb *strings.Builder) {
	if n == nil {
		return
	}
	if n.left == nil && n.right == nil {
		sb.WriteString(n.data)
		return
	}
	n.left.writeToString(sb)
	n.right.writeToString(sb)
}

// appendTextRange appends text within a range to the provided StringBuilder.
func (n *RopeNode) appendTextRange(sb *strings.Builder, start, end int) {
	if n == nil || start >= end {
		return
	}
	if n.left == nil && n.right == nil {
		// Leaf node: extract the substring within the range.
		gr := uniseg.NewGraphemes(n.data)
		count := 0
		for gr.Next() {
			if count >= start && count < end {
				sb.WriteString(gr.Str())
			}
			count++
			if count >= end {
				break
			}
		}
		return
	}
	if start < n.weight {
		// range starts in the left subtree.
		leftEnd := min(end, n.weight)
		n.left.appendTextRange(sb, start, leftEnd)
	}
	if end > n.weight {
		// The range extends into the right subtree.
		rightStart := max(0, start-n.weight)
		rightEnd := end - n.weight
		n.right.appendTextRange(sb, rightStart, rightEnd)
	}
}

// totalGraphemes returns the total number of grapheme clusters in the node.
func (n *RopeNode) totalGraphemes() int {
	if n == nil {
		return 0
	}
	if n.left == nil && n.right == nil {
		return n.weight
	}
	return n.weight + n.right.totalGraphemes()
}

// CountGraphemes counts the number of grapheme clusters in a string.
func CountGraphemes(s string) int {
	gr := uniseg.NewGraphemes(s)
	count := 0
	for gr.Next() {
		count++
	}
	return count
}

// splitIntoLeaves splits the input string into chunks of up to maxSize grapheme clusters.
func splitIntoLeaves(s string, maxSize int) []*RopeNode {
	var leaves []*RopeNode
	gr := uniseg.NewGraphemes(s)
	var sb strings.Builder
	count := 0
	for gr.Next() {
		sb.WriteString(gr.Str())
		count++
		if count >= maxSize {
			leafData := sb.String()
			leaves = append(leaves, &RopeNode{
				data:   leafData,
				weight: count,
			})
			sb.Reset()
			count = 0
		}
	}
	if sb.Len() > 0 {
		leafData := sb.String()
		leaves = append(leaves, &RopeNode{
			data:   leafData,
			weight: count,
		})
	}
	return leaves
}

// buildBalancedTree builds a balanced rope tree from a list of leaf nodes.
func buildBalancedTree(leaves []*RopeNode) *RopeNode {
	if len(leaves) == 0 {
		return nil
	}
	if len(leaves) == 1 {
		return leaves[0]
	}
	mid := len(leaves) / 2
	left := buildBalancedTree(leaves[:mid])
	right := buildBalancedTree(leaves[mid:])
	return &RopeNode{
		left:   left,
		right:  right,
		weight: left.totalGraphemes(),
	}
}

// concatenateNodes concatenates two RopeNodes into a new parent node.
func concatenateNodes(left, right *RopeNode) *RopeNode {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	return &RopeNode{
		left:   left,
		right:  right,
		weight: left.totalGraphemes(),
	}
}
