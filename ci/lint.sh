#!/usr/bin/env bash

set -e

echo "Running $0 $*..."

linter_config=(
  --vendor
  --line-length 120
  --disable-all
  --enable deadcode
  --enable errcheck
  --enable goconst
  --enable gofmt
  --enable gosec
  --enable gosimple
  --enable ineffassign
  --enable interfacer
  --enable lll
  --enable maligned
  --enable megacheck
  --enable misspell
  --enable nakedret
  --enable staticcheck
  --enable structcheck
  --enable unconvert
  --enable unparam
  --enable unused
  --enable varcheck
  --enable vet
  --enable vetshadow
  # not compatible with our code style.
  --disable golint
  # gives a lot of false positives and does the same job as 'go build'
  --disable gotype
  --disable gotypex
  # gives many false positives and is useless considering the use of 'unused' linter
  --disable goimports
  # gives a lot of false positives in the tests
  --disable dupl
  # this linter is useless in our approach to the development process
  --disable gocyclo
  # golang tests should be run from the project CI
  --disable test
  --disable testify
  --exclude examples/weather.*
)

echo gometalinter \
  "${linter_config[@]}" \
  ./...

gometalinter \
  "${linter_config[@]}" \
  ./...
