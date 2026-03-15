# StagentHand (`shand`)

![StagentHand Banner](assets/banner.png)

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **CLI-first AI short drama pipeline — fully automated, agent-driven production.**

> **CLI-first AI 短劇製作 Pipeline — 全自動、專為 Agent 設計的產線。**

---

## Pipeline Flow / 管線流程

```
Story Prompt
  ↓ story-to-outline       (LLM)
Outline JSON
  ↓ outline-to-storyboard  (LLM)
Storyboard JSON
  ↓ storyboard-to-panels   (LLM)
Panel[] JSON
  ↓ panels-to-images       (Nano Banana 2 / Nova Canvas, concurrent)
Panel[] + image_url
  ↓ TTS                    (Amazon Polly Neural + SSML)
Panel[] + audio_url
  ↓ BGM                    (Jamendo API)
  ↓ storyboard-to-remotion-props
RemotionProps JSON
  ↓ remotion-render        (npx remotion)
output.mp4
  ↓ critic                 (Amazon Nova Pro, multimodal)
APPROVE / REJECT
```

---

## Features / 功能特色

### Core Pipeline / 核心管線

**EN:** End-to-end pipeline from a raw story prompt to a rendered MP4. Every stage reads from stdin and writes to stdout as JSON, composable with standard Unix tools.

**中文：** 從一句故事描述直接產出 MP4。每個階段以 JSON 從 stdin 讀取、stdout 輸出，可任意與 Unix 工具組合。

### LLM Support / LLM 支援

**EN:** Three providers supported out of the box. Priority: flag > env > `~/.shand/config.yaml` > defaults.

**中文：** 三個提供商開箱即用，優先順序：flag > 環境變數 > config 檔 > 預設值。

| Provider | Config value |
|---|---|
| AWS Bedrock (Claude / Nova) | `llm.provider: bedrock` |
| OpenAI-compatible (Gemini, local) | `llm.provider: openai` or `gemini` |
| Google Gemini | `llm.provider: gemini` |

### Image Generation / 圖像生成

**EN:** Two providers. Nano Banana 2 supports character reference images for cross-panel consistency. Nova Canvas is the AWS Bedrock option.

**中文：** 兩個圖像提供商。Nano Banana 2 支援角色參考圖保持跨鏡頭一致性；Nova Canvas 為 AWS Bedrock 選項。

| Provider | Config value |
|---|---|
| Nano Banana 2 (Gemini-based) | `image.provider: nanobanana` |
| AWS Nova Canvas | `image.provider: nova` |

### Text-to-Speech / 語音合成

**EN:** Amazon Polly Neural (voice: Zhiyu, Mandarin Chinese). Dialogue is automatically wrapped in SSML. Whisper cues (`Whisper: ...`) are detected and mapped to Polly's whispered effect. Speech rate fixed at 90% for a dramatic, non-rushed delivery.

**中文：** Amazon Polly Neural（Zhiyu 中文語音）。對白自動包裝成 SSML；偵測 `Whisper:` 標記並轉為 Polly 悄聲效果；語速鎖定 90% 避免急促感。

### Background Music / 背景音樂

**EN:** Jamendo API integration. Tags are driven by the `BGMTags` directive (e.g. `cinematic+dark`). The pipeline searches, picks the first match, and downloads the MP3 automatically.

**中文：** 整合 Jamendo API，由 `BGMTags` directive 驅動（如 `cinematic+dark`）。自動搜索、選取第一首並下載 MP3。

### AI Critic / AI 評審

**EN:** Post-render evaluation using Amazon Nova Pro (multimodal). The critic watches the actual MP4 and scores across 4 dimensions. Hard-stop thresholds: `visual_score ≥ 8`, `audio_sync_score ≥ 8`, total `≥ 32/40`.

**中文：** 使用 Amazon Nova Pro 多模態模型對渲染後的 MP4 進行評審。4 個維度評分，強制閾值：視覺 ≥ 8、音視頻同步 ≥ 8、總分 ≥ 32/40。

| Dimension | Description |
|---|---|
| Visual Coherence (A) | Character consistency, subtitle cleanliness |
| Audio-Visual Sync (B) | BGM ducking, voice naturalness, subtitle timing |
| Directive Adherence (C) | BGM mood match, visual directive compliance |
| Narrative Tone (D) | Pacing, dramatic breathing room, story closure |

### Directives System / Directives 配置系統

**EN:** Two global directives injected into the pipeline via JSON:

- `style_prompt`: Prepended to every panel's image generation prompt for visual consistency.
- `bgm_tags`: Passed to Jamendo for music mood selection.

Additional per-panel `PanelDirective` fields control camera motion (`ken_burns_in`, `pan_left`, etc.), transition type, subtitle position, and font size.

**中文：** 兩個全域 directive 透過 JSON 注入：

- `style_prompt`：自動前置到每個 panel 的圖像生成 prompt，確保視覺風格統一。
- `bgm_tags`：傳給 Jamendo 控制音樂情境。

另有 per-panel `PanelDirective`，可控制鏡頭動效、轉場類型、字幕位置與字體大小。

### Smart Resume / 智能恢復機制

**EN:** Asset-aware caching. If a pipeline run is interrupted, re-running skips panels whose `image_url` or `audio_url` files already exist on disk. No duplicate API calls, no duplicate costs.

**中文：** 檔案感知快取。管線中途失敗後重啟，自動跳過磁碟上已存在的 `image_url` / `audio_url` 資產。不重複呼叫 API，不浪費費用。

### Human-in-the-Loop / 人類監控

**EN:** Four HITL checkpoints: `outline`, `storyboard`, `images`, `final`. All three approval channels write to the same SQLite record.

**中文：** 四個 HITL 檢查點：`outline`、`storyboard`、`images`、`final`。三種審核管道都寫入同一個 SQLite 記錄。

```
story → [outline ⏸] → [storyboard ⏸] → [images ⏸] → [final ⏸] → mp4
```

| Channel | How |
|---|---|
| CLI | `shand checkpoint approve <id>` |
| Discord | Webhook → bot reply |
| HTTP API | `POST :28080/checkpoints/:id/approve` |

### Agent Friendly / Agent 友好設計

**EN:** Built with AI agents as first-class consumers. Strict input sanitization blocks path traversal (`../../../.ssh`), double-encoding (`%2e%2e`), and control character injection. Non-zero exit codes and structured stderr errors let agents retry predictably.

**中文：** 以 AI Agent 為第一優先使用者。嚴格的輸入防護阻擋目錄穿越、雙重編碼與控制字元注入。非零 exit code 加上結構化 stderr 讓 Agent 可預測地進行 retry。

---

## Quick Start / 快速開始

### Prerequisites / 環境需求

```bash
# Go 1.23+, Node.js 20+, FFmpeg, AWS CLI
brew install awscli ffmpeg node
go build -o shand .
```

### End-to-end run / 全流程執行

```bash
echo "機器人找到了一朵會發光的花" | ./shand pipeline --skip-hitl
```

### Resume from existing panels / 從現有 panels 恢復

```bash
cat ~/.shand/projects/my-id/remotion_props.json | ./shand pipeline --skip-hitl
```

### Render only / 只執行渲染

```bash
cat remotion_props.json | ./shand remotion-render --output ./final.mp4
```

### Run AI Critic / 執行 AI 評審

```bash
./shand critic --video ./final.mp4 --props ./remotion_props.json
```

---

## Configuration / 配置

**EN:** Default config path: `~/.shand/config.yaml`. Env vars use `SHAND_` prefix (e.g. `SHAND_LLM_API_KEY`). Flags take highest priority.

**中文：** 預設配置路徑：`~/.shand/config.yaml`。環境變數使用 `SHAND_` 前綴（如 `SHAND_LLM_API_KEY`）。CLI flag 優先級最高。

```yaml
llm:
  provider: bedrock          # bedrock | openai | gemini
  model: amazon.nova-pro-v1:0
  aws_access_key_id: ""
  aws_secret_access_key: ""
  aws_region: us-east-1
  # For openai/gemini:
  # api_key: ${GOOGLE_API_KEY}
  # base_url: ""             # Leave empty for default; any OpenAI-compatible URL works

image:
  provider: nanobanana        # nanobanana | nova
  api_key: ${GOOGLE_API_KEY}
  width: 1024
  height: 576
  # For nova (AWS Bedrock):
  # provider: nova
  # access_key_id: ""
  # secret_key: ""
  # region: us-east-1

audio:
  voice_provider: polly       # polly (default)
  music_provider: jamendo     # jamendo (default)
  jamendo_client_id: ""       # Leave empty to use public test key

remotion:
  template_path: ./remotion-template
  composition: ShortDrama

notify:
  discord_webhook: ${DISCORD_WEBHOOK_URL}

store:
  db_path: ~/.shand/shand.db

server:
  port: 28080                 # HTTP API for agent / Discord bot checkpoint approval
```

---

## Commands Reference / 指令參考

**EN:** All commands read JSON from stdin and write JSON to stdout unless noted. Use `--dry-run` for any command to validate without calling external APIs.

**中文：** 所有指令從 stdin 讀取 JSON，輸出到 stdout（另有說明除外）。所有指令支援 `--dry-run` 驗證，不呼叫外部 API。

| Command | Description (EN) | 說明 |
|---|---|---|
| `shand pipeline` | Full pipeline: story → mp4 | 全流程：故事 → mp4 |
| `shand story-to-outline` | Story prompt → Outline JSON (LLM) | 故事描述 → 大綱 JSON |
| `shand outline-to-storyboard` | Outline JSON → Storyboard JSON (LLM) | 大綱 → 分鏡腳本 |
| `shand storyboard-to-panels` | Storyboard JSON → Panel[] JSON (LLM) | 分鏡腳本 → 畫格列表 |
| `shand panel-to-image` | Generate image for a single panel | 生成單一畫格圖像 |
| `shand panels-to-images` | Batch image generation (concurrent) | 批量並發圖像生成 |
| `shand storyboard-to-remotion-props` | Panel[] → RemotionProps JSON | 畫格列表 → Remotion 配置 |
| `shand remotion-render` | Render MP4 via Remotion | 渲染 MP4 |
| `shand remotion-preview` | Open Remotion Studio (blocking) | 開啟 Remotion Studio 預覽 |
| `shand critic` | AI Critic multimodal video evaluation | AI 多模態視頻品質評審 |
| `shand checkpoint list` | List all HITL checkpoints | 列出所有 HITL 檢查點 |
| `shand checkpoint approve <id>` | Approve a checkpoint | 批准檢查點 |
| `shand checkpoint reject <id>` | Reject a checkpoint | 拒絕檢查點 |
| `shand checkpoint wait <id>` | Poll until checkpoint resolves | 輪詢直到檢查點完成 |
| `shand status <job-id>` | Query job status | 查詢任務狀態 |

### Key flags / 常用 flags

| Flag | Applies to | Effect |
|---|---|---|
| `--dry-run` | All commands | Skip external API calls, return mock JSON |
| `--skip-hitl` | `pipeline` | Disable all 4 HITL pause points |
| `--output <path>` | `remotion-render` | Output MP4 path |
| `--video <path>` | `critic` | Path to rendered MP4 |
| `--props <path>` | `critic` | Path to `remotion_props.json` |
| `--config <path>` | All commands | Override default config file path |

---

## Architecture / 架構設計

**EN:** SOLID-based layered architecture. The `cmd/` layer is thin: IO only, no business logic. All external services are accessed through interfaces, injected at construction time.

**中文：** 基於 SOLID 原則的分層架構。`cmd/` 層只負責 IO，不含業務邏輯。所有外部服務均透過 interface 存取，在建構時注入。

```
cmd/                   Thin layer: IO + dependency injection
internal/
  domain/              Pure data structs, zero external dependencies
  llm/                 LLMClient interface + Bedrock / OpenAI-compatible / Mock
  image/               ImageClient interface + NanoBanana / NovaCanvas / Mock
  audio/               Polly TTS + Jamendo BGM client
  video/               AI Critic (Nova Pro multimodal)
  store/               Repository pattern: JobRepo + CheckpointRepo (SQLite/gorm)
  notify/              Notifier interface + Discord webhook
  remotion/            RemotionExecutor interface + exec npx remotion
  pipeline/            Orchestrator — depends on all interfaces only
config/                viper loader: flag > env > yaml > defaults
remotion-template/     React + Remotion (ShortDrama composition)
```

### SOLID at a glance

| Principle | Implementation |
|---|---|
| Single Responsibility | Each package owns exactly one domain |
| Open/Closed | New provider = implement interface, touch nothing else |
| Liskov Substitution | Every Mock is a drop-in replacement, same behavioral contract |
| Interface Segregation | `LLMClient`, `ImageClient`, `AudioBatcher`, `MusicBatcher` are separate |
| Dependency Inversion | `cmd/` depends on interfaces; concrete types injected via constructors |

---

## For AI Agents / 給 AI Agent 的使用指引

**EN:** `shand` is designed to be controlled by AI agents without human intervention.

**中文：** `shand` 設計為可由 AI Agent 在無人工干預的情況下完全控制。

```bash
# Full automated run — agent controls everything
echo "太空飛行員愛上了外星植物學家" | ./shand pipeline --skip-hitl

# Agent approves a HITL checkpoint via HTTP
curl -X POST http://localhost:28080/checkpoints/<id>/approve

# Agent reads structured exit codes
./shand pipeline --skip-hitl
echo $?   # 0=success, 1=failed, 2=waiting_hitl

# Agent pipes stages independently
echo "story text" \
  | ./shand story-to-outline \
  | ./shand outline-to-storyboard \
  | ./shand storyboard-to-panels \
  | ./shand panels-to-images \
  | ./shand storyboard-to-remotion-props \
  | ./shand remotion-render --output ./out.mp4
```

**Input hardening / 輸入防護:**

All user-supplied strings (IDs, file paths, prompts) pass through `internal/domain` sanitization before use. The pipeline rejects path traversal sequences, double-encoded characters, and control characters. Agents are treated as untrusted sources.

所有使用者提供的字串（ID、路徑、prompt）在使用前都會通過 `internal/domain` 的淨化邏輯。管線會拒絕目錄穿越序列、雙重編碼字元和控制字元。Agent 被視為不可信來源。

---

## Development Status / 開發狀態

| Phase | Status | Deliverables |
|---|---|---|
| Phase 1 | Done | CLI skeleton, viper config, domain types, SQLite/gorm, status/checkpoint |
| Phase 2 | Done | LLM interface, story-to-outline / outline-to-storyboard / storyboard-to-panels |
| Phase 3 | Done | Image interface, panel-to-image / panels-to-images, Discord notify |
| Phase 4 | Done | Remotion template, storyboard-to-remotion-props, render/preview |
| Phase 5 | Done | Pipeline orchestrator, 4-node HITL, end-to-end tests |
| Phase 6 | Done | AWS Bedrock LLM/Image, Amazon Polly Neural TTS + SSML, audio sync |
| Phase 7 | Done | AI Critic (multimodal), Jamendo BGM, subtitle sanitization, dynamic duration |
| Phase 8 | Done | Directives system (StylePrompt / BGMTags), Smart Resume |
| **Phase 9** | **In progress** | — |

---

## License / Credits

**EN:** MIT License. See [LICENSE](LICENSE).

Built by **Castle Studio**. Developed using a dual-model workflow: Claude (implementation) + Codex (review).

**中文：** MIT 授權。詳見 [LICENSE](LICENSE)。

由 **Castle Studio** 開發。採用雙模型工作流：Claude（施工）+ Codex（審核）。

---

*StagentHand — Part of the Castle Studio C3A ecosystem.*
*Binary: `shand` | Module: `github.com/baochen10luo/stagenthand`*
