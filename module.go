package archiver

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(CaddyArchiver{})
}

type CaddyArchiver struct {
	Root   string `json:"root"`
	logger *zap.Logger
}

func (CaddyArchiver) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.archiver",
		New: func() caddy.Module { return new(CaddyArchiver) },
	}
}

func (a *CaddyArchiver) Provision(context caddy.Context) error {
	a.logger = context.Logger(a)

	a.logger.Debug("Provisioning done")
	info, err := os.Stat(a.Root)
	if err != nil {
		return fmt.Errorf("error opening root folder %v %v", a.Root, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("the specified root is not a valid directory %v", info.Name())
	}
	a.Root, _ = filepath.Abs(a.Root)

	return nil
}
