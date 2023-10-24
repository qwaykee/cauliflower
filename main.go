package cauliflower

import (
	"gopkg.in/telebot.v3"
	"time"
	"errors"
)

var (
	ErrTimeoutExceeded = errors.New("cauliflower: Didn't receive a message before the end of the timeout")
	ErrBotIsNil = errors.New("cauliflower: Settings.Bot can't be nil")
	ErrContextIsNil = errors.New("cauliflower: Parameters.Context can't be nil")
)

type Instance struct {
	Bot 		*telebot.Bot
	Timeout 	time.Duration
	Channel 	map[int64](*chan *telebot.Message)
}

type Settings struct {
	Bot 				*telebot.Bot

	// Default timeout for every Listen() call
	// Will be overriden if Listen() call has Parameters.Timeout field filled
	// Optional, default: 1 * time.Minute
	Timeout 			time.Duration

	// List of dummy handlers to create in order to make Listen() work
	// Will be overriden if instance is created before creating another handle
	// Recommended, default: telebot.OnText
	Handlers 			[]string

	// Automatically install middleware instead of doing it manually
	// Execute: Bot.Use(i.Middleware())
	// Optional, default: false
	InstallMiddleware 	bool
}

type Parameters struct {
	Context telebot.Context

	// Timeout before listener is cancelled
	// Optional, default: Instance.Settings.Timeout
	Timeout time.Duration

	// Message to send in chat before listener starts
	// Optional
	Message string
}

func NewInstance(settings Settings) (*Instance, error) {
	if settings.Bot == nil {
		return nil, ErrBotIsNil
	}

	if settings.Timeout == 0 {
		settings.Timeout = 1 * time.Minute
	}

	if len(settings.Handlers) == 0 {
		settings.Handlers = []string{telebot.OnText}
	}

	for _, handler := range settings.Handlers {
		settings.Bot.Handle(handler, func(c telebot.Context) error {return nil})
	}
	
	i := Instance{
		Bot: 		settings.Bot,
		Timeout: 	settings.Timeout,
		Channel: 	make(map[int64](*chan *telebot.Message)),
	}

	if settings.InstallMiddleware {
		settings.Bot.Use(i.Middleware())
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

func (i *Instance) Listen(params Parameters) (*telebot.Message, error) {
	if params.Context == nil {
		return &telebot.Message{}, ErrContextIsNil
	}

	if params.Timeout == 0 {
		params.Timeout = i.Timeout
	}

	if params.Message != "" {
		params.Context.Send(params.Message)
	}

	ch := make(chan *telebot.Message)

	i.Channel[params.Context.Chat().ID] = &ch

	message := make(chan *telebot.Message)
	error := make(chan error)

	go func() {
		select {
		case response := <-ch:
			message <- response
			error <- nil
		case <-time.After(params.Timeout):
			message <- &telebot.Message{}
			error <- ErrTimeoutExceeded
		}
	}()

	response := <-message
	err := <-error

	if err != nil {
		return &telebot.Message{}, err
	}

	return response, nil
}