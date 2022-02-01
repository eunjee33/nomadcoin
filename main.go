package main

import (
	"github.com/eunjee33/nomadcoin/cli"
	"github.com/eunjee33/nomadcoin/db"
)

func main() {
	defer db.Close()
	cli.Start()
}