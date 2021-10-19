#!/usr/bin/env bash -e

go test -v && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
