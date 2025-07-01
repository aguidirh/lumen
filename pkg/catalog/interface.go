//go:generate mockgen -source=interface.go -destination=mock/interface_generated.go -package=mock

package catalog

import (
	"io"

	"github.com/opencontainers/go-digest"
)

// Logger defines the interface this package expects for logging.
type Logger interface {
	Infof(format string, args ...interface{})
	Info(args ...interface{})
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
}

// Imager defines the interface this package expects for image operations.
type Imager interface {
	RemoteInfo(imageRef string) (string, string, digest.Digest, error)
	CopyToOci(imageRef, ociDir string) (string, error)
}

// FsIO defines the interface this package expects for filesystem I/O.
type FsIO interface {
	CopyDirectory(src, dst string) error
	UntarFromStream(r io.Reader, dest string) error
}
