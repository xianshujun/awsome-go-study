// Package mobile exposes a small gomobile-bind compatible local engine lifecycle.
package mobile

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/xianshujun/awsome-go-study/app/backend"
	"github.com/xianshujun/awsome-go-study/app/web"
)

var lifecycle struct {
	sync.Mutex
	server *backend.Server
}

// Start uses the app-private directory supplied by Android and returns a tokenized loopback URL.
func Start(dataDir string) (string, error) {
	lifecycle.Lock()
	defer lifecycle.Unlock()
	if lifecycle.server != nil {
		return lifecycle.server.URL(), nil
	}
	if strings.TrimSpace(dataDir) == "" {
		return "", errors.New("请提供应用私有数据目录")
	}
	server, err := backend.Start(backend.Options{DataDir: dataDir}, web.Handler())
	if err != nil {
		return "", err
	}
	lifecycle.server = server
	return server.URL(), nil
}
func Stop() {
	lifecycle.Lock()
	defer lifecycle.Unlock()
	if lifecycle.server == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = lifecycle.server.Stop(ctx)
	lifecycle.server = nil
}
