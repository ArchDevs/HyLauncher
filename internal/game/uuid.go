package game

import (
	"strings"

	"github.com/google/uuid"
)

func OfflineUUID(nick string) uuid.UUID {
	data := []byte("OfflinePlayer:" + strings.TrimSpace(nick))
	return uuid.NewMD5(uuid.Nil, data)
}
