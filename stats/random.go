package stats

import (
	"crypto/rand"
	"hash/crc32"
)

const (
	dict = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func RandomID() (string, error) {
	return RandomIDFromDict(dict)
}

func GetRand16Byte() (uint32, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return 0, err
	}
	return crc32.ChecksumIEEE(uuid), nil
}

func RandomIDFromDict(dict string) (string, error) {
	instanceId, err := GetRand16Byte()
	if err != nil {
		return "", err
	}

	if instanceId == 0 {
		return "0", nil
	}

	instanceIdStr := make([]byte, 0, 8)
	for instanceId > 0 {
		ch := dict[instanceId%64]
		instanceIdStr = append(instanceIdStr, byte(ch))
		instanceId /= 64
	}
	return string(instanceIdStr), nil
}
