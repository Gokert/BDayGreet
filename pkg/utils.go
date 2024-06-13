package utils

import (
	"crypto/sha512"
	"math/rand"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func HashPassword(password string) []byte {
	hashPassword := sha512.Sum512([]byte(password))
	passwordByteSlice := hashPassword[:]
	return passwordByteSlice
}

func RandStringRunes(seed int) string {
	symbols := make([]rune, seed)
	for i := range symbols {
		symbols[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(symbols)
}

const MaxRetries = 5

var HeaderBirthdayEmp = "Поздравление"
var BodyBirthdayFromEmp = "Сегодня день рождения у %s, не забудьте поздравить его!"
var BodyBirthdayToEmp = "Поздравляем Вас с днём рождения!"
