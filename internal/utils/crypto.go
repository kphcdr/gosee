package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// EncryptAESGCM 使用 AES-256-GCM 加密明文。
// keyHex 为 64 位十六进制字符串（解码后 32 字节）。
// 返回值为 hex(nonce + ciphertext)。
func EncryptAESGCM(plaintext, keyHex string) (string, error) {
	key, err := decodeKey(keyHex)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil) // nonce 前置
	return hex.EncodeToString(sealed), nil
}

// DecryptAESGCM 解密 EncryptAESGCM 产出的密文。
func DecryptAESGCM(ciphertextHex, keyHex string) (string, error) {
	key, err := decodeKey(keyHex)
	if err != nil {
		return "", err
	}
	data, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("密文不是合法的十六进制: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("密文长度不足")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}
	return string(plaintext), nil
}

func decodeKey(keyHex string) ([]byte, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, errors.New("加密密钥不是合法的十六进制字符串")
	}
	if len(key) != 32 {
		return nil, errors.New("加密密钥必须解码后为 32 字节（64 位十六进制）")
	}
	return key, nil
}
