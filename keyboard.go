package cauliflower

import (
	"errors"
	"gopkg.in/telebot.v3"
	"math/rand"
	"reflect"
)

const (
	Reply  string = "reply"
	Inline string = "inline"
)

var (
	ErrRowsConflict    = errors.New("cauliflower: only one of Rows, Row or DataRows can be set")
	ErrNoRowsProvided  = errors.New("cauliflower: neither Rows or Row has been set")
	ErrInvalidKeyboard = errors.New("cauliflower: Keyboard is neither Inline or Reply")
	ErrOptionsConflict = errors.New("cauliflower: Rows and Split can not be set together")
)

type KeyboardOptions struct {
	// Optional, Custom ReplyMarkup to use
	// Default: telebot.ReplyMarkup{}
	ReplyMarkup telebot.ReplyMarkup

	// Optional, Keyboard type (cauliflower.Reply or cauliflower.Inline)
	// Default: cauliflower.Inline
	Keyboard string

	// Text and value which will be used to create buttons
	// Can be used with Split
	// Required to use one between DataRow, Rows and Row
	DataRow map[string]string

	// Rows of rows of string which will be used as
	// text and data when creating the buttons
	// Can not be used with Split
	// Required to use one between DataRow, Rows and Row
	Rows [][]string

	// Rows of string if Rows is not used
	// Can be used with Split
	// Required to use one between DataRow, Rows and Row
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
		return &m, ErrNoOptionsProvided
	}

	if len(opts.Rows) == 0 && len(opts.Row) == 0 && len(opts.DataRow) == 0 {
		return &m, ErrNoRowsProvided
	}

	if opts.Keyboard != Inline && opts.Keyboard != Reply {
		return &m, ErrInvalidKeyboard
	}

	if len(opts.Rows) > 0 && opts.Split != 0 {
		return &m, ErrOptionsConflict
	}

	count := 0

	if len(opts.DataRow) > 0 {
		count++
	}

	if len(opts.Rows) > 0 {
		count++
	}

	if len(opts.Row) > 0 {
		count++
	}

	if count > 1 {
		return &m, ErrRowsConflict
	}

	// handle defaults
	if !reflect.DeepEqual(opts.ReplyMarkup, telebot.ReplyMarkup{}) {
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
	var row []telebot.Btn

	switch {
	case opts.DataRow != nil:
		for text, value := range opts.DataRow {
			var button telebot.Btn

			if opts.Keyboard == Inline {
				button = m.Data(text, randomString(16), value)
			} else {
				button = m.Text(text)
			}

			if opts.Handler != nil {
				i.Bot.Handle(&button, opts.Handler)
			}

			row = append(row, button)
		}
	case opts.Rows != nil:
		for _, r := range opts.Rows {
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
	case opts.Row != nil:
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
	default:
		return &m, ErrRowsConflict
	}

	if opts.Split != 0 {
		rows = m.Split(opts.Split, row)
	} else {
		rows = append(rows, m.Row(row...))
	}

	if opts.Keyboard == Inline {
		m.Inline(rows...)
	} else {
		m.Reply(rows...)
	}

	return &m, nil
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
