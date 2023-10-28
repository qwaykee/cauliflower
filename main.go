package cauliflower

import (
	"errors"
	"gopkg.in/telebot.v3"
	"time"
)

var (
	ErrTimeoutExceeded = errors.New("cauliflower: Didn't receive a message before the end of the timeout")
	ErrCancelCommand   = errors.New("cauliflower: Listen function has been canceled")
	ErrBotIsNil        = errors.New("cauliflower: Settings.Bot can't be nil")
	ErrChatIsNil       = errors.New("cauliflower: Parameters.Chat can't be nil")
)

type Instance struct {
	Bot            *telebot.Bot
	Timeout        time.Duration
	Cancel         string
	TimeoutHandler func(telebot.Context) error
	CancelHandler  func(telebot.Context) error
	Channel        map[int64](*chan *telebot.Message)
}

type Settings struct {
	Bot *telebot.Bot

	// Default timeout for every Listen() call
	// Optional, default: 1 * time.Minute
	Timeout time.Duration

	// Default cancel command for Listen()
	// Optional
	Cancel string

	// Default function to execute in case of timeout error
	// Optional
	TimeoutHandler func(telebot.Context) error

	// Default function to execute in case of cancel error
	// Optional
	CancelHandler func(telebot.Context) error

	// List of dummy handlers to create in order to make Listen() work
	// Will be overridden if instance is created before creating another handle
	// Recommended, default: telebot.OnText
	Handlers []string

	// Automatically install middleware instead of doing it manually
	// Execute: Bot.Use(i.Middleware())
	// Optional, default: false
	InstallMiddleware bool
}

type Parameters struct {
	// Required
	Context telebot.Context

	// Timeout before listener is cancelled
	// Optional, default: Instance.Settings.Timeout
	Timeout time.Duration

	// Function to execute in case of timeout error
	// Optional
	TimeoutHandler func(telebot.Context) error

	// Function to execute in case of cancel error
	// Optional
	CancelHandler func(telebot.Context) error

	// Cancel command to cancel listening
	// Optional
	Cancel string

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

	i := Instance{
		Bot:            settings.Bot,
		Timeout:        settings.Timeout,
		Cancel:         settings.Cancel,
		TimeoutHandler: settings.TimeoutHandler,
		CancelHandler:  settings.CancelHandler,
		Channel:        make(map[int64](*chan *telebot.Message)),
	}

	if settings.InstallMiddleware {
		settings.Bot.Use(i.Middleware())
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

func (i *Instance) Listen(params Parameters) (*telebot.Message, *telebot.Message, error) {
	var sentMessage *telebot.Message

	if params.Context.Chat().ID == 0 {
		return sentMessage, &telebot.Message{}, ErrChatIsNil
	}

	if params.Timeout == 0 {
		params.Timeout = i.Timeout
	}

	if params.Cancel == "" {
		params.Cancel = i.Cancel
	}

	if params.TimeoutHandler == nil {
		params.TimeoutHandler = i.TimeoutHandler
	}

	if params.CancelHandler == nil {
		params.CancelHandler = i.CancelHandler
	}

	if params.Message != "" {
		var err error
		sentMessage, err = i.Bot.Send(params.Context.Chat(), params.Message)

		if err != nil {
			return sentMessage, &telebot.Message{}, err
		}
	}

	messageChannel := make(chan *telebot.Message)

	i.Channel[params.Context.Chat().ID] = &messageChannel

	select {
	case response := <-messageChannel:
		delete(i.Channel, params.Context.Chat().ID)

		if response.Text == params.Cancel {
			if params.CancelHandler != nil {
				params.CancelHandler(params.Context)
			}
			return sentMessage, response, ErrCancelCommand
		}

		return sentMessage, response, nil
	case <-time.After(params.Timeout):
		if params.TimeoutHandler != nil {
			params.TimeoutHandler(params.Context)
		}
		return sentMessage, &telebot.Message{}, ErrTimeoutExceeded
	}
}
