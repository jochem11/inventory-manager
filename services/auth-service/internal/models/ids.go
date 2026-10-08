package models

import "github.com/segmentio/ksuid"

// newID returns a KSUID: 27 base62 chars that sort by creation time, the same
// IDs the user-service uses. ID columns use
// `char(27) character set ascii collate ascii_bin` so ordering and
// uniqueness are case-sensitive, as the KSUID spec requires.
func newID() string {
	return ksuid.New().String()
}
