# Kopi FileArchiver

## What This Is

Kopi 是一个用于本地文件归档的 Go 工具，提供 CLI 和 Web UI 两种入口。它支持按分类与模板进行重命名归档、批量上传处理，以及可选 AI 分类能力，目标用户是需要快速整理文档的个人或小团队。

## Core Value

文件归档必须稳定、可预测且不会错误写入到非预期路径。

## Requirements

### Validated

- ✓ 支持 CLI 单文件归档（带版本号递增）— existing
- ✓ 支持 Web 批量上传并逐文件返回结果 — existing
- ✓ 支持配置分类映射和归档模板 — existing
- ✓ 支持 OpenAI 兼容接口进行文件名分类 — existing

### Active

- [ ] 完成全量代码审计并修复高风险输入校验缺陷
- [ ] 为核心归档和 Web API 增加自动化测试覆盖
- [ ] 明确性能基线并优化批量归档路径

### Out of Scope

- 移动端 App — 当前目标是本地工具优先
- 分布式多节点归档服务 — 现阶段不需要跨机器协调

## Context

- 项目语言为 Go 1.20，当前模块结构清晰：`internal/archiver` 与 `internal/webui`。
- 原始仓库此前未初始化 git，已在本次初始化时创建仓库。
- 当前审计发现的主要风险集中在：配置合法性校验不足、模板渲染结果缺少路径安全约束、Web JSON 请求体无显式上限。
- README 存在编码异常（展示中文乱码），属于文档质量问题，需要在后续阶段修复。

## Constraints

- **Compatibility**: 保持现有 CLI/Web 用法不变 — 避免影响已有使用方式
- **Security**: 输入需默认拒绝危险值 — 防止路径穿越和异常请求导致故障
- **Performance**: 批量归档场景需保证稳定吞吐 — 不能引入明显回退

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 作为 brownfield 项目直接初始化 `.planning` 并并行审计代码 | 用户目标是现有项目审计优化，不应阻塞在前置映射流程 | ✓ Good |
| 优先修复输入校验与请求解析稳定性 | 这类问题风险最高且改动小收益大 | ✓ Good |
| 先补关键测试再扩展功能优化 | 降低后续迭代引入回归的概率 | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-13 after initialization and first audit pass*
