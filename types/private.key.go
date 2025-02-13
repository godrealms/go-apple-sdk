package types

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// ParsePrivateKey Parse private keys in PEM format, supporting PKCS#8 and EC private keys
func ParsePrivateKey(pemKey string) (*ecdsa.PrivateKey, error) {
	// 解析 PEM 格式
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, errors.New("failed to decode PEM block: invalid format or empty key")
	}

	// 检查私钥类型
	var privateKey *ecdsa.PrivateKey
	var err error

	switch block.Type {
	case "PRIVATE KEY": // PKCS#8 格式
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.New("failed to parse PKCS#8 private key: " + err.Error())
		}
		// 确保是 EC 私钥
		if ecKey, ok := key.(*ecdsa.PrivateKey); ok {
			privateKey = ecKey
		} else {
			return nil, errors.New("parsed key is not an ECDSA private key")
		}
	case "EC PRIVATE KEY": // EC 私钥格式
		privateKey, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.New("failed to parse EC private key: " + err.Error())
		}
	default:
		return nil, errors.New("unsupported private key type: " + block.Type)
	}

	return privateKey, nil
}
