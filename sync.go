package main

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func syncLanguage() string   { return "en" }
func syncConfigPath() string { return filepath.Join(filepath.Dir(configPath()), "sync.json") }
func syncConfig() M          { return obj(maybeJSON(syncConfigPath(), M{})) }
func iCloudDirectory() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	base := filepath.Join(homeDir(), "Library", "Mobile Documents", "com~apple~CloudDocs")
	if !isDir(base) {
		return ""
	}
	return filepath.Join(base, "Language Exchange")
}

// Only portable learning facts travel. Runtime workers, credentials and raw Voice caches stay local.
func syncFiles(root string) M {
	files := M{}
	total := 0
	for _, p := range sourceFiles(root) {
		rel, err := filepath.Rel(root, p)
		must(err)
		name := filepath.ToSlash(rel)
		if name != "profile.json" && name != "Runtime/scene-history.json" && !strings.HasPrefix(name, "Sessions/") && !strings.HasPrefix(name, "Evidence/") && !strings.HasPrefix(name, "Archive/") && !strings.HasPrefix(name, "Practice/") && !strings.HasPrefix(name, "Pending/") && !strings.HasPrefix(name, "Context/") {
			continue
		}
		b := readFile(p)
		total += len(b)
		require(total <= maxBackup, "学习档案超过同步上限；原件保留。")
		files[name] = string(b)
	}
	return files
}
func syncDigest(files M) string { return hash(encode(files)) }
func syncWriteFiles(root string, files M) {
	for name, value := range files {
		atomicWrite(filepath.Join(root, filepath.FromSlash(name)), []byte(str(value)))
	}
	for _, d := range []string{"Sessions", "Evidence", "Pending"} {
		mkdir(filepath.Join(root, d))
	}
}
func syncInfo(root string) M {
	cfg := syncConfig()
	return M{"enabled": cfg["root"] == absolute(root), "directory": cfg["directory"], "suggested_directory": iCloudDirectory(), "last_result": cfg["last_result"], "last_checked": cfg["last_checked"], "last_error": cfg["last_error"], "message": "iCloud 负责设备间传输；这里显示本机可见的版本，不代表另一台电脑已收到。"}
}
func syncID(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func syncReadMeta(blob []byte) M {
	inspectBackup(blob)
	z, err := zip.NewReader(bytes.NewReader(blob), int64(len(blob)))
	must(err)
	for _, f := range z.File {
		if f.Name != "data/SyncInfo.json" {
			continue
		}
		r, err := f.Open()
		must(err)
		b, err := io.ReadAll(io.LimitReader(r, 16385))
		r.Close()
		must(err)
		require(len(b) <= 16384, "同步版本信息过大。")
		m := parseObject(string(b))
		require(m["format"] == "language-exchange-sync-v1" && m["language"] == syncLanguage() && syncID(str(m["id"])) && syncID(str(m["digest"])), "同步文件不是当前语言的学习版本。")
		checkStrings(m["parents"], "parents")
		for _, p := range arr(m["parents"]) {
			require(syncID(str(p)) && p != m["id"], "同步版本来源无效。")
		}
		return m
	}
	panic("同步文件缺少版本信息；未导入。")
}

// ponytail: scan immutable snapshots; add an index when hundreds of lessons make scanning slow.
func syncSnapshots(dir string) (M, A) {
	nodes := M{}
	parents := M{}
	entries, err := os.ReadDir(dir)
	must(err)
	require(len(entries) <= 2000, "同步版本过多，请先归档旧版本。")
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".icloud") {
			panic("iCloud 文件尚未下载，请在 Finder 中下载并重试。")
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		} // Atomic-write temporary files are not published versions.
		require(!e.IsDir() && strings.HasSuffix(e.Name(), ".zip") && e.Type()&os.ModeSymlink == 0, "同步文件夹含未知内容，已保留并停止同步。")
		p := filepath.Join(dir, e.Name())
		info, err := e.Info()
		must(err)
		require(info.Size() <= maxBackup, "同步备份过大。")
		m := syncReadMeta(readFile(p))
		id := str(m["id"])
		require(nodes[id] == nil, "同步文件出现重复版本，请检查 iCloud 冲突副本。")
		m["file"] = p
		nodes[id] = m
		for _, v := range arr(m["parents"]) {
			parents[str(v)] = true
		}
	}
	for id := range parents {
		require(nodes[id] != nil, "部分历史版本尚未到达本机，请等待 iCloud 下载后重试。")
	}
	heads := A{}
	for id := range nodes {
		if parents[id] == nil {
			heads = append(heads, id)
		}
	}
	sort.Slice(heads, func(i, j int) bool { return str(heads[i]) < str(heads[j]) })
	require(len(nodes) == 0 || len(heads) > 0, "同步版本关系无效。")
	return nodes, heads
}
func syncPublish(dir string, files M, parents A) M {
	staging, err := os.MkdirTemp("", "coach-sync-export-")
	must(err)
	defer os.RemoveAll(staging)
	syncWriteFiles(staging, files)
	m := M{"format": "language-exchange-sync-v1", "language": syncLanguage(), "id": token(), "parents": parents, "digest": syncDigest(files), "created_at": now()}
	writeJSON(filepath.Join(staging, "SyncInfo.json"), m)
	blob, _ := backupBytes(staging, "", false)
	p := filepath.Join(dir, str(m["id"])+".zip")
	require(!exists(p), "同步版本已存在；未覆盖。")
	atomicWrite(p, blob)
	m["file"] = p
	return m
}
func syncImportedFiles(node M) M {
	blob := readFile(str(node["file"]))
	meta := syncReadMeta(blob)
	require(meta["id"] == node["id"] && meta["digest"] == node["digest"], "同步版本在读取期间发生变化。")
	staging, err := os.MkdirTemp("", "coach-sync-check-")
	must(err)
	defer os.RemoveAll(staging)
	extractBackup(blob, staging)
	files := syncFiles(staging)
	require(syncDigest(files) == str(node["digest"]), "学习内容校验失败；本地档案未改变。")
	if exists(filepath.Join(staging, "Practice", "progress-help.json")) {
		lessonView(staging, today())
	}
	return files
}
func syncFresh(root string) bool {
	for name := range syncFiles(root) {
		if name != "profile.json" && name != "Archive/legacy-v1.md" {
			return false
		}
	}
	if len(glob(filepath.Join(root, "Sessions", "*.md"))) > 0 || len(glob(filepath.Join(root, "Evidence", "*.md"))) > 0 || len(glob(filepath.Join(root, "Pending", "*.json"))) > 0 || len(glob(filepath.Join(root, "Practice", "*.json"))) > 0 {
		return false
	}
	state := buildState(root)
	if len(arr(state["sessions"])) > 0 || len(arr(state["expressions"])) > 0 || len(arr(state["concepts"])) > 0 {
		return false
	}
	p := obj(readJSON(filepath.Join(root, "profile.json")))
	return equal(omit(p, "updated"), omit(profileDefault(), "updated"))
}
func syncApply(root string, node M, expected string) string {
	assertIdle(root)
	files := syncImportedFiles(node)
	target := filepath.Join(filepath.Dir(configPath()), "sync-restored", str(node["id"])+"-"+token()[:8])
	// Restore into a new local directory; the old archive survives even a crash before activation.
	syncWriteFiles(target, files)
	rebuild(target)
	archiveCounts(target)
	withLock(filepath.Join(root, ".write.lock"), func() {
		require(syncDigest(syncFiles(root)) == expected, "恢复前本地有新记录，已保留双方，请重试。")
		assertIdle(root)
		saveConfiguration(target, "")
	})
	return target
}
func syncRun(root, action, directory, choice string) M {
	require(has(stringsA("connect", "now", "push", "resolve", "versions", "disconnect"), action), "未知同步操作。")
	var result M
	withLock(filepath.Join(filepath.Dir(configPath()), ".sync.lock"), func() {
		cfg := syncConfig()
		root = absolute(root)
		if action == "disconnect" {
			writeJSON(syncConfigPath(), M{})
			result = M{"status": "disconnected", "message": "已停止自动同步，云端和本地版本均保留。"}
			return
		}
		if action == "connect" {
			if directory == "" {
				directory = iCloudDirectory()
			}
			require(directory != "", "未找到 iCloud Drive，请先启用，或指定已同步到本机的文件夹。")
			directory = checkRoot(directory)
			require(!within(directory, root) && !within(root, directory), "同步文件夹必须与本地档案分开。")
			require(isDir(filepath.Dir(directory)), "同步文件夹的上级目录不存在，请先等待 iCloud 可用。")
			if cfg["directory"] != directory || cfg["root"] != root {
				cfg = M{"directory": directory, "root": root, "baseline": "", "head": ""}
			}
			mkdir(filepath.Join(directory, syncLanguage()))
		} else if cfg["root"] != root {
			result = M{"status": "disabled"}
			return
		}
		dir := filepath.Join(str(cfg["directory"]), syncLanguage())
		require(isDir(dir), "同步文件夹暂时不可用；本地记录仍保留。")
		resolved, err := filepath.EvalSymlinks(dir)
		must(err)
		require(absolute(resolved) == absolute(dir), "同步目录不能通过符号链接跳转。")
		nodes, heads := syncSnapshots(dir)
		if action == "versions" {
			items := A{}
			for _, id := range heads {
				items = append(items, pick(obj(nodes[str(id)]), "id", "parents", "created_at", "digest", "file"))
			}
			result = M{"status": "versions", "versions": items}
			return
		}
		var files M
		withLock(filepath.Join(root, ".write.lock"), func() { files = syncFiles(root) })
		local := syncDigest(files)
		base := str(cfg["baseline"])
		head := str(cfg["head"])
		if head != "" {
			require(nodes[head] != nil, "上次同步版本尚未下载；没有覆盖任何内容。")
		}
		changed := base != "" && local != base
		publish := func(parents A) {
			n := syncPublish(dir, files, parents)
			head = str(n["id"])
			cfg["head"], cfg["baseline"] = head, local
			nodes[head] = n
		}
		status := "up_to_date"
		switch {
		case action == "resolve":
			require(choice == "local" || nodes[choice] != nil, "请选择现存版本或 local；旧版本将全部保留。")
			assertIdle(root)
			if changed {
				publish(stringsA(str(cfg["head"])))
				_, heads = syncSnapshots(dir)
			}
			if choice != "local" {
				files = syncImportedFiles(obj(nodes[choice]))
				local = syncDigest(files)
			}
			publish(heads)
			if choice != "local" {
				root = syncApply(root, obj(nodes[head]), syncDigest(syncFiles(root)))
			}
			status = "resolved"
		case len(heads) == 0:
			publish(A{})
			status = "published"
		case len(heads) == 1 && obj(nodes[str(heads[0])])["digest"] == local:
			cfg["head"], cfg["baseline"] = heads[0], local
		case len(heads) == 1 && str(heads[0]) == head:
			if changed {
				publish(heads)
				status = "published"
			}
		default:
			canImport := len(heads) == 1 && (!changed && base != "" || base == "" && syncFresh(root))
			if canImport && action != "push" {
				node := obj(nodes[str(heads[0])])
				root = syncApply(root, node, local)
				cfg["head"], cfg["baseline"] = heads[0], node["digest"]
				status = "restored"
			} else if canImport {
				status = "remote_available"
			} else {
				if changed || base == "" && !syncFresh(root) {
					parents := A{}
					if head != "" {
						parents = append(parents, head)
					}
					publish(parents)
				}
				status = "conflict"
			}
		}
		cfg["root"], cfg["last_checked"], cfg["last_result"] = root, now(), status
		delete(cfg, "last_error")
		writeJSON(syncConfigPath(), cfg)
		messages := M{"up_to_date": "本机档案与当前可见同步版本一致。", "published": "已保存到同步文件夹；请等待 iCloud 完成上传。", "restored": "已恢复学习进度，原本地档案仍保留。", "remote_available": "有其他电脑的新进度，下次练习前会恢复。", "conflict": "发现不同设备的进度分支，双方版本均保留。请让教练协助选择，未自动覆盖。", "resolved": "已采用所选进度；所有旧版本仍保留。"}
		result = M{"status": status, "message": messages[status], "data_root": root, "directory": cfg["directory"]}
		if status == "conflict" {
			_, heads = syncSnapshots(dir)
			result["versions"] = heads
		}
	})
	return result
}
func syncAfterSave(root string) M {
	if syncConfig()["root"] != absolute(root) {
		return M{}
	}
	var result M
	err := attempt(func() { result = syncRun(root, "push", "", "") })
	if err != nil {
		withLock(filepath.Join(filepath.Dir(configPath()), ".sync.lock"), func() {
			cfg := syncConfig()
			if cfg["root"] == absolute(root) {
				cfg["last_result"], cfg["last_error"] = "pending", err.Error()
				writeJSON(syncConfigPath(), cfg)
			}
		})
		return M{"sync_warning": "本地已保存；同步待重试：" + err.Error()}
	}
	return M{"sync": result}
}
func syncBeforePractice(w M) (M, M) {
	root := str(w["data_root"])
	if syncConfig()["root"] != absolute(root) {
		return w, M{"status": "disabled"}
	}
	// An ongoing call keeps its archive pinned; new data is read between practices.
	if attempt(func() { assertIdle(root) }) != nil {
		return w, M{"status": "busy", "message": "当前练习结束后再读取其他设备进度。"}
	}
	var result M
	err := attempt(func() { result = syncRun(root, "now", "", "") })
	if err != nil {
		return w, M{"status": "offline", "message": err.Error()}
	}
	require(result["status"] != "conflict", str(result["message"]))
	if str(result["data_root"]) != root {
		w = workspace("")
	}
	return w, result
}
