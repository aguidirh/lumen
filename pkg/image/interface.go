//go:generate mockgen -source=interface.go -destination=mock/interface_generated.go -package=mock

package image

// Logger defines the interface this package expects for logging.
type Logger interface {
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}
