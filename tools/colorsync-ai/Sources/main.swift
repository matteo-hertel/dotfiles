import Foundation
import FoundationModels

@Generable(description: "A terminal color theme with background, foreground, cursor, and 16 ANSI colors")
struct GeneratedTheme {
    @Guide(description: "A short hyphenated name for the theme, e.g. 'autumn-dusk'")
    var name: String

    @Guide(description: "Background color as a 7-character hex string like #1a1b26. For dark themes use a dark color, for light themes use a light color.")
    var background: String

    @Guide(description: "Foreground/text color as a 7-character hex string like #cdd6f4. Should contrast well with the background.")
    var foreground: String

    @Guide(description: "Cursor color as a 7-character hex string. Usually same as foreground or an accent color.")
    var cursor: String

    @Guide(description: """
        Exactly 16 ANSI terminal colors as 7-character hex strings (#rrggbb). \
        Index meanings: 0=black, 1=red, 2=green, 3=yellow, 4=blue, 5=magenta, 6=cyan, 7=white, \
        8=bright black, 9=bright red, 10=bright green, 11=bright yellow, 12=bright blue, \
        13=bright magenta, 14=bright cyan, 15=bright white. \
        Colors should be harmonious and match the theme description.
        """)
    @Guide(.count(16))
    var colors: [String]
}

func isValidHex(_ s: String) -> Bool {
    s.count == 7 && s.first == "#" && s.dropFirst().allSatisfy(\.isHexDigit)
}

@main
struct ColorsyncAI {
    static func main() async throws {
        let args = CommandLine.arguments
        guard args.count >= 2 else {
            FileHandle.standardError.write(Data("Usage: colorsync-ai <description>\n".utf8))
            exit(1)
        }

        let description = args.dropFirst().joined(separator: " ")

        let model = SystemLanguageModel.default
        guard model.isAvailable else {
            FileHandle.standardError.write(Data("Error: Apple Intelligence is not available on this device.\n".utf8))
            exit(2)
        }

        let session = LanguageModelSession()
        let prompt = """
            Generate a terminal color theme based on this description: \(description)

            The theme needs a name, background color, foreground color, cursor color, \
            and exactly 16 ANSI colors. All colors must be 7-character hex strings starting with #. \
            The colors should be harmonious, visually appealing, and match the description. \
            Ensure good contrast between background and foreground.
            """

        let response = try await session.respond(
            to: prompt,
            generating: GeneratedTheme.self
        )

        let theme = response.content

        // Validate all hex colors before outputting
        let allColors = [theme.background, theme.foreground, theme.cursor] + theme.colors
        for color in allColors {
            guard isValidHex(color) else {
                FileHandle.standardError.write(Data("Error: model produced invalid hex color: \(color)\n".utf8))
                exit(3)
            }
        }

        // Build JSON matching Go's palette.Theme format
        let dict: [String: Any] = [
            "name": theme.name,
            "background": theme.background,
            "foreground": theme.foreground,
            "cursor": theme.cursor,
            "colors": theme.colors
        ]

        let jsonData = try JSONSerialization.data(withJSONObject: dict, options: [.prettyPrinted, .sortedKeys])
        FileHandle.standardOutput.write(jsonData)
        FileHandle.standardOutput.write(Data("\n".utf8))
    }
}
