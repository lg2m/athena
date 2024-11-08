package buffer

import "errors"

// BufferManager manages multiple buffers.
type BufferManager struct {
	buffers map[string]*Buffer // Keyed by buffer ID or file path
	current *Buffer
}

// NewBufferManager initializes a new BufferManager.
func NewBufferManager() *BufferManager {
	return &BufferManager{
		buffers: make(map[string]*Buffer),
	}
}

// AddBuffer adds a buffer with a given name.
func (bm *BufferManager) AddBuffer(buffer *Buffer) {
	bm.buffers[buffer.name] = buffer
	bm.current = buffer
}

// GetBuffer retrieves a buffer by name.
func (bm *BufferManager) GetBuffer(name string) (*Buffer, bool) {
	buffer, exists := bm.buffers[name]
	return buffer, exists
}

// SetCurrentBuffer sets the current active buffer.
func (bm *BufferManager) SetCurrentBuffer(name string) error {
	buffer, exists := bm.buffers[name]
	if !exists {
		return errors.New("buffer not found")
	}
	bm.current = buffer
	return nil
}

// GetCurrentBuffer returns the current active buffer.
func (bm *BufferManager) GetCurrentBuffer() *Buffer {
	return bm.current
}
