# Japanese Speaking Coach

个人日语口语练习：日常生活、工作沟通与求职面试。左侧用 Codex Voice 对话，右侧查看日中字幕、可展开的假名、复盘和表达卡。

这是 Ccryrt 基于 yomage-ai/english-speaking-coach `a2c798c` 制作的独立日语衍生版本，保留原始许可证。修改包括日语教学规则、18 个场景、整句日中翻译、假名/语域记录和独立运行配置；不是上游官方日语版。

英语版仍位于仓库根目录，其文件、默认端口和安装保持原样。

| | 英语版 | 日语版 |
|---|---|---|
| 插件 | english-speaking-coach | japanese-speaking-coach |
| 本地网页 | 127.0.0.1:8897 | 127.0.0.1:8898 |
| 学习目录 | ~/.codex/english-speaking-coach/data | ~/.codex/japanese-speaking-coach/data |
| 备份 | 原英语备份格式 | 独立日语备份格式 |

## 开发与安装

需要 Go 1.26.2+ 和 Python 3 来构建。安装后的运行程序不依赖 Go/Python。

```sh
cd japanese-speaking-coach
python3 scripts/build.py
(cd runtime && go test ./...)
```

构建产生当前平台的 `skills/japanese-speaking-coach/bin/japanese-coach`（Windows 为 `.exe`）和校验文件，不提交二进制或学习数据。将此目录作为名为 `japanese-speaking-coach` 的本地插件添加到 Codex 后，在新任务使用 `$japanese-speaking-coach`。英语插件可以同时启用。

例如：“练日语点餐”“练日语汇报项目进度”“练日语面试，介绍我的项目经历”。打开 Voice 后才有实时字幕；文字练习也可以保存复盘，但不生成虚构的 Voice。

`--topic daily|work|interview|mixed` 可以选择本次场景类别。中文求助不会被禁止，默认以日语继续对话。假名和礼貌建议是模型生成的学习辅助，不是发音评分；面试回答只采用学习者真实提供的经历。

## 验证

日语测试覆盖整句/纯汉字/混合语种字幕、错误结果隔离、原文修订、假名、复盘入库、页面接口、备份与英语档案隔离。`JAPANESE_COACH_MODEL_TEST=1 go test -run TestLiveModelJapanese -v` 可通过现有 Codex 登录对虚构材料做真实翻译与复盘模型检查，不写入用户档案。

真实 Voice 转写和宿主实际日语回应仍需在启用 Voice 的任务中验证，模型测试和网页演示不能替代。

<!-- ponytail: pinned independent runtime copy keeps English untouched; consolidate shared code only after Japanese behavior is accepted and an English refactor is authorized. -->
当前有意保留一份固定的运行代码副本，避免本轮修改英语版；后续合并上游修复需分别评估。未复制旧 Python 运行实现和历史发布流程。

许可与版权见 [LICENSE](LICENSE)。
