package cauliflower

import (
	"gopkg.in/telebot.v3"
	"errors"
	"math/rand"
)

const (
	Reply string = "reply"
	Inline string = "inline"
)

var (
	ErrRowsConflict = errors.New("cauliflower: only one of KeyboardOptions.Rows and KeyboardOptions.Row can be set")
	ErrNoRowsProvided = errors.New("cauliflower: neither KeyboardOptions.Rows or KeyboardOptions.Row has been set")
	ErrInvalidKeyboard = errors.New("cauliflower: KeyboardOptions.Keyboard is neither Inline or Reply")
)

type KeyboardOptions struct {
	// Optional, Custom ReplyMarkup to use
	// Default: telebot.ReplyMarkup{}
	ReplyMarkup *telebot.ReplyMarkup

	// Optional, Keyboard type (cauliflower.Reply or cauliflower.Inline)
	// Default: cauliflower.Inline
	Keyboard string

	// Rows of rows of string which will be used as
	// text and data when creating the buttons
	// Required if Row isn't set
	Rows [][]string

	// Rows of string if Rows is not used
	// Use with Split
	// Required if Rows isn't set
	Row []string

	// Will be used with ReplyMarkup.Split()
	// If zero, will create a single row
	Split int

	// Optional, Function to call when any button is called
	Handler func(c telebot.Context) error
}

func (i *Instance) Keyboard(opts *KeyboardOptions) (*telebot.ReplyMarkup, error) {
	m := i.DefaultKeyboard.ReplyMarkup

	// handle errors
	if opts == nil {
		return m, ErrNoOptionsProvided
	}

	if len(opts.Rows) > 0 && len(opts.Row) > 0 {
		return m, ErrRowsConflict
	}

	if len(opts.Rows) == 0 && len(opts.Row) == 0 {
		return m, ErrNoRowsProvided
	}

	if opts.Keyboard != Inline && opts.Keyboard != Reply {
		return m, ErrInvalidKeyboard
	}

	// handle defaults
	if opts.ReplyMarkup != nil {
		m = opts.ReplyMarkup
	}
	
	if opts.Keyboard == "" {
		opts.Keyboard = i.DefaultKeyboard.Keyboard
	}

	if opts.Split == 0 {
		opts.Split = i.DefaultKeyboard.Split
	}

	// actual code
	var rows []telebot.Row

	if opts.Rows != nil {
		for _, r := range opts.Rows {
			var row []telebot.Btn

			for _, text := range r {
				var button telebot.Btn

				if opts.Keyboard == Inline {
					button = m.Data(text, randomString(16), text)
				} else {
					button = m.Text(text)
				}

				if opts.Handler != nil {
					i.Bot.Handle(&button, opts.Handler)
				}

				row = append(row, button)
			}

			rows = append(rows, row)
		}
	} else {
		var row []telebot.Btn

		for _, text := range opts.Row {
			var button telebot.Btn

			if opts.Keyboard == Inline {
				button = m.Data(text, randomString(16), text)
			} else {
				button = m.Text(text)
			}

			if opts.Handler != nil {
				i.Bot.Handle(&button, opts.Handler)
			}

			row = append(row, button)
		}

		if opts.Split != 0 {
			rows = m.Split(opts.Split, row)
		} else {
			rows = append(rows, m.Row(row...))
		}
	}

	if opts.Keyboard == Inline {
		m.Inline(rows...)
	} else {
		m.Reply(rows...)
	}

	return m, nil
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}