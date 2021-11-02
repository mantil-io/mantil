// Package kit is project standard library
//
// The packages that belong to the kit project need to be designed with the
// highest levels of portability in mind. They should not depend on other
// project packages.
//
// Idea from: https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html
//
// Check dependencises with `run list-packages`, current dependencies are:
// $ run list-packages ./kit
// bytes
// compress/gzip
// context
// crypto/ed25519
// crypto/rand
// encoding/json
// errors
// fmt
// github.com/alecthomas/jsonschema
// github.com/go-git/go-git/v5
// github.com/go-git/go-git/v5/plumbing/transport
// github.com/kataras/jwt
// github.com/pkg/errors
// github.com/qri-io/jsonschema
// gopkg.in/yaml.v2
// io
// io/fs
// io/ioutil
// log
// os
// os/exec
// os/signal
// path/filepath
// strings
// syscall
// time
package kit
