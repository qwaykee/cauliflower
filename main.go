package cauliflower

import (
	"errors"
	"gopkg.in/telebot.v3"
	"reflect"
	"time"
)

var (
	// global error
	ErrNoOptionsProvided = errors.New("cauliflower: no options provided")

	// main.go error
	ErrBotIsNil = errors.New("cauliflower: Settings.Bot can't be nil")
)

type Instance struct {
	Bot             *telebot.Bot
	DefaultListen   *ListenOptions
	DefaultKeyboard *KeyboardOptions
	Channel         map[int64](*chan *telebot.Message)
}

type Settings struct {
	// Required
	Bot *telebot.Bot

	// Automatically install middleware instead of doing it manually
	// Execute: Bot.Use(i.Middleware())
	// Optional, default: false
	InstallMiddleware bool

	// Default options for Listen()
	// Optional
	DefaultListen *ListenOptions

	// Default options for Keyboard()
	// Optional
	DefaultKeyboard *KeyboardOptions

	// List of dummy handlers to create in order to make Listen() work
	// Will be overridden if instance is created before creating another handle
	// Optional, default: telebot.OnText
	Handlers []string
}

func NewInstance(settings *Settings) (*Instance, error) {
	// handle errors
	if settings == nil {
		return nil, ErrNoOptionsProvided
	}

	if settings.Bot == nil {
		return nil, ErrBotIsNil
	}

	// handle listen defaults
	if settings.DefaultListen == nil {
		settings.DefaultListen = &ListenOptions{}
	}

	if settings.DefaultListen.Timeout == 0 {
		settings.DefaultListen.Timeout = 1 * time.Minute
	}

	// handle keyboard defaults
	if settings.DefaultKeyboard == nil {
		settings.DefaultKeyboard = &KeyboardOptions{}
	}

	if reflect.DeepEqual(settings.DefaultKeyboard.ReplyMarkup, telebot.ReplyMarkup{}) {
		settings.DefaultKeyboard.ReplyMarkup = telebot.ReplyMarkup{}
	}

	if settings.DefaultKeyboard.Keyboard == "" {
		settings.DefaultKeyboard.Keyboard = Inline
	}

	i := Instance{
		Bot:             settings.Bot,
		DefaultListen:   settings.DefaultListen,
		DefaultKeyboard: settings.DefaultKeyboard,
		Channel:         make(map[int64](*chan *telebot.Message)),
	}

	if settings.InstallMiddleware {
		settings.Bot.Use(i.Middleware())
	}

	if len(settings.Handlers) == 0 {
		settings.Handlers = []string{telebot.OnText}
	}

	for _, handler := range settings.Handlers {
		settings.Bot.Handle(handler, func(c telebot.Context) error { return nil })
	}

	return &i, nil
}

func (i *Instance) Middleware() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			if ch, ok := i.Channel[c.Chat().ID]; ok {
				*ch <- c.Message()
			}
			return next(c)
		}
	}
}
