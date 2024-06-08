package cauliflower

import (
	"gopkg.in/telebot.v3"
	"math/rand"
)

type (
	KeyboardType string
	ButtonType string

	Keyboard struct {
		KeyboardType KeyboardType
		Split int

		buttons []Button
		instance *Instance
	}

	Button struct {
		ButtonType ButtonType
		Handler telebot.HandlerFunc
		Text string
		Data []string
		Option any
	}
)

const (
	KeyboardInline	KeyboardType = "inline"
	KeyboardReply	KeyboardType = "reply"

	ButtonContact	ButtonType = "contact"
	ButtonLocation	ButtonType = "location"
	ButtonText		ButtonType = "text"
	ButtonData		ButtonType = "data"
	ButtonQuery		ButtonType = "query"
	ButtonQueryChat	ButtonType = "querychat"
	ButtonURL		ButtonType = "url"
	ButtonChat		ButtonType = "chat"
	ButtonLogin		ButtonType = "login"
	ButtonPoll		ButtonType = "poll"
	ButtonUser		ButtonType = "user"
	ButtonWebApp	ButtonType = "webapp"
)

func NoHandler(c telebot.Context) error {
	return nil
}

func (i *Instance) NewKeyboard(keyboardType KeyboardType, split int) *Keyboard {
	return &Keyboard{
		KeyboardType: keyboardType,
		Split: split,
		instance: i,
	}
}

func (k *Keyboard) Add(buttonType ButtonType, handler telebot.HandlerFunc, text string, data ...string) *Button {
	b := Button{
		ButtonType: buttonType,
		Handler: handler,
		Text: text,
		Data: data,
	}

	k.buttons = append(k.buttons, b)

	return &b
}

func (b *Button) WithOption(option any) *Button {
	b.Option = option

	return b
}

func (k *Keyboard) Convert() *telebot.ReplyMarkup {
	rm := &telebot.ReplyMarkup{}

	var buttons []telebot.Btn

	for _, b := range k.buttons {
		var button telebot.Btn

		switch b.ButtonType {
		case ButtonContact:
			button = rm.Contact(b.Text)
		case ButtonLocation:
			button = rm.Location(b.Text)
		case ButtonText:
			button = rm.Text(b.Text)
		case ButtonData:
			button = rm.Data(b.Text, randomString(32), b.Data...)
		case ButtonQuery:
			button = rm.Query(b.Text, b.Data[0])
		case ButtonQueryChat:
			button = rm.QueryChat(b.Text, b.Data[0])
		case ButtonURL:
			button = rm.URL(b.Text, b.Data[0])
		case ButtonChat:
			opt := b.Option.(telebot.ReplyRecipient)
			button = rm.Chat(b.Text, &opt)
		case ButtonLogin:
			opt := b.Option.(telebot.Login)
			button = rm.Login(b.Text, &opt)
		case ButtonPoll:
			button = rm.Poll(b.Text, b.Option.(telebot.PollType))
		case ButtonUser:
			opt := b.Option.(telebot.ReplyRecipient)
			button = rm.User(b.Text, &opt)
		case ButtonWebApp:
			opt := b.Option.(telebot.WebApp)
			button = rm.WebApp(b.Text, &opt)
		}

		buttons = append(buttons, button)
		k.instance.bot.Handle(&button, b.Handler)
	}

	rows := rm.Split(k.Split, buttons)

	switch k.KeyboardType {
	case KeyboardInline:
		rm.Inline(rows...)
	case KeyboardReply:
		rm.Reply(rows...)
	}

	return rm
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}

	return string(s)
}