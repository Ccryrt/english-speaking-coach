package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func TestLanguageSelectionAndArchiveRouting(t *testing.T) {
	t.Setenv("CODEX_HOME", t.TempDir())
	root := testRoot(t)
	commitRecord(root, fixtureRecord())
	profileBefore := hash(readFile(filepath.Join(root, "profile.json")))
	s := newServer(root, true)
	s.port = 18997
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/overview" && r.URL.Path != "/api/storage/backup" {
			t.Errorf("Wrong Japanese path: %s", r.URL.Path)
		}
		if r.Method == "POST" {
			if r.Header.Get("Origin") != "http://"+r.Host || r.Header.Get("X-Coach-Token") != "japanese-token" {
				t.Error("Proxy lost origin or language-specific token")
			}
		}
		fmt.Fprint(w, `{"language":"ja","japanese":"日本語の記録"}`)
	}))
	defer backend.Close()
	target, _ := url.Parse(backend.URL)
	s.japanese = japaneseProxy(target)
	request := func(method, path, body, token, origin string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:18997"+path, strings.NewReader(body))
		if method == "POST" {
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("X-Coach-Token", token)
		}
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		return w
	}
	if selectedLanguage() != "en" {
		t.Fatal("Existing users must default to English")
	}
	if got := request("GET", "/", "", "", "").Header().Get("Location"); got != "/en/" {
		t.Fatal(got)
	}
	for _, bad := range []struct{ body, token, origin string }{{`{"language":"ja"}`, "wrong", ""}, {`{"language":"ja"}`, s.archive.token, "https://example.com"}, {`{"language":"xx"}`, s.archive.token, ""}, {`{"language":"ja","root":"/tmp"}`, s.archive.token, ""}} {
		if w := request("POST", "/api/language", bad.body, bad.token, bad.origin); w.Code < 400 {
			t.Fatal("Invalid switch accepted")
		}
		if selectedLanguage() != "en" {
			t.Fatal("Failed switch changed choice")
		}
	}
	if w := request("POST", "/api/language", `{"language":"ja"}`, s.archive.token, ""); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if languageCommand([]string{"language"})["language"] != "ja" {
		t.Fatal("CLI did not observe webpage selection")
	}
	if got := request("GET", "/", "", "", "").Header().Get("Location"); got != "/ja/" {
		t.Fatal(got)
	}
	en := request("GET", "/en/api/overview", "", "", "")
	ja := request("GET", "/ja/api/overview", "", "", "")
	if en.Code != 200 || !strings.Contains(en.Body.String(), "Fictional bakery") || strings.Contains(en.Body.String(), "日本語の記録") {
		t.Fatal("English archive changed with selection")
	}
	if ja.Code != 200 || !strings.Contains(ja.Body.String(), "日本語の記録") || strings.Contains(ja.Body.String(), "Fictional bakery") {
		t.Fatal("Japanese archive not routed")
	}
	if w := request("POST", "/ja/api/storage/backup", "{}", "japanese-token", "http://127.0.0.1:18997"); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := request("POST", "/ja/api/storage/backup", "{}", "japanese-token", "https://example.com"); w.Code != 403 {
		t.Fatal("Proxy accepted foreign origin")
	}
	if hash(readFile(filepath.Join(root, "profile.json"))) != profileBefore {
		t.Fatal("Language selection mutated English teaching profile")
	}
	saveLanguage("en")
	if selectedLanguage() != "en" {
		t.Fatal("Cannot switch back")
	}
	reject(t, func() { saveLanguage("zh") })
	// Explicit pins remain stable even when the next-practice preference changes.
	saveLanguage("ja")
	paths := obj(runCLI([]string{"paths", "--language", "en", "--root", root}))
	if paths["data_root"] != absolute(root) {
		t.Fatal("Explicit English pin ignored")
	}
	got := withoutLanguage([]string{"resume", "--language=ja", "--topic", "work"})
	if strings.Join(got, " ") != "resume --topic work" {
		t.Fatal(got)
	}
	// A Japanese bilingual profile must not be accepted by the English writer.
	profile := profileDefault()
	profile["target_language"] = "ja"
	profile["practice_language"] = "bilingual"
	writeJSON(filepath.Join(root, "profile.json"), profile)
	reject(t, func() { workspace(root) })
}
