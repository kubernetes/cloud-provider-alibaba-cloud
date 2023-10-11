package base

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-cmd/cmd"
	ctrlCfg "k8s.io/cloud-provider-alibaba-cloud/pkg/config"
	prvd "k8s.io/cloud-provider-alibaba-cloud/pkg/provider"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	AddonTokenFilePath = "/var/addon/token-config"
)

type DefaultToken struct {
	Region          string
	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
}

// TokenAuth is an interface of Token auth method
type TokenAuth interface {
	NextToken() (*DefaultToken, error)
}

// AkAuthToken implement ak auth
type AkAuthToken struct {
	Region string
}

func (f *AkAuthToken) NextToken() (*DefaultToken, error) {
	key, secret, err := LoadAK()
	if err != nil {
		return nil, err
	}
	return &DefaultToken{
		AccessKeyId:     key,
		AccessKeySecret: secret,
		Region:          f.Region,
	}, nil
}

type RamRoleToken struct {
	meta prvd.IMetaData
}

func (f *RamRoleToken) NextToken() (*DefaultToken, error) {
	roleName, err := f.meta.RoleName()
	if err != nil {
		return nil, fmt.Errorf("role name: %s", err.Error())
	}
	// use instance ram file way.
	role, err := f.meta.RamRoleToken(roleName)
	if err != nil {
		return nil, fmt.Errorf("ramrole Token retrieve: %s", err.Error())
	}
	region, err := f.meta.Region()
	if err != nil {
		return nil, fmt.Errorf("read region error: %s", err.Error())
	}

	return &DefaultToken{
		Region:          region,
		AccessKeyId:     role.AccessKeyId,
		AccessKeySecret: role.AccessKeySecret,
		SecurityToken:   role.SecurityToken,
	}, nil
}

// ServiceToken is an implementation of service account auth
type ServiceToken struct {
	Region   string
	ExecPath string
}

func (f *ServiceToken) NextToken() (*DefaultToken, error) {
	key, secret, err := LoadAK()
	if err != nil {
		return nil, err
	}
	status := <-cmd.NewCmd(
		filepath.Join(f.ExecPath, "servicetoken"),
		fmt.Sprintf("--uid=%s", ctrlCfg.CloudCFG.Global.UID),
		fmt.Sprintf("--key=%s", key),
		fmt.Sprintf("--secret=%s", secret),
		fmt.Sprintf("--region=%s", f.Region),
	).Start()
	if status.Error != nil {
		return nil, fmt.Errorf("invoke servicetoken: %s", status.Error.Error())
	}

	st := struct {
		AccessSecret string `json:"accessSecret,omitempty"`
		UID          string `json:"uid,omitempty"`
		Token        string `json:"token,omitempty"`
		AccessKey    string `json:"accesskey,omitempty"`
		Expiration   string `json:"expiration,omitempty"`
	}{}
	if err = json.Unmarshal([]byte(strings.Join(status.Stdout, "")), &st); err != nil {
		return nil, fmt.Errorf("unmarshal ServiceToken output %+v error: %s", status, err.Error())
	}

	return &DefaultToken{
		Region:          f.Region,
		AccessKeyId:     st.AccessKey,
		AccessKeySecret: st.AccessSecret,
		SecurityToken:   st.Token,
	}, nil
}

type AddonToken struct {
	Region string `json:"region,omitempty"`
}

func (f *AddonToken) NextToken() (*DefaultToken, error) {
	fileBytes, err := os.ReadFile(AddonTokenFilePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s error: %s", AddonTokenFilePath, err.Error())
	}
	akInfo := struct {
		AccessKeyId     string `json:"access.key.id,omitempty"`
		AccessKeySecret string `json:"access.key.secret,omitempty"`
		SecurityToken   string `json:"security.token,omitempty"`
		Expiration      string `json:"expiration,omitempty"`
		Keyring         string `json:"keyring,omitempty"`
	}{}
	if err = json.Unmarshal(fileBytes, &akInfo); err != nil {
		return nil, fmt.Errorf("unmarshal AddonToken [%s] error: %s", string(fileBytes), err.Error())
	}

	keyring := akInfo.Keyring
	ak, err := Decrypt(akInfo.AccessKeyId, []byte(keyring))
	if err != nil {
		return nil, fmt.Errorf("failed to decode ak, err: %v", err)
	}

	sk, err := Decrypt(akInfo.AccessKeySecret, []byte(keyring))
	if err != nil {
		return nil, fmt.Errorf("failed to decode sk, err: %v", err)
	}

	token, err := Decrypt(akInfo.SecurityToken, []byte(keyring))
	if err != nil {
		return nil, fmt.Errorf("failed to decode token, err: %v", err)
	}

	t, err := time.Parse("2006-01-02T15:04:05Z", akInfo.Expiration)
	if err != nil {
		log.Error(err, "Expiration parse error")
	} else {
		if t.Before(time.Now()) {
			return nil, fmt.Errorf("invalid token which is expired")
		}
	}

	return &DefaultToken{
		Region:          f.Region,
		AccessKeyId:     string(ak),
		AccessKeySecret: string(sk),
		SecurityToken:   string(token),
	}, nil
}

func LoadAK() (string, string, error) {
	var keyId, keySecret string
	log.V(5).Info(fmt.Sprintf("load cfg from file: %s", ctrlCfg.ControllerCFG.CloudConfigPath))
	if err := ctrlCfg.CloudCFG.LoadCloudCFG(); err != nil {
		return "", "", fmt.Errorf("load cloud config %s error: %v",
			ctrlCfg.ControllerCFG.CloudConfigPath, err.Error())
	}

	if ctrlCfg.CloudCFG.Global.AccessKeyID != "" && ctrlCfg.CloudCFG.Global.AccessKeySecret != "" {
		key, err := base64.StdEncoding.DecodeString(ctrlCfg.CloudCFG.Global.AccessKeyID)
		if err != nil {
			return "", "", err
		}
		keyId = string(key)
		secret, err := base64.StdEncoding.DecodeString(ctrlCfg.CloudCFG.Global.AccessKeySecret)
		if err != nil {
			return "", "", err
		}
		keySecret = string(secret)
	}

	if keyId == "" || keySecret == "" {
		log.V(5).Info("LoadAK: cloud config does not have keyId or keySecret. " +
			"try environment ACCESS_KEY_ID ACCESS_KEY_SECRET")
		keyId = os.Getenv(AccessKeyID)
		keySecret = os.Getenv(AccessKeySecret)
		if keyId == "" || keySecret == "" {
			return "", "", fmt.Errorf("cloud config and env do not have keyId or keySecret, load AK failed")
		}
	}
	return keyId, keySecret, nil
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func Decrypt(s string, keyring []byte) ([]byte, error) {
	cdata, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string, err: %s", err.Error())
	}
	block, err := aes.NewCipher(keyring)
	if err != nil {
		return nil, fmt.Errorf("failed to new cipher, err: %s", err.Error())
	}
	blockSize := block.BlockSize()

	iv := cdata[:blockSize]
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(cdata)-blockSize)

	blockMode.CryptBlocks(origData, cdata[blockSize:])

	origData = PKCS5UnPadding(origData)
	return origData, nil
}
