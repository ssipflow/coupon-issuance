package util

import (
	"github.com/google/uuid"
	"strings"
)

func GenerateCouponCode() string {
	return strings.ToUpper(uuid.New().String()[:10])
}
