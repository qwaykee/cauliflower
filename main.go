package cauliflower

import (
	"gopkg.in/telebot.v3"
	"time"
	"sync"
)

type (
	Instance struct {
		DefaultListenOptions   	ListenOptions

		bot             		*telebot.Bot
		channel         		map[int64](*chan *telebot.Message)
		mutex 					*sync.Mutex

		messageCount			int
		responseTime			[]time.Duration
		usersLanguage			map[int64]string
	}

	Settings struct {
		// Avoid having to install the middleware manually
		InstallMiddleware bool

		DefaultListenOptions ListenOptions

		// List of dummy handlers to create in order to make Listen() work
		// Will be overridden if instance is created before creating another handle
		// Default: telebot.OnText
		Handlers []string
	}
)

func NewInstance(b *telebot.Bot, s *Settings) *Instance {
	if s.DefaultListenOptions.Timeout == 0 {
		s.DefaultListenOptions.Timeout = time.Minute
	}

	i := Instance{
		DefaultListenOptions:   s.DefaultListenOptions,
		bot:             		b,
		channel:         		make(map[int64](*chan *telebot.Message)),
		mutex: 					&sync.Mutex{},
		usersLanguage:         	make(map[int64]string),
	}

	if s.InstallMiddleware {
		i.bot.Use(i.Middleware())
	}

	if len(s.Handlers) == 0 {
		s.Handlers = []string{telebot.OnText}
	}

	for _, handler := range s.Handlers {
		i.bot.Handle(handler, func(c telebot.Context) error { return nil })
	}

	return &i
}