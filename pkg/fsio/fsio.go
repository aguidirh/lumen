package fsio

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyDirectory recursively copies a directory from src to dst.
func CopyDirectory(src, dst string) error {
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
		return CopyFile(path, dstPath)
	})
}

// CopyFile copies a single file from src to dst.
func CopyFile(src, dst string) error {
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
func UntarFromStream(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err == nil {
		// If it's a valid gzip stream, use the gzip reader.
		r = gzr
		defer gzr.Close()
	}
	// If it's not a gzip stream, err will be non-nil, and we'll proceed with the original reader `r`,
	// effectively treating it as a plain tar stream.

	tr := tar.NewReader(r)

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
