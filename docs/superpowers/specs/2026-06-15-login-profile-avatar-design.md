# 微信小程序登录 & 用户头像昵称 — 技术规格

| 文档版本 | 日期 | 作者 | 说明 |
|---------|------|------|------|
| V1.0 | 2026-06-15 | 后端开发 | 基于 PRD V1.0 + 微信官方登录规范 |

## 1. 背景

微信于 2022年10月 收回 `wx.getUserProfile` 接口，2025-2026年该接口已彻底不可用（仅返回灰色默认头像 + "微信用户"）。当前官方推荐方案为「头像昵称填写能力」：
- 头像：`<button open-type="chooseAvatar">` → 临时路径 → 前端上传到后端
- 昵称：`<input type="nickname">` → 键盘上方弹出微信昵称供用户点选

同时，第三方插件（如「安全注册」`wxc7b7f914565de923`）可一键获取真实微信头像和昵称，作为可选前端方案。

**本规格定义后端需支持的完整登录-资料-头像链路。**

## 2. 设计目标

- 用户登录后立即展示头像和微信名（零等待）
- 新用户自动生成首字符头像（无冷启动空白头像问题）
- 用户可在个人中心更换头像和昵称
- 阿里云 OSS 存储头像，工厂模式预留 COS/MinIO 切换能力

## 3. 完整登录流程

```
┌──────────┐          ┌──────────────┐          ┌──────────┐
│ 小程序    │          │  API Server  │          │  微信API  │
└────┬─────┘          └──────┬───────┘          └────┬─────┘
     │                       │                       │
     │ ① wx.login() → code   │                       │
     │ ② 获取昵称头像         │                       │
     │   (官方填写/插件)      │                       │
     │                       │                       │
     │ ③ POST /auth/wechat/login                     │
     │ {code, nickname?, avatar_url?}                 │
     │──────────────────────▶│                       │
     │                       │ code2session          │
     │                       │──────────────────────▶│
     │                       │◀── openid ────────────│
     │                       │                       │
     │                       │ 查找/创建用户          │
     │                       │  ┌──────────────────┐ │
     │                       │  │ 新用户:           │ │
     │                       │  │ · 有昵称→首字符头像│ │
     │                       │  │ · 无昵称→"微信用户"│ │
     │                       │  │ · 自动上传OSS     │ │
     │                       │  └──────────────────┘ │
     │                       │                       │
     │                       │ 签发 JWT              │
     │                       │                       │
     │◀── {token, user} ────│                       │
     │                       │                       │
     │ ④ [可选] PUT /user/profile                    │
     │ multipart/form-data                           │
     │ {nickname?, avatar_file?}                     │
     │──────────────────────▶│                       │
     │                       │ 更新/重新生成头像     │
     │◀── {user} ───────────│                       │
```

**关键行为：**
- 登录接口 `nickname` 和 `avatar_url` 均为可选字段
- 新用户**始终自动生成首字符头像**（无论是否提供了昵称），8KB 量级，秒级完成
- 新用户无昵称时默认 `"微信用户"`
- 已有用户（openid 已注册）：登录不覆盖头像和昵称，返回已有资料
- 用户后续可通过 `PUT /user/profile` 修改昵称或上传自定义头像

## 4. API 端点

### 4.1 登录（修改已有接口）

**POST** `/api/v1/auth/wechat/login`

```json
// Request
{
  "code": "0b3...",           // 必填，wx.login() 返回
  "nickname": "李志明",        // 可选
  "avatar_url": "https://..." // 可选，前端插件返回的永久URL
}

// Response (200)
{
  "code": 0,
  "msg": "success",
  "data": {
    "token": "eyJhbG...",
    "user": {
      "id": 1,
      "openid": "oABC...",
      "nickname": "李志明",
      "avatar_url": "https://oss-cn-hangzhou.aliyuncs.com/avatars/1_1718400000.png"
    }
  }
}
```

### 4.2 获取用户资料（新增）

**GET** `/api/v1/user/profile`

Header: `x-token: <jwt>`

```json
// Response (200)
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "nickname": "李志明",
    "avatar_url": "https://oss-cn-hangzhou...",
    "phone": null,
    "default_reminder_minutes": 30,
    "week_start_day": 1,
    "onboarding_completed": false,
    "created_at": "2026-06-15T10:00:00Z"
  }
}
```

### 4.3 更新用户资料（新增）

**PUT** `/api/v1/user/profile`

Header: `x-token: <jwt>`
Content-Type: `multipart/form-data`

| 字段 | 类型 | 说明 |
|------|------|------|
| nickname | text | 可选，新昵称。修改后自动重新生成首字符头像（除非同时上传了自定义头像） |
| avatar | file | 可选，自定义头像图片。支持 PNG/JPG，最大 2MB |
| default_reminder_minutes | text | 可选，默认提醒分钟数 |
| onboarding_completed | text | 可选，"true"/"false" |

```json
// Response (200)
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "nickname": "李志明",
    "avatar_url": "https://oss-cn-hangzhou...",
    "default_reminder_minutes": 15,
    "onboarding_completed": true
  }
}
```

**头像更新规则：**
- 上传了 `avatar` 文件 → 使用上传的图片，上传到 OSS，更新 `avatar_url`
- 仅修改 `nickname`、未上传 `avatar` → 重新生成首字符头像（因为首字符可能变了）
- 两者都未提供 → 不做任何头像变更

## 5. 头像自动生成

### 5.1 生成规则

```
输入: nickname = "李志明" (或 "微信用户")
输出: 256×256 PNG, ~6KB
```

**算法：**
1. 提取首字符：取 nickname 的第一个 rune（中文单字或英文首字母）
2. 选背景色：`palette[user.ID % len(palette)]`，16 色预设调色板
3. 绘制：
   - 填充背景色
   - 白色文字居中渲染（使用系统自带中文字体）
   - 字号 = 128pt，水平垂直居中
4. 编码为 PNG（`image/png` encoder）

**调色板（16 色）：**
```go
var palette = []color.RGBA{
    {0xE5, 0x39, 0x35}, // Red
    {0xD8, 0x1B, 0x60}, // Pink
    {0x8E, 0x24, 0xAA}, // Purple
    {0x5E, 0x35, 0xB1}, // Deep Purple
    {0x39, 0x49, 0xAB}, // Indigo
    {0x1E, 0x88, 0xE5}, // Blue
    {0x03, 0x9B, 0xE5}, // Light Blue
    {0x00, 0xAC, 0xC1}, // Cyan
    {0x00, 0x89, 0x7B}, // Teal
    {0x43, 0xA0, 0x47}, // Green
    {0x7C, 0xB3, 0x42}, // Light Green
    {0xC0, 0xCA, 0x33}, // Lime
    {0xFB, 0xB0, 0x0F}, // Yellow (avoid: poor contrast)
    {0xFB, 0x8C, 0x00}, // Amber
    {0xF4, 0x51, 0x1E}, // Orange
    {0x6D, 0x4C, 0x41}, // Brown
}
```

### 5.2 中文字体

Go 的 `golang.org/x/image/font` 标准库不内置中文字体。方案：

- **MVP 策略**：首次登录时，取首字符的 Unicode 码点，用 [`golang.org/x/image/font/gofont/gobold`](https://pkg.go.dev/golang.org/x/image/font/gofont/gobold) 渲染英文字符；中文退回英文字母缩写方案：

  中文昵称 → 转为拼音首字母（最多 2 个字母）渲染
  "李志明" → "L"
  "微信用户" → "W"

- **V1.1 增强**（后续）：嵌入一款开源中文字体（如 Noto Sans SC Regular，~5MB），用 `opentype` 库渲染真实中文

  当前方案是务实选择——避免 5MB 字体文件嵌入二进制导致编译产物过大，且首字母头像已满足辨识需求。

## 6. 存储层 — 工厂模式

### 6.1 目录结构

```
utils/storage/
├── storage.go       # FileStorage 接口 + New() 工厂
├── aliyun_oss.go    # 阿里云 OSS 实现
└── local.go         # 本地存储（开发/调试用）
```

### 6.2 接口定义

```go
type FileStorage interface {
    Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error)
}
```

### 6.3 配置

```yaml
# config.yaml
oss:
  type: aliyun          # aliyun | local（后续扩展: tencent | minio）
  endpoint: "oss-cn-hangzhou.aliyuncs.com"
  access-key-id: ""
  access-key-secret: ""
  bucket-name: "pa-assistant"
  bucket-url: "https://pa-assistant.oss-cn-hangzhou.aliyuncs.com"
  base-path: "avatars/"
```

### 6.4 全局单例

```go
// global.go
GVA_STORAGE storage.FileStorage

// 初始化（main.go / core/server.go）
global.GVA_STORAGE, err = storage.New(global.GVA_CONFIG.Oss)
```

### 6.5 OSS 文件路径

```
{base-path}{user_id}_{timestamp}.{ext}
→ "avatars/123_1718457600.png"
```

上传时使用 `PutObject` + 公共读 ACL，返回拼接后的完整 URL。

## 7. 涉及文件清单

| 操作 | 文件 |
|------|------|
| **新增** | `config/oss.go` |
| **修改** | `config/config.go` — 加 `Oss Oss` 字段 |
| **新增** | `utils/storage/storage.go` — 接口 + 工厂 |
| **新增** | `utils/storage/aliyun_oss.go` |
| **新增** | `utils/storage/local.go` |
| **新增** | `utils/avatar/avatar.go` — 首字符头像生成 |
| **修改** | `global/global.go` — 加 `GVA_STORAGE` |
| **修改** | `service/auth/wechat.go` — 新用户自动生成头像 |
| **新增** | `service/user/user.go` — 用户资料查询/更新 |
| **新增** | `api/v1/user/user.go` — GET/PUT /user/profile |
| **修改** | `api/v1/auth/wechat.go` — 接受 nickname/avatar_url |
| **修改** | `api/v1/enter.go` — 加 `UserApi` |
| **新增** | `router/user.go` — 注册用户路由 |
| **修改** | `initialize/router.go` — 挂载 userRouter |
| **修改** | `core/server.go` — 初始化 OSS 存储 |
| **修改** | `config.yaml` / `config.docker.yaml` — 加 oss 配置段 |
| **修改** | `go.mod` — 加 `github.com/aliyun/aliyun-oss-go-sdk` |

## 8. 验收标准

1. **新用户首次登录** → 自动生成首字符头像，返回完整 `{token, user{nickname, avatar_url}}`
2. **老用户再次登录** → 不覆盖已有头像和昵称，直接返回
3. **GET /user/profile** → 返回完整用户资料（含设置项）
4. **PUT /user/profile（改昵称）** → 头像自动重新生成（首字符变了）
5. **PUT /user/profile（上传自定义头像）** → 使用上传图片，OSS 返回永久 URL
6. **OSS 类型切换** → 修改配置 `oss.type: local` 即可切到本地存储，无需改代码

## 9. 不在此范围的

- 腾讯 COS / MinIO 实现 — 接口已预留，后续按需添加
- 头像删除/历史版本管理 — 无需求
- 前端头像裁剪组件 — 前端负责，后端只接受并存储
- 头像 CDN 加速 — 阿里云 OSS 自带，不做额外配置
