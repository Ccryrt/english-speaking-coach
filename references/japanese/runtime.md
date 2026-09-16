# Runtime operations

The Japanese service uses `http://127.0.0.1:8898`, a separate service identifier and `~/.codex/japanese-speaking-coach/workspace.json`. The default archive is `~/.codex/japanese-speaking-coach/data`. English stays on 8897 with its existing files. Never point the Japanese writer at an English archive.

Use the shared `scripts/coach` entry with `--language ja` on every command. User-facing pages use `http://127.0.0.1:8897/ja/`; 8898 is an internal worker:

- `version`, `paths`, `doctor [--probe]`: identify the build, storage, current Voice and model connection separately. `--probe` connects using existing Codex login; it is not an actual audio test.
- `init`: initialize only the Japanese archive. `validate`, `rebuild`, `recover`: validate/rebuild derived views or recover pending records.
- `service start|status|stop|resume`: use the native process manager. A conflicting port is preserved. Check no practice is active before a maintenance restart. An intentional stop needs `resume`.
- `live status --run <id>`: check exact source and translation state. No Voice ID means no valid binding; zero source utterances cannot be fixed by translating again.
- `review-context --thread-id <task> --voice-id <voice> --with-transcript`: read-only evidence check for a known Voice. `review-begin --thread-id <task> --voice-id <voice> --retry` retries a failed review for that identity.
- `recover-voice --thread-id <task> --voice-id <voice>`: recovery requires a genuinely closed, identified Voice. Never use another session to fill missing evidence.
- `storage backup --output <zip> --include-live`: preserves portable records and optional subtitle cache. Japanese backup format is distinct; an English backup must be rejected. The webpage also provides backup/restore controls with local request checks.

When a learner explicitly changes a persistent preference, read the current profile and its SHA-256 immediately before `set-preferences --input <patch.json> --expected-profile-sha256 <hash>`. Include real decision `source_ids` and `updated` date, preserve unrelated fields. Conflict means reread/merge, not overwrite. Available topic values: mixed/daily/work/interview; practice_language: japanese_first/bilingual; help_language: zh-CN/ja; mode: conversation/roleplay/focused; correction: in_character/after_scene/light/detailed. Default is Japanese-first, Chinese help, roleplay and concise in-character correction. Interviews defer optional corrections until the answer is complete.

Build the combined plugin from the repository root with `go run ./internal/release -targets darwin/arm64` (choose the target platform). Japanese runtime source is in `languages/ja/runtime/`. A missing or invalid binary is a build/install issue: do not fetch the upstream English release. The build script requires Go only on the developer machine. Once built, the plugin runs without Go or Python. Keep the source, binary checksum and installed build consistent; updated skills load in a new Codex task.
