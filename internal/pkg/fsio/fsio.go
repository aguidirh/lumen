// Package fsio (file system I/O) provides functions for file system operations.
package fsio

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

// FsIO provides methods for file system operations.
type FsIO struct{}

// NewFsIO creates a new FsIO instance.
func NewFsIO() *FsIO {
	return &FsIO{}
}

// CopyDirectory recursively copies a directory from src to dst.
func (f *FsIO) CopyDirectory(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return f.CopyFile(path, dstPath)
	})
}

// CopyFile copies a single file from src to dst.
func (f *FsIO) CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// UntarFromStream reads a tar stream (potentially gzipped) and extracts it to a destination directory.
func (f *FsIO) UntarFromStream(r io.Reader, dest string) error {
	// We need to peek at the first few bytes to determine if it's a gzipped stream.
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}

	// Create a new reader that prepends the peeked bytes to the original reader.
	multiReader := io.MultiReader(bytes.NewReader(buf[:n]), r)

	var tarReader io.Reader
	contentType := http.DetectContentType(buf)
	if contentType == "application/x-gzip" {
		gzr, err := gzip.NewReader(multiReader)
		if err != nil {
			return err
		}
		defer gzr.Close()
		tarReader = gzr
	} else {
		tarReader = multiReader
	}

	tr := tar.NewReader(tarReader)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil // End of archive
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Ensure the directory exists.
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			// Create the file.
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// Copy the file contents.
			if _, err := io.Copy(f, tr); err != nil {
				f.Close() // Close file on copy error.
				return err
			}
			f.Close() // Close file successfully.
		}
	}
}
