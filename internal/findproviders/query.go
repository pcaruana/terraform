package findproviders

import (
	"fmt"

	"github.com/apparentlymart/go-versions/versions"

	"github.com/hashicorp/terraform/addrs"
)

// AvailableVersions returns all of the versions available for the provider
// with the given address, or an error if that result cannot be determined.
//
// If the request fails, the returned error might be an value of
// ErrHostNoProviders, ErrHostUnreachable, ErrUnauthenticated,
// ErrProviderNotKnown, or ErrQueryFailed. Callers must be defensive and
// expect errors of other types too, to allow for future expansion.
func (s *Source) AvailableVersions(provider addrs.Provider) (VersionList, error) {
	client, err := s.registryClient(provider.Hostname)
	if err != nil {
		return nil, err
	}

	versionStrs, err := client.ProviderVersions(provider)
	if err != nil {
		return nil, err
	}

	if len(versionStrs) == 0 {
		return nil, nil
	}

	ret := make(versions.List, len(versionStrs))
	for i, str := range versionStrs {
		v, err := versions.ParseVersion(str)
		if err != nil {
			return nil, ErrQueryFailed{
				Provider: provider,
				Wrapped:  fmt.Errorf("registry response includes invalid version string %q: %s", str, err),
			}
		}
		ret[i] = v
	}
	ret.Sort() // lowest precedence first, preserving order when equal precedence
	return ret, nil
}

// DownloadLocation returns metadata about the location and capabilities of
// a distribution package for a particular provider at a particular version
// targeting a particular platform.
//
// Callers of DownloadLocation should first call AvailableVersions and pass
// one of the resulting versions to this function. This function cannot
// distinguish between a version that is not available and an unsupported
// target platform, so if it encounters either case it will return an error
// suggesting that the target platform isn't supported under the assumption
// that the caller already checked that the version is available at all.
//
// To find a package suitable for the platform where the provider installation
// process is running, set the "target" argument to
// findproviders.CurrentPlatform.
//
// If the request fails, the returned error might be an value of
// ErrHostNoProviders, ErrHostUnreachable, ErrUnauthenticated,
// ErrPlatformNotSupported, or ErrQueryFailed. Callers must be defensive and
// expect errors of other types too, to allow for future expansion.
func (s *Source) DownloadLocation(provider addrs.Provider, version Version, target Platform) (PackageMeta, error) {
	client, err := s.registryClient(provider.Hostname)
	if err != nil {
		return PackageMeta{}, err
	}

	return client.PackageMeta(provider, version, target)
}
