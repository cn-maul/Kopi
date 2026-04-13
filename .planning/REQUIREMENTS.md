# Requirements: Kopi FileArchiver

**Defined:** 2026-04-13
**Core Value:** 文件归档必须稳定、可预测且不会错误写入到非预期路径。

## v1 Requirements

### Core Archiving

- [ ] **ARCH-01**: 用户可以通过 CLI 归档单文件并得到递增版本号文件名
- [ ] **ARCH-02**: 用户可以通过 Web 批量上传多个文件并逐个返回处理结果
- [ ] **ARCH-03**: 模板渲染后的前缀必须拒绝路径分隔符与危险字符
- [ ] **ARCH-04**: 分类缩写必须通过格式校验，防止非法目录片段

### Configuration

- [ ] **CONF-01**: 缺失配置文件时系统自动创建默认 `config.yaml`
- [ ] **CONF-02**: 配置保存时必须校验 AI 参数完整性（url/apiKey/modelName）
- [ ] **CONF-03**: 配置读取时必须对关键字段执行安全校验

### API Robustness

- [ ] **API-01**: `/api/config` 与 `/api/ai/test` JSON 请求体限制为 1MB
- [ ] **API-02**: JSON 解析拒绝未知字段与多余 token
- [ ] **API-03**: 错误响应使用统一 JSON 结构返回

### Quality

- [ ] **TEST-01**: 为 `config` 模块增加默认配置与非法缩写测试
- [ ] **TEST-02**: 为模板安全校验增加路径字符测试
- [ ] **TEST-03**: 为 Web JSON 解码增加超限与非法字段测试

## v2 Requirements

### Usability

- **DOC-01**: 修复 README 中文编码问题并补充常见故障排查
- **UX-01**: Web 上传页增加批量结果导出（CSV/JSON）

### Performance

- **PERF-01**: 增加批量归档性能基准（1k 文件）
- **PERF-02**: 评估并实现可选并发归档执行策略

## Out of Scope

| Feature | Reason |
|---------|--------|
| 云端对象存储适配（S3/OSS） | 当前定位是本地工具，避免范围膨胀 |
| 用户权限系统/多租户 | 单机工具场景不需要 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| ARCH-01 | Phase 1 | Pending |
| ARCH-02 | Phase 1 | Pending |
| ARCH-03 | Phase 1 | Pending |
| ARCH-04 | Phase 1 | Pending |
| CONF-01 | Phase 1 | Pending |
| CONF-02 | Phase 1 | Pending |
| CONF-03 | Phase 1 | Pending |
| API-01 | Phase 1 | Pending |
| API-02 | Phase 1 | Pending |
| API-03 | Phase 1 | Pending |
| TEST-01 | Phase 1 | Pending |
| TEST-02 | Phase 1 | Pending |
| TEST-03 | Phase 1 | Pending |

**Coverage:**
- v1 requirements: 13 total
- Mapped to phases: 13
- Unmapped: 0 ✓

---
*Requirements defined: 2026-04-13*
*Last updated: 2026-04-13 after initial audit and optimization*
