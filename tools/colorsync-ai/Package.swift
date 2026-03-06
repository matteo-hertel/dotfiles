// swift-tools-version: 6.2

import PackageDescription

let package = Package(
    name: "colorsync-ai",
    platforms: [.macOS(.v26)],
    targets: [
        .executableTarget(name: "colorsync-ai")
    ]
)
