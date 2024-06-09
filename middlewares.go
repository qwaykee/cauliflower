package cauliflower

import (
	"gopkg.in/telebot.v3"
	"time"
	//"log"
)

func (i *Instance) Statistics() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			start := time.Now()

			err := next(c)

			timeTaken := time.Now().Sub(start)

			i.mutex.Lock()
			defer i.mutex.Unlock()
			
			i.responseTime = append(i.responseTime, timeTaken)
			i.messageCount++

			return err
		}
	}
}

func (i *Instance) GetResponseTime() []time.Duration {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.responseTime
}

func (i *Instance) GetAverageResponseTime() time.Duration {
	i.mutex.Lock()
	rt := i.responseTime
	i.mutex.Unlock()

	if len(rt) == 0 {
		return 0 * time.Second
	}

	var totalResponseTime time.Duration

	for _, responseTime := range rt {
		totalResponseTime += responseTime
	}

	return time.Duration(int64(totalResponseTime) / int64(len(rt)))
}

func (i *Instance) GetMessageCount() int {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.messageCount
}

func (i *Instance) Languages() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			i.mutex.Lock()

			if _, ok := i.usersLanguage[c.Chat().ID]; !ok {
				i.usersLanguage[c.Chat().ID] = c.Sender().LanguageCode
			}

			i.mutex.Unlock()

			return next(c)
		}
	}
}

func (i *Instance) GetChatLanguage(chatID int64) string {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	
	return i.usersLanguage[chatID]
}

func (i *Instance) SetChatLanguage(chatID int64, languageCode string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	
	i.usersLanguage[chatID] = languageCode
}

func (i *Instance) GetUserCount() int {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	
	return len(i.usersLanguage)
}