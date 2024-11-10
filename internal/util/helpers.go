package util

import (
	"path/filepath"
	"strings"
)

func GetFileName(filePath string, withExt bool) string {
	fileName := filepath.Base(filePath)
	if !withExt {
		fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	}
	return fileName
}

func GetFileExt(filePath string) string {
	ext := filepath.Ext(filePath)
	return strings.TrimPrefix(ext, ".")
}

func CalcProgress(tot, curr int) int {
	if tot == 0 {
		return 0
	}
	return int(((float64(curr) / float64(tot)) * 100) + 0.5)
}
