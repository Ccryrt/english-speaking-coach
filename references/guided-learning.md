# Guided learning / 引导学习

Use this route when the learner asks for “引导学习”, “带我学英语/日语”, beginner teaching, or the “说明进度并请求帮助” micro-lesson. Preserve free conversation when that is requested. Pin `--language en|ja` on every command. This route overrides the free-conversation English-only, no-choices and no-pre-model rules: short Chinese explanations, examples and temporary choices are allowed while teaching. Maintenance never starts a lesson.

The first lesson has one Can-do and two expressions: describe current work and ask a colleague for specific help. Use real learner details without confidential customer information. Demonstration examples are fictional, never claimed as learner experience. Use simple English or Japanese です・ます, one useful point per turn, Chinese when needed. A short comprehension choice can support teaching but never counts as independent speech. Do not add a full interview or more expressions to this lesson.

## Start or continue

`coach` below means `sh "<skill-root>/scripts/coach" --language en|ja` with the chosen language.

1. `coach lesson show` returns `stage` and `revision`. `coach lesson start --expected-revision <revision>` starts a new lesson or resumes existing progress without erasing it. Never advance simply because the page opened.
2. Text: `coach resume --phase guided`, then `coach open --page guided --no-browser`. Active Voice: `coach prepare --phase guided --opening --with-project`. Preparation must verify the actual Voice before claiming it is ready; retain `review_url` and `caption_url`. Use the returned `url` for the right panel. It shows the lesson, while captions remain at `caption_url`. No Voice means text practice only; do not create a fake binding.
3. Follow the saved stage. On later turns read `lesson show`; it restores progress and automatically returns `retest` when due. Never use free-scene setup or its English-only `turn_guidance` in this route.

## Teach, withdraw help, then retest

- `demonstrate`: show/model the first expression and explain it briefly in Chinese if useful. Check meaning before the second expression. Reading or echoing is supported practice, never mastery. Once both are understood, call `lesson next --expected-revision <latest revision>`.
- `supported`: ask the learner to substitute their own task and actual help need, one expression per turn. Supply a model or keyword as needed. After they have tried both, call `lesson next` with the latest revision to enter `independent`.
- `independent`: reopen/refresh the guided page and verify the “独立尝试” step is visible and model answers are absent **before** asking. Do a brief unrelated exchange first. Then ask separately what they are working on and what help they need, without a candidate answer. Hide notes or captions that would give the answer; if they remain visible or are consulted, record `model` support, not `none`. A repetition immediately after a model is also `model` even if hidden. Do not prevent help: provide it when requested and record the actual support for that expression.
- `waiting`: report only what they managed and where help remained, plus `next_review`. Do not repeat the same-day test as a next-day success. They can stop, switch to free conversation or ask for extra supported practice; extra teaching is not a new independent evidence entry.
- `retest`: begin with a different real task and different context from the prior attempt, without re-showing the old model. Keep the same two abilities, not exact memorized wording. Verify the test page has no answers. Record each actual response and support level. Failure/help schedules tomorrow; an independent retest schedules seven days later. This is a simple schedule, not a validated mastery or forgetting model.

Pausing/stopping leaves the stage intact; do not invent attempts for unfinished lessons. A free-conversation request leaves this route and uses the language's original rules. On resuming guided learning, continue the saved stage. If learning materials were used again before a pending test, record the resulting support honestly.

## Save actual attempts

After both independent attempts, use `lesson finish --expected-revision <latest revision> --input <JSON file>`:

```json
{
  "context": "Actual task and colleague/context used in this attempt",
  "source": {"kind": "text", "reference": "Actual task ID and learner message references"},
  "responses": [
    {"id": "progress", "quote": "Exact learner words", "support": "none", "success": true},
    {"id": "help", "quote": "Exact learner words", "support": "keyword", "success": true}
  ]
}
```

`support` is `none`, `keyword`, or `model`; `success` means the communicative goal was met, not perfect grammar. For no response, explicitly record “未作答” with `success:false`. Use `voice_transcript` only with real source-bound transcript references; it does not establish pronunciation or unassisted listening. Never use example JSON as actual evidence. The runtime validates structure, dates and transitions; the Agent must check quote/reference accuracy against observed messages. These are coaching observations, not automatically verified mastery scores. Never mark support-free success just because the learner echoed a sentence or clicked the page.

Guided state is saved in the current language's `Practice/progress-help.json`, included in normal backups. It is not a replacement for the existing source-checked session review: keep the language's normal text/Voice closeout, end-call authorization and evidence rules. Do not create a second competing Voice review. Report actual Voice delivery as unverified unless heard in a real call.
