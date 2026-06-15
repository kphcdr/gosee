package auth

import (
	"errors"

	"gorm.io/gorm"

	"gosee/internal/config"
	"gosee/internal/repository"
	"gosee/internal/utils"
)

// Service 认证服务
type Service struct {
	userRepo *repository.UserRepository
	jwt      *config.JWTConfig
}

func NewService(userRepo *repository.UserRepository, jwtCfg *config.JWTConfig) *Service {
	return &Service{userRepo: userRepo, jwt: jwtCfg}
}

// UserInfo 不含敏感字段的用户信息
type UserInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}

// LoginResult 登录返回
type LoginResult struct {
	Token    string   `json:"token"`
	ExpireIn int      `json:"expire_in"` // 秒
	User     UserInfo `json:"user"`
}

// Login 校验账密并签发 JWT
func (s *Service) Login(username, password string) (*LoginResult, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}
	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}
	token, err := utils.GenerateToken(user.ID, user.Username, s.jwt.Secret, s.jwt.Issuer, s.jwt.ExpireHours)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}
	_ = s.userRepo.UpdateLastLogin(user.ID)

	return &LoginResult{
		Token:    token,
		ExpireIn: s.jwt.ExpireHours * 3600,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Email:    user.Email,
		},
	}, nil
}

// Profile 获取当前用户信息
func (s *Service) Profile(userID int64) (*UserInfo, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	return &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
	}, nil
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(userID int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("原密码错误")
	}
	if len(newPassword) < 6 {
		return errors.New("新密码长度至少 6 位")
	}
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}
	return s.userRepo.UpdatePassword(userID, hashed)
}
