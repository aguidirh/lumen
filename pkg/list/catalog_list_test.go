package list

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/aguidirh/lumen/pkg/image"
	"github.com/opencontainers/go-digest"
)

// Use a known public catalog for all tests to ensure consistency and avoid auth issues.
const testCatalog = "registry.redhat.io/redhat/community-operator-index:v4.16"

// TODO: create fake catalog image for testing. The current tests are depending on internet to pull the catalogs.
func TestMain(m *testing.M) {
	// Let the user know which catalog is being used.
	// In a real-world CI/CD, this could be configured.
	fmt.Printf("INFO: Using catalog image for tests: %s\n", testCatalog)
	// Run all tests
	exitCode := m.Run()

	// Clean up the cache directory generated during tests.
	os.RemoveAll("working-dir")

	// Exit with the appropriate code
	os.Exit(exitCode)
}

func TestList_Operators(t *testing.T) {
	opts := ListOptions{
		Catalog: testCatalog,
	}
	listImpl := NewListImpl(opts)
	results, err := listImpl.List()
	if err != nil {
		t.Fatalf("List(operators) error = %v", err)
	}

	if len(results.Packages) == 0 {
		t.Fatal("expected operators, got none")
	}

	// Spot check for a few well-known operators in this catalog version.
	expectedOperators := []string{
		"cert-manager",
		"grafana-operator",
		"prometheus",
	}

	var operatorNames []string
	for _, op := range results.Packages {
		operatorNames = append(operatorNames, op.Name)
	}

	for _, expected := range expectedOperators {
		if !slices.Contains(operatorNames, expected) {
			t.Errorf("expected operator %q not found in catalog", expected)
		}
	}
}

func TestList_Channels(t *testing.T) {
	pkgName := "prometheus" // This package is known to be in the test catalog.
	opts := ListOptions{
		Catalog:     testCatalog,
		PackageName: pkgName,
	}
	listImpl := NewListImpl(opts)
	results, err := listImpl.List()
	if err != nil {
		t.Fatalf("List(channels) for package %s error = %v", pkgName, err)
	}

	if len(results.Channels) == 0 {
		t.Fatalf("expected to find channels for package %s, but got none", pkgName)
	}

	// Spot check for the 'beta' channel.
	var channelNames []string
	for _, ch := range results.Channels {
		channelNames = append(channelNames, ch.Name)
	}
	if !slices.Contains(channelNames, "beta") {
		t.Errorf("expected to find channel 'beta' for package %s, but it was not found", pkgName)
	}
}

func TestList_VersionsInChannel(t *testing.T) {
	pkgName := "prometheus"
	channelName := "beta" // The 'beta' channel is a good candidate for this test.
	opts := ListOptions{
		Catalog:     testCatalog,
		PackageName: pkgName,
		ChannelName: channelName,
	}

	listImpl := NewListImpl(opts)
	results, err := listImpl.List()
	if err != nil {
		t.Fatalf("List(versions) for package %s in channel %s error = %v", pkgName, channelName, err)
	}

	if len(results.Versions) == 0 {
		t.Errorf("expected at least one version in package %s, channel %s, but got none", pkgName, channelName)
	}
}

func TestList_Catalogs(t *testing.T) {
	// Temporarily replace the real RemoteInfo function with a mock for this test.
	originalRemoteInfo := image.RemoteInfoFunc
	defer func() { image.RemoteInfoFunc = originalRemoteInfo }()

	// Mock RemoteInfo to simulate which catalogs exist.
	image.RemoteInfoFunc = func(imageRef string) (string, string, digest.Digest, error) {
		// Pretend only these two catalogs exist for the test.
		if imageRef == "registry.redhat.io/redhat/redhat-operator-index:v4.16" ||
			imageRef == "registry.redhat.io/redhat/community-operator-index:v4.16" {
			return "", "", "", nil // Success
		}
		// For all other catalogs, return an error to simulate that they don't exist.
		return "", "", "", fmt.Errorf("image not found")
	}

	opts := ListOptions{
		Catalogs:   true,
		OCPVersion: "4.16",
	}

	listImpl := NewListImpl(opts)
	results, err := listImpl.List()
	if err != nil {
		t.Fatalf("List(catalogs) returned an unexpected error: %v", err)
	}

	expectedCatalogs := []string{
		"registry.redhat.io/redhat/redhat-operator-index:v4.16",
		"registry.redhat.io/redhat/community-operator-index:v4.16",
	}

	if len(results.Catalogs) != len(expectedCatalogs) {
		t.Errorf("expected %d catalogs, but got %d", len(expectedCatalogs), len(results.Catalogs))
	}

	for _, expected := range expectedCatalogs {
		if !slices.Contains(results.Catalogs, expected) {
			t.Errorf("expected catalog %q not found in results", expected)
		}
	}
}
