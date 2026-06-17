# 时间偏好学习 — 技术规格

| 文档版本 | 日期 | 作者 | 说明 |
|---------|------|------|------|
| V1.0 | 2026-06-16 | 后端开发 | 基于 PRD V1.5 — F-017 偏好学习 |

## 1. 背景

PRD V1.5 规划了「偏好学习」（F-017）和「智能建议」（F-015）两项功能。本规格定义 V1.5 第一阶段：**时间偏好学习**——从用户历史日程中归纳时间习惯，在对话创建日程时注入 LLM 上下文，使 AI 回复更贴合用户个人规律。

完整知识引擎分两期交付：
- **V1.5**：时间偏好学习（本规格），轻量实时检索
- **V2.0**：外部知识注入（文档解析 + 向量检索 + RAG）

## 2. 设计目标

- 从用户历史日程中实时检索相似记录，归纳时间偏好
- 偏好上下文注入 Agent Server 的 LLM Prompt，不影响对话主流程
- 任何环节出错静默降级，不阻塞对话
- 冷启动保护：不足 5 条日程时不注入偏好
- 零新依赖、零数据库变更

## 3. 架构位置

```
用户消息 "周三下午开个评审会"
      │
      ▼
┌─ API Server ─────────────────────────────────────┐
│                                                   │
│  conversation/service                             │
│    │                                              │
│    ├─ ① 查询冲突日程（已有逻辑）                    │
│    │                                              │
│    ├─ ② 新增：偏好检索                            │
│    │   SELECT title, tags, start_time, end_time    │
│    │   FROM schedules                             │
│    │   WHERE user_id = ? AND deleted_at IS NULL   │
│    │   AND start_time > NOW() - INTERVAL 90 DAY   │
│    │   ORDER BY start_time DESC LIMIT 50          │
│    │   → Go 内存中按标签/时段/星期聚类            │
│    │                                              │
│    └─ ③ 组装 gRPC Request                        │
│        ProcessMessageRequest {                    │
│          existing_schedules: [...]                │
│          preference_context: "..."  ← 新增字段     │
│        }                                          │
└──────────────────┬───────────────────────────────┘
                   │ gRPC
                   ▼
┌─ Agent Server ───────────────────────────────────┐
│  系统 Prompt 注入偏好上下文 → LLM 推理             │
└──────────────────────────────────────────────────┘
```

**关键决策**：Agent Server 不碰 MySQL。检索和聚类在 API Server 完成，通过 gRPC 的 `preference_context` 字段传给 Agent。

## 4. 检索与聚类

### 4.1 检索 SQL

```sql
SELECT title, tags, start_time, end_time, DAYOFWEEK(start_time) AS dow
FROM schedules
WHERE user_id = ?
  AND deleted_at IS NULL
  AND start_time >= DATE_SUB(NOW(), INTERVAL 90 DAY)
ORDER BY start_time DESC
LIMIT 50
```

- 查询窗口：90 天
- 返回上限：50 条
- 性能：单次索引扫描，P99 < 5ms（MySQL `idx_user_time` 索引覆盖）

### 4.2 时段划分

| 时段 | 范围 |
|------|------|
| 上午 | 06:00–12:00 |
| 下午 | 12:00–18:00 |
| 晚上 | 18:00–22:00 |
| 深夜 | 22:00–06:00 |

### 4.3 聚类逻辑

50 条日程 → 按 `tags` 字段分组 → 无标签的用标题关键词匹配（"会"→工作 / "课"→学习 / "健身"/"跑步"→运动 / "评审"→工作）。

每组输出：
- 时段分布百分比
- 星期分布（按 DAYOFWEEK 统计）
- 平均时长（分钟）

### 4.4 匹配当前消息

从用户消息中提取关键词，匹配标签组：
- 命中 → 输出该组偏好摘要
- 未命中 → 输出全局摘要（Top 3 活跃时段 + 星期分布）
- 历史日程 <5 条 → 不检索，直接返回空

## 5. Prompt 注入

### 5.1 注入位置

在 Agent Server 系统 Prompt 中插入偏好区块：

```
# 角色
你是"个人AI小助手"的日程管理AI……

# 当前时间
{{current_time}}

# 已有日程（冲突感知）
{{existing_schedules}}

# 用户时间偏好（来自历史习惯分析）
{{preference_context}}              ← 新增

# 输出要求
……
```

### 5.2 注入规则

| 条件 | 注入内容 |
|------|---------|
| 匹配到标签组 + ≥5 条 | 完整偏好摘要（时段% + 星期% + 平均时长） |
| 匹配到标签组但 <5 条 | 仅输出条目数 + 粗略倾向 |
| 无匹配 / 全局查询 | 全局摘要（Top 3 活跃时段 + 星期分布） |
| 用户 <5 条历史日程 | **不注入**（`preference_context` 为空） |

### 5.3 注入示例

匹配到「工作」组（18 条）：

```
# 用户时间偏好（来自历史习惯分析）
用户过去90天内的18条工作类日程：
- 时段倾向：上午55%（8-12点）、下午35%（13-18点）、晚上10%
- 星期分布：周二三最密集（各25%）、周四一五递减
- 平均时长：1.5小时
建议：工作类日程安排在上午或下午，避开周末和晚上。
```

### 5.4 LLM 约束

系统 Prompt 中追加护栏：

> 以上偏好仅供参考，用户可能临时改变习惯。请以用户当前消息的显式意图为准。

### 5.5 Token 预算

| 场景 | 偏好摘要 token |
|------|---------------|
| 有匹配标签组 | ~80-150 |
| 全局摘要 | ~120-200 |
| 冷启动 | 0 |

## 6. 错误处理与降级

核心原则：**偏好系统是锦上添花，任何环节出错静默降级为空，绝不阻塞对话。**

### 6.1 降级链

| 环节 | 失败处理 |
|------|---------|
| 检索 SQL 异常 | 记录 warn 日志，`preference_context = ""`，继续对话 |
| 检索 SQL 超时（>200ms） | context canceled，同上 |
| 聚类计算 panic | recover 捕获，同上 |
| 空结果（0 条匹配） | 冷启动处理，同上 |
| 字段超长（>2KB） | 截断到前 2KB |
| gRPC 传输失败 | 走已有重试/超时机制，空值无影响 |

### 6.2 检索超时保护

```go
ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
defer cancel()
```

200ms 硬上限，超时即放弃本次偏好检索，对话继续。

### 6.3 监控埋点

| 指标 | 告警阈值 |
|------|---------|
| `pref_retrieval_latency_ms` | P95 > 200ms |
| `pref_retrieval_error_rate` | >5% |
| `pref_injection_rate` | 反映覆盖率（非空比例） |
| `pref_cold_start_rate` | 衡量新用户占比 |

## 7. Proto 变更

**文件**：`personal-assistant-proto/proto/agent/v1/agent.proto`

```protobuf
message ProcessMessageRequest {
  string user_id = 1;
  string conversation_id = 2;
  string message = 3;
  repeated HistoryEntry history = 4;
  repeated ScheduleSummary existing_schedules = 5;
  string preference_context = 6;  // 新增：偏好上下文文本
}
```

重新 `buf generate`，两个服务各自 `go mod tidy`。

## 8. 涉及文件清单

| 操作 | 文件 | 说明 |
|------|------|------|
| **修改** | `personal-assistant-proto/proto/agent/v1/agent.proto` | 加 `preference_context` 字段 |
| **新增** | `service/preference/preference.go` | 检索 + 聚类 + 匹配 + 格式化 |
| 修改 | `service/enter.go` | 加 `PreferenceService` |
| 修改 | `service/conversation/conversation.go` | 调 gRPC 前先调偏好服务 |
| 修改 | `personal-assistant-agent/internal/prompt/templates.go` | 模板加 `{{preference_context}}` 占位 |
| 修改 | `personal-assistant-agent/internal/handler/agent_handler.go` | 读取 `req.PreferenceContext` |

**无需变更**：MySQL 表结构、Redis、config.yaml、前端

## 9. 后续迭代

### 9.1 V1.5 内优化

- **Token 压缩**：当用户日程量 >200 时，切换为一行统计摘要（~60 token）替代完整偏好文本
- **标签分类增强**：配合 F-023 引入正式的分类标签体系，替代关键词匹配

### 9.2 V2.0 演进

- 外部知识注入（文档解析 + embedding + 向量检索）
- 复用 `preference_context` 字段，将知识库检索结果一并注入
- 引入专门的向量数据库或 PostgreSQL pgvector

## 10. 验收标准

1. 用户有 ≥5 条标签为「工作」的日程 → 对话中 LLM 回复体现对工作类日程的时间偏好理解
2. 用户新建日程时说「安排运动」→ 偏好上下文包含运动类历史习惯
3. 新用户（<5 条日程）→ 对话正常进行，无偏好注入，无报错
4. 检索 SQL 超时 → 对话正常完成，日志中有 warn 记录
5. 偏好注入不改变 gRPC 通信的 P95 延迟（<5s）

## 11. 不在此范围的

- 周期性规律自动发现（V2.0）
- 冲突偏好学习（用户决策模式）
- 外部文档上传与解析（V2.0）
- 向量检索 / embedding（V2.0）
- 前端偏好设置 UI
