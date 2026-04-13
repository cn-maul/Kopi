# Roadmap: Kopi FileArchiver

## Milestone v1.0 — Stability and Audit Hardening

### Phase 1: 审计修复与稳定性加固

**Goal:** 完成首轮代码审计并修复高优先级风险，建立测试护栏。

**Plans:**
1. 配置与模板安全校验
2. Web API 请求解析稳健性增强
3. 单元测试与回归验证

**Requirements Covered:**
- ARCH-01, ARCH-02, ARCH-03, ARCH-04
- CONF-01, CONF-02, CONF-03
- API-01, API-02, API-03
- TEST-01, TEST-02, TEST-03

**Success Criteria:**
- `go test ./...` 全通过
- `go vet ./...` 无告警
- 非法配置与危险模板输入被拒绝
- JSON 请求体超限与未知字段被拒绝

### Phase 2: 文档与可用性修复

**Goal:** 提升可维护性与用户上手质量。

**Plans:**
1. 修复 README 中文乱码与说明缺失
2. 增加错误案例与配置示例
3. 校验脚本在 Linux/Windows 的一致性

### Phase 3: 性能观测与优化

**Goal:** 给出可量化的批量归档性能基线并优化热点路径。

**Plans:**
1. 增加 benchmark 与压测样例
2. 评估并发归档策略与 I/O 开销
3. 输出性能优化报告与参数建议

## Notes

- 当前项目属于 brownfield，优先保护兼容性，避免破坏已有 CLI/Web 用法。
- 每个阶段完成后应更新 `.planning/PROJECT.md` 与 `.planning/REQUIREMENTS.md` 的状态。

---
*Last updated: 2026-04-13 after new-project initialization*
