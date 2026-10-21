// Command cahut runs a Cahut home: the server that keeps, shares and logs
// what AIs make.
package main

import (
	"os"

	"cahut.cloud/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
