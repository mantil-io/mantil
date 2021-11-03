#!/usr/bin/env bash -e

go list -f '{{range $imp := .Imports}}{{printf "%s\n" $imp}}{{end}}' ./$1/... | sort | uniq
