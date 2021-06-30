// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func Test() error {
	args := []string{"test", "./...", "-cover"}
	if mg.Verbose() {
		args = append(args, "-v")
	}

	return sh.Run("go", args...)
}

func Lint() error {
	return sh.Run("golangci-lint", "run", "--timeout=10m")
}

func UnitTests() error {
	args := []string{"test", "./...", "-cover", "-short"}
	if mg.Verbose() {
		args = append(args, "-v")
	}

	return sh.Run("go", args...)
}
