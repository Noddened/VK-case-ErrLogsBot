package adapters

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/Noddened/ErrLogsBot/internal/config"
	"github.com/Noddened/ErrLogsBot/internal/utils"
	"gopkg.in/telebot.v4"
)

// реализует интерфейс TelegramSender для отправки сообщений
type TelegramAdapter struct {
	bot             *telebot.Bot
	subscribers     map[int64]bool
	mu              sync.RWMutex
	subscribersFile string
}

func NewTelegramAdapter(cfg *config.Config) (*TelegramAdapter, error) {
	pref := telebot.Settings{
		Token:  cfg.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, err
	}

	adapter := &TelegramAdapter{
		bot:             bot,
		subscribers:     make(map[int64]bool),
		subscribersFile: "configs/subscribers.json",
	}

	// Парсинг пользователей с диска
	if err := adapter.LoadSubscribers(); err != nil {
		slog.Warn("Не удалось загрузить юзеров", "error", err)
	}
	return adapter, nil
}

func (a *TelegramAdapter) LoadSubscribers() error {
	data, err := os.ReadFile(a.subscribersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // если файла нету, то все норм
		}
		return err
	}

	var ids []int64
	if err := json.Unmarshal(data, &ids); err != nil {
		return err
	}

	a.mu.Lock()
	for _, id := range ids {
		a.subscribers[id] = true
	}
	a.mu.Unlock()

	slog.Info("Загружено юзеров", "count", len(ids))
	return nil
}

func (a *TelegramAdapter) SaveSubscribers() error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	ids := make([]int64, 0, len(a.subscribers))
	for id := range a.subscribers {
		ids = append(ids, id)
	}

	data, err := json.MarshalIndent(ids, "", "   ")
	if err != nil {
		return err
	}

	return os.WriteFile(a.subscribersFile, data, 0644)
}

// Добавить юзера
func (a *TelegramAdapter) Subscribe(chatID int64) error {
	a.mu.Lock()
	a.subscribers[chatID] = true
	a.mu.Unlock()
	return a.SaveSubscribers()
}

// Убрать юзера
func (a *TelegramAdapter) Unsubscribe(chatID int64) error {
	a.mu.Lock()
	delete(a.subscribers, chatID)
	a.mu.Unlock()
	return a.SaveSubscribers()
}

func (a *TelegramAdapter) GetSubscribers() []int64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ids := make([]int64, 0, len(a.subscribers))
	for id := range a.subscribers {
		ids = append(ids, id)
	}
	return ids
}

// Отправка сообщений всем юзерам бота
func (a *TelegramAdapter) Broadcast(line string) {
	// Запускаем отправку в отдельной горутине, чтобы не блокировать монитор логов
	go func() {
		subs := a.GetSubscribers()
		if len(subs) == 0 {
			return
		}
		text := utils.EscapeMarkdownV2(line)

		for _, chatID := range subs {
			_, err := a.bot.Send(&telebot.User{ID: chatID}, text, &telebot.SendOptions{
				ParseMode:             telebot.ModeMarkdownV2,
				DisableWebPagePreview: true,
			})

			if err != nil {
				slog.Error("Не удалось отправить пользователю", "chat_id", chatID, "error", err)
				// если бота блокнули, то удаляем юзера из списка
				if strings.Contains(err.Error(), "blocked") || strings.Contains(err.Error(), "not found") {
					slog.Info("Удаляем заблокировавшего бота пользователя", "chat_id", chatID)
					a.Unsubscribe(chatID)
				}
			} else {
				slog.Debug("Сообщение отправлено", "chat_id", chatID)
			}
		}
	}()
}

func (a *TelegramAdapter) Start() {
	a.bot.Handle("/start", func(c telebot.Context) error {
		chatID := c.Chat().ID
		if err := a.Subscribe(chatID); err != nil {
			// Не дай бог я это на прод залью
			return c.Send("Ошибка подписки :Ъ")
		}
		return c.Send(
			"Теперь ты будешь получать логи.\n\n" +
				"Чтобы прекратить получение сообщений напиши /stop\n",
		)
	})
	a.bot.Handle("/stop", func(c telebot.Context) error {
		chatID := c.Chat().ID
		if err := a.Unsubscribe(chatID); err != nil {
			return c.Send("АХАХАХАХАХ не отпишешься)))))))))")
		}
		return c.Send("Прекращение рассылки, больше логи не будут приходить.\n")
	})
	slog.Info("Telegram бот запущен")

	a.bot.Start()
}
