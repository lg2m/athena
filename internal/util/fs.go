package util

import (
	"path/filepath"
	"strings"
)

// FileSystem interface defines the file system operations we need.
type FileSystem interface {
	Base(path string) string
	Ext(path string) string
}

// DefaultFileSystem implements FileSystem using the standard library.
type DefaultFileSystem struct{}

func (fs DefaultFileSystem) Base(path string) string {
	return filepath.Base(path)
}

func (fs DefaultFileSystem) Ext(path string) string {
	return filepath.Ext(path)
}

// FileUtil contains the file utility methods.
type FileUtil struct {
	fs FileSystem
}

// NewFileUtil creates a new FileUtil with the given FileSystem.
func NewFileUtil(fs FileSystem) *FileUtil {
	if fs == nil {
		fs = DefaultFileSystem{}
	}
	return &FileUtil{fs: fs}
}

// GetFileName returns the base name of the file path.
func (fu *FileUtil) GetFileName(filePath string, withExt bool) string {
	fileName := fu.fs.Base(filePath)
	if !withExt {
		fileName = strings.TrimSuffix(fileName, fu.fs.Ext(fileName))
	}
	return fileName
}

// GetFileExt returns the file extension without the dot.
func (fu *FileUtil) GetFileExt(filePath string) string {
	ext := fu.fs.Ext(filePath)
	return strings.TrimPrefix(ext, ".")
}
