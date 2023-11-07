package cauliflower

import (
	"gopkg.in/telebot.v3"
	"errors"
	"time"
)

var (
	ErrTimeoutExceeded   = errors.New("cauliflower: didn't receive a message before the end of the timeout")
	ErrCancelCommand     = errors.New("cauliflower: Listen function has been canceled")
	ErrChatIsNil         = errors.New("cauliflower: ListenOptions.Chat can't be nil")
)

type ListenOptions struct {
	// Required
	Context telebot.Context

	// Optional, timeout before listener is cancelled
	// Default: 1 * time.Minute
	Timeout time.Duration

	// Optional, function to execute in case of timeout error
	TimeoutHandler func(telebot.Context) error

	// Optional, function to execute in case of cancel error
	CancelHandler func(telebot.Context) error

	// Optional, cancel command to cancel listening
	Cancel string

	// Optional, message to send in chat before listener starts
	Message string

	// Optional, will edit the message instead of sending a new one
	Edit telebot.Editable
}

func (i *Instance) Listen(opts *ListenOptions) (*telebot.Message, *telebot.Message, error) {
	var sentMessage *telebot.Message

	// handle error
	if opts == nil {
		return sentMessage, &telebot.Message{}, ErrNoOptionsProvided
	}

	if opts.Context.Chat().ID == 0 {
		return sentMessage, &telebot.Message{}, ErrChatIsNil
	}

	// handle defaults
	if opts.Timeout == 0 {
		opts.Timeout = i.DefaultListen.Timeout
	}

	if opts.Cancel == "" {
		opts.Cancel = i.DefaultListen.Cancel
	}

	if opts.TimeoutHandler == nil {
		opts.TimeoutHandler = i.DefaultListen.TimeoutHandler
	}

	if opts.CancelHandler == nil {
		opts.CancelHandler = i.DefaultListen.CancelHandler
	}

	if opts.Message == "" {
		opts.Message = i.DefaultListen.Message
	}

	// actual code
	if opts.Message != "" {
		var err error

		if opts.Edit != nil {
			sentMessage, err = i.Bot.Edit(opts.Edit, opts.Message)
		} else {
			sentMessage, err = i.Bot.Send(opts.Context.Chat(), opts.Message)
		}

		if err != nil {
			return sentMessage, &telebot.Message{}, err
		}
	}

	messageChannel := make(chan *telebot.Message)

	i.Channel[opts.Context.Chat().ID] = &messageChannel

	select {
	case response := <-messageChannel:
		delete(i.Channel, opts.Context.Chat().ID)

		if response.Text == opts.Cancel {
			if opts.CancelHandler != nil {
				opts.CancelHandler(opts.Context)
			}
			return sentMessage, response, ErrCancelCommand
		}

		return sentMessage, response, nil
	case <-time.After(opts.Timeout):
		if opts.TimeoutHandler != nil {
			opts.TimeoutHandler(opts.Context)
		}

		return sentMessage, &telebot.Message{}, ErrTimeoutExceeded
	}
}