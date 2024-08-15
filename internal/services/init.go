package service

var (
	Telegram iTelegramService
)

// Initialisation for all services
func Init() error {
	Telegram = initTelegram()

	return nil
}
