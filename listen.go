package cauliflower

import (
	"gopkg.in/telebot.v3"
	"errors"
	"time"
)

var (
	ErrTimeoutExceeded = errors.New("cauliflower: didn't receive a message before the end of the timeout")
	ErrCancelCommand   = errors.New("cauliflower: Listen function has been canceled")
	ErrChatIsNil       = errors.New("cauliflower: Chat can't be nil")
)

type (
	InputType string

	ListenOptions struct {
		// Default: 1 * time.Minute
		Timeout time.Duration
		CancelCommand string

		TimeoutHandler telebot.HandlerFunc
		CancelHandler telebot.HandlerFunc
	}
)

const (
	TextInput InputType = "text"
	PhotoInput InputType = "photo"
	AudioInput InputType = "audio"
	VideoInput InputType = "video"
	DocumentInput InputType = "document"
	StickerInput InputType = "sticker"
	VoiceInput InputType = "voice"
	VideoNoteInput InputType = "videonote"
	AnimationInput InputType = "animation"
	ContactInput InputType = "contact"
	LocationInput InputType = "location"
	VenueInput InputType = "venue"
	PollInput InputType = "poll"
	GameInput InputType = "game"
	DiceInput InputType = "dice"
	AnyInput InputType = "any"
)

func (i *Instance) Listen(c telebot.Context, inputType InputType, opts *ListenOptions) (*telebot.Message, error) {
	var answer *telebot.Message

	if c.Chat() == nil {
		return answer, ErrChatIsNil
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

	var err error

	answer, err = i.listen(c.Chat().ID, opts.Timeout)

	if inputType != AnyInput {
		loop:
		for {
			switch inputType {
			case TextInput:
				if answer.Text != "" { break loop }
			case PhotoInput:
				if answer.Photo != nil { break loop }
			case AudioInput:
				if answer.Audio != nil { break loop }
			case VideoInput:
				if answer.Video != nil { break loop }
			case DocumentInput:
				if answer.Document != nil { break loop }
			case StickerInput:
				if answer.Sticker != nil { break loop }
			case VoiceInput:
				if answer.Voice != nil { break loop }
			case VideoNoteInput:
				if answer.VideoNote != nil { break loop }
			case AnimationInput:
				if answer.Animation != nil { break loop }
			case ContactInput:
				if answer.Contact != nil { break loop }
			case LocationInput:
				if answer.Location != nil { break loop }
			case VenueInput:
				if answer.Venue != nil { break loop }
			case PollInput:
				if answer.Poll != nil { break loop }
			case GameInput:
				if answer.Game != nil { break loop }
			case DiceInput:
				if answer.Dice != nil { break loop }
			}

			answer, err = i.listen(c.Chat().ID, opts.Timeout)
		}
	}

	if err == ErrTimeoutExceeded {
		if opts.TimeoutHandler != nil {
			opts.TimeoutHandler(c)
		}
		return answer, err
	}

	if answer.Text == opts.CancelCommand {
		if opts.CancelHandler != nil {
			opts.CancelHandler(c)
		}
		return answer, ErrCancelCommand
	}

	return answer, nil
}

func (i *Instance) listen(chatID int64, timeout time.Duration) (*telebot.Message, error) {
	messageChannel := make(chan *telebot.Message)

	i.mutex.Lock()

	i.channel[chatID] = &messageChannel

	i.mutex.Unlock()

	select {
	case response := <-messageChannel:
		delete(i.channel, chatID)

		return response, nil
	case <-time.After(timeout):
		return &telebot.Message{}, ErrTimeoutExceeded
	}
}

func (i *Instance) Middleware() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			if ch, ok := i.channel[c.Chat().ID]; ok {
				*ch <- c.Message()
			}
			return next(c)
		}
	}
}