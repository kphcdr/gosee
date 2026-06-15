package server

import (
	"errors"
	"time"

	"gosee/internal/config"
	"gosee/internal/model"
	"gosee/internal/repository"
	"gosee/internal/sshclient"
	"gosee/internal/utils"
)

// Service 服务器管理服务
type Service struct {
	repo      *repository.ServerRepository
	groupRepo *repository.ServerGroupRepository
	sec       *config.SecurityConfig
	sshTimeout time.Duration
}

func NewService(repo *repository.ServerRepository, groupRepo *repository.ServerGroupRepository, sec *config.SecurityConfig, sshTimeout time.Duration) *Service {
	return &Service{repo: repo, groupRepo: groupRepo, sec: sec, sshTimeout: sshTimeout}
}

// SaveInput 创建/编辑服务器入参。私钥、密码为明文，仅创建或修改时传入。
type SaveInput struct {
	ID         int64  `json:"id"`
	Name       string `json:"name" binding:"required"`
	GroupID    *int64 `json:"group_id"`
	Host       string `json:"host" binding:"required"`
	Port       int    `json:"port"`
	Username   string `json:"username" binding:"required"`
	AuthType   string `json:"auth_type"`
	PrivateKey string `json:"private_key"`
	Passphrase string `json:"passphrase"`
	Password   string `json:"password"`
	Remark     string `json:"remark"`
	Enabled    *int8  `json:"enabled"`
}

// List 分页列表
func (s *Service) List(q repository.ServerListQuery) ([]model.Server, int64, error) {
	return s.repo.List(q)
}

// Get 详情
func (s *Service) Get(id int64) (*model.Server, error) {
	return s.repo.FindByID(id)
}

// Create 新增
func (s *Service) Create(in SaveInput) (*model.Server, error) {
	server := &model.Server{
		Name:     in.Name,
		GroupID:  in.GroupID,
		Host:     in.Host,
		Port:     orDefault(in.Port, 22),
		Username: in.Username,
		AuthType: normAuthType(in.AuthType),
		Remark:   in.Remark,
		Status:   model.ServerStatusUnknown,
		Enabled:  1,
	}
	if in.Enabled != nil {
		server.Enabled = *in.Enabled
	}
	if err := s.applyCredentials(server, in); err != nil {
		return nil, err
	}
	if err := s.repo.Create(server); err != nil {
		return nil, err
	}
	return server, nil
}

// Update 编辑。未提供新凭证时保留旧值。
func (s *Service) Update(in SaveInput) (*model.Server, error) {
	server, err := s.repo.FindByID(in.ID)
	if err != nil {
		return nil, errors.New("服务器不存在")
	}
	server.Name = in.Name
	server.GroupID = in.GroupID
	server.Host = in.Host
	server.Port = orDefault(in.Port, server.Port)
	server.Username = in.Username
	if in.AuthType != "" {
		server.AuthType = normAuthType(in.AuthType)
	}
	server.Remark = in.Remark
	if in.Enabled != nil {
		server.Enabled = *in.Enabled
	}
	if err := s.applyCredentials(server, in); err != nil {
		return nil, err
	}
	if err := s.repo.Update(server); err != nil {
		return nil, err
	}
	return server, nil
}

// applyCredentials 仅在传入新凭证时加密覆盖；否则保留旧值
func (s *Service) applyCredentials(server *model.Server, in SaveInput) error {
	if in.PrivateKey != "" {
		enc, err := utils.EncryptAESGCM(in.PrivateKey, s.sec.EncryptionKey)
		if err != nil {
			return errors.New("私钥加密失败: " + err.Error())
		}
		server.PrivateKeyEncrypted = &enc
	}
	if in.Password != "" {
		enc, err := utils.EncryptAESGCM(in.Password, s.sec.EncryptionKey)
		if err != nil {
			return errors.New("密码加密失败: " + err.Error())
		}
		server.PasswordEncrypted = &enc
	}
	return nil
}

// Delete 删除
func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}

// ResolveSSHConfig 读取服务器记录并解密凭证，构造明文 SSH 配置
func (s *Service) ResolveSSHConfig(id int64) (sshclient.Config, error) {
	server, err := s.repo.FindByID(id)
	if err != nil {
		return sshclient.Config{}, errors.New("服务器不存在")
	}
	return s.buildSSHConfig(server)
}

// MarkStatus 更新服务器状态与最近采集信息
func (s *Service) MarkStatus(id int64, status string, lastErr string) error {
	return s.repo.UpdateStatus(id, status, lastErr, true)
}

// TestSSH 测试 SSH 连接，依据结果更新服务器状态
func (s *Service) TestSSH(id int64) error {
	cfg, err := s.ResolveSSHConfig(id)
	if err != nil {
		return err
	}
	if err := sshclient.TestConnection(cfg); err != nil {
		_ = s.MarkStatus(id, model.ServerStatusOffline, err.Error())
		return errors.New("SSH 连接失败: " + err.Error())
	}
	_ = s.MarkStatus(id, model.ServerStatusNormal, "")
	return nil
}

// buildSSHConfig 由数据库记录构造明文 SSH 配置
func (s *Service) buildSSHConfig(server *model.Server) (sshclient.Config, error) {
	cfg := sshclient.Config{
		Host:           server.Host,
		Port:           server.Port,
		Username:       server.Username,
		AuthType:       server.AuthType,
		ConnectTimeout: s.sshTimeout,
	}
	switch server.AuthType {
	case model.AuthTypePassword:
		if server.PasswordEncrypted == nil {
			return cfg, errors.New("该服务器未配置密码")
		}
		pw, err := utils.DecryptAESGCM(*server.PasswordEncrypted, s.sec.EncryptionKey)
		if err != nil {
			return cfg, errors.New("密码解密失败")
		}
		cfg.Password = pw
	default: // private_key
		if server.PrivateKeyEncrypted == nil {
			return cfg, errors.New("该服务器未配置私钥")
		}
		pk, err := utils.DecryptAESGCM(*server.PrivateKeyEncrypted, s.sec.EncryptionKey)
		if err != nil {
			return cfg, errors.New("私钥解密失败")
		}
		cfg.PrivateKey = pk
	}
	return cfg, nil
}

func orDefault(port, def int) int {
	if port <= 0 {
		return def
	}
	return port
}

func normAuthType(t string) string {
	if t == model.AuthTypePassword {
		return model.AuthTypePassword
	}
	return model.AuthTypePrivateKey
}
