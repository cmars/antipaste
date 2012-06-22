package main

import (
	"fmt"
	"os"
	"github.com/cmars/antipaste"
)

func main() {
	app := antipaste.NewApp()
	err := app.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
