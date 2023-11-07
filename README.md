# Cauliflower [![Go Report Card](https://goreportcard.com/badge/github.com/qwaykee/cauliflower)](https://goreportcard.com/report/github.com/qwaykee/cauliflower) [![GoDoc](https://godoc.org/github.com/qwaykee/cauliflower?status.svg)](https://godoc.org/github.com/qwaykee/cauliflower)

A simple [telebot](https://github.com/tucnak/telebot) monkeypatcher

Functions: Listen for incoming messages, create keyboard easily

## Example

##### Installation:

`go get github.com/qwaykee/cauliflower`

##### Code:

```golang
package main

import (
    "gopkg.in/telebot.v3"
    "github.com/qwaykee/cauliflower"
)

func main() {
    b, _ := telebot.NewBot(telebot.Settings{ ... })

    i, _ := cauliflower.NewInstance(&cauliflower.Settings{
        Bot: b,
        InstallMiddleware: true,
    })

    markup, _ := i.Keyboard(&cauliflower.KeyboardOptions{
        Keyboard: cauliflower.Inline,
        Row: []string{"abc", "def", "ghi", "jkl"},
        Split: 2,
        Handler: func (c telebot.Context) error {
            c.Respond()
            return c.Send(c.Callback().Data)
        },
    })

    b.Handle("/echo", func (c telebot.Context) error {
        msg, response, _ := i.Listen(&cauliflower.ListenOptions{
            Context: c,
            Message: "Enter message: ",
        })

        _, err := b.Edit(msg, response.Text, markup)

        return err
    })
}
```

## Detailed usage

### Create a new instance

What it does:
- This will initialize the channels and handlers that cauliflower needs in order to function.

Fields explanation:
- DefaultListen: Default parameters for Listen()
- DefaultKeyboard: Default parameters for Keyboard()
- Handlers: The type of messages you want to use with the Listen functions
- InstallMiddleware: Automatically install middleware instead of doing it manually

```golang
// will return *cauliflower.Instance, error
i, err := cauliflower.NewInstance(&cauliflower.Settings{
    Bot:                *telebot.Bot,                   // required
    DefaultListen       *cauliflower.ListenOptions      // optional
    DefaultKeyboard     *cauliflower.KeyboardOptions    // optional
    Handlers:           []string,                       // optional, default: []string{telebot.OnText}
    InstallMiddleware:  bool,                           // optional, default: false
})
if err != nil {
    panic(err) // possible error: ErrNoOptionsProvided, ErrBotIsNil
}
```

### Listen for an incoming message ([Source](listen.go))

What it does:
- This will wait for an incoming message, any code after this will be blocked until a message has arrived or the timeout has exceeded

Fields explanation:
- Timeout: The maximum time to wait for the message
- Cancel: Text to cancel the Listen function
- TimeoutHandler: Function to execute in case of timeout error
- CancelHandler: Function to execute in case of cancel error
- Message: A message to send before listening
- Edit: Edit the message instead of sending a new one

```golang
// will return *telebot.Message, *telebot.Message, error
msg, answer, err := i.Listen(&cauliflower.ListenOptions{
    Chat:           *telebot.Chat,                  // required
    Timeout:        time.Duration,                  // optional, default: Instance.DefaultListen.Timeout
    Cancel:         string,                         // optional, default: Instance.DefaultListen.Cancel
    TimeoutHandler: func(telebot.Context) error,    // optional, default: Instance.DefaultListen.TimeoutHandler
    CancelHandler:  func(telebot.Context) error,    // optional, default: Instance.DefaultListen.CancelHandler
    Message:        string,                         // optional, default: nil
    Edit:           telebot.Editable,               // optional, default: nil
})
if err == cauliflower.ErrTimeoutExceeded {
    return c.Send("You didn't type anything, please rerun the command :/")
    // possible error: ErrNoOptionsProvided, ErrContextIsNil, ErrTimeoutExceeded, ErrCancelCommand, telebot error (bot.Send)
}
```

### Create a keyboard ([Source](keyboard.go))

What it does:
- Will help you create a keyboard that will trigger the same function for each buttons

Fields explanation:
- ReplyMarkup: Use a custom markup instead of the default one
- Keyboard: Type of keyboard to create (cauliflower.Inline or cauliflower.Reply)
- Rows: Text of each button already splitted
- Row: Text of each button
- Split: Will split the buttons if Row is used
- Handler: Function to call for each button

```golang
markup, err := i.Keyboard(&cauliflower.KeyboardOptions{
    ReplyMarkup: *telebot.ReplyMarkup,             // optional, default: Instance.DefaultKeyboard.ReplyMarkup
    Keyboard:    string,                           // optional, default: Instance.DefaultKeyboard.Keyboard
    Rows:        [][]string,                       // required if Row isn't set
    Row:         []string,                         // required if Rows isn't set
    Split:       int,                              // optional
    Handler:     func(c telebot.Context) error,    // optional
})
if err != nil {
    panic(err)
    // possible error: ErrNoOptionsProvided, ErrRowsConflict, ErrNoRowsProvided, ErrInvalidKeyboard
}
```

### Install middleware manually

What it does:
- Allows the instance to work properly, you can use Settings.InstallMiddleware: true, to do it automatically

```golang
b, _ := telebot.NewBot(telebot.Settings{ ... })

i, _ := cauliflower.NewInstance(&cauliflower.Settings{ ... })

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