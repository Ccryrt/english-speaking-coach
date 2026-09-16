package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testThread = "11111111-1111-4111-8111-111111111111"
const testVoice = "22222222-2222-4222-8222-222222222222"

func testRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "日语学习")
	initialize(root)
	rebuild(root)
	return root
}

func fixtureRecord() M {
	return M{"id": "SES-20260916-001", "date": "2026-09-16", "title": "测试示例 · 向上司报告延期", "summary": "虚构测试资料：练习说明完成时间，不代表用户学习记录。", "source_ids": stringsA("fictional:test"), "topics": stringsA("工作沟通"), "expressions": A{M{"id": "EXP-20260916-001", "original": "あと二日必要。", "japanese": "あと２日ほどかかる見込みです。", "chinese": "预计还需要两天。", "reading": "あとふつかほどかかるみこみです。", "register": "向上司汇报，使用礼貌而不过分正式的表达。", "correction_kind": "register", "note": "测试示例，课后建议，不代表用户已掌握。", "mastery": "not_tested", "next_review": "2026-09-17"}}}
}

func fixtureVoice(t *testing.T, closed bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.jsonl")
	rows := A{M{"type": "session_meta", "payload": M{"id": testThread}}, M{"type": "realtime_item", "timestamp": "2026-09-16T01:00:00Z", "payload": M{"type": "realtime_session_started", "realtime_session_id": testVoice}}}
	for _, s := range []M{{"id": "u1", "role": "user", "text": "あと二日必要。"}, {"id": "a1", "role": "assistant", "text": "あと２日ほどかかる見込みです、と言えます。"}, {"id": "u2", "role": "user", "text": "ありがとうございます。"}} {
		rows = append(rows, M{"type": "realtime_item", "timestamp": "2026-09-16T01:00:01Z", "payload": merge(s, M{"type": "transcript_segment", "realtime_session_id": testVoice})})
	}
	if closed {
		rows = append(rows, M{"type": "realtime_item", "timestamp": "2026-09-16T01:01:00Z", "payload": M{"type": "realtime_session_closed", "realtime_session_id": testVoice}})
	}
	var data strings.Builder
	for _, row := range rows {
		data.WriteString(compact(row) + "\n")
	}
	atomicWrite(path, []byte(data.String()))
	return path
}

func TestJapaneseCaptionsAndFailureIsolation(t *testing.T) {
	segments := A{M{"id": "a", "text": "無料"}, M{"id": "b", "text": "APIの仕様を確認したいです。"}, M{"id": "c", "text": "我想说延期，日语怎么说？"}}
	units := translationUnits(segments)
	if len(units) != 3 || obj(units[1])["text"] != obj(segments[1])["text"] {
		t.Fatal("mixed Japanese was split or kanji skipped")
	}
	translated := A{M{"id": "a/japanese/0", "kind": "translation", "chinese": "免费", "reading": "むりょう"}, M{"id": "b/japanese/0", "kind": "translation", "chinese": "想确认 API 的规格。", "reading": "APIのしようをかくにんしたいです。"}, M{"id": "c/japanese/0", "kind": "translation", "chinese": "我想说延期，日语怎么说？"}}
	rows, rejected := partialTranslations(segments, units, translated)
	if len(rows) != 3 || len(rejected) != 0 || obj(rows[0])["reading"] != "むりょう" {
		t.Fatalf("bad translations: %v / %v", rows, rejected)
	}
	obj(translated[1])["chinese"] = "APIの仕様を確認したいです。"
	rows, rejected = partialTranslations(segments, units, translated)
	if len(rows) != 2 || rejected["b"] == nil {
		t.Fatal("one invalid translation poisoned valid utterances or echoed Japanese")
	}
	obj(translated[0])["reading"] = "muryou"
	rows, _ = partialTranslations(segments, units, translated)
	if obj(rows[0])["reading"] != nil {
		t.Fatal("romaji presented as kana")
	}
}

func TestLiveRevisionsAndKanjiRequireTranslation(t *testing.T) {
	root := testRoot(t)
	l := openLive(root)
	defer l.close()
	sqlExec(l.db, "INSERT INTO segments(run,id,role,text,status,ingested_at) VALUES(?,?,?,?,?,?)", "fixture", "s1", "user", "無料", "pending", now())
	l.copyLocalTranscripts("fixture")
	if obj(query(l.db, "SELECT status FROM segments")[0])["status"] != "pending" {
		t.Fatal("kanji wrongly treated as Chinese")
	}
	l.translated("fixture", A{M{"id": "s1", "chinese": "免费", "reading": "むりょう"}}, 0, M{"s1": "無料"})
	sqlExec(l.db, "UPDATE segments SET text=?,chinese=NULL,reading=NULL,status='pending' WHERE id=?", "有料", "s1")
	l.translated("fixture", A{M{"id": "s1", "chinese": "免费", "reading": "むりょう"}}, 0, M{"s1": "無料"})
	row := obj(query(l.db, "SELECT * FROM segments")[0])
	if row["chinese"] != nil || row["reading"] != nil || row["status"] != "pending" {
		t.Fatal("stale reading/translation replaced a revised original")
	}
}

func TestJapaneseArchiveAndPage(t *testing.T) {
	root := testRoot(t)
	record := fixtureRecord()
	commitRecord(root, record)
	if commitRecord(root, record)["status"] != "already_saved" {
		t.Fatal("same record not idempotent")
	}
	saved := extract(filepath.Join(root, "Sessions", str(record["id"])+".md"), "speaking-record-v2")
	if obj(arr(saved["expressions"])[0])["reading"] != "あとふつかほどかかるみこみです。" || !truth(validateArchive(root)["ok"]) {
		t.Fatal("kana not retained")
	}
	s := newServer(root, false)
	s.port = 8898
	for _, path := range []string{"/api/overview", "/api/terms", "/api/sessions/SES-20260916-001", "/api/live", "/"} {
		r := httptest.NewRequest("GET", "http://127.0.0.1:8898"+path, nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
		if path == "/api/terms" && !strings.Contains(w.Body.String(), "あとふつか") {
			t.Fatal("card API dropped kana")
		}
	}
	bad := httptest.NewRequest("GET", "http://example.com/api/overview", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, bad)
	if w.Code != 403 {
		t.Fatal("local host gate missing")
	}
	before := fingerprint(root)
	obj(arr(record["expressions"])[0])["reading"] = "ato futsuka"
	if attempt(func() { commitRecord(root, record) }) == nil || fingerprint(root) != before {
		t.Fatal("invalid/conflicting record overwritten")
	}
}

func TestJapaneseClosedVoiceReview(t *testing.T) {
	root := testRoot(t)
	source := fixtureVoice(t, true)
	d := M{"title": "报告进度", "summary": "虚构对话中，学习者询问如何报告延期。", "expressions": A{M{"source_turn_ids": stringsA("u1"), "original": "あと二日必要。", "japanese": "あと２日ほどかかる見込みです。", "chinese": "预计还需要两天。", "reading": "あとふつかほどかかるみこみです。", "register": "向上司汇报", "correction_kind": "register", "note": "课后保存的礼貌说法。"}}, "omitted_turns": M{"u2": "礼貌致谢"}, "priority_indices": A{0}, "reading_omissions": M{"0": "已提供假名读音，不强行切分短句"}, "word_checks": A{M{"segment_id": "u1", "needs_word_help": false, "reason": "礼貌表达", "concept_indices": A{}}, M{"segment_id": "u2", "needs_word_help": false, "reason": "致谢", "concept_indices": A{}}}, "concept_observations": A{}}
	result := finishReview(root, reconcileReview(d, "zh-CN"), testThread, testVoice, source)
	if result["status"] != "saved" {
		t.Fatalf("review not saved: %v", result)
	}
	if len(arr(buildState(root)["sessions"])) != 1 {
		t.Fatal("missing reviewed lesson")
	}
	if attempt(func() { snapshotVoice(fixtureVoice(t, false), testThread, testVoice) }) == nil {
		t.Fatal("open Voice accepted as ended")
	}
}

func TestTopicsAndEnglishIsolation(t *testing.T) {
	root := testRoot(t)
	for _, category := range []string{"daily", "work", "interview"} {
		c := resumeContext(root, today(), "", nil, category)
		scene := chooseScene(root, c)
		if scene["category"] != category || !kanaRE.MatchString(str(scene["opening_line"])) {
			t.Fatalf("wrong topic: %v", scene)
		}
		rememberScene(root, "fixture-"+category, scene)
		if chooseScene(root, c)["goal"] == scene["goal"] {
			t.Fatal("same scenario repeated despite alternatives")
		}
		policy := obj(speakingContext(obj(c["profile"]), true, "", scene)["policy"])
		if policy["response_language"] != "japanese_first" || category == "interview" && policy["correction_timing"] != "after_answer" {
			t.Fatal("wrong coaching policy")
		}
	}
	textContext := resumeContext(root, today(), "", nil, "interview")
	if obj(textContext["policy"])["correction_timing"] != "after_answer" {
		t.Fatal("text interview interrupted by roleplay correction policy")
	}
	if obj(readJSON(filepath.Join(root, "profile.json")))["topic"] != "mixed" {
		t.Fatal("temporary topic rewrote preference")
	}
	english := filepath.Join(t.TempDir(), "english")
	mkdir(english)
	p := profileDefault()
	p["practice_language"] = "english_first"
	delete(p, "target_language")
	writeJSON(filepath.Join(english, "profile.json"), p)
	before := string(readFile(filepath.Join(english, "profile.json")))
	if attempt(func() { initialize(english) }) == nil || string(readFile(filepath.Join(english, "profile.json"))) != before || exists(filepath.Join(english, "Archive")) {
		t.Fatal("Japanese initialization touched English archive")
	}
	p["practice_language"] = "bilingual"
	writeJSON(filepath.Join(english, "profile.json"), p)
	if attempt(func() { runCLI([]string{"init", "--root", english}) }) == nil || exists(filepath.Join(english, ".write.lock")) {
		t.Fatal("CLI wrote to a bilingual English archive")
	}
	t.Setenv("CODEX_HOME", t.TempDir())
	if !strings.Contains(configPath(), "japanese-speaking-coach") {
		t.Fatal("config collides with English")
	}
	d := M{"application": "english-speaking-coach", "data_root": root}
	h := httptest.NewServer(httpJSON(d))
	defer h.Close()
	if attempt(func() { checkService(h.URL, root, false) }) == nil {
		t.Fatal("English service was accepted as Japanese")
	}
}

func TestJapaneseBackupBoundary(t *testing.T) {
	root := testRoot(t)
	commitRecord(root, fixtureRecord())
	blob, _ := backupBytes(root, "", false)
	info := inspectBackup(blob)
	if str(info["format"]) != "japanese-speaking-coach-backup" {
		t.Fatalf("wrong backup: %v", info)
	}
	dest := filepath.Join(t.TempDir(), "restored")
	restoreBackup(blob, dest, false)
	if !truth(validateArchive(dest)["ok"]) || len(arr(buildState(dest)["expressions"])) != 1 {
		t.Fatal("backup lost Japanese record")
	}
	var foreign bytes.Buffer
	writer := zip.NewWriter(&foreign)
	f, e := writer.Create("manifest.json")
	must(e)
	_, e = f.Write([]byte(`{"format":"english-speaking-coach-backup","version":1}`))
	must(e)
	must(writer.Close())
	if attempt(func() { inspectBackup(foreign.Bytes()) }) == nil {
		t.Fatal("English backup accepted")
	}
}

func TestLiveModelJapanese(t *testing.T) {
	if os.Getenv("JAPANESE_COACH_MODEL_TEST") != "1" {
		t.Skip("opt-in real-account fictional translation/review check")
	}
	c := newModelClient(defaultCaptionModel, 45*time.Second)
	defer c.close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	segments := A{M{"id": "1", "role": "assistant", "text": "無料"}, M{"id": "2", "role": "user", "text": "APIの仕様を確認したいです。"}, M{"id": "3", "role": "user", "text": "我想说还需要两天，日语怎么说？"}}
	rows, rejected, _ := c.translate(ctx, segments, nil)
	if len(rows) != 3 || len(rejected) != 0 {
		t.Fatalf("translation failed: %v / %v", rows, rejected)
	}
	t.Logf("fictional caption output: %s", compact(rows))
	root := testRoot(t)
	source := fixtureVoice(t, true)
	job := enqueueReview(root, testThread, testVoice, source, false)
	processReview(ctx, root, job)
	status := reviewStatus(root, testThread, testVoice)
	if status["status"] != "saved" {
		t.Fatalf("model review did not save: %s", compact(status))
	}
	state := buildState(root)
	if len(arr(state["expressions"])) == 0 {
		t.Fatal("no useful Japanese expression saved")
	}
	t.Logf("fictional review: %s", compact(state["expressions"]))
}

// Deliberately separate from user storage; used to inspect actual rendered cards.
func TestBrowserFixture(t *testing.T) {
	root := os.Getenv("JAPANESE_COACH_QA_DIR")
	if root == "" {
		t.Skip("opt-in isolated browser fixture")
	}
	if !filepath.IsAbs(root) || !emptyDir(root) {
		t.Fatal("fixture needs an empty absolute directory")
	}
	initialize(root)
	rebuild(root)
	commitRecord(root, fixtureRecord())
	l := openLive(root)
	defer l.close()
	s := M{"id": "fictional-ja", "demo": true, "status": "ended", "desired": "stopped", "created_at": now(), "translation_status": "ready"}
	sqlExec(l.db, "INSERT INTO runs VALUES(?,?,?)", "fictional-ja", "stopped", compact(s))
	sqlExec(l.db, "INSERT INTO segments(run,id,role,text,timestamp,chinese,reading,status,ingested_at) VALUES(?,?,?,?,?,?,?,?,?)", "fictional-ja", "1", "user", "あと２日ほどかかる見込みです。", now(), "预计还需要两天。", "あとふつかほどかかるみこみです。", "translated", now())
}

// httptest's dynamic port is intentional: no real service is touched.
type jsonHandler M

func httpJSON(d M) jsonHandler                                         { return jsonHandler(d) }
func (d jsonHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) { json.NewEncoder(w).Encode(d) }

func TestRequestedTopicSurvivesRecentScenes(t *testing.T) {
	for _, topic := range []string{"work", "interview"} {
		root := testRoot(t)
		for _, v := range arr(contracts["scenes"]) {
			row := arr(v)
			if row[6] == topic {
				rememberScene(root, str(row[0]), M{"setting": row[1], "goal": row[4]})
			}
		}
		for i := 0; i < 12; i++ {
			scene := chooseScene(root, resumeContext(root, today(), "", nil, topic))
			if scene["category"] != topic {
				t.Fatalf("requested %s, selected %v", topic, scene)
			}
			rememberScene(root, fmt.Sprint("repeat-", i), scene)
		}
	}
}
