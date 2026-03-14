# StagentHand (`shand`)

![StagentHand Banner](assets/banner.png)

> CLI-first AI 短劇製作 Pipeline — 讓 AI 分身 (Agent) 在幕後全自動運行的產線。

---

## 核心設計理念 (Filosofy)

`shand` 不只是一個工具，它是一套符合 **"Good Taste"** (Linus 式品味) 的 Unix 工具鏈。
1. **Never break userspace**: 設定檔與 JSON 介面保持穩定。
2. **數據驅動 (Data-Driven)**: 整個管線是透過 JSON 傳遞，這讓機器人 (Agent) 能極易操縱它。
3. **人類可控 (Human-in-the-loop)**: 在產生的關鍵節點暫停，讓你（或你的 Agent 監督者）決定品質。
4. **省錢智慧 (Smart Resume)**: 內建檔案感知的快取機制，產過的圖、生過的語音，絕不花第二次錢。

---

## 管線流程

```text
故事描述 → 大綱 → 分鏡腳本 → 產出畫格 → 圖像生成 → 語音合成 (TTS) → Remotion 合成 → mp4
```

---

## 核心功能特色

### 🎙️ 語音旁白整合 (Audio/TTS)
整合 **Amazon Polly (Neural)**。現在影片不再是默片，具備自然的人聲口白，並與畫面精確對齊。

### 💰 智能恢復機制 (Smart Resume)
**最省錢的產線。** 如果產圖或語音在中間出錯，修正後重啟，`shand` 會自動跳過已存在的資產檔案，只針對缺失部分呼叫 API。

### 🤖 Agent Friendly
專為 AI Agent（如 Claude, GPT-4, Gemini）設計。如果你是 Agent，你可以：
- 透過 `shand pipeline` 讀取 stdout 的 JSON 進行下一步決策。
- 透過 `remotion_props.json` 直接注入自定義的分鏡內容。
- 藉由 `shand checkpoint approve` 實現自動化 QC。

---

## 快速開始

### 1. 安裝環境
```bash
# 需要 Go 1.23+, Node.js 20+, FFmpeg
# 需要安裝 AWS CLI 以支援 Polly TTS
brew install awscli
go build -o shand .
```

### 2. 配置 (`~/.shand/config.yaml`)
```yaml
llm:
  provider: bedrock  # 或 openai, gemini
  aws_access_key_id: "YOUR_KEY"
  aws_secret_access_key: "YOUR_SECRET"
  aws_region: "us-east-1"
remotion:
  template_path: "./remotion-template"
```

### 3. 指令集

#### 全流程運行
```bash
# 從一個故事點子直接產出影片
echo "機器人找到了一朵會發光的花" | ./shand pipeline --skip-hitl
```

#### 局部更新 (Resume)
如果你已經有一個專案想要追加聲音：
```bash
# 讀取現有的 props，自動補齊音軌並合成
cat ~/.shand/projects/my-id/remotion_props.json | ./shand pipeline --skip-hitl
```

#### 影片渲染
```bash
# 單獨執行渲染
cat remotion_props.json | ./shand remotion-render --output ./final.mp4
```

---

## 開發者與 Agent 指引

### 數據結構規範
- **Panel**: 核心單元。包含 `description` (產圖用), `dialogue` (旁白用), `image_url` 與 `audio_url`。
- **RemotionProps**: 最終輸入。所有絕對路徑會被 `./shand` 自動 normalize 為虛構路徑 `/shand/...` 以配合 Remotion 的 Security Policy。

### 任務清單與進度 (Phase 4 & 5 完成)
- [x] **Phase 4** — Remotion 模板實作 + Ken Burns 鏡頭動效
- [x] **Phase 5** — Pipeline Orchestrator + **Smart Resume 快取系統**
- [x] **Phase 6** — **Amazon Polly 語音串聯** + 跨平台路徑標準化
- [x] **Phase 7** — **AWS CLI 整合與權限自動化**

---

## 專案命名

**StagentHand** = Stage + Agent + Hand (舞台工作人員)。
在幕後默默耕耘，把 AI 的創意拼湊成最終的傑作。

---
*Created by Castle Studio. Design with Good Taste.*
