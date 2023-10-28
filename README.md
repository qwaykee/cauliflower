# Cauliflower [![Go Report Card](https://goreportcard.com/badge/github.com/qwaykee/cauliflower)](https://goreportcard.com/report/github.com/qwaykee/cauliflower) [![GoDoc](https://godoc.org/github.com/qwaykee/cauliflower?status.svg)](https://godoc.org/github.com/qwaykee/cauliflower)

A simple telebot monkeypatcher

Functions: Listen for incoming messages

**Update:** Cauliflower now works with Middleware instead of bot.Handle() allowing to create handlers

## Quickstart

### Installation:

`go get github.com/qwaykee/cauliflower`

### Code:

```golang
package main

import (
	"gopkg.in/telebot.v3"
	"github.com/qwaykee/cauliflower"
)

func main() {
	b, _ := telebot.NewBot(telebot.Settings{ ... })

	i, _ := cauliflower.NewInstance(cauliflower.Settings{
		Bot: b,
		InstallMiddleware: true,
	})

	b.Handle("/echo", func (c telebot.Context) error {
		msg, answer, err := i.Listen(cauliflower.Parameters{
			Chat: c.Chat(),
			Message: "Please enter a text:",
		})
		if err == cauliflower.ErrTimeoutExceeded {
			return c.Send("You didn't type anything, please rerun the command :/")
		}

		return b.Edit(msg, answer.Text)
	})
}
```

## Detailed usage

### Create a new instance

What it does:
- This will initialize the channels and handlers that cauliflower needs in order to function.

Fields explanation:
- Timeout: A default timeout for all Listen functions that will be called without the Timeout field
- Cancel: A default cancel command for all Listen functions that will be called without the Cancel field
- Handlers: The type of messages you want to use with the Listen functions
- InstallMiddleware: Automatically install middleware instead of doing it manually

```golang
i, err := cauliflower.NewInstance(cauliflower.Settings{
	Bot: 				*telebot.Bot, 	// required
	Timeout: 			time.Duration, 	// optional, default: 1 * time.Minute
	Cancel: 			string, 		// optional
	Handlers: 			[]string, 		// optional, default: []string{telebot.OnText}
	InstallMiddleware: 	bool,			// optional, default: false
}) // will return *cauliflower.Instance, error
if err != nil {
	panic(err)
	// Possible error: ErrBotIsNil
}
```

### Listen for an incoming message

What it does:
- This will wait for an incoming message, any code after this will be blocked until a message has arrived or the timeout has exceeded

Fields explanation:
- Timeout: The maximum time to wait for the message
- Cancel: Text to cancel the Listen function
- Message: A message to send before listening

```golang
answer, err := i.Listen(cauliflower.Parameters{
	Chat:    *telebot.Chat, 	// required
	Timeout: time.Duration,		// optional, default: Instance.Settings.Timeout
	Cancel:  string, 			// optional
	Message: string, 			// optional, default: nil
}) // will return *telebot.Message, error
if err == cauliflower.ErrTimeoutExceeded {
	return c.Send("You didn't type anything, please rerun the command :/")
	// Possible error: ErrContextIsNil, ErrTimeoutExceeded, ErrCancelCommand, telebot error (bot.Send)
}
```

### Install middleware manually

What it does:
- Allows the instance to work properly, you can use Settings.InstallMiddleware: true, to do it automatically

```golang
b, _ := telebot.NewBot(telebot.Settings{ ... })

i, _ := cauliflower.NewInstance(cauliflower.Settings{ ... })

b.Use(i.Middleware())
```

## Troubleshooting

### Why can't I use bot.Handle()

You have to create the instance before using bot.Handle()

E.g:
```golang
b, _ := telebot.NewBot(telebot.Settings{ ... })

b.Handle(telebot.OnText, func(c telebot.Context) { ... }) // will NOT work

i, _ := cauliflower.NewInstance(cauliflower.Settings{ ... })

b.Handle(telebot.OnText, func(c telebot.Context) { ... }) // will work
```

### Listen() doesn't work

Make sure you've set the Handlers field when using cauliflower.NewInstance() to the types of message you want to listen to