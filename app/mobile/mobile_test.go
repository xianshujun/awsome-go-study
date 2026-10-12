package mobile

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestLifecycle(t *testing.T) {
	Stop()
	if _, err := Start(""); err == nil {
		t.Fatal("private data directory required")
	}
	address, err := Start(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer Stop()
	again, err := Start(t.TempDir())
	if err != nil || again != address {
		t.Fatal("start must be idempotent")
	}
	u, err := url.Parse(address)
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(u.Fragment, "token=")
	u.Fragment = ""
	u.Path = "/api/bootstrap"
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}
	Stop()
	Stop()
	second, err := Start(t.TempDir())
	if err != nil || second == address {
		t.Fatal("restart must rotate token")
	}
}
