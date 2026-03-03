# StagentHand Development Guide

## Environment
- Go 1.25+
- SQLite3
- Environment Variables: GEMINI_API_KEY (optional for dry-run)

## Common Commands
- Build: `go build -o shand`
- Test: `go test ./...`
- Run Server: `./shand server`
- Run Job (Dry Run): `./shand run --dry-run`

## Workflow
1. Story → Outline (LLM)
2. Outline → Storyboard (LLM)
3. Storyboard → Panels (LLM)
4. Panels → Images (Imagen/NanaBanana)
5. Images → Video (Grok/Remotion)

## HITL (Human-In-The-Loop)
- Reviewers (Jared or Human) can approve/reject via CLI (`shand approve <id>`) or HTTP POST to `:28080/checkpoints/<id>/approve`.
