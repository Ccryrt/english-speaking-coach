# Selected text-practice records

Only save actual practice after an observed end; technical troubleshooting is not a lesson. For Voice, use the source-verified worker instead. For text, the agent verifies actual quoted user and coach messages before writing; structural checks cannot prove those messages happened.

`coach resume --compact` provides the current archive context. Inspect existing session/expression IDs before allocating the next three-digit suffix for that date. Reuse an expression ID for the same canonical wording/meaning. Never overwrite a different record with the same ID.

Create a JSON payload with:

- `id`: `SES-YYYYMMDD-NNN`; `date`: actual practice date, `title`, `summary` in Chinese; `source_ids`: the actual chat/task references; `expressions`: selected expressions, possibly empty.
- Optional `practiced_at` with the known timezone, `scenarios`, `topics`, `next_focus`, `progress`, `coaching_notes`, `evidence_note`. Do not invent times.
- Each expression: `id` (`EXP-YYYYMMDD-NNN`), `original` (exact learner quote), `japanese` (recommended form), `chinese` (meaning), `note` (what happened or a review-only suggestion), `mastery`, `next_review` (date).
- Optional `reading` (kana; preserve necessary acronyms/numbers), `register` (Chinese audience/politeness note), `correction_kind`: error/naturalness/register/already_correct. Omit uncertain readings. No romaji or pitch-scoring claim.

`mastery: not_tested` is appropriate for a review-only suggestion without a learner attempt. Actual attempts need `review_prompt` and `review_result`: source_text/success for repeating a model; keywords/success for prompted use; independent with none/success; transfer with changed_context/transfer_success. Preserve exact attempt evidence when available. Reading or seeing a card is not an attempt. Inspect runtime validation errors rather than weakening evidence requirements.

Write via `coach add-session --input <file> --check`, inspect the saved session ID and validation result, then `coach open --page sessions/<saved-id>` and check that record. Files in `Sessions/` are authoritative Markdown with embedded JSON; `state.json`, indexes and HTML are derived. Private learning records never belong in the source repository or plugin package.
