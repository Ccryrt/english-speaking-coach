package main

import (
	"net"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuidedLearningLifecycle(t *testing.T) {
	root := testRoot(t)
	day := "2026-09-16"
	command := func(action string, args M) M {
		args = merge(M{"expected-revision": lessonView(root, day)["revision"]}, args)
		cliArgs := []string{"lesson", action, "--root", root, "--today", day, "--expected-revision", str(args["expected-revision"])}
		if str(args["input"]) != "" {
			cliArgs = append(cliArgs, "--input", str(args["input"]))
		}
		return obj(runCLI(cliArgs))
	}
	initial := lessonView(root, day)
	if initial["stage"] != "new" || initial["models"] != nil {
		t.Fatal(initial)
	}
	started := command("start", nil)
	if len(arr(started["models"])) != 2 {
		t.Fatal(started)
	}
	// Read-only page access and repeated starts do not complete teaching.
	command("start", nil)
	if lessonView(root, day)["stage"] != "demonstrate" {
		t.Fatal("page/start advanced learning")
	}
	command("next", nil)
	if attempt(func() { lessonCommand(root, day, "next", M{"expected-revision": initial["revision"]}) }) == nil {
		t.Fatal("stale update accepted")
	}
	command("next", nil)
	for _, v := range []M{lessonView(root, day), newArchive(root).query("/api/lesson", nil)} {
		if v["stage"] != "independent" || v["models"] != nil || v["last_attempt"] != nil {
			t.Fatal("answers leaked", v)
		}
	}
	evidence := M{"context": "Fictional login task with colleague A", "source": M{"kind": "text", "reference": "fixture/task/message-1"}, "responses": A{
		M{"id": "progress", "quote": "fictional learner response", "support": "none", "success": true},
		M{"id": "help", "quote": "fictional learner response after a hint", "support": "keyword", "success": true},
	}}
	file := filepath.Join(t.TempDir(), "evidence.json")
	writeJSON(file, evidence)
	saved := command("finish", M{"input": file})
	if saved["next_review"] != "2026-09-17" || truth(obj(saved["last_attempt"])["independent"]) {
		t.Fatal(saved)
	}
	if attempt(func() { command("finish", M{"input": file}) }) == nil {
		t.Fatal("same-day/repeated finish accepted")
	}
	// Backups retain observations; viewing the test never changes its history.
	_, manifest := backupBytes(root, "", false)
	if obj(manifest["files"])["data/Practice/progress-help.json"] == nil {
		t.Fatal("lesson not backed up")
	}
	day = "2026-09-17"
	due := lessonView(root, day)
	if due["stage"] != "retest" || due["models"] != nil || due["last_attempt"] != nil {
		t.Fatal(due)
	}
	if attempt(func() { command("finish", M{"input": file}) }) == nil {
		t.Fatal("same-context retest accepted")
	}
	evidence["context"] = "Fictional deployment task with colleague B"
	obj(arr(evidence["responses"])[1])["support"] = "none"
	writeJSON(file, evidence)
	saved = command("finish", M{"input": file})
	if saved["next_review"] != "2026-09-24" || saved["attempt_count"] != 2 || !truth(obj(saved["last_attempt"])["independent"]) {
		t.Fatal(saved)
	}
	// A later failed retest shortens the interval again.
	day = "2026-09-24"
	evidence["context"] = "Fictional report task with colleague C"
	obj(arr(evidence["responses"])[1])["success"] = false
	writeJSON(file, evidence)
	if command("finish", M{"input": file})["next_review"] != "2026-09-25" {
		t.Fatal("failed retest not scheduled tomorrow")
	}
	if strings.Contains(str(lessonContext(root, day)["voice_brief"]), "english_only") {
		t.Fatal("free policy leaked")
	}
}

func TestGuidedPreparationKeepsSourceAndOwnPolicy(t *testing.T) {
	t.Setenv("CODEX_HOME", t.TempDir())
	root := testRoot(t)
	initial := lessonView(root, today())
	lessonCommand(root, today(), "start", M{"expected-revision": initial["revision"]})
	resumed := obj(runCLI([]string{"resume", "--root", root, "--phase", "guided"}))
	if obj(resumed["policy"])["guided_learning"] != true || obj(resumed["lesson"])["stage"] != "demonstrate" || str(resumed["voice_brief"]) == "" {
		t.Fatal(resumed)
	}
	source := voiceLog(t, false)
	server := newServer(root, false)
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	server.port = httpServer.Listener.Addr().(*net.TCPAddr).Port
	args := M{"root": root, "thread-id": testThread, "source": source, "phase": "guided", "opening": true, "auto-scene": true, "service-url": httpServer.URL}
	prepared := preparePractice(args)
	if !truth(prepared["conversation_may_start"]) || obj(prepared["binding"])["voice_id"] != testVoice || !strings.HasSuffix(str(prepared["url"]), "/#guided") || !strings.Contains(str(prepared["caption_url"]), "#live") || !strings.Contains(str(prepared["review_url"]), testVoice) || prepared["turn_guidance"] != nil || len(sceneHistory(root)) != 0 {
		t.Fatal(prepared)
	}
	if obj(prepared["policy"])["guided_learning"] != true || obj(prepared["lesson"])["stage"] != "demonstrate" {
		t.Fatal(prepared)
	}
}
