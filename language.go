package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func languagePath() string { return filepath.Join(filepath.Dir(configPath()), "language.json") }
func selectedLanguage() string {
	language := textOr(obj(maybeJSON(languagePath(), M{}))["language"], "en")
	require(language == "en" || language == "ja", "Invalid saved language; choose en or ja explicitly")
	return language
}
func saveLanguage(language string) M {
	require(language == "en" || language == "ja", "Language must be en or ja")
	state := M{"language": language, "applies_to": "next_practice"}
	writeJSON(languagePath(), state)
	return state
}
func languageCommand(commands []string) M {
	if len(commands) == 1 {
		return M{"language": selectedLanguage(), "applies_to": "next_practice"}
	}
	require(len(commands) == 3 && commands[1] == "set", "Use language or language set en|ja")
	return saveLanguage(commands[2])
}
func japaneseCommand(args []string) M {
	root := filepath.Join(skillRoot(), "languages", "ja")
	binary := filepath.Join(root, "bin", "japanese-coach_"+version+"_"+runtime.GOOS+"_"+runtime.GOARCH)
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	require(exists(binary) && exists(binary+".sha256"), "Japanese runtime is missing; rebuild/reinstall the combined plugin")
	require(hash(readFile(binary)) == strings.TrimSpace(string(readFile(binary+".sha256"))), "Japanese runtime checksum mismatch; nothing executed")
	must(os.Chmod(binary, 0755))
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	// Explicit arguments override inherited settings without changing the parent process.
	args = append(args, "--skill-root", root, "--codex-home", codexHome())
	if c := os.Getenv("ENGLISH_COACH_CODEX"); c != "" {
		args = append(args, "--codex-executable", c)
	}
	command := exec.CommandContext(ctx, binary, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	require(err == nil, "Japanese coach: "+strings.TrimSpace(stderr.String())+fmt.Sprint(err))
	var result M
	must(json.Unmarshal(output, &result))
	return result
}
func withoutLanguage(args []string) []string {
	out := []string{}
	for i := 0; i < len(args); i++ {
		if args[i] == "--language" {
			i++
			continue
		}
		if strings.HasPrefix(args[i], "--language=") {
			continue
		}
		out = append(out, args[i])
	}
	return out
}
func languageURLs(result M, from, to string) M {
	for _, key := range []string{"url", "review_url", "caption_url"} {
		if address := str(result[key]); strings.HasPrefix(address, from+"/") {
			result[key] = to + strings.TrimPrefix(address, from)
		}
	}
	return result
}
func runJapanese(args, commands []string, options M) M {
	require(commands[0] != "serve", "Use the shared service; do not run an independent Japanese page")
	result := japaneseCommand(withoutLanguage(args))
	if (commands[0] == "prepare" || commands[0] == "open") && str(result["url"]) != "" && str(options["service-url"]) == "" {
		w := workspace("")
		ensureWorkspace(w)
		base := str(serviceStart(str(w["data_root"]), true)["url"])
		languageURLs(result, "http://127.0.0.1:8898", base+"/ja")
	}
	result["language"] = "ja"
	return result
}
func (s *Server) japaneseHandler() http.Handler {
	s.gate.Lock()
	defer s.gate.Unlock()
	if s.japanese == nil {
		require(s.managed, "Japanese view needs the managed workspace; an explicit English preview cannot open personal Japanese data")
		result := japaneseCommand([]string{"service", "start"})
		target, err := url.Parse(str(result["url"]))
		must(err)
		require(target.String() == "http://127.0.0.1:8898", "Unexpected Japanese service URL")
		// Keep the established language engines and archive formats; the UI and skill share one entry.
		s.japanese = japaneseProxy(target)
	}
	return s.japanese
}
func japaneseProxy(target *url.URL) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(target)
	director := proxy.Director
	proxy.Director = func(r *http.Request) {
		director(r)
		r.Host = target.Host
		r.Header.Set("X-Coach-Gateway", "1")
		if r.Header.Get("Origin") != "" {
			r.Header.Set("Origin", target.String())
		}
	}
	return proxy
}
func (s *Server) languageRoute(w http.ResponseWriter, r *http.Request) (handled bool) {
	path := r.URL.Path
	if path != "/" && path != "/language.js" && path != "/api/language" && path != "/en" && path != "/ja" && !strings.HasPrefix(path, "/en/") && !strings.HasPrefix(path, "/ja/") {
		return false
	}
	handled = true
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	defer func() {
		if v := recover(); v != nil {
			sendJSON(w, r, 400, M{"error": fmt.Sprint(v)})
		}
	}()
	allowed := r.Host == fmt.Sprintf("127.0.0.1:%d", s.port) || r.Host == fmt.Sprintf("localhost:%d", s.port)
	if !allowed || (r.Header.Get("Origin") != "" && r.Header.Get("Origin") != "http://"+r.Host) {
		sendJSON(w, r, 403, M{"error": "请从本机学习网页操作。"})
		return
	}
	if path == "/api/language" {
		s.gate.Lock()
		localToken := s.archive.token
		s.gate.Unlock()
		if r.Method == "POST" {
			if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Coach-Token")), []byte(localToken)) != 1 {
				sendJSON(w, r, 403, M{"error": "请刷新学习页面后重试。"})
				return
			}
			require(strings.Split(r.Header.Get("Content-Type"), ";")[0] == "application/json", "Expected JSON")
			data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
			must(err)
			var body M
			must(json.Unmarshal(data, &body))
			require(exactKeys(body, "language") && has(stringsA("en", "ja"), body["language"]), "Language must be en or ja")
			if body["language"] == "ja" {
				s.japaneseHandler()
			}
			sendJSON(w, r, 200, saveLanguage(str(body["language"])))
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			sendJSON(w, r, 405, M{"error": "Method not allowed"})
			return
		}
		sendJSON(w, r, 200, M{"language": selectedLanguage(), "token": localToken})
		return
	}
	if path == "/" || path == "/en" || path == "/ja" {
		if r.Method != "GET" && r.Method != "HEAD" {
			sendJSON(w, r, 405, M{"error": "Method not allowed"})
			return
		}
		destination := path + "/"
		if path == "/" {
			destination = "/en/"
			if s.managed {
				destination = "/" + selectedLanguage() + "/"
			}
		}
		http.Redirect(w, r, destination, http.StatusTemporaryRedirect)
		return
	}
	if path == "/language.js" {
		sendBytes(w, r, 200, readAsset("assets/library/language.js"), "text/javascript; charset=utf-8")
		return
	}
	if strings.HasPrefix(path, "/ja/") {
		http.StripPrefix("/ja", s.japaneseHandler()).ServeHTTP(w, r)
		return
	}
	http.StripPrefix("/en", http.HandlerFunc(s.serveEnglish)).ServeHTTP(w, r)
	return
}
