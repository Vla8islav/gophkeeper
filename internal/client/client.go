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
		return errors.New("usage: gophkeeper [flags] <command> [args]\n" +
			"commands: register, login, list, add, get, update, delete, sync",
		)
	}

	cmd, cmdArgs := rest[0], rest[1:]
	switch cmd {
	case "register":
		return runRegister(cfg, cmdArgs)
	case "login":
		return runLogin(cfg, cmdArgs)
	case "list":
		return runList(cfg, cmdArgs)
	case "add":
		return runAdd(cfg, cmdArgs)
	case "get":
		return runGet(cfg, cmdArgs)
	case "sync":
		return runSync(cfg, cmdArgs)
	case "update":
		return runUpdate(cfg, cmdArgs)
	case "delete":
		return runDelete(cfg, cmdArgs)
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}
