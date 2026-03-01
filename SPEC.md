# waoo — CLI-first AI 影視製作 Pipeline

> 建檔日期：2026-03-01
> 負責人：005 瑪勒列（開發主管）
> 狀態：Phase 1 準備中

## 背景

參考 waoowaoo / huobao-drama 的流程，從頭用 Go 實作。不 fork，重新設計。

核心哲學：Unix philosophy、Agent-friendly、HITL、模型可配置

## 技術棧
- Go 1.22+ / cobra / Gin / SQLite(gorm) / viper
- Remotion（npx exec）/ Discord webhook

## 實作階段
- Phase 1：CLI 骨架 + config + SQLite store + status/checkpoint
- Phase 2：LLM interface + story-to-outline / outline-to-storyboard / storyboard-to-panels
- Phase 3：image gen interface + panel-to-image / panels-to-images + Discord notify
- Phase 4：remotion-template + storyboard-to-remotion-props + render/preview
- Phase 5：waoo pipeline orchestrator + HITL 四節點 + e2e test

完整規格見 SPEC_FULL.md
