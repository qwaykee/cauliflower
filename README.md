# Cauliflower [![Go Report Card](https://goreportcard.com/badge/github.com/qwaykee/cauliflower)](https://goreportcard.com/report/github.com/qwaykee/cauliflower) [![GoDoc](https://godoc.org/github.com/qwaykee/cauliflower?status.svg)](https://godoc.org/github.com/qwaykee/cauliflower)

A simple [telebot](https://github.com/tucnak/telebot) monkeypatcher

Functions: Listen for incoming messages, create keyboard easily, create forms

## Example

##### Installation:

`go get github.com/qwaykee/cauliflower`

##### [Code:](example/echo.go)

```golang
package main

import (
    "gopkg.in/telebot.v3"
    cl "github.com/qwaykee/cauliflower"
)

func main() {
    b, _ := telebot.NewBot(telebot.Settings{ ... })

    i := cauliflower.NewInstance(b, &cauliflower.Settings{
        InstallMiddleware: true,
    })

    b.Handle("/echo", echoHandler)
}

func echoHandler(c telebot.Context) error {
    c.Send("Send a text that will be put into a button...")

    answer, err := i.Listen(c, cl.TextInput, &cl.ListenOptions{})
    if err != nil {
        return c.Send(err)
    }

    menu := i.NewKeyboard(cl.KeyboardInline, 2)
    
    menu.Add(cl.ButtonData, echoHandler, "Retry")
    menu.Add(cl.ButtonURL, cl.NoHandler, "Open Google", "https://google.com")
    menu.Add(cl.ButtonText, cl.NoHandler, answer.Text)

    return c.Send("Here is the requested message with a custom keyboard", menu.Convert())
}
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