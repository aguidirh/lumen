// Package fsio_test contains tests for the fsio package.
package fsio_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/aguidirh/lumen/internal/pkg/fsio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFsIO(t *testing.T) {
	f := fsio.NewFsIO()
	assert.NotNil(t, f)
}

func TestFsIO_CopyFile(t *testing.T) {
	f := fsio.NewFsIO()
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcFile := filepath.Join(srcDir, "test.txt")
	dstFile := filepath.Join(dstDir, "test.txt")
	content := []byte("hello world")

	err := os.WriteFile(srcFile, content, 0644)
	require.NoError(t, err)

	err = f.CopyFile(srcFile, dstFile)
	require.NoError(t, err)

	copiedContent, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, content, copiedContent)
}

func TestFsIO_CopyDirectory(t *testing.T) {
	f := fsio.NewFsIO()
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create a source directory structure
	err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("file1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("file2"), 0644)
	require.NoError(t, err)

	err = f.CopyDirectory(srcDir, dstDir)
	require.NoError(t, err)

	// Verify the destination directory structure
	assert.FileExists(t, filepath.Join(dstDir, "file1.txt"))
	assert.FileExists(t, filepath.Join(dstDir, "subdir", "file2.txt"))

	content, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
	require.NoError(t, err)
	assert.Equal(t, []byte("file1"), content)

	content, err = os.ReadFile(filepath.Join(dstDir, "subdir", "file2.txt"))
	require.NoError(t, err)
	assert.Equal(t, []byte("file2"), content)
}

func TestFsIO_UntarFromStream(t *testing.T) {
	f := fsio.NewFsIO()
	destDir := t.TempDir()

	// Create a test tar archive in memory
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err := tw.WriteHeader(&tar.Header{Name: "testdir/", Typeflag: tar.TypeDir, Mode: 0755})
	require.NoError(t, err)
	err = tw.WriteHeader(&tar.Header{Name: "testdir/test.txt", Typeflag: tar.TypeReg, Size: 11, Mode: 0644})
	require.NoError(t, err)
	_, err = tw.Write([]byte("hello world"))
	require.NoError(t, err)
	err = tw.Close()
	require.NoError(t, err)

	// Test with plain tar
	err = f.UntarFromStream(bytes.NewReader(buf.Bytes()), destDir)
	require.NoError(t, err)

	assert.DirExists(t, filepath.Join(destDir, "testdir"))
	assert.FileExists(t, filepath.Join(destDir, "testdir", "test.txt"))
	content, err := os.ReadFile(filepath.Join(destDir, "testdir", "test.txt"))
	require.NoError(t, err)
	assert.Equal(t, []byte("hello world"), content)
}

func TestFsIO_UntarFromStream_Gzipped(t *testing.T) {
	f := fsio.NewFsIO()
	destDir := t.TempDir()

	// Create a test gzipped tar archive in memory
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)
	err := tw.WriteHeader(&tar.Header{Name: "testdir/", Typeflag: tar.TypeDir, Mode: 0755})
	require.NoError(t, err)
	err = tw.WriteHeader(&tar.Header{Name: "testdir/test.txt", Typeflag: tar.TypeReg, Size: 11, Mode: 0644})
	require.NoError(t, err)
	_, err = tw.Write([]byte("hello world"))
	require.NoError(t, err)
	err = tw.Close()
	require.NoError(t, err)
	err = gzw.Close()
	require.NoError(t, err)

	// Test with gzipped tar
	err = f.UntarFromStream(bytes.NewReader(buf.Bytes()), destDir)
	require.NoError(t, err)

	assert.DirExists(t, filepath.Join(destDir, "testdir"))
	assert.FileExists(t, filepath.Join(destDir, "testdir", "test.txt"))
	content, err := os.ReadFile(filepath.Join(destDir, "testdir", "test.txt"))
	require.NoError(t, err)
	assert.Equal(t, []byte("hello world"), content)
}

func TestFsIO_CopyDirectory_NonExistentSource(t *testing.T) {
	f := fsio.NewFsIO()
	srcDir := filepath.Join(t.TempDir(), "nonexistent")
	dstDir := t.TempDir()

	err := f.CopyDirectory(srcDir, dstDir)
	// filepath.Walk returns an error when the root path doesn't exist.
	require.Error(t, err)
}

func TestFsIO_CopyFile_NonExistentSource(t *testing.T) {
	f := fsio.NewFsIO()
	srcFile := filepath.Join(t.TempDir(), "nonexistent.txt")
	dstFile := filepath.Join(t.TempDir(), "dst.txt")

	err := f.CopyFile(srcFile, dstFile)
	require.Error(t, err)
}
