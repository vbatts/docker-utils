package sum

import "github.com/docker/docker/pkg/tarsum"

var (
	// mapping for flag parsing
	tarsumVersions = map[string]tarsum.Version{
		"Version0":   tarsum.Version0,
		"Version1":   tarsum.Version1,
		"VersionDev": tarsum.VersionDev,
		"0":          tarsum.Version0,
		"1":          tarsum.Version1,
		"dev":        tarsum.VersionDev,
	}
)

func DetermineVersion(vstr string) (tarsum.Version, error) {
	for key, val := range tarsumVersions {
		if key == vstr {
			return val, nil
		}
	}
	return tarsum.Version(-1), tarsum.ErrVersionNotImplemented
}
