package main

import (
	cmd "github.com/shima-park/lotus/pkg/cmd/lotussrv"
	_ "github.com/shima-park/lotus/pkg/component/include"
)

func main() {
	cmd.Execute()
}
