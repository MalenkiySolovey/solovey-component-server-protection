// Package serverprotection supplies the in-process Server Protection source descriptor.
package serverprotection

import _ "embed"

//go:embed component.json
var manifest []byte

// ManifestJSON returns an isolated copy of the canonical contribution descriptor.
func ManifestJSON() []byte { return append([]byte(nil), manifest...) }
