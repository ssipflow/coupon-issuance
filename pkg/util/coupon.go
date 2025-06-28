package util

import (
	"math/rand"
	"time"
)

var (
	koreanRunes = []rune("가나다라마바사아자차카타파하거너더러머버서어저처커터퍼허")
	numberRunes = []rune("0123456789")
	fullRunes   = append(koreanRunes, numberRunes...)
)

func GenerateCouponCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	code := make([]rune, 10)
	for i := range code {
		code[i] = fullRunes[r.Intn(len(fullRunes))]
	}
	return string(code)
}
