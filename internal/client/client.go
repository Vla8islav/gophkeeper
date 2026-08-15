package client

import (
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/Vla8islav/gophkeeper/internal/config"
)

func Execute() error {
	cfg, rest, err := config.ReadFlagsClient(os.Args[1:], zap.NewNop())
	if err != nil {
		return err
	}
	if len(rest) == 0 {
		return errors.New("usage: gophkeeper [flags] <command> [args]\ncommands: register, login, list")
	}

	cmd, cmdArgs := rest[0], rest[1:]
	switch cmd {
	case "register":
		return runRegister(cfg, cmdArgs)
	case "login":
		return runLogin(cfg, cmdArgs)
	case "list":
		return runList(cfg, cmdArgs)
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}
