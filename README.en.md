# 语言交流 / Language Exchange

This branch adds a Japanese option to the existing English coach. Use the **Practice language / 练习语言** selector on the shared learning page at http://127.0.0.1:8897/. The one `$english-speaking-coach` skill reads the saved selection; explicit language requests override it. English teaching rules remain unchanged.

Japanese covers everyday life, workplace communication and interviews, with Chinese support, kana readings and contextual register notes. Each language keeps its original archive and backup format. Switching applies to the next practice; start another Voice call to change a running session's language.

Build the combined plugin with `go run ./internal/release -targets darwin/arm64` (or another supported OS/architecture). Run `go test ./...` and `(cd languages/ja/runtime && go test ./...)`. The ZIP in `dist/` includes both verified native runtimes and one discoverable skill, with no runtime/compiler installation required for the learner. Do not replace it with the upstream English-only release.

The Japanese runtime lives in `languages/ja/runtime/` and is an internal component, reachable through the same page under `/ja/`. Its existing data directory is preserved. Native Voice behavior requires actual voice testing; browser/model tests are not audio validation.

Derived from yomage-ai/english-speaking-coach. See LICENSE for original terms.
