package list

import (
	"fmt"
	"sync"

	"github.com/opencontainers/go-digest"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
)

// imager defines the interface this package expects for image operations.
type imager interface {
	RemoteInfo(imageRef string) (string, string, digest.Digest, error)
}

// Package represents a package (operator) in the catalog.
type Package struct {
	Name           string
	DefaultChannel string
}

// Channel represents a channel in a package.
type Channel struct {
	Name string
	Head string
}

// ChannelEntry represents a bundle version in a channel.
type ChannelEntry struct {
	Name string
}

// loger defines the interface this package expects for logging operations.
type loger interface {
	Infof(format string, args ...interface{})
	Info(args ...interface{})
	Debugf(format string, args ...interface{})
}

// cataloger defines the interface this package expects for catalog operations.
type cataloger interface {
	CatalogConfig(imageRef string) (*declcfg.DeclarativeConfig, error)
}

// CatalogLister holds dependencies for listing operations.
type CatalogLister struct {
	log       loger
	cataloger cataloger
	imager    imager
}

// NewCatalogLister creates a new Lister.
func NewCatalogLister(log loger, cataloger cataloger, imager imager) *CatalogLister {
	return &CatalogLister{
		log:       log,
		cataloger: cataloger,
		imager:    imager,
	}
}

func (c *CatalogLister) Catalogs(version string) ([]string, error) {
	if len(version) == 0 {
		return nil, fmt.Errorf("a version is required when listing catalogs")
	}
	c.log.Debugf("Searching for operator catalogs for version %s...", version)
	repos := []string{
		"registry.redhat.io/redhat/redhat-operator-index",
		"registry.redhat.io/redhat/certified-operator-index",
		"registry.redhat.io/redhat/community-operator-index",
		"registry.redhat.io/redhat/redhat-marketplace-index",
	}

	tag := "v" + version
	var wg sync.WaitGroup
	catalogsCh := make(chan string, len(repos))

	for _, repo := range repos {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			imageRef := fmt.Sprintf("%s:%s", repo, tag)
			if _, _, _, err := c.imager.RemoteInfo(imageRef); err == nil {
				catalogsCh <- imageRef
			} else {
				c.log.Debugf("Catalog %s not found, skipping...", imageRef)
			}
		}(repo)
	}

	wg.Wait()
	close(catalogsCh)

	var catalogs []string
	for catalog := range catalogsCh {
		catalogs = append(catalogs, catalog)
	}

	if len(catalogs) == 0 {
		return nil, fmt.Errorf("no catalogs found for version %s", version)
	}

	c.log.Debugf("Found %d catalogs.", len(catalogs))
	return catalogs, nil
}

func (c *CatalogLister) PackagesByCatalog(catalogRef string) ([]Package, error) {
	if catalogRef == "" {
		return nil, fmt.Errorf("catalog reference is required")
	}
	c.log.Debugf("Listing packages for catalog %s...", catalogRef)
	cfg, err := c.cataloger.CatalogConfig(catalogRef)
	if err != nil {
		return nil, err
	}

	var packages []Package
	for _, pkg := range cfg.Packages {
		packages = append(packages, Package{
			Name:           pkg.Name,
			DefaultChannel: pkg.DefaultChannel,
		})
	}

	c.log.Debugf("Found %d packages.", len(packages))
	return packages, nil
}

func (c *CatalogLister) ChannelsByPackage(catalogRef, pkgName string) ([]Channel, error) {
	if catalogRef == "" || pkgName == "" {
		return nil, fmt.Errorf("catalog reference and package name are required")
	}
	c.log.Debugf("Listing channels for package %s in catalog %s...", pkgName, catalogRef)
	cfg, err := c.cataloger.CatalogConfig(catalogRef)
	if err != nil {
		return nil, err
	}

	var channels []Channel
	for _, ch := range cfg.Channels {
		if ch.Package == pkgName {
			var head string
			if len(ch.Entries) > 0 {
				head = ch.Entries[len(ch.Entries)-1].Name
			}
			channels = append(channels, Channel{
				Name: ch.Name,
				Head: head,
			})
		}
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("package %q not found in catalog", pkgName)
	}

	c.log.Debugf("Found %d channels.", len(channels))
	return channels, nil
}

func (c *CatalogLister) BundleVersionsByChannel(catalogRef, pkgName, channelName string) ([]ChannelEntry, error) {
	if catalogRef == "" || pkgName == "" || channelName == "" {
		return nil, fmt.Errorf("catalog reference, package name, and channel name are required")
	}
	c.log.Debugf("Listing bundle versions for channel %s in package %s, catalog %s...", channelName, pkgName, catalogRef)
	cfg, err := c.cataloger.CatalogConfig(catalogRef)
	if err != nil {
		return nil, err
	}

	for _, ch := range cfg.Channels {
		if ch.Package == pkgName && ch.Name == channelName {
			var entries []ChannelEntry
			for _, entry := range ch.Entries {
				entries = append(entries, ChannelEntry{
					Name: entry.Name,
				})
			}
			c.log.Debugf("Found %d bundle versions.", len(entries))
			return entries, nil
		}
	}

	return nil, fmt.Errorf("channel %q for package %q not found", channelName, pkgName)
}
