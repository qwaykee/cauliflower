package cauliflower

import (
	"gopkg.in/telebot.v3"
	"time"
)

type (
	StepType string

	ValidationFunction func(f *Form, c telebot.Context, m *telebot.Message)
	FormFunction func(f *Form, c telebot.Context)

	Form struct {
		// the timeout will be applied to every step individually
		Timeout time.Duration
		TimeoutHandler telebot.HandlerFunc

		// delay after executing AddMessage() step
		// useful to avoid sending multiple message at once
		MessageDelay time.Duration
		Steps []Step
		CurrentStep int
		Answers map[string]*telebot.Message

		instance *Instance
		stopChannel chan bool
		continueChannel chan bool
		repeatChannel chan bool
		skipChannel chan int
	}

	Step struct {
		StepType StepType
		InputType InputType
		UniqueID string
		Message any
		Wait time.Duration
		Validate ValidationFunction
		Function FormFunction

		form *Form
	}
)

const (
	Input StepType = "input"
	Message StepType = "message"
	Wait StepType = "wait"
	Function StepType = "function"
)

// form creation components
func (i *Instance) NewForm(timeout, messageDelay time.Duration) *Form {
	return &Form{
		Timeout: timeout,
		MessageDelay: messageDelay,
		Answers: make(map[string]*telebot.Message),
		instance: i,
		stopChannel: make(chan bool),
		continueChannel: make(chan bool),
		repeatChannel: make(chan bool),
		skipChannel: make(chan int),
	}
}

func (f *Form) AddInput(inputType InputType, uniqueID string, validationFunc ValidationFunction) *Form {
	s := Step{
		StepType: Input,
		InputType: inputType,
		UniqueID: uniqueID,
		Validate: validationFunc,
		form: f,
	}

	f.Steps = append(f.Steps, s)

	return f
}

func (f *Form) AddMessage(message any) *Form {
	s := Step{
		StepType: Message,
		Message: message,
		form: f,
	}

	f.Steps = append(f.Steps, s)

	return f
}

func (f *Form) AddWait(duration time.Duration) *Form {
	s := Step{
		StepType: Wait,
		Wait: duration,
		form: f,
	}

	f.Steps = append(f.Steps, s)

	return f
}

func (f *Form) AddFunction(fu FormFunction) *Form {
	s := Step{
		StepType: Function,
		Function: fu,
		form: f,
	}

	f.Steps = append(f.Steps, s)

	return f
}

// form usage components
func (f *Form) Send(c telebot.Context) {
	for {
	    if f.CurrentStep >= len(f.Steps) {
	        break
	    }

	    f.Steps[f.CurrentStep].Execute(c)

	    select {
	    case <-f.stopChannel:
	        return
	    case <-f.continueChannel:
	        f.CurrentStep++
	    case <-f.repeatChannel:
	        // Don't increment the step, just re-execute the current step
	    case stepIndex := <-f.skipChannel:
	        f.CurrentStep = stepIndex
	    case <-time.After(f.Timeout):
	        return
	    }
	}
}

func (f *Form) GetAnswer(uniqueID string) *telebot.Message {
	if answer, ok := f.Answers[uniqueID]; ok {
		return answer
	}

	return &telebot.Message{}
}

// form actions components
func (f *Form) Next() {
	go func() {
		f.continueChannel <- true
	}()
}

func (f *Form) Repeat() {
	go func() {
		f.repeatChannel <- true
	}()
}

func (f *Form) Stop() {
	go func() {
		f.stopChannel <- true
	}()
}

func (f *Form) Skip(toStep int) {
	go func() {
		f.skipChannel <- toStep
	}()
}

func (s *Step) Execute(c telebot.Context) {
	switch s.StepType {
	case Input:
		answer, err := s.form.instance.Listen(c, s.InputType, &ListenOptions{ Timeout: s.form.Timeout, })

		if err == ErrTimeoutExceeded {
			s.form.TimeoutHandler(c)
		}

		s.Validate(s.form, c, answer) // s.Validate should set f.Next(), f.Repeat() or f.Stop()

		s.form.Answers[s.UniqueID] = answer
	case Message:
		c.Send(s.Message)
		time.Sleep(s.form.MessageDelay)
		s.form.Next()
	case Wait:
		time.Sleep(s.Wait)
		s.form.Next()
	case Function:
		s.Function(s.form, c)
		s.form.Next()
	}
}

func NoInputVerification(f *Form, c telebot.Context, m *telebot.Message) {
	f.Next()
}