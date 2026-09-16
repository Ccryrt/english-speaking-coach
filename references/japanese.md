# Japanese practice / 日语练习

## Start the requested practice

Run the bundled entry with `sh "<skill-root>/scripts/coach" --language ja` (PowerShell: `& "<skill-root>/scripts/coach.ps1" --language ja`). Use the existing English coach entry with `--language ja` on every command. On first use `coach init` initializes a separate Japanese archive. If it reports an existing unavailable archive, recover that location instead of creating a replacement. Commands below use `coach` as shorthand for the shared entry with `--language ja`.

Choose the requested topic: `daily`, `work`, `interview`, or `mixed` when unspecified. A topic is a per-practice choice; it does not rewrite saved preferences. For a specific user scenario, provide `--scene <json>` containing `setting`, `learner_role`, `partner_role`, `goal`, `introduction`, `opening_line` and `category` (`daily`, `work`, `interview`). Use Japanese scene content and an open first question. Honor user-selected roles and facts.

- **Active Voice:** `coach prepare --auto-scene --opening --with-project --topic <topic>`. This verifies the current task and actual active Voice before binding. If `conversation_may_start` is true, open the returned `url` once in the right browser panel, retain `review_url`, and briefly introduce the scene/roles/goal followed by one Japanese question. Continue an existing scene instead of repeating its opening. Do not narrate setup aloud. A text task must never be presented as active Voice.
- **Text practice:** `coach resume --compact --with-project --topic <topic>`. Select a matching scene and start in chat. `coach open --page overview` opens the record page; text chat does not produce live captions. Never fabricate a `voice_id`.
- **No active Voice:** explain the actual missing capability briefly; text practice can continue. A running server or an empty page is not proof of working captions.

## Teaching

- Begin with one or two short Japanese sentences and one question. Adapt difficulty from actual answers and saved evidence, not an assumed JLPT level. Use short natural clauses rather than a word-count target.
- Daily/work situations: give at most one useful correction at a time, then continue the role. Distinguish incorrect meaning/grammar, optional naturalness, and inappropriate register. Subjects omitted naturally in Japanese and valid short answers are not automatically errors.
- Default to です・ます. Match politeness to the partner (shop staff, colleague, manager, interviewer); do not replace every sentence with elaborate keigo.
- If the learner speaks Chinese, provide one learner-ready Japanese phrase with brief Chinese support if useful. Ask for clarification only when meaning is ambiguous. Do not force a confirmation loop just because languages are mixed. Allow English technical names and acronyms.
- In interviews, listen to the complete answer, give one focused point, and ask a relevant follow-up. Immediate help is available when requested. Build answers from the learner's true experience; ask about missing context rather than inventing companies, achievements, timelines or numbers.
- On reading help, show kana and demonstrate a short meaning group if Voice is available. Romaji is off by default. Do not infer pitch accent, pronunciation quality or audio accuracy from transcript text. The model's kana is a suggestion, not acoustic evidence.
- Follow requests to slow down, pause, switch topic or stop. Maintenance suspends the exercise until the learner asks to resume. Scene completion alone does not authorize ending Voice.

## Companion and review

The page preserves raw utterances and adds Chinese translations and optional collapsible kana. Translation failure must leave original text available. Speech and autonomous Voice behavior remain the host's responsibility; a skill cannot guarantee every spoken reply follows these instructions.

For explicit Voice end requests, use the available end-call tool before review paperwork. After a real close, open the returned `review_url`; the local worker prepares a review for that exact task and Voice. `coach review-begin --thread-id <task> --voice-id <voice>` queues/reuses that closed session if needed. Check the saved status; provisional suggestions are not a saved lesson. Never create a competing manual review while a worker owns it.

For ended text practice, follow [the record contract](japanese/records.md), selecting actual chat evidence and calling `coach add-session --input <file> --check`. Open that exact saved lesson. Maintenance creates no lesson; viewing or flipping a card never counts as speaking or mastery.

Review notes use Chinese. Save the original attempt separately from the recommended `japanese` expression and its `chinese` meaning. Optional `reading` uses kana, `register` explains the audience/politeness, and `correction_kind` distinguishes error/naturalness/register/already_correct. Use evidence to separate prompted repetition from independent use. Do not fabricate missed teaching as something the learner received.

For diagnosis, preference changes, backup or recovery, read [runtime operations](japanese/runtime.md). The shared webpage uses `/ja/`; English and Japanese records remain separate.
