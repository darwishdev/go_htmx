import SwiftUI
import WebKit

/// A SwiftUI wrapper around WKWebView. This is the "native shell" — the place
/// where you'd later add a JS<->Swift bridge for push, biometrics, share, etc.
struct WebContainerView: UIViewRepresentable {
    let url: URL

    func makeCoordinator() -> Coordinator { Coordinator() }

    func makeUIView(context: Context) -> WKWebView {
        let config = WKWebViewConfiguration()
        config.allowsInlineMediaPlayback = true
        config.defaultWebpagePreferences.allowsContentJavaScript = true

        let webView = WKWebView(frame: .zero, configuration: config)
        webView.navigationDelegate = context.coordinator
        webView.allowsBackForwardNavigationGestures = true   // swipe-back
        webView.scrollView.contentInsetAdjustmentBehavior = .never
        webView.isOpaque = false
        webView.backgroundColor = UIColor(red: 0.06, green: 0.09, blue: 0.16, alpha: 1) // #0f172a
        return webView
    }

    func updateUIView(_ webView: WKWebView, context: Context) {
        // Load once; avoids reloading on every SwiftUI update.
        if webView.url == nil {
            webView.load(URLRequest(url: url))
        }
    }

    /// Keeps in-app navigation inside the web view (only the initial host).
    final class Coordinator: NSObject, WKNavigationDelegate {
        func webView(_ webView: WKWebView,
                     decidePolicyFor navigationAction: WKNavigationAction,
                     decisionHandler: @escaping (WKNavigationActionPolicy) -> Void) {
            decisionHandler(.allow)
        }
    }
}
