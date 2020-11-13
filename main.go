package main

import (
	"fmt"
	"os"
)

func run() error {
	cli, err := newCLI("")
	if err != nil {
		return err
	}

	if err := cli.Run(); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}
