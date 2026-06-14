#!/usr/bin/env bash
# Build and run the Counter app on the iOS Simulator.
# Requires the FULL Xcode app (not just Command Line Tools).
set -euo pipefail

cd "$(dirname "$0")"

SCHEME="Counter"
BUNDLE_ID="com.exploremelon.trigger"
SIM_NAME="${SIM_NAME:-iPhone 15}"

# 0. Sanity: full Xcode must be selected.
if ! xcodebuild -version >/dev/null 2>&1; then
  echo "ERROR: full Xcode is required. Install it from the App Store, then run:"
  echo "  sudo xcode-select -s /Applications/Xcode.app/Contents/Developer"
  echo "  sudo xcodebuild -license accept"
  exit 1
fi

# 1. (Re)generate the project from project.yml.
command -v xcodegen >/dev/null && xcodegen generate

# 2. Boot the simulator (ignore error if already booted).
xcrun simctl boot "$SIM_NAME" 2>/dev/null || true
open -a Simulator

# 3. Build for the simulator.
DERIVED="build"
xcodebuild \
  -project Counter.xcodeproj \
  -scheme "$SCHEME" \
  -sdk iphonesimulator \
  -destination "platform=iOS Simulator,name=$SIM_NAME" \
  -derivedDataPath "$DERIVED" \
  build

# 4. Install and launch on the booted simulator.
APP_PATH="$DERIVED/Build/Products/Debug-iphonesimulator/$SCHEME.app"
xcrun simctl install booted "$APP_PATH"
xcrun simctl launch booted "$BUNDLE_ID"
echo "Launched $BUNDLE_ID on $SIM_NAME"
