package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/model"
	"personal-assistant-server/service/email"
	"personal-assistant-server/service/sms"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterService handles email/phone registration and credential login
type RegisterService struct {
	emailService email.EmailService
	smsService   sms.SMSService
}

func NewRegisterService() *RegisterService {
	return &RegisterService{
		emailService: email.NewEmailService(),
		smsService:   sms.NewSMSService(),
	}
}

// ==================== Request/Response Types ====================

type SendEmailCodeRequest struct {
	Email string `json:"email"`
}

type SendSMSCodeRequest struct {
	Phone string `json:"phone"`
}

type RegisterByEmailRequest struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

type RegisterByPhoneRequest struct {
	Phone    string `json:"phone"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

type CredentialLoginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

// ==================== Send Verification Codes ====================

func (s *RegisterService) SendEmailCode(ctx context.Context, req *SendEmailCodeRequest) error {
	if !isEmail(req.Email) {
		return errors.New("邮箱格式不正确")
	}

	var count int64
	global.GVA_DB.Model(&model.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return errors.New("该邮箱已注册")
	}

	if err := s.checkResendCooldown(ctx, req.Email, "email"); err != nil {
		return err
	}

	code, err := s.generateAndStoreEmailCode(ctx, req.Email)
	if err != nil {
		return err
	}

	return s.emailService.SendVerificationCode(ctx, req.Email, code)
}

func (s *RegisterService) SendSMSCode(ctx context.Context, req *SendSMSCodeRequest) error {
	if !isPhone(req.Phone) {
		return errors.New("手机号格式不正确")
	}

	var count int64
	global.GVA_DB.Model(&model.User{}).Where("phone = ?", req.Phone).Count(&count)
	if count > 0 {
		return errors.New("该手机号已注册")
	}

	if err := s.checkResendCooldown(ctx, req.Phone, "phone"); err != nil {
		return err
	}

	code, err := s.generateAndStoreSMSCode(ctx, req.Phone)
	if err != nil {
		return err
	}

	return s.smsService.SendVerificationCode(ctx, req.Phone, code)
}

// ==================== Registration ====================

func (s *RegisterService) RegisterByEmail(ctx context.Context, req *RegisterByEmailRequest) (*LoginResponse, error) {
	if !isEmail(req.Email) {
		return nil, errors.New("邮箱格式不正确")
	}
	if err := ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	if err := s.verifyEmailCode(ctx, req.Email, req.Code); err != nil {
		return nil, err
	}

	var count int64
	global.GVA_DB.Model(&model.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return nil, errors.New("该邮箱已注册")
	}

	return s.createUser(ctx, req.Email, "", req.Password, "email")
}

func (s *RegisterService) RegisterByPhone(ctx context.Context, req *RegisterByPhoneRequest) (*LoginResponse, error) {
	if !isPhone(req.Phone) {
		return nil, errors.New("手机号格式不正确")
	}
	if err := ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	if err := s.verifySMSCode(ctx, req.Phone, req.Code); err != nil {
		return nil, err
	}

	var count int64
	global.GVA_DB.Model(&model.User{}).Where("phone = ?", req.Phone).Count(&count)
	if count > 0 {
		return nil, errors.New("该手机号已注册")
	}

	return s.createUser(ctx, "", req.Phone, req.Password, "phone")
}

// ==================== Credential Login ====================

func (s *RegisterService) LoginByCredential(ctx context.Context, req *CredentialLoginRequest) (*LoginResponse, error) {
	if req.Account == "" || req.Password == "" {
		return nil, errors.New("邮箱/手机号或密码错误")
	}

	var user model.User
	var err error

	if isEmail(req.Account) {
		err = global.GVA_DB.Where("email = ? AND auth_method != ?", req.Account, "wechat").First(&user).Error
	} else if isPhone(req.Account) {
		err = global.GVA_DB.Where("phone = ? AND auth_method != ?", req.Account, "wechat").First(&user).Error
	} else {
		return nil, errors.New("邮箱/手机号或密码错误")
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("邮箱/手机号或密码错误")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("邮箱/手机号或密码错误")
	}

	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	authSvc := &AuthService{}
	deviceID := fmt.Sprintf("web_cred_%s", uuid.New().String()[:16])
	return authSvc.buildLoginResponse(&user, deviceID, "Web 网页端（账号密码）")
}

// ==================== Internal Methods ====================

func (s *RegisterService) createUser(ctx context.Context, emailAddr, phone, password, authMethod string) (*LoginResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	username := GenerateUniqueUsername()

	user := model.User{
		Username:     username,
		Nickname:     username,
		PasswordHash: string(hash),
		AuthMethod:   authMethod,
		Status:       1,
	}

	if emailAddr != "" {
		user.Email = &emailAddr
	}
	if phone != "" {
		user.Phone = phone
	}

	if err := global.GVA_DB.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	authSvc := &AuthService{}
	avatarURL, err := authSvc.generateAndUploadAvatar(user.ID, username)
	if err != nil {
		global.GVA_LOG.Error("生成头像失败: " + err.Error())
	} else if avatarURL != "" {
		global.GVA_DB.Model(&user).Update("avatar_url", avatarURL)
		user.AvatarURL = avatarURL
	}

	deviceID := fmt.Sprintf("web_%s_%s", authMethod, uuid.New().String()[:16])
	return authSvc.buildLoginResponse(&user, deviceID, "Web 网页端")
}

func (s *RegisterService) checkResendCooldown(ctx context.Context, target, targetType string) error {
	var lastCreatedAt time.Time
	switch targetType {
	case "email":
		var ev model.EmailVerification
		if err := global.GVA_DB.Where("email = ?", target).Order("created_at DESC").First(&ev).Error; err == nil {
			lastCreatedAt = ev.CreatedAt
		}
	case "phone":
		var sv model.SmsVerification
		if err := global.GVA_DB.Where("phone = ?", target).Order("created_at DESC").First(&sv).Error; err == nil {
			lastCreatedAt = sv.CreatedAt
		}
	}
	if !lastCreatedAt.IsZero() && time.Since(lastCreatedAt) < 60*time.Second {
		return errors.New("发送太频繁，请60秒后再试")
	}
	return nil
}

func (s *RegisterService) generateAndStoreEmailCode(ctx context.Context, emailAddr string) (string, error) {
	code := generateVerificationCode()
	ev := model.EmailVerification{
		Email:     emailAddr,
		Code:      code,
		Purpose:   "register",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if err := global.GVA_DB.Create(&ev).Error; err != nil {
		return "", fmt.Errorf("存储验证码失败: %w", err)
	}
	return code, nil
}

func (s *RegisterService) generateAndStoreSMSCode(ctx context.Context, phoneNum string) (string, error) {
	code := generateVerificationCode()
	sv := model.SmsVerification{
		Phone:     phoneNum,
		Code:      code,
		Purpose:   "register",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if err := global.GVA_DB.Create(&sv).Error; err != nil {
		return "", fmt.Errorf("存储验证码失败: %w", err)
	}
	return code, nil
}

func (s *RegisterService) verifyEmailCode(ctx context.Context, emailAddr, code string) error {
	var ev model.EmailVerification
	err := global.GVA_DB.Where("email = ? AND code = ?", emailAddr, code).
		Order("created_at DESC").First(&ev).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("验证码错误")
		}
		return fmt.Errorf("查询验证码失败: %w", err)
	}

	if ev.Verified {
		return errors.New("验证码已使用，请重新获取")
	}
	if time.Now().After(ev.ExpiresAt) {
		return errors.New("验证码已过期，请重新获取")
	}

	var attemptCount int64
	global.GVA_DB.Model(&model.EmailVerification{}).
		Where("email = ? AND created_at > ?", emailAddr, ev.CreatedAt.Add(-5*time.Minute)).
		Count(&attemptCount)
	if attemptCount > 3 {
		return errors.New("验证次数过多，请重新获取验证码")
	}

	global.GVA_DB.Model(&ev).Update("verified", true)
	return nil
}

func (s *RegisterService) verifySMSCode(ctx context.Context, phoneNum, code string) error {
	var sv model.SmsVerification
	err := global.GVA_DB.Where("phone = ? AND code = ?", phoneNum, code).
		Order("created_at DESC").First(&sv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("验证码错误")
		}
		return fmt.Errorf("查询验证码失败: %w", err)
	}

	if sv.Verified {
		return errors.New("验证码已使用，请重新获取")
	}
	if time.Now().After(sv.ExpiresAt) {
		return errors.New("验证码已过期，请重新获取")
	}

	var attemptCount int64
	global.GVA_DB.Model(&model.SmsVerification{}).
		Where("phone = ? AND created_at > ?", phoneNum, sv.CreatedAt.Add(-5*time.Minute)).
		Count(&attemptCount)
	if attemptCount > 3 {
		return errors.New("验证次数过多，请重新获取验证码")
	}

	global.GVA_DB.Model(&sv).Update("verified", true)
	return nil
}

// ==================== Utility Functions ====================

func isEmail(s string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(s)
}

func isPhone(s string) bool {
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{6,14}$`)
	return phoneRegex.MatchString(s)
}

func generateVerificationCode() string {
	code := make([]byte, 6)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			code[i] = byte('0' + i%10)
			continue
		}
		code[i] = byte('0' + n.Int64())
	}
	return string(code)
}
