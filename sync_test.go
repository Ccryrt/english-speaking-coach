package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncAcrossDevicesAndConflicts(t *testing.T) {
	cloud := filepath.Join(t.TempDir(), "Cloud", "Language Exchange")
	mkdir(filepath.Dir(cloud))
	aHome, bHome := t.TempDir(), t.TempDir()
	t.Setenv("CODEX_HOME", aHome)
	a := testRoot(t)
	saveConfiguration(a, "")
	lessonCommand(a, today(), "start", M{"expected-revision": lessonView(a, today())["revision"]})
	if syncRun(a, "connect", cloud, "")["status"] != "published" {
		t.Fatal("first device did not publish")
	}
	t.Setenv("CODEX_HOME", bHome)
	b := testRoot(t)
	saveConfiguration(b, "")
	result := syncRun(b, "connect", cloud, "")
	if result["status"] != "restored" {
		t.Fatal(result)
	}
	b = str(result["data_root"])
	if lessonView(b, today())["stage"] != "demonstrate" {
		t.Fatal("micro-lesson progress lost")
	}
	if !exists(filepath.Join(str(syncConfig()["root"]), "profile.json")) {
		t.Fatal("missing restored archive")
	}
	// Offline A and B both change. Every branch remains a complete, recoverable ZIP.
	lessonCommand(b, today(), "next", M{"expected-revision": lessonView(b, today())["revision"]})
	t.Setenv("CODEX_HOME", aHome)
	lessonCommand(a, today(), "next", M{"expected-revision": lessonView(a, today())["revision"]})
	lessonCommand(a, today(), "next", M{"expected-revision": lessonView(a, today())["revision"]})
	result = syncRun(a, "now", "", "")
	if result["status"] != "conflict" || len(arr(result["versions"])) != 2 {
		t.Fatal(result)
	}
	if lessonView(a, today())["stage"] != "independent" {
		t.Fatal("conflict overwrote local progress")
	}
	count := len(glob(filepath.Join(cloud, syncLanguage(), "*.zip")))
	syncRun(a, "now", "", "")
	if len(glob(filepath.Join(cloud, syncLanguage(), "*.zip"))) != count {
		t.Fatal("unchanged conflict repeatedly uploaded")
	}
	resolved := syncRun(a, "resolve", "", "local")
	if resolved["status"] != "resolved" {
		t.Fatal(resolved)
	}
	t.Setenv("CODEX_HOME", bHome)
	result = syncRun(b, "now", "", "")
	b = str(result["data_root"])
	if result["status"] != "restored" || lessonView(b, today())["stage"] != "independent" {
		t.Fatal(result)
	}
	if len(glob(filepath.Join(cloud, syncLanguage(), "*.zip"))) <= count {
		t.Fatal("resolution removed history")
	}
	if syncRun(b, "now", "", "")["status"] != "up_to_date" {
		t.Fatal("restored content immediately changed")
	}
	// A missing parent represents partially downloaded iCloud history, never an empty cloud.
	cfg := syncConfig()
	parent := str(cfg["head"])
	p := filepath.Join(cloud, syncLanguage(), parent+".zip")
	blob := readFile(p)
	must(os.Remove(p))
	if attempt(func() { syncRun(b, "now", "", "") }) == nil {
		t.Fatal("missing known version silently accepted")
	}
	atomicWrite(p, blob)
	// A failed upload cannot fail a successful local micro-lesson save.
	moved := cloud + "-offline"
	must(os.Rename(cloud, moved))
	writeJSON(filepath.Join(b, "Practice", "offline-note.json"), M{"fixture": true})
	if syncAfterSave(b)["sync_warning"] == nil {
		t.Fatal("offline save hid pending sync")
	}
	if !exists(filepath.Join(b, "Practice", "offline-note.json")) {
		t.Fatal("offline save lost data")
	}
	must(os.Rename(moved, cloud))
	if syncRun(b, "now", "", "")["status"] != "published" {
		t.Fatal("offline changes did not retry")
	}
}

func TestSyncRejectsPartialOrWrongLanguageAndKeepsPrivateRuntime(t *testing.T) {
	t.Setenv("CODEX_HOME", t.TempDir())
	root := testRoot(t)
	writeJSON(filepath.Join(root, "Runtime", "private.json"), M{"fixture_secret": "never sync this"})
	cloud := filepath.Join(t.TempDir(), "Language Exchange")
	syncRun(root, "connect", cloud, "")
	nodes, heads := syncSnapshots(filepath.Join(cloud, syncLanguage()))
	node := obj(nodes[str(heads[0])])
	files := syncImportedFiles(node)
	if files["Runtime/private.json"] != nil || files["SyncInfo.json"] != nil {
		t.Fatal("machine metadata leaked into learning facts")
	}
	before := syncDigest(syncFiles(root))
	atomicWrite(filepath.Join(cloud, syncLanguage(), ".waiting.zip.icloud"), nil)
	if attempt(func() { syncRun(root, "now", "", "") }) == nil {
		t.Fatal("undownloaded files accepted")
	}
	if syncDigest(syncFiles(root)) != before {
		t.Fatal("failed sync modified local archive")
	}
	must(os.Remove(filepath.Join(cloud, syncLanguage(), ".waiting.zip.icloud")))
	atomicWrite(str(node["file"]), []byte("broken zip"))
	if attempt(func() { syncRun(root, "now", "", "") }) == nil {
		t.Fatal("corrupt snapshot accepted")
	}
	if syncDigest(syncFiles(root)) != before {
		t.Fatal("corruption overwrote local archive")
	}
}

func TestSyncWebConnectAuthorizationAndAutomaticResume(t *testing.T) {
	cloud := filepath.Join(t.TempDir(), "Language Exchange")
	firstHome, secondHome := t.TempDir(), t.TempDir()
	t.Setenv("CODEX_HOME", firstHome)
	first := testRoot(t)
	saveConfiguration(first, "")
	lessonCommand(first, today(), "start", M{"expected-revision": lessonView(first, today())["revision"]})
	syncRun(first, "connect", cloud, "")
	t.Setenv("CODEX_HOME", secondHome)
	second := testRoot(t)
	saveConfiguration(second, "")
	server := newServer(second, true)
	server.port = 18997
	defer server.stopBackground()
	request := func(method, path, body, token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:18997"+path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Coach-Token", token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	body := compact(M{"action": "connect", "directory": cloud})
	if request("POST", "/api/sync", body, "wrong").Code != 403 || len(syncConfig()) != 0 {
		t.Fatal("untrusted webpage connected sync")
	}
	response := request("POST", "/api/sync", body, server.archive.token)
	if response.Code != 200 || !strings.Contains(response.Body.String(), "restored") {
		t.Fatal(response.Body.String())
	}
	if server.archive.root == second {
		t.Fatal("web service did not follow restored archive")
	}
	if !exists(filepath.Join(second, "profile.json")) {
		t.Fatal("original archive was removed")
	}
	response = request("GET", "/api/lesson", "", "")
	if response.Code != 200 || !strings.Contains(response.Body.String(), "demonstrate") {
		t.Fatal(response.Body.String())
	}
	// A future change on A is picked up by ordinary resume on B, without a manual import.
	t.Setenv("CODEX_HOME", firstHome)
	lessonCommand(first, today(), "next", M{"expected-revision": lessonView(first, today())["revision"]})
	t.Setenv("CODEX_HOME", secondHome)
	resumed := obj(runCLI([]string{"resume", "--phase", "guided"}))
	if obj(resumed["sync"])["status"] != "restored" || obj(resumed["lesson"])["stage"] != "supported" {
		t.Fatal(resumed)
	}
	response = request("GET", "/api/lesson", "", "")
	if response.Code != 200 || !strings.Contains(response.Body.String(), "supported") {
		t.Fatal("web did not follow CLI restore", response.Body.String())
	}
}
