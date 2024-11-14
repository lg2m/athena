package util

import "testing"

// MockFileSystem is a mock implementation of FileSystem
type MockFileSystem struct {
	BaseFunc func(path string) string
	ExtFunc  func(path string) string
}

func (m MockFileSystem) Base(path string) string {
	return m.BaseFunc(path)
}

func (m MockFileSystem) Ext(path string) string {
	return m.ExtFunc(path)
}

func TestFileUtilWithMocks(t *testing.T) {
	// Test GetFileName with mocks
	t.Run("GetFileName with mocks", func(t *testing.T) {
		mockFS := MockFileSystem{
			BaseFunc: func(path string) string {
				return "file.txt"
			},
			ExtFunc: func(path string) string {
				return ".txt"
			},
		}

		fu := NewFileUtil(mockFS)

		// Test with extension
		got := fu.GetFileName("any/path/file.txt", true)
		want := "file.txt"
		if got != want {
			t.Errorf("GetFileName() with ext = %v, want %v", got, want)
		}

		// Test without extension
		got = fu.GetFileName("any/path/file.txt", false)
		want = "file"
		if got != want {
			t.Errorf("GetFileName() without ext = %v, want %v", got, want)
		}
	})

	// Test GetFileExt with mocks
	t.Run("GetFileExt with mocks", func(t *testing.T) {
		mockFS := MockFileSystem{
			ExtFunc: func(path string) string {
				return ".txt"
			},
		}

		fu := NewFileUtil(mockFS)

		got := fu.GetFileExt("any/path/file.txt")
		want := "txt"
		if got != want {
			t.Errorf("GetFileExt() = %v, want %v", got, want)
		}
	})
}
