package main

import (
	"context"
	"os"

	"github.com/shellcell/cnvrt/internal/bootstrap"
)

func main() {
	app := bootstrap.New()
	os.Exit(app.Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
