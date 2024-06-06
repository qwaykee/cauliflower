package main

import (
	cl "github.com/qwaykee/cauliflower"
	"gopkg.in/telebot.v3"

	"log"
	"time"
)

var (
	token 	= "SET YOUR BOT TOKEN HERE"
	b 		*telebot.Bot
	i 		*cl.Instance
)

func main() {
	var err error

	b, err = telebot.NewBot(telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	i = cl.NewInstance(b, &cl.Settings{
		InstallMiddleware: true,
	})

	b.Handle("/echo", echoHandler)

	log.Println("starting...")

	b.Start()
}

func echoHandler(c telebot.Context) error {
	c.Send("listening")

	msg, err := i.Listen(c, &cl.ListenOptions{})
	if err != nil {
		return c.Send(err)
	}

	menu := i.NewKeyboard(cl.KeyboardInline, 3)
	menu.Add(cl.ButtonData, echoHandler, "retry")

	return c.Send(msg.Text, menu.Convert())
}