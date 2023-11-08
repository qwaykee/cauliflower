# Todo

- [x] handle defaults in main.go
- [x] doc readme.md
- [x] doc comments listen.go
- [ ] add dictionnary keyboard

# Defaults

## Listen

Context none -> required
Timeout 1 * time.Minute -> optional
TimeoutHandler none -> optional
CancelHandler none -> optional
Cancel none -> optional
Message none -> optional
Edit none -> optional

## Keyboard

ReplyMarkup telebot.ReplyMarkup{} -> optional
Keyboard cauliflower.Inline -> optional
Rows none -> required
Row none -> required
Split none -> required
Handler none -> optional