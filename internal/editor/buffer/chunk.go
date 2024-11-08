package buffer

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/lg2m/athena/internal/rope"
)

const (
	DefaultChunkSize = 1024 * 64 // 64KB chunks
	MaxLoadedChunks  = 50        // Maximum number of chunks to keep in memory
)

// Chunk represents a portion of the file content
type Chunk struct {
	ID       int
	Data     *rope.Rope
	Dirty    bool
	LastUsed time.Time
	mu       sync.RWMutex
}

func (c *Chunk) Read() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.LastUsed = time.Now()
	return c.Data.ToString()
}

func (c *Chunk) Write(data string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data = rope.NewRope(data)
	c.Dirty = true
	c.LastUsed = time.Now()
}

// ChunkManager handles loading, unloading, and caching of chunks
type ChunkManager struct {
	chunks       map[int]*Chunk
	chunkSize    int
	loadedChunks int
	totalSize    int64
	filePath     string
	file         *os.File
	mu           sync.RWMutex
}

func NewChunkManager(filePath string) (*ChunkManager, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	cm := &ChunkManager{
		chunks:    make(map[int]*Chunk),
		chunkSize: DefaultChunkSize,
		totalSize: info.Size(),
		filePath:  filePath,
		file:      file,
	}

	return cm, nil
}

// GetChunk retrieves a chunk, loading it if necessary
func (cm *ChunkManager) GetChunk(id int) (*Chunk, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if chunk is already loaded
	if chunk, exists := cm.chunks[id]; exists {
		chunk.LastUsed = time.Now()
		return chunk, nil
	}

	// Ensure we don't exceed max loaded chunks
	if cm.loadedChunks >= MaxLoadedChunks {
		if err := cm.evictOldestChunk(); err != nil {
			return nil, err
		}
	}

	// Load the chunk
	chunk, err := cm.loadChunk(id)
	if err != nil {
		return nil, err
	}

	cm.chunks[id] = chunk
	cm.loadedChunks++
	return chunk, nil
}

// loadChunk loads a chunk from disk
func (cm *ChunkManager) loadChunk(id int) (*Chunk, error) {
	offset := int64(id * cm.chunkSize)
	if offset >= cm.totalSize {
		return nil, io.EOF
	}

	// Calculate size to read
	size := cm.chunkSize
	if remaining := cm.totalSize - offset; remaining < int64(cm.chunkSize) {
		size = int(remaining)
	}

	data := make([]byte, size)
	_, err := cm.file.ReadAt(data, offset)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &Chunk{
		ID:       id,
		Data:     rope.NewRope(string(data)),
		LastUsed: time.Now(),
	}, nil
}

// evictOldestChunk removes the least recently used chunk
func (cm *ChunkManager) evictOldestChunk() error {
	var oldest *Chunk
	var oldestID int

	for id, chunk := range cm.chunks {
		if oldest == nil || chunk.LastUsed.Before(oldest.LastUsed) {
			oldest = chunk
			oldestID = id
		}
	}

	if oldest != nil && oldest.Dirty {
		if err := cm.saveChunk(oldest); err != nil {
			return err
		}
	}

	delete(cm.chunks, oldestID)
	cm.loadedChunks--
	return nil
}

// saveChunk writes a chunk back to disk
func (cm *ChunkManager) saveChunk(chunk *Chunk) error {
	offset := int64(chunk.ID * cm.chunkSize)
	data := []byte(chunk.Read())
	_, err := cm.file.WriteAt(data, offset)
	if err != nil {
		return err
	}
	chunk.Dirty = false
	return nil
}
