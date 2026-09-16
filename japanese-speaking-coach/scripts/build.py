#!/usr/bin/env python3
"""Build the standalone Japanese plugin for this developer's current platform."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[1]
runtime = root / "runtime"
output = root / "skills/japanese-speaking-coach/bin"
output.mkdir(parents=True, exist_ok=True)
manifest = json.loads((root / ".codex-plugin/plugin.json").read_text())
sources = sorted(p for p in runtime.rglob("*") if p.is_file())
digest = hashlib.sha256()
for path in sources:
    digest.update(str(path.relative_to(runtime)).encode())
    digest.update(path.read_bytes())
revision = digest.hexdigest()[:16]
binary = output / ("japanese-coach.exe" if os.name == "nt" else "japanese-coach")
with tempfile.TemporaryDirectory(dir=output, prefix=".build-") as temporary:
    built = Path(temporary) / binary.name
    subprocess.run(["go", "build", "-trimpath", "-ldflags",
                    f"-s -w -X main.version=v{manifest['version']} -X main.revision={revision}",
                    "-o", str(built), "."], cwd=runtime, check=True)
    checksum = hashlib.sha256(built.read_bytes()).hexdigest()
    os.replace(built, binary)
    binary.with_suffix(binary.suffix + ".sha256").write_text(checksum + "\n")
print(f"Built {binary.name}, version {manifest['version']}, source {revision}")
