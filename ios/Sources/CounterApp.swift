import SwiftUI

/// Entry point. The whole app is a single full-screen web view that loads the
/// same deployed site the Android TWA points at:
/// https://trigger.exploremelon.com
@main
struct CounterApp: App {
    private let appURL = URL(string: "https://trigger.exploremelon.com")!

    var body: some Scene {
        WindowGroup {
            WebContainerView(url: appURL)
                .ignoresSafeArea()              // edge-to-edge, no browser chrome
                .preferredColorScheme(.dark)    // matches the app's dark theme
        }
    }
}
