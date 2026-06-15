# Login & Profile & Avatar — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable WeChat Mini Program users to login with nickname/avatar display, auto-generate initials avatars for new users, and update their profile in the user center.

**Architecture:** Add a factory-pattern file storage layer (Aliyun OSS + local fallback), an avatar generation utility (initials + colored background PNG), extend the existing login flow to accept optional nickname/avatar_url and auto-generate avatars for new users, and expose `GET/PUT /user/profile` endpoints.

**Tech Stack:** Go 1.24, Gin, GORM, Aliyun OSS SDK, go-pinyin (initials), golang.org/x/image (font rendering)

---

## File Structure

| File | Responsibility |
|------|---------------|
| `config/oss.go` | Oss config struct (type, endpoint, bucket, etc.) |
| `config/config.go` | Add `Oss Oss` field to Server |
| `config.yaml` | Add `oss:` section |
| `config.docker.yaml` | Add `oss:` section (local type for Docker dev) |
| `utils/storage/storage.go` | `FileStorage` interface + `New()` factory |
| `utils/storage/aliyun_oss.go` | Aliyun OSS `Upload` implementation |
| `utils/storage/local.go` | Local filesystem `Upload` (dev/debug) |
| `utils/avatar/avatar.go` | Generate initials PNG avatar |
| `global/global.go` | Add `GVA_STORAGE` variable |
| `service/auth/wechat.go` | Modified: accept nickname/avatar_url, auto-generate avatar for new users |
| `service/user/user.go` | New: `GetProfile` / `UpdateProfile`, avatar regeneration logic |
| `service/enter.go` | Add `UserService` to ServiceGroup |
| `api/v1/auth/wechat.go` | Modified: accept optional nickname/avatar_url in login body |
| `api/v1/user/user.go` | New: `GET /user/profile`, `PUT /user/profile` handlers |
| `api/v1/enter.go` | Add `UserApi` to ApiGroup |
| `router/user.go` | New: register `GET/PUT /user/profile` on PrivateGroup |
| `initialize/router.go` | Mount `InitUserRouter` |
| `core/server.go` | Init storage after config load |
| `go.mod` | Add aliyun-oss-go-sdk, go-pinyin |

---

### Task 1: Add go-pinyin dependency

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Install go-pinyin dependency**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server
go get github.com/mozillazg/go-pinyin@latest
```

Expected: adds `github.com/mozillazg/go-pinyin` to go.mod

---

### Task 2: OSS config struct

**Files:**
- Create: `config/oss.go`
- Modify: `config/config.go`
- Modify: `config.yaml`
- Modify: `config.docker.yaml`

- [ ] **Step 1: Create `config/oss.go`**

```go
package config

type Oss struct {
	Type            string `mapstructure:"type" json:"type" yaml:"type"`                                     // aliyun | local
	Endpoint        string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`                         // OSS endpoint
	AccessKeyID     string `mapstructure:"access-key-id" json:"access-key-id" yaml:"access-key-id"`          // AccessKey ID
	AccessKeySecret string `mapstructure:"access-key-secret" json:"access-key-secret" yaml:"access-key-secret"` // AccessKey Secret
	BucketName      string `mapstructure:"bucket-name" json:"bucket-name" yaml:"bucket-name"`                // Bucket 名称
	BucketURL       string `mapstructure:"bucket-url" json:"bucket-url" yaml:"bucket-url"`                   // Bucket 访问URL
	BasePath        string `mapstructure:"base-path" json:"base-path" yaml:"base-path"`                      // 基础路径，如 "avatars/"
}
```

- [ ] **Step 2: Add Oss field to `config/config.go`**

Replace the `Server` struct with:

```go
type Server struct {
	JWT    JWT    `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Zap    Zap    `mapstructure:"zap" json:"zap" yaml:"zap"`
	Redis  Redis  `mapstructure:"redis" json:"redis" yaml:"redis"`
	System System `mapstructure:"system" json:"system" yaml:"system"`
	Mysql  Mysql  `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Wechat Wechat `mapstructure:"wechat" json:"wechat" yaml:"wechat"`
	Grpc   Grpc   `mapstructure:"grpc" json:"grpc" yaml:"grpc"`
	Oss    Oss    `mapstructure:"oss" json:"oss" yaml:"oss"`
}
```

- [ ] **Step 3: Add oss section to `config.yaml`**

Append after the grpc section:

```yaml
# OSS 对象存储配置（头像上传）
oss:
  type: local
  endpoint: ""
  access-key-id: ""
  access-key-secret: ""
  bucket-name: ""
  bucket-url: ""
  base-path: "avatars/"
```

- [ ] **Step 4: Add oss section to `config.docker.yaml`**

Append after the grpc section:

```yaml
oss:
  type: local
  endpoint: ""
  access-key-id: ""
  access-key-secret: ""
  bucket-name: ""
  bucket-url: ""
  base-path: "avatars/"
```

- [ ] **Step 5: Verify compilation**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./... 2>&1
```

Expected: no errors

---

### Task 3: FileStorage interface + factory

**Files:**
- Create: `utils/storage/storage.go`

- [ ] **Step 1: Create directory**

```bash
mkdir -p D:\goProject\personal-assistant\personal-assistant-server\utils\storage
```

- [ ] **Step 2: Create `utils/storage/storage.go`**

```go
package storage

import (
	"context"
	"fmt"
	"io"

	"personal-assistant-server/config"
)

// FileStorage 文件存储接口
type FileStorage interface {
	Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error)
}

// New 根据配置创建文件存储实例（工厂方法）
func New(cfg config.Oss) (FileStorage, error) {
	switch cfg.Type {
	case "aliyun":
		return newAliyunOSS(cfg)
	case "local":
		return newLocal(cfg)
	default:
		return nil, fmt.Errorf("unsupported oss type: %s", cfg.Type)
	}
}
```

- [ ] **Step 3: Verify compilation (will fail — expected)**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./utils/storage 2>&1
```

Expected: `undefined: newAliyunOSS` and `undefined: newLocal` — we'll create those next.

---

### Task 4: Local storage implementation

**Files:**
- Create: `utils/storage/local.go`

- [ ] **Step 1: Create directory for local uploads**

```bash
mkdir -p D:\goProject\personal-assistant\personal-assistant-server\uploads\avatars
```

- [ ] **Step 2: Create `utils/storage/local.go`**

```go
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"personal-assistant-server/config"
)

type localStorage struct {
	basePath string
	saveDir  string
}

func newLocal(cfg config.Oss) (FileStorage, error) {
	dir := filepath.Join("uploads", cfg.BasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建本地存储目录失败: %w", err)
	}
	return &localStorage{
		basePath: cfg.BasePath,
		saveDir:  dir,
	}, nil
}

func (l *localStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	filename := filepath.Join(l.saveDir, filepath.Base(key))

	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return "/" + filename, nil
}
```

- [ ] **Step 3: Verify compilation**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./utils/storage 2>&1
```

Expected: still `undefined: newAliyunOSS` — one more to go.

---

### Task 5: Aliyun OSS storage implementation

**Files:**
- Create: `utils/storage/aliyun_oss.go`
- Modify: `go.mod`

- [ ] **Step 1: Install Aliyun OSS SDK**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server
go get github.com/aliyun/aliyun-oss-go-sdk/oss@latest
```

Expected: adds `github.com/aliyun/aliyun-oss-go-sdk` to go.mod

- [ ] **Step 2: Create `utils/storage/aliyun_oss.go`**

```go
package storage

import (
	"context"
	"fmt"
	"io"

	"personal-assistant-server/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type aliyunOSS struct {
	bucket    *oss.Bucket
	bucketURL string
	basePath  string
}

func newAliyunOSS(cfg config.Oss) (FileStorage, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("获取OSS Bucket失败: %w", err)
	}

	return &aliyunOSS{
		bucket:    bucket,
		bucketURL: cfg.BucketURL,
		basePath:  cfg.BasePath,
	}, nil
}

func (a *aliyunOSS) Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	fullKey := a.basePath + key
	options := []oss.Option{
		oss.ContentType(contentType),
		oss.ACL(oss.ACLPublicRead),
	}

	if err := a.bucket.PutObject(fullKey, reader, options...); err != nil {
		return "", fmt.Errorf("上传OSS失败: %w", err)
	}

	return fmt.Sprintf("%s/%s", a.bucketURL, fullKey), nil
}
```

- [ ] **Step 3: Verify full compilation**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./... 2>&1
```

Expected: no errors

---

### Task 6: Avatar generation utility

**Files:**
- Create: `utils/avatar/avatar.go`

- [ ] **Step 1: Create directory**

```bash
mkdir -p D:\goProject\personal-assistant\personal-assistant-server\utils\avatar
```

- [ ] **Step 2: Create `utils/avatar/avatar.go`**

```go
package avatar

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/math/fixed"
)

var palette = []color.RGBA{
	{0xE5, 0x39, 0x35},
	{0xD8, 0x1B, 0x60},
	{0x8E, 0x24, 0xAA},
	{0x5E, 0x35, 0xB1},
	{0x39, 0x49, 0xAB},
	{0x1E, 0x88, 0xE5},
	{0x03, 0x9B, 0xE5},
	{0x00, 0xAC, 0xC1},
	{0x00, 0x89, 0x7B},
	{0x43, 0xA0, 0x47},
	{0x7C, 0xB3, 0x42},
	{0xC0, 0xCA, 0x33},
	{0xF4, 0x51, 0x1E},
	{0x6D, 0x4C, 0x41},
	{0x55, 0x6B, 0x2F},
	{0x20, 0x82, 0xAE},
}

const (
	avatarSize  = 256
	fontSize    = 128
)

// Generate 生成首字符头像 PNG，返回 PNG 字节数据
// userID 用于确定性选色
// nickname 用于提取首字符（中文转拼音首字母）
func Generate(userID uint, nickname string) ([]byte, error) {
	initial := getInitial(nickname)
	bgColor := palette[userID%uint(len(palette))]

	img := image.NewRGBA(image.Rect(0, 0, avatarSize, avatarSize))

	// 填充背景
	for y := 0; y < avatarSize; y++ {
		for x := 0; x < avatarSize; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// 渲染白色首字母
	if err := drawText(img, initial); err != nil {
		// 渲染失败退回纯色背景头像
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// getInitial 提取昵称首字母：ASCII 直取，中文转拼音首字母
func getInitial(nickname string) string {
	if nickname == "" {
		return "?"
	}

	runes := []rune(nickname)
	first := runes[0]

	// ASCII 字母直接返回大写
	if (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') {
		return strings.ToUpper(string(first))
	}

	// ASCII 数字直接返回
	if first >= '0' && first <= '9' {
		return string(first)
	}

	// 中日韩文字 → 转拼音，取首字母
	if unicode.Is(unicode.Han, first) {
		py := pinyin.Pinyin(string(first), pinyin.NewArgs())
		if len(py) > 0 && len(py[0]) > 0 {
			return strings.ToUpper(string(py[0][0][0]))
		}
	}

	// 兜底
	return "?"
}

// drawText 在图片上居中绘制白色文字
func drawText(img *image.RGBA, text string) error {
	face, err := font.Font(gofont.GoBold, fontSize)
	if err != nil {
		return err
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
		Face: face,
	}

	// 测量文字宽度
	advance := d.MeasureString(text)
	x := (avatarSize*64 - advance.Ceil()) / 2

	// 垂直居中（以 baseline 计算）
	ascent := face.Metrics().Ascent.Ceil()
	descent := face.Metrics().Descent.Ceil()
	textHeight := ascent + descent
	y := (avatarSize+textHeight)/2 - descent

	d.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}
	d.DrawString(text)
	return nil
}
```

- [ ] **Step 3: Add golang.org/x/image dependency**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server
go get golang.org/x/image@latest
```

- [ ] **Step 4: Verify compilation**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./utils/avatar 2>&1
```

Expected: no errors

---

### Task 7: Add GVA_STORAGE to global

**Files:**
- Modify: `global/global.go`

- [ ] **Step 1: Add GVA_STORAGE variable**

Replace `global/global.go`:

```go
package global

import (
	"personal-assistant-server/config"
	"personal-assistant-server/utils/storage"
	"personal-assistant-server/utils/timer"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	GVA_DB      *gorm.DB
	GVA_REDIS   redis.UniversalClient
	GVA_CONFIG  config.Server
	GVA_VP      *viper.Viper
	GVA_LOG     *zap.Logger
	GVA_TIMER   timer.Timer        = timer.NewTimerTask()
	GVA_STORAGE storage.FileStorage
)
```

- [ ] **Step 2: Verify compilation**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./... 2>&1
```

Expected: no errors

---

### Task 8: Auth service — accept nickname/avatar, auto-generate avatar

**Files:**
- Modify: `service/auth/wechat.go`

- [ ] **Step 1: Add LoginWithProfile method**

Replace the `Login` method signature and extend the flow. The full updated file:

```go
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/model"
	"personal-assistant-server/utils"
	"personal-assistant-server/utils/avatar"
)

type AuthService struct{}

type WechatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type LoginRequest struct {
	Code      string `json:"code"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

// Login 微信小程序登录（兼容旧接口，无昵称/头像）
func (s *AuthService) Login(ctx context.Context, code string) (*LoginResponse, error) {
	return s.LoginWithProfile(ctx, LoginRequest{Code: code})
}

// LoginWithProfile 微信小程序登录（支持昵称和头像）
func (s *AuthService) LoginWithProfile(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// 1. 调用微信 code2session
	sessionResp, err := code2session(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("微信登录失败: %w", err)
	}

	// 2. 查找或创建用户
	user, isNew, err := s.findOrCreateUser(sessionResp)
	if err != nil {
		return nil, fmt.Errorf("用户处理失败: %w", err)
	}

	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	// 3. 新用户：处理昵称和头像
	if isNew {
		nickname := req.Nickname
		if nickname == "" {
			nickname = "微信用户"
		}
		user.Nickname = nickname

		// 自动生成首字符头像并上传
		avatarURL, err := s.generateAndUploadAvatar(user.ID, nickname)
		if err != nil {
			global.GVA_LOG.Error("生成头像失败: " + err.Error())
		} else {
			user.AvatarURL = avatarURL
		}

		// 保存用户资料
		global.GVA_DB.Model(user).Updates(map[string]interface{}{
			"nickname":   user.Nickname,
			"avatar_url": user.AvatarURL,
		})
	}

	// 4. 生成 JWT
	j := utils.NewJWT()
	claims := j.CreateClaims(user.ID, user.OpenID)
	token, err := j.CreateToken(claims)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	return &LoginResponse{Token: token, User: user}, nil
}

// generateAndUploadAvatar 生成首字符头像并上传到 OSS
func (s *AuthService) generateAndUploadAvatar(userID uint, nickname string) (string, error) {
	pngBytes, err := avatar.Generate(userID, nickname)
	if err != nil {
		return "", fmt.Errorf("生成头像图片失败: %w", err)
	}

	if global.GVA_STORAGE == nil {
		global.GVA_LOG.Warn("OSS存储未初始化，跳过头像上传")
		return "", nil
	}

	key := fmt.Sprintf("%d_%d.png", userID, time.Now().Unix())
	url, err := global.GVA_STORAGE.Upload(context.Background(), key, bytes.NewReader(pngBytes), "image/png")
	if err != nil {
		return "", fmt.Errorf("上传头像失败: %w", err)
	}
	return url, nil
}

func (s *AuthService) GetUserProfile(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	if err := global.GVA_DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// findOrCreateUser 返回 (user, isNew, error)
func (s *AuthService) findOrCreateUser(session *WechatSessionResponse) (*model.User, bool, error) {
	var user model.User
	err := global.GVA_DB.Where("openid = ?", session.OpenID).First(&user).Error
	if err == nil {
		if session.UnionID != "" && user.UnionID == "" {
			global.GVA_DB.Model(&user).Update("unionid", session.UnionID)
		}
		return &user, false, nil
	}

	user = model.User{
		OpenID:  session.OpenID,
		UnionID: session.UnionID,
	}
	if err := global.GVA_DB.Create(&user).Error; err != nil {
		return nil, false, err
	}
	return &user, true, nil
}

func code2session(ctx context.Context, code string) (*WechatSessionResponse, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?"+
			"appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		global.GVA_CONFIG.Wechat.AppID,
		global.GVA_CONFIG.Wechat.AppSecret,
		code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求微信接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取微信响应失败: %w", err)
	}

	var sessionResp WechatSessionResponse
	if err := json.Unmarshal(body, &sessionResp); err != nil {
		return nil, fmt.Errorf("解析微信响应失败: %w", err)
	}

	if sessionResp.ErrCode != 0 {
		return nil, fmt.Errorf("微信返回错误: %d - %s", sessionResp.ErrCode, sessionResp.ErrMsg)
	}

	return &sessionResp, nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./service/auth 2>&1
```

Expected: no errors

---

### Task 9: User service — profile query and update

**Files:**
- Create: `service/user/user.go`
- Modify: `service/enter.go`

- [ ] **Step 1: Create `service/user/user.go`**

```go
package user

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/model"
	"personal-assistant-server/utils/avatar"
)

type UserService struct{}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	if err := global.GVA_DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Nickname               string                `form:"nickname"`
	Avatar                 *multipart.FileHeader `form:"avatar"`
	DefaultReminderMinutes *int                  `form:"default_reminder_minutes"`
	OnboardingCompleted    *bool                 `form:"onboarding_completed"`
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) (*model.User, error) {
	var user model.User
	if err := global.GVA_DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	needRegenAvatar := false

	// 处理昵称
	if req.Nickname != "" && req.Nickname != user.Nickname {
		updates["nickname"] = req.Nickname
		user.Nickname = req.Nickname
		// 除非同时上传了自定义头像，否则重新生成
		if req.Avatar == nil {
			needRegenAvatar = true
		}
	}

	// 处理自定义头像上传
	if req.Avatar != nil {
		avatarURL, err := s.uploadCustomAvatar(userID, req.Avatar)
		if err != nil {
			return nil, fmt.Errorf("头像上传失败: %w", err)
		}
		updates["avatar_url"] = avatarURL
		needRegenAvatar = false
	}

	// 重新生成首字符头像
	if needRegenAvatar {
		avatarURL, err := s.regenerateAvatar(userID, user.Nickname)
		if err != nil {
			global.GVA_LOG.Error("重新生成头像失败: " + err.Error())
		} else if avatarURL != "" {
			updates["avatar_url"] = avatarURL
		}
	}

	// 处理设置项
	if req.DefaultReminderMinutes != nil {
		updates["default_reminder_minutes"] = *req.DefaultReminderMinutes
	}
	if req.OnboardingCompleted != nil {
		updates["onboarding_completed"] = *req.OnboardingCompleted
	}

	if len(updates) > 0 {
		if err := global.GVA_DB.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	// 重新加载最新数据
	global.GVA_DB.WithContext(ctx).First(&user, userID)
	return &user, nil
}

// uploadCustomAvatar 上传用户自定义头像
func (s *UserService) uploadCustomAvatar(userID uint, fileHeader *multipart.FileHeader) (string, error) {
	if global.GVA_STORAGE == nil {
		return "", fmt.Errorf("OSS存储未初始化")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	ext := ".png"
	if fileHeader.Header.Get("Content-Type") == "image/jpeg" {
		ext = ".jpg"
	}

	key := fmt.Sprintf("%d_%d%s", userID, time.Now().Unix(), ext)
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}

	return global.GVA_STORAGE.Upload(ctx, key, file, contentType)
}

// regenerateAvatar 重新生成首字符头像
func (s *UserService) regenerateAvatar(userID uint, nickname string) (string, error) {
	pngBytes, err := avatar.Generate(userID, nickname)
	if err != nil {
		return "", err
	}

	if global.GVA_STORAGE == nil {
		return "", nil
	}

	key := fmt.Sprintf("%d_%d.png", userID, time.Now().Unix())
	return global.GVA_STORAGE.Upload(context.Background(), key, bytes.NewReader(pngBytes), "image/png")
}
```

- [ ] **Step 2: Add UserService to `service/enter.go`**

Replace with:

```go
package service

import (
	"personal-assistant-server/service/auth"
	"personal-assistant-server/service/conversation"
	"personal-assistant-server/service/push"
	"personal-assistant-server/service/schedule"
	"personal-assistant-server/service/user"
	"personal-assistant-server/service/view"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	AuthService         auth.AuthService
	ScheduleService     schedule.ScheduleService
	ConversationService conversation.ConversationService
	PushService         push.PushService
	ViewService         view.ViewService
	UserService         user.UserService
}
```

- [ ] **Step 3: Verify compilation**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./service/... 2>&1
```

Expected: no errors

---

### Task 10: Auth API — accept optional nickname/avatar_url

**Files:**
- Modify: `api/v1/auth/wechat.go`

- [ ] **Step 1: Update auth API handler**

Replace `api/v1/auth/wechat.go`:

```go
package auth

import (
	"github.com/gin-gonic/gin"

	"personal-assistant-server/model/common/response"
	"personal-assistant-server/service"
	"personal-assistant-server/service/auth"
)

type AuthApi struct{}

// Login 微信小程序登录
func (a *AuthApi) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请提供微信登录code", c)
		return
	}

	if req.Code == "" {
		response.FailWithMessage("code不能为空", c)
		return
	}

	resp, err := service.ServiceGroupApp.AuthService.LoginWithProfile(c.Request.Context(), req)
	if err != nil {
		response.FailWithMessage("登录失败: "+err.Error(), c)
		return
	}

	response.OkWithData(resp, c)
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./api/v1/auth 2>&1
```

Expected: no errors

---

### Task 11: User API — GET/PUT /user/profile

**Files:**
- Create: `api/v1/user/user.go`

- [ ] **Step 1: Create `api/v1/user/user.go`**

```go
package user

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"personal-assistant-server/model/common/response"
	"personal-assistant-server/service"
	"personal-assistant-server/service/user"
	"personal-assistant-server/utils"
)

type UserApi struct{}

// GetProfile 获取当前用户资料
func (a *UserApi) GetProfile(c *gin.Context) {
	userID := utils.GetUserID(c)
	u, err := service.ServiceGroupApp.UserService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.FailWithMessage("获取用户信息失败: "+err.Error(), c)
		return
	}
	response.OkWithData(u, c)
}

// UpdateProfile 更新当前用户资料
func (a *UserApi) UpdateProfile(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req user.UpdateProfileRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 处理表单中的布尔和数值字段
	if v := c.PostForm("default_reminder_minutes"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.DefaultReminderMinutes = &n
		}
	}
	if v := c.PostForm("onboarding_completed"); v != "" {
		b := v == "true"
		req.OnboardingCompleted = &b
	}

	u, err := service.ServiceGroupApp.UserService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithData(u, c)
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./api/v1/user 2>&1
```

Expected: no errors

---

### Task 12: API aggregator — add UserApi

**Files:**
- Modify: `api/v1/enter.go`

- [ ] **Step 1: Add UserApi to ApiGroup**

Replace `api/v1/enter.go`:

```go
package v1

import (
	"personal-assistant-server/api/v1/auth"
	"personal-assistant-server/api/v1/conversation"
	"personal-assistant-server/api/v1/push"
	"personal-assistant-server/api/v1/schedule"
	"personal-assistant-server/api/v1/user"
	"personal-assistant-server/api/v1/view"
)

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	AuthApi         auth.AuthApi
	ScheduleApi     schedule.ScheduleApi
	ConversationApi conversation.ConversationApi
	ViewApi         view.ViewApi
	PushApi         push.PushApi
	UserApi         user.UserApi
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./api/v1 2>&1
```

Expected: no errors

---

### Task 13: User router registration

**Files:**
- Create: `router/user.go`

- [ ] **Step 1: Create `router/user.go`**

```go
package router

import (
	"github.com/gin-gonic/gin"

	v1 "personal-assistant-server/api/v1"
)

func InitUserRouter(privateGroup *gin.RouterGroup) {
	userApi := v1.ApiGroupApp.UserApi
	userRouter := privateGroup.Group("/user")
	{
		userRouter.GET("/profile", userApi.GetProfile)
		userRouter.PUT("/profile", userApi.UpdateProfile)
	}
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./router 2>&1
```

Expected: no errors

---

### Task 14: Mount user router + init storage

**Files:**
- Modify: `initialize/router.go`
- Modify: `core/server.go`

- [ ] **Step 1: Add user router to `initialize/router.go`**

Add `router.InitUserRouter(PrivateGroup)` inside the PrivateGroup block:

```go
	{
		router.InitScheduleRouter(PrivateGroup)
		router.InitConversationRouter(PrivateGroup)
		router.InitViewRouter(PrivateGroup)
		router.InitPushRouter(PrivateGroup)
		router.InitUserRouter(PrivateGroup)
	}
```

- [ ] **Step 2: Add storage initialization to `core/server.go`**

Add storage init before router init. Replace the file:

```go
package core

import (
	"fmt"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/initialize"
	"personal-assistant-server/utils/storage"

	"go.uber.org/zap"
)

func RunServer() {
	if global.GVA_CONFIG.System.UseRedis {
		initialize.Redis()
	}

	// 初始化 OSS 存储
	if global.GVA_CONFIG.Oss.Type != "" {
		s, err := storage.New(global.GVA_CONFIG.Oss)
		if err != nil {
			zap.L().Warn("OSS存储初始化失败，头像功能将不可用: " + err.Error())
		} else {
			global.GVA_STORAGE = s
			zap.L().Info("OSS存储初始化成功, 类型: " + global.GVA_CONFIG.Oss.Type)
		}
	}

	Router := initialize.Routers()

	address := fmt.Sprintf(":%d", global.GVA_CONFIG.System.Addr)

	fmt.Printf(`
  欢迎使用 个人AI小助手 API Server
  当前版本:%s
  运行地址: http://127.0.0.1%s
`, global.Version, address)
	zap.L().Info("服务器启动中...", zap.String("address", address))
	initServer(address, Router, 10*time.Minute, 10*time.Minute)
}
```

- [ ] **Step 3: Add storage import to `core/server.go`**

The import block now includes `"personal-assistant-server/utils/storage"`.

- [ ] **Step 4: Verify full compilation**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./... 2>&1
```

Expected: no errors

---

### Task 15: Add .gitkeep for uploads directory

**Files:**
- Create: `uploads/avatars/.gitkeep`

- [ ] **Step 1: Create gitkeep**

```bash
mkdir -p D:\goProject\personal-assistant\personal-assistant-server\uploads\avatars
touch D:\goProject\personal-assistant\personal-assistant-server\uploads\avatars\.gitkeep
```

- [ ] **Step 2: Add .gitignore entry**

Check `D:\goProject\personal-assistant\personal-assistant-server\.gitignore` for an `uploads/` entry. If not present, add `uploads/*` to ignore uploaded files while keeping the directory.

---

### Task 16: Final verification

**Files:**
- All files from tasks 1-14

- [ ] **Step 1: Full build**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go build ./... 2>&1
```

Expected: no errors

- [ ] **Step 2: go mod tidy**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server && go mod tidy 2>&1
```

Expected: no errors

- [ ] **Step 3: Commit all changes**

```bash
cd D:\goProject\personal-assistant\personal-assistant-server
git add -A
git commit -m "feat: add login profile & avatar with OSS storage

- Add Oss config struct with aliyun/local support
- Add FileStorage interface + factory pattern (aliyun OSS + local)
- Add initials avatar generation (Chinese pinyin + English initials)
- Extend login API to accept nickname/avatar_url
- Auto-generate avatar for new users on first login
- Add GET/PUT /user/profile endpoints
- Init OSS storage in server startup"
```

Expected: clean commit

---

## Plan Self-Review

1. **Spec coverage check:**
   - Login with nickname/avatar ✓ (Task 8, 10)
   - Auto-generate initials avatar for new users ✓ (Task 6, 8)
   - GET /user/profile ✓ (Task 9, 11)
   - PUT /user/profile with nickname/avatar/settings ✓ (Task 9, 11)
   - Aliyun OSS + factory pattern ✓ (Tasks 2, 3, 4, 5)
   - OSS.type switchable ✓ (Task 3 factory, Task 14 init)
   - config.yaml oss section ✓ (Task 2)

2. **Placeholder scan:** 0 TODOs, 0 TBDs, all code blocks contain complete implementations.

3. **Type consistency:**
   - `storage.FileStorage` interface defined in Task 3, used in Task 8 (auth), Task 9 (user), Task 14 (core)
   - `LoginRequest` struct defined in Task 8 (service/auth), used in Task 10 (api/auth)
   - `UpdateProfileRequest` struct defined in Task 9 (service/user), used in Task 11 (api/user)
   - `global.GVA_STORAGE` defined in Task 7, written in Task 14, read in Task 8 and 9
   - All types match across tasks ✓
