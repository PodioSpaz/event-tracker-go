#!/usr/bin/env fish

set APP_NAME "Event Tracker"
set BUNDLE_DIR "dist/$APP_NAME.app"
set VERSION "0.1.0"

echo "Building Event Tracker v$VERSION for macOS..."

# Clean previous app bundle only (preserve other dist files)
rm -rf "dist/$APP_NAME.app"

# Build for macOS ARM64
echo "Building for Apple Silicon (ARM64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$VERSION" -o bin/event-tracker-arm64 ./cmd/app

# Create bundle structure
mkdir -p "$BUNDLE_DIR/Contents/MacOS"
mkdir -p "$BUNDLE_DIR/Contents/Resources"

# Copy binary
cp bin/event-tracker-arm64 "$BUNDLE_DIR/Contents/MacOS/event-tracker"
chmod +x "$BUNDLE_DIR/Contents/MacOS/event-tracker"

# Create Info.plist
echo '<?xml version="1.0" encoding="UTF-8"?>' > "$BUNDLE_DIR/Contents/Info.plist"
echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '<plist version="1.0">' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '<dict>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <key>CFBundleExecutable</key>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <string>event-tracker</string>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <key>CFBundleIdentifier</key>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <string>com.podiospaz.event-tracker</string>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <key>CFBundleName</key>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <string>Event Tracker</string>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <key>CFBundleDisplayName</key>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <string>Event Tracker</string>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <key>CFBundleVersion</key>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <string>0.1.0</string>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <key>CFBundleShortVersionString</key>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <string>0.1.0</string>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <key>CFBundlePackageType</key>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <string>APPL</string>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <key>LSMinimumSystemVersion</key>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <string>11.0</string>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <key>NSHighResolutionCapable</key>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '    <true/>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '</dict>' >> "$BUNDLE_DIR/Contents/Info.plist"
echo '</plist>' >> "$BUNDLE_DIR/Contents/Info.plist"

echo "✓ macOS app bundle created at $BUNDLE_DIR"
echo ""
echo "To run: open \"$BUNDLE_DIR\""
echo "To test: ./\"$BUNDLE_DIR/Contents/MacOS/event-tracker\""
