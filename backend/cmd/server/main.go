// Command server runs the calculator REST API.
//
// This file is the composition root: it only wires dependencies together.
package main

import (
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("calculator service scaffold")
}
