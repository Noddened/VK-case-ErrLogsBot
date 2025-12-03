package validators

import (
	"errors"
	"strings"
)

func ValidateBotToken(token string) error {
	if strings.TrimSpace(token) == "" {
		return errors.New("токен бота не может быть пустым")
	}

	if !strings.Contains(token, ":") {
		return errors.New("неверный формат токена бота")
	}

	if len(token) < 20 || len(token) > 100 {
		return errors.New("неверная длинна токена бота")
	}

	return nil
}
