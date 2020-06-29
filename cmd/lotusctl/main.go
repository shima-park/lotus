package main

import (
	cmd "github.com/shima-park/lotus/pkg/cmd/lotusctl"
)

var (
	VERSION string
	BRANCH  string
	COMMIT  string
	BUILT   string
)

func main() {
	cmd.Execute(VERSION, BRANCH, COMMIT, BUILT)
}
