# 语言交流 / Language Exchange

一个 Codex 插件、一张学习网页。在侧栏“练习语言”选择英语或日语，之后对 Codex 说“带我练口语”。也可直接说“练英语”或“练日语面试”，明确请求优先。

- 自由对话保留原有英语教学规则；日语支持日常生活、工作沟通和面试共 18 个场景，卡住时中文辅助。
- 同一网页显示对应语言的字幕、复盘与卡片；日语包含可展开的假名与礼貌建议。
- 选择用于下一次练习，正在进行的 Voice 和课后复盘保持原语言。换语言请另开一次语音。
- 英语与日语档案分别保存在原有本地目录，已有记录不迁移、不混合。网页备份只导出当前语言。

打开 [学习页](http://127.0.0.1:8897/)；首次安装/更新后，在新 Codex 任务使用 `$english-speaking-coach`。Voice 开启后才有实时字幕；文字练习不会伪造语音记录。

## 引导学习

首页点“引导学习”，在左侧说“带我学英语：说明进度并请求帮助”或“带我学日语：说明进度并请求帮助”。每种语言先提供这一节微课：简短示范 → 替换成自己的任务 → 收起答案独立尝试 → 隔天换情境复测。教学时可用中文求助，两句表达分别记录提示程度。

网页随步骤更新；独立尝试和复测接口不返回示范答案。首次尝试后隔天复测；仍需帮助时隔天再练，独立完成复测后七天再测。它记录教练观察，不宣称掌握、发音合格或无字幕听懂。原生 Voice 的实际教学行为仍需真实练习验收。

练习进度分别存于各语言档案的 `Practice/progress-help.json`，纳入现有备份。“自由练英语/日语”继续原有对话，切换不会清除微课进度。

## 构建与验证

```sh
go test ./...
(cd languages/ja/runtime && go test ./...)
go run ./internal/release -targets darwin/arm64
```

构建产物在 `dist/english-speaking-coach_v1.0.2_plugin.zip`，同时包含两个已校验的运行程序，仅暴露一个 Skill；运行无需 Go/Python。发布器支持原有 macOS、Linux、Windows 目标。请构建本分支插件包，不能用上游纯英语二进制替换。

统一入口：`scripts/coach language` 查看选择，`scripts/coach language set en|ja` 保存选择。每次练习的全部命令显式带 `--language en|ja`，以免网页选择改变正在处理的旧课次。维护主网页服务用 `service start|status|stop|resume`；加 `--language ja` 管理内部日语服务。网页会按需启动内部服务，不接管冲突端口或主动停止的服务。

保留原有英语数据路径与格式，日语继续使用独立的 `~/.codex/japanese-speaking-coach/data`。内部日语进程使用 8898，用户从 8897 的语言选项进入。两种语言的数据恢复相互拒绝。

<!-- ponytail: keep the tested language engines separate to preserve English teaching behavior; consolidate their shared storage code when upstream updates make maintaining both costly. -->
本分支基于 yomage-ai/english-speaking-coach。英语自由对话保持上游规则，日语功能与语言选择由本分支追加；许可见 LICENSE。真实 Voice 的转写、发音和宿主回应仍需实际语音验收。
