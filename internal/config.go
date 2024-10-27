package internal

import (
	"fmt"
	"log/slog"
	"os"
)

func getIcon(icon string) []byte {
	iconBytes, err := icons.ReadFile(fmt.Sprintf("resources/%s.ico", icon))
	if err != nil {
		slog.Error("Unable to load icon", "error", err)
		os.Exit(1)
	}
	return iconBytes
}
