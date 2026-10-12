#!/usr/bin/env bash
set -euo pipefail
source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/common.sh"

if [[ "${1:-}" == '--help' ]]; then
    printf '%s\n' 'Usage: build-android.sh [debug|release]' \
        'Requires Go 1.25+, Java 17, Node.js 20.19+ or 22.12+, npm,' \
        'ANDROID_HOME with platforms;android-35 and build-tools;35.0.0,' \
        'and a compatible NDK (recommended 27.2.12479018) via ANDROID_NDK_HOME.' \
        'First run scripts/install-mobile-tools.sh to install pinned binding tools.' \
        'Builds all Android ABIs with min API 26; uses the checked-in Gradle wrapper.' \
        'Debug output: release/GoStudy-debug.apk (development signing).' \
        'Release output: release/GoStudy-release-unsigned.apk (sign before installation).'
    exit 0
fi
[[ $# -le 1 ]] || fail 'Use --help for usage.'
MODE="${1:-debug}"
case "$MODE" in debug) TASK=assembleDebug ;; release) TASK=assembleRelease ;; *) fail 'Choose debug or release.' ;; esac
require_go
if [[ -n "${JAVA_HOME:-}" ]]; then export PATH="$JAVA_HOME/bin:$PATH"; fi
require_command java
JAVA_VERSION="$(java -version 2>&1)"
[[ "$JAVA_VERSION" =~ version\ \"17[\.\"] ]] || fail 'Set JAVA_HOME and PATH to a Java 17 JDK.'
require_command javac
export ANDROID_HOME="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"
[[ -n "$ANDROID_HOME" && -d "$ANDROID_HOME" ]] || fail 'Set ANDROID_HOME to an existing Android SDK.'
[[ -f "$ANDROID_HOME/platforms/android-35/android.jar" ]] || fail 'Install platforms;android-35 with sdkmanager.'
[[ -d "$ANDROID_HOME/build-tools/35.0.0" ]] || fail 'Install build-tools;35.0.0 with sdkmanager.'
export ANDROID_NDK_HOME="${ANDROID_NDK_HOME:-${ANDROID_NDK_ROOT:-$ANDROID_HOME/ndk/27.2.12479018}}"
[[ -f "$ANDROID_NDK_HOME/source.properties" && -d "$ANDROID_NDK_HOME/toolchains/llvm/prebuilt" ]] \
    || fail 'Set ANDROID_NDK_HOME to a compatible NDK; recommended sdkmanager "ndk;27.2.12479018".'
[[ -x "$TOOLS_BIN/gomobile" && -x "$TOOLS_BIN/gobind" ]] \
    || fail 'Run scripts/install-mobile-tools.sh to install the pinned gomobile and gobind.'
for tool in gomobile gobind; do
    TOOL_INFO="$(go version -m "$TOOLS_BIN/$tool")"
    [[ "$TOOL_INFO" == *"golang.org/x/mobile"*"$MOBILE_VERSION"* ]] \
        || fail "$tool does not match the pinned version; rerun scripts/install-mobile-tools.sh."
done
export PATH="$TOOLS_BIN:$PATH"
export GOTOOLCHAIN=local
# 该版本的绑定生成器需要旧式类型别名表示。
export GODEBUG="${GODEBUG:+$GODEBUG,}gotypesalias=0"
[[ -d "$APP_DIR/mobile" && -d "$APP_DIR/backend" ]] || fail 'Go backend/mobile sources are missing.'
build_frontend
mkdir -p -- "$SCRIPT_DIR/.tools" "$APP_DIR/android/app/libs" "$APP_DIR/release"
STAGING="$(mktemp -d "$SCRIPT_DIR/.tools/android-bind.XXXXXX")"
trap 'rm -rf -- "$STAGING"' EXIT
# 在副本增加绑定依赖，防止 gomobile 改写主工程的 go.mod 和 go.sum。
cp -- "$APP_DIR/go.mod" "$STAGING/go.mod"
if [[ -f "$APP_DIR/go.sum" ]]; then cp -- "$APP_DIR/go.sum" "$STAGING/go.sum"; fi
cp -R -- "$APP_DIR/backend" "$APP_DIR/mobile" "$APP_DIR/web" "$STAGING/"
(
    cd -- "$STAGING"
    go get "golang.org/x/mobile/bind@$MOBILE_VERSION"
    GOFLAGS=-mod=mod gomobile bind -target=android -androidapi 26 \
        -o "$APP_DIR/android/app/libs/gostudy.aar" ./mobile
)
(cd -- "$APP_DIR/android"; ./gradlew --no-daemon "$TASK")
if [[ "$MODE" == debug ]]; then
    SOURCE_APK="$APP_DIR/android/app/build/outputs/apk/debug/app-debug.apk"
    DEST_APK="$APP_DIR/release/GoStudy-debug.apk"
else
    SOURCE_APK="$APP_DIR/android/app/build/outputs/apk/release/app-release-unsigned.apk"
    DEST_APK="$APP_DIR/release/GoStudy-release-unsigned.apk"
fi
cp -- "$SOURCE_APK" "$DEST_APK"
printf 'Android package: %s\n' "$DEST_APK"
