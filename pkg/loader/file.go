package loader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// File is a document loader that reads files from the local filesystem.
// It supports text files and can be extended for other formats.
type File struct{}

// NewFile creates a new File loader.
func NewFile() *File {
	return &File{}
}

// Load implements DocumentLoader. The source can be a single file path
// or a directory (all .txt and .md files will be loaded).
func (f *File) Load(ctx context.Context, source string) ([]core.Document, error) {
	_ = ctx

	info, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("file loader: %w", err)
	}

	if info.IsDir() {
		return f.loadDir(source)
	}
	return f.loadFile(source)
}

func (f *File) loadFile(path string) ([]core.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("file loader: read %s: %w", path, err)
	}

	return []core.Document{
		{
			ID:      filepath.Base(path),
			Content: string(data),
			Metadata: map[string]any{
				"source": path,
				"size":   len(data),
			},
		},
	}, nil
}

func (f *File) loadDir(dir string) ([]core.Document, error) {
	var docs []core.Document

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".txt" && ext != ".md" {
			return nil
		}
		loaded, err := f.loadFile(path)
		if err != nil {
			return err
		}
		docs = append(docs, loaded...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("file loader: walk %s: %w", dir, err)
	}

	return docs, nil
}
