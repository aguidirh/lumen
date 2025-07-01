//go:generate mockgen -source=interface.go -destination=mock/interface_generated.go -package=mock

package printer

// Logger defines the interface this package expects for logging.
type Logger interface {
	Debugf(format string, args ...interface{})
}
