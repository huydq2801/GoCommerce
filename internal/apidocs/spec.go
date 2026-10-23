package apidocs

import _ "embed"

// Spec is served without relying on the process working directory.
//
//go:embed openapi.json
var Spec []byte
