package main

import (
	cl "github.com/qwaykee/cauliflower"
	"gopkg.in/telebot.v3"

	"log"
	"time"
	"strconv"
)

var (
	token 	= "SET YOUR TOKEN HERE"
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

	b.Handle("/start", startHandler)
	b.Handle("/echo", echoHandler)
	b.Handle("/form", formHandler)

	log.Println("starting...")

	b.Start()
}

func startHandler(c telebot.Context) error {
	menu := i.NewKeyboard(cl.KeyboardInline, 3)
	
	menu.Add(cl.ButtonData, echoHandler, "Echo")
	menu.Add(cl.ButtonData, formHandler, "Form")

	return c.Send("Type /echo to try the listen and keyboard components\nType /form to try the form component", menu.Convert())
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

func formHandler(c telebot.Context) error {
	f := i.NewForm(10 * time.Second, 500 * time.Millisecond).
		AddMessage("Please enter a number, type 5 if you wish to restart the form when it achieves its end.").
		AddInput(cl.TextInput, "unique-id", verifyIsNumber).
		AddMessage("Thanks!").
		AddFunction(logFormFunction).
		AddMessage("I just called a function which logged something, look at your terminal.")

	f.Send(c)

	return c.Send("Your entry is " + f.GetAnswer("unique-id").Text)
}

func verifyIsNumber(f *cl.Form, c telebot.Context, m *telebot.Message) {
	number, err := strconv.Atoi(m.Text)

	if err != nil {
		c.Send("You didn't send a number, please retry.")
		f.Repeat()
		return // needed after f.Repeat() call to avoid errors
	}

	c.Send(number)

	f.Next()
}

func logFormFunction(f *cl.Form, c telebot.Context) {
	log.Println("You could do anything here instead of just a log call, for example, restart the whole form...")
	
	number, _ := strconv.Atoi(f.GetAnswer("unique-id").Text)

	if number == 5 {
		log.Println("...And it looks like the form will be restarted, look at your telegram client.")
		f.Skip(0)
	}
}