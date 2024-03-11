//go:build !local

package main

import (
	"context"

	"github.com/GolangUA/telegram-butler/pkg/module/logger"
)

func preSetup(_ context.Context, _ logger.Logger) error {
	return nil
}
