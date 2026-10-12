package io.gostudy.local;

import android.app.Activity;
import android.graphics.Color;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.view.Gravity;
import android.view.View;
import android.view.WindowInsets;
import android.webkit.RenderProcessGoneDetail;
import android.webkit.WebResourceRequest;
import android.webkit.WebResourceResponse;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.FrameLayout;
import android.widget.TextView;

import java.io.ByteArrayInputStream;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import mobile.Mobile;

public final class MainActivity extends Activity {
    // 与前端 --paper 保持一致，避免启动瞬间闪一次旧底色。
    private static final int BACKGROUND = Color.rgb(246, 247, 247);
    // 启动和停止串行执行；旋转屏幕不会启动第二个 Go 引擎。
    private static final ExecutorService ENGINE = Executors.newSingleThreadExecutor();
    private static String engineUrl;

    private FrameLayout container;
    private WebView webView;
    private boolean destroyed;

    @Override
    public void onCreate(Bundle state) {
        super.onCreate(state);
        getWindow().setStatusBarColor(BACKGROUND);
        getWindow().setNavigationBarColor(BACKGROUND);
        getWindow().getDecorView().setSystemUiVisibility(
                View.SYSTEM_UI_FLAG_LIGHT_STATUS_BAR | View.SYSTEM_UI_FLAG_LIGHT_NAVIGATION_BAR);
        if (Build.VERSION.SDK_INT >= 29) {
            getWindow().setNavigationBarContrastEnforced(false);
        }
        if (Build.VERSION.SDK_INT >= 30) {
            getWindow().setDecorFitsSystemWindows(false);
        }
        container = new FrameLayout(this);
        container.setBackgroundColor(BACKGROUND);
        // Android 15 强制边到边布局；给 WebView 留出系统栏和键盘空间。
        container.setOnApplyWindowInsetsListener((view, insets) -> {
            if (Build.VERSION.SDK_INT >= 30) {
                android.graphics.Insets safe = insets.getInsets(
                        WindowInsets.Type.systemBars() | WindowInsets.Type.displayCutout()
                                | WindowInsets.Type.ime());
                view.setPadding(safe.left, safe.top, safe.right, safe.bottom);
            }
            return insets;
        });
        setContentView(container);
        container.requestApplyInsets();
        showStatus(R.string.starting);

        String dataDir = getFilesDir().getAbsolutePath();
        ENGINE.execute(() -> {
            try {
                if (engineUrl == null) {
                    String started = Mobile.start(dataDir);
                    if (!validLaunchUrl(Uri.parse(started))) {
                        Mobile.stop();
                        throw new IllegalStateException("Invalid local engine URL");
                    }
                    engineUrl = started;
                }
                String url = engineUrl;
                runOnUiThread(() -> {
                    if (!destroyed) {
                        openLocalPage(url);
                    }
                });
            } catch (Exception failure) {
                // 不记录含鉴权令牌的地址或可能包含用户数据的错误。
                runOnUiThread(() -> {
                    if (!destroyed) {
                        showStatus(R.string.startup_failed);
                    }
                });
            }
        });
    }

    private static boolean validLaunchUrl(Uri uri) {
        String fragment = uri.getFragment();
        return "http".equals(uri.getScheme())
                && "127.0.0.1".equals(uri.getHost())
                && uri.getPort() > 0 && uri.getPort() <= 65535
                && uri.getUserInfo() == null
                && fragment != null && fragment.startsWith("token=")
                && fragment.length() > "token=".length();
    }

    private static boolean sameOrigin(Uri candidate, Uri origin) {
        return origin.getScheme().equals(candidate.getScheme())
                && origin.getHost().equals(candidate.getHost())
                && origin.getPort() == candidate.getPort()
                && candidate.getUserInfo() == null;
    }

    private void showStatus(int resource) {
        container.removeAllViews();
        TextView message = new TextView(this);
        message.setText(resource);
        message.setTextColor(Color.rgb(17, 24, 27));
        message.setTextSize(17);
        message.setGravity(Gravity.CENTER);
        message.setPadding(32, 32, 32, 32);
        container.addView(message, new FrameLayout.LayoutParams(-1, -1));
    }

    private void openLocalPage(String url) {
        Uri origin = Uri.parse(url);
        WebView.setWebContentsDebuggingEnabled(BuildConfig.DEBUG);
        webView = new WebView(this);
        webView.setBackgroundColor(BACKGROUND);
        WebSettings settings = webView.getSettings();
        settings.setUserAgentString(settings.getUserAgentString() + " GoStudyAndroid/1.0");
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        settings.setAllowFileAccess(false);
        settings.setAllowContentAccess(false);
        settings.setAllowFileAccessFromFileURLs(false);
        settings.setAllowUniversalAccessFromFileURLs(false);
        settings.setMixedContentMode(WebSettings.MIXED_CONTENT_NEVER_ALLOW);
        settings.setJavaScriptCanOpenWindowsAutomatically(false);
        settings.setSupportMultipleWindows(false);
        settings.setMediaPlaybackRequiresUserGesture(true);
        settings.setCacheMode(WebSettings.LOAD_NO_CACHE);
        // 使用系统确认弹窗支持页面内的重置操作；不启用文件选择等原生扩展。
        webView.setWebChromeClient(new android.webkit.WebChromeClient());
        webView.setWebViewClient(new WebViewClient() {
            @Override
            public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest request) {
                return !sameOrigin(request.getUrl(), origin);
            }

            @Override
            public boolean shouldOverrideUrlLoading(WebView view, String target) {
                return !sameOrigin(Uri.parse(target), origin);
            }

            @Override
            public WebResourceResponse shouldInterceptRequest(WebView view, WebResourceRequest request) {
                if (sameOrigin(request.getUrl(), origin)) {
                    return null;
                }
                // 外链和远程资源不能绕过导航检查；AI 请求由 Go 后端发起。
                return new WebResourceResponse("text/plain", "UTF-8", 403, "Forbidden",
                        java.util.Collections.emptyMap(),
                        new ByteArrayInputStream("Blocked external resource".getBytes(StandardCharsets.UTF_8)));
            }

            @Override
            public void onReceivedError(WebView view, WebResourceRequest request,
                                        android.webkit.WebResourceError error) {
                if (request.isForMainFrame()) {
                    showStatus(R.string.page_failed);
                }
            }

            @Override
            public boolean onRenderProcessGone(WebView view, RenderProcessGoneDetail detail) {
                container.removeView(view);
                view.destroy();
                webView = null;
                showStatus(R.string.page_failed);
                return true;
            }
        });
        // 不暴露 JavaScript 接口，也不接受文件选择或外部 Intent。
        container.removeAllViews();
        container.addView(webView, new FrameLayout.LayoutParams(-1, -1));
        webView.loadUrl(url);
    }

    @Override
    public void onBackPressed() {
        if (webView != null && webView.canGoBack()) {
            webView.goBack();
        } else {
            super.onBackPressed();
        }
    }

    @Override
    protected void onDestroy() {
        destroyed = true;
        if (webView != null) {
            container.removeView(webView);
            webView.stopLoading();
            webView.destroy();
            webView = null;
        }
        // 旋转和后台切换保留引擎；只有真正关闭页面时才释放监听端口。
        if (isFinishing()) {
            ENGINE.execute(() -> {
                try {
                    Mobile.stop();
                } finally {
                    engineUrl = null;
                }
            });
        }
        super.onDestroy();
    }
}
