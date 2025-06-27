package list

import (
	"fmt"
	"sync"

	"github.com/aguidirh/lumen/pkg/catalog"
	"github.com/aguidirh/lumen/pkg/image"
)

// Package represents a package (operator) in the catalog.
type Package struct {
	Name           string
	DefaultChannel string
	Content        []*ChannelEntry `json:"content"`
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

// Results holds the structured output from a List operation.
type Results struct {
	Packages []Package
	Channels []Channel
	Versions []ChannelEntry
	Catalogs []string
}

type ListOptions struct {
	Catalog     string
	Catalogs    bool
	OCPVersion  string
	PackageName string
	ChannelName string
}

type ListImpl struct {
	opts ListOptions
}

func NewListImpl(opts ListOptions) *ListImpl {
	return &ListImpl{opts: opts}
}

// TODO: add a log to tell the user what is being done.
// List lists catalog contents based on the provided options.
func (l *ListImpl) List() (Results, error) {
	if !l.opts.Catalogs && l.opts.Catalog == "" {
		return Results{}, fmt.Errorf("catalog reference is required, unless --catalogs is specified")
	}

	var results Results
	var err error

	switch {
	case l.opts.Catalogs:
		if len(l.opts.OCPVersion) == 0 {
			return Results{}, fmt.Errorf("a version is required when listing catalogs")
		}
		results.Catalogs, err = catalogs(l.opts.OCPVersion)
	case len(l.opts.PackageName) > 0:
		if len(l.opts.ChannelName) > 0 {
			results.Versions, err = bundleVersionsByChannel(l.opts.Catalog, l.opts.PackageName, l.opts.ChannelName)
		} else {
			results.Channels, err = channelsByPackage(l.opts.Catalog, l.opts.PackageName)
		}
	case len(l.opts.Catalog) > 0:
		results.Packages, err = packagesByCatalog(l.opts.Catalog)
	default:
		return Results{}, fmt.Errorf("invalid set of options provided")
	}

	if err != nil {
		return Results{}, err
	}
	return results, nil
}

func catalogs(version string) ([]string, error) {
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
			if _, _, _, err := image.RemoteInfoFunc(imageRef); err == nil {
				catalogsCh <- imageRef
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

	return catalogs, nil
}

func packagesByCatalog(catalogRef string) ([]Package, error) {
	cfg, err := catalog.CatalogConfig(catalogRef)
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

	return packages, nil
}

func channelsByPackage(catalogRef, pkgName string) ([]Channel, error) {
	cfg, err := catalog.CatalogConfig(catalogRef)
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

	return channels, nil
}

func bundleVersionsByChannel(catalogRef, pkgName, channelName string) ([]ChannelEntry, error) {
	cfg, err := catalog.CatalogConfig(catalogRef)
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
			return entries, nil
		}
	}

	return nil, fmt.Errorf("channel %q for package %q not found", channelName, pkgName)
}
