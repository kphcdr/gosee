package sshclient

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config SSH 连接参数
type Config struct {
	Host           string
	Port           int
	Username       string
	AuthType       string // private_key | password
	PrivateKey     string // 私钥明文（PEM）
	Passphrase     string // 私钥口令（可选）
	Password       string // 密码（auth_type=password 时）
	ConnectTimeout time.Duration
}

// Connect 建立 SSH 连接
func Connect(cfg Config) (*ssh.Client, error) {
	authMethods, err := buildAuth(cfg)
	if err != nil {
		return nil, err
	}
	sshCfg := &ssh.ClientConfig{
		User:            cfg.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 第一版不校验主机指纹
		Timeout:         cfg.ConnectTimeout,
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return ssh.Dial("tcp", addr, sshCfg)
}

// TestConnection 建立连接并执行 `echo ok` 验证可用性，随后关闭
func TestConnection(cfg Config) error {
	client, err := Connect(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	return session.Run("echo ok")
}

// RunCommand 在已建立的连接上执行命令，返回标准输出
func RunCommand(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout
	if err := session.Run(cmd); err != nil {
		return stdout.String(), err
	}
	return stdout.String(), nil
}

// RunCommandWithTimeout 执行命令，超时后发送 SIGKILL 中止
func RunCommandWithTimeout(client *ssh.Client, cmd string, timeout time.Duration) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout

	done := make(chan error, 1)
	go func() { done <- session.Run(cmd) }()

	if timeout <= 0 {
		if err := <-done; err != nil {
			return stdout.String(), err
		}
		return stdout.String(), nil
	}

	select {
	case err := <-done:
		return stdout.String(), err
	case <-time.After(timeout):
		_ = session.Signal(ssh.SIGKILL)
		return stdout.String(), errors.New("命令执行超时")
	}
}

func buildAuth(cfg Config) ([]ssh.AuthMethod, error) {
	switch cfg.AuthType {
	case "private_key", "":
		if cfg.PrivateKey == "" {
			return nil, fmt.Errorf("私钥认证未提供私钥")
		}
		var (
			signer ssh.Signer
			err    error
		)
		if cfg.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(cfg.PrivateKey), []byte(cfg.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
		}
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
	case "password":
		if cfg.Password == "" {
			return nil, fmt.Errorf("密码认证未提供密码")
		}
		return []ssh.AuthMethod{ssh.Password(cfg.Password)}, nil
	default:
		return nil, fmt.Errorf("不支持的认证方式: %s", cfg.AuthType)
	}
}
