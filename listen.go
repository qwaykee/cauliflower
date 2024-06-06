package cauliflower

import (
	"gopkg.in/telebot.v3"
	"errors"
	"time"
	"sync"
)

var (
	ErrTimeoutExceeded = errors.New("cauliflower: didn't receive a message before the end of the timeout")
	ErrCancelCommand   = errors.New("cauliflower: Listen function has been canceled")
	ErrChatIsNil       = errors.New("cauliflower: Chat can't be nil")
)

type (
	ListenOptions struct {
		// Default: 1 * time.Minute
		Timeout time.Duration
		CancelCommand string

		TimeoutHandler telebot.HandlerFunc
		CancelHandler telebot.HandlerFunc
	}
)

func (i *Instance) Listen(ctx telebot.Context, opts *ListenOptions) (*telebot.Message, error) {
	var receivedMessage *telebot.Message
	var err error

	if ctx.Chat() == nil {
		return receivedMessage, ErrChatIsNil
	}

	// handle defaults
	if opts.Timeout == 0 {
		opts.Timeout = i.DefaultListenOptions.Timeout
	}

	if opts.CancelCommand == "" {
		opts.CancelCommand = i.DefaultListenOptions.CancelCommand
	}

	if opts.TimeoutHandler == nil {
		opts.TimeoutHandler = i.DefaultListenOptions.TimeoutHandler
	}

	if opts.CancelHandler == nil {
		opts.CancelHandler = i.DefaultListenOptions.CancelHandler
	}

	receivedMessage, err = i.listen(ctx, opts.Timeout)

	if err == ErrTimeoutExceeded {
		if opts.TimeoutHandler != nil {
			opts.TimeoutHandler(ctx)
		}
		return receivedMessage, ErrTimeoutExceeded
	}

	if receivedMessage.Text == opts.CancelCommand {
		if opts.CancelHandler != nil {
			opts.CancelHandler(ctx)
		}
		return receivedMessage, ErrCancelCommand
	}

	return receivedMessage, nil
}

func (i *Instance) listen(ctx telebot.Context, timeout time.Duration) (*telebot.Message, error) {
	messageChannel := make(chan *telebot.Message)

	i.mutex.Lock()

	i.channel[ctx.Chat().ID] = &messageChannel

	i.mutex.Unlock()

	select {
	case response := <-messageChannel:
		delete(i.channel, ctx.Chat().ID)

		return response, nil
	case <-time.After(timeout):
		return &telebot.Message{}, ErrTimeoutExceeded
	}
}