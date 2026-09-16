# Continue learning on another computer

Use this workflow for iCloud sync, “换电脑继续学习”, a fresh installation with existing learning history, or a reported sync conflict. The plugin uses a local archive and immutable, verified backup snapshots in a user-selected cloud folder. It does not sync Codex accounts, raw chats, ongoing calls or login credentials.

## First connection

Install the matching released plugin from **Ccryrt/english-speaking-coach**, following [plugin-installation.md](plugin-installation.md). A new Codex installation still needs this plugin. Do not install the upstream English-only release. New tasks pick up installed skills.

On Mac, the default is iCloud Drive / Language Exchange. Both computers use the same Apple Account with iCloud Drive enabled; wait for files to download (Finder “Keep Downloaded” helps). On Windows, let the user select the downloaded iCloud Drive folder; do not guess its location. Sync does not log into Apple or change iCloud settings.

After the user requests this setup, run the bundled entry for both languages:

- `coach --language en sync connect`
- `coach --language ja sync connect`

Here `coach` means the installed `scripts/coach` through `sh` or `coach.ps1` on Windows. Supply `--directory <absolute path>` for a chosen folder. Both languages share that folder but use separate `en/` and `ja/` snapshots. A new empty archive adopts the available history. Existing differing records are preserved as a conflict, not overwritten. The webpage's “本地学习数据 → 换电脑继续学习” offers the same connection and sync controls per language.

Connect once per computer. `sync status` shows its folder and last local result. `published` only means the complete snapshot was written to the local cloud folder: never claim iCloud uploaded it or the second device received it without observing that device. The first computer should finish review and allow iCloud to upload before shutdown.

## Normal practice

`resume` and source-verified `prepare` check sync before a new practice. Saved sessions, preferences, evidence and micro-lesson steps publish automatically. An ongoing call stays bound to its existing archive; never change language or archive in the middle. `sync` in the returned result reports restored/current/busy/offline state. Explain pending/offline sync briefly in the user's language and keep local work; never invent successful remote delivery.

Before guided `lesson show/start`, run `coach --language <pinned> sync now` if connected, then read fresh lesson state/revision. The later `resume --phase guided` also checks updates. A restored archive has a new local path: reread `paths`, do not reuse an old explicit `--root` or revision. Normal managed web services follow the validated directory change. Fixed-directory previews remain fixed.

The learner can say “检查同步” to run `sync now`. No recurring automation is required. Snapshots contain profile, sessions, evidence, legacy archive, micro-lesson progress, pending records, copied context and scene history. Runtime jobs, live transcript databases and external linked project files are excluded. Preserve locally retained old archives; do not clean them up without a separate request.

## Conflicts and recovery

- Offline, corrupt, partial or not-yet-downloaded files: keep the local archive, explain the pending sync and retry after download. Do not delete suspect files or initialize over an unavailable configured archive.
- `conflict`: stop automatic import; both device branches remain in the sync folder. Use `sync versions` to list current heads and their source ZIP files. Inspect them with existing `storage inspect` / restore into isolated directories if comparing content is needed. Never claim an automatic merge.
- After the learner explicitly chooses the desired history, `sync resolve --choose <version-id>` or `sync resolve --choose local` publishes that choice as a new version retaining all prior branches. This chooses one history; it is not a field-by-field merge. If they want both, reconcile records using existing evidence contracts before choosing local. Do not choose merely by timestamp.
- `sync disconnect` stops automatic sync without deleting local or cloud data.

The synchronizer scans immutable snapshots; this first version is intended for personal learning histories, not concurrent group editing. It detects diverged device versions without relying on iCloud as a distributed lock. It cannot see a change that iCloud has not delivered yet, so fresh sessions must surface the actual local sync status honestly.
