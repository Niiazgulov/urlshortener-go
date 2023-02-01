package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
)

var key = []byte("encryption key")

func NewUserSign(id string) (string, error) {
	idint, err := strconv.Atoi(id)
	if err != nil {
		return "", fmt.Errorf("unable to convert Id (NewUserSign func): %w", err)
	}
	iduint32 := uint32(idint)
	databyte := binary.BigEndian.AppendUint32(nil, iduint32)
	hash := hmac.New(sha256.New, key)
	hash.Write(databyte)
	sign := hash.Sum(nil)
	databyte = append(databyte, sign...)
	newsign := hex.EncodeToString(databyte)
	return newsign, nil
}

// func GetUserSign(s string) (string, error) {
// 	decodedbyte, err := hex.DecodeString(s)
// 	if err != nil {
// 		return "", fmt.Errorf("unable to Decode String (GetUserSign func): %w", err)
// 	}
// 	signuint32 := binary.BigEndian.Uint32(decodedbyte[:4])
// 	hash := hmac.New(sha256.New, key)
// 	hash.Write(decodedbyte[:4])
// 	usersign := strconv.Itoa(int(signuint32))
// 	return usersign, nil
// }

func GetUserSign(s string) (string, bool, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return "0", false, err
	}
	id := binary.BigEndian.Uint32(data[:4])
	usersign := strconv.Itoa(int(id))
	h := hmac.New(sha256.New, key)
	h.Write(data[:4])
	sign := h.Sum(nil)
	return usersign, hmac.Equal(sign, data[4:]), nil
}

// func GetUserSign2(s string) (string, bool, error) {
// 	decodedbyte, err := hex.DecodeString(s)
// 	if err != nil {
// 		return "", fmt.Errorf("unable to Decode String (GetUserSign func): %w", err)
// 	}
// 	signuint32 := binary.BigEndian.Uint32(decodedbyte[:4])
// 	hash := hmac.New(sha256.New, key)
// 	hash.Write(decodedbyte[:4])
// 	usersign := strconv.Itoa(int(signuint32))
// 	return usersign, nil
// }
