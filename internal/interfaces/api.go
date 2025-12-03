package interfaces

import "context"

type TelegramSender interface {
	SendMessage(ctx context.Context, message string) error
}
