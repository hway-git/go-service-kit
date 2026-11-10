package rsa

import (
	"crypto"
	"crypto/rand"
	stdrsa "crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

type Config struct {
	Enabled        bool   `json:"enabled"`
	PublicKeyFile  string `json:"public_key_file"`
	PrivateKeyFile string `json:"-"`
}

// Keys 将验签或加密公钥与签名或解密私钥分开，支持只加载其中一把。
type Keys struct {
	public  *stdrsa.PublicKey
	private *stdrsa.PrivateKey
}

var (
	errPublicKey  = errors.New("RSA 公钥缺失或无效，须至少 2048 位")
	errPrivateKey = errors.New("RSA 私钥缺失或无效，须至少 2048 位")
)

// New 从 PEM 文件加载启用的密钥，错误中不输出文件路径或内容。
func New(cfg Config) (*Keys, error) {
	if cfg.PublicKeyFile == "" && cfg.PrivateKeyFile == "" {
		return nil, errors.New("启用 RSA 后至少配置一个密钥文件")
	}
	keys := &Keys{}
	if cfg.PublicKeyFile != "" {
		key, err := loadPublicKey(cfg.PublicKeyFile)
		if err != nil {
			return nil, err
		}
		keys.public = key
	}
	if cfg.PrivateKeyFile != "" {
		key, err := loadPrivateKey(cfg.PrivateKeyFile)
		if err != nil {
			return nil, err
		}
		keys.private = key
	}
	return keys, nil
}

func loadPublicKey(path string) (*stdrsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errPublicKey
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errPublicKey
	}
	var key *stdrsa.PublicKey
	switch block.Type {
	case "PUBLIC KEY":
		parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, errPublicKey
		}
		key, _ = parsed.(*stdrsa.PublicKey)
	case "RSA PUBLIC KEY":
		key, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, errPublicKey
		}
	}
	if key == nil || key.N.BitLen() < 2048 {
		return nil, errPublicKey
	}
	return key, nil
}

func loadPrivateKey(path string) (*stdrsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errPrivateKey
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errPrivateKey
	}
	var key *stdrsa.PrivateKey
	switch block.Type {
	case "PRIVATE KEY":
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, errPrivateKey
		}
		key, _ = parsed.(*stdrsa.PrivateKey)
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, errPrivateKey
		}
	}
	if key == nil || key.N.BitLen() < 2048 || key.Validate() != nil {
		return nil, errPrivateKey
	}
	return key, nil
}

// Encrypt 使用 RSA-OAEP 和 SHA-256 加密短数据。
func (k *Keys) Encrypt(plaintext []byte) ([]byte, error) {
	if k == nil || k.public == nil {
		return nil, errPublicKey
	}
	return stdrsa.EncryptOAEP(sha256.New(), rand.Reader, k.public, plaintext, nil)
}

// Decrypt 解密 RSA-OAEP 密文。
func (k *Keys) Decrypt(ciphertext []byte) ([]byte, error) {
	if k == nil || k.private == nil {
		return nil, errPrivateKey
	}
	return stdrsa.DecryptOAEP(sha256.New(), rand.Reader, k.private, ciphertext, nil)
}

// Sign 使用 RSA-PSS 签名，盐长度与 SHA-256 摘要等长。
func (k *Keys) Sign(message []byte) ([]byte, error) {
	if k == nil || k.private == nil {
		return nil, errPrivateKey
	}
	digest := sha256.Sum256(message)
	return stdrsa.SignPSS(rand.Reader, k.private, crypto.SHA256, digest[:], &stdrsa.PSSOptions{SaltLength: stdrsa.PSSSaltLengthEqualsHash})
}

// Verify 校验 RSA-PSS 签名。
func (k *Keys) Verify(message, signature []byte) error {
	if k == nil || k.public == nil {
		return errPublicKey
	}
	digest := sha256.Sum256(message)
	return stdrsa.VerifyPSS(k.public, crypto.SHA256, digest[:], signature, &stdrsa.PSSOptions{SaltLength: stdrsa.PSSSaltLengthEqualsHash})
}
