//go:generate mockgen -source=interface.go -destination=mock/interface_generated.go -package=mock

package list

import (
	"github.com/opencontainers/go-digest"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
)

// Logger defines the interface this package expects for logging operations.
type Logger interface {
	Infof(format string, args ...interface{})
	Info(args ...interface{})
	Debugf(format string, args ...interface{})
}

// Imager defines the interface this package expects for image operations.
type Imager interface {
	RemoteInfo(imageRef string) (string, string, digest.Digest, error)
}

// Cataloger defines the interface this package expects for catalog operations.
type Cataloger interface {
	CatalogConfig(imageRef string) (*declcfg.DeclarativeConfig, error)
}
