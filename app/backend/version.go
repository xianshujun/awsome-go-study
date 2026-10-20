package backend

import (
	"strconv"
	"strings"
	"time"
)

/*
Version 由构建脚本用 -ldflags 注入：

	-X github.com/xianshujun/awsome-go-study/app/backend/Version=0.11.1

源码里直接 go run 时是 dev，更新检查会跳过程序本身的比较（内容更新仍然可用）。
*/
var Version = "dev"

// parseVersion 把 v0.8.1 / 0.8.1 解析成可比较的数字；dev、带后缀的构建都返回 false。
func parseVersion(raw string) ([3]int, bool) {
	var out [3]int
	raw = strings.TrimPrefix(strings.TrimSpace(raw), "v")
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return out, false
		}
		out[i] = value
	}
	return out, true
}

// versionNewer 判断 candidate 是否比 current 新；任一侧不可解析时都返回 false，
// 宁可不提示，也不要让 dev 构建或异常版本号把用户引到错误的方向。
func versionNewer(candidate, current string) bool {
	next, ok := parseVersion(candidate)
	if !ok {
		return false
	}
	now, ok := parseVersion(current)
	if !ok {
		return false
	}
	for i := 0; i < 3; i++ {
		if next[i] != now[i] {
			return next[i] > now[i]
		}
	}
	return false
}

/** 版本信息里给界面看的那一份：当前版本 + 构建时间（代理用）。 */
func versionInfo() string {
	if Version == "" {
		return "dev"
	}
	return Version
}

// 自动检查的最小间隔：一天最多一次，避免每次启动都联网。
const autoCheckInterval = 24 * time.Hour
