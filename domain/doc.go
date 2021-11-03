// Package domain provides core mantil domain model. This package is distilled
// domain knowledge. It should use shared language that allows developers and
// domain experts to collaborate effectively.
//
// This is the place where all domain decisions are made. It should not depend
// on other packages, except for the standard library, and project kit package.
//
// Current dependencies:
// $ run list-packages domain
// crypto/rand
// encoding/base32
// encoding/json
// fmt
// github.com/json-iterator/go
// github.com/mantil-io/mantil/kit/gz
// github.com/mantil-io/mantil/kit/schema
// github.com/mantil-io/mantil/kit/token
// github.com/pkg/errors
// gopkg.in/yaml.v2
// io
// io/ioutil
// os
// os/user
// path
// path/filepath
// reflect
// regexp
// strings
// time
package domain
