package main

import (
	"log"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/xumxum/telegram-bot-api"
)

var gBot *tgbotapi.BotAPI
var gBotMutex sync.Mutex

const telegramWorkersCount = 4

var gTelegramActiveChats map[int64]bool
var gTelegramActiveChatsMutex sync.Mutex

// 2 functions to see if we are usign the llm to block same telegram user asking multiple questions at the same time
// it will f u context...
func telegramChatIsActive(chatId int64) (rez bool) {
	gTelegramActiveChatsMutex.Lock()
	rez = gTelegramActiveChats[chatId]
	gTelegramActiveChatsMutex.Unlock()
	return
}

func telegramChatSetActive(chatId int64, active bool) {
	//log.Println("Setting ", chatId, " -> ", active)
	gTelegramActiveChatsMutex.Lock()
	gTelegramActiveChats[chatId] = active
	gTelegramActiveChatsMutex.Unlock()
	return
}

func runTelegramBot() {
	if gConfig.TelegramToken == "" {
		log.Fatal("You must specify a valig telegram Token in config file")
	}

	gTelegramActiveChats = make(map[int64]bool)

	bot, err := tgbotapi.NewBotAPI(gConfig.TelegramToken)
	gBot = bot
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true

	INF("Running bot on telegram account '" + bot.Self.UserName + "'")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	tgUpdatesChannel := bot.GetUpdatesChan(u)

	workerChan := make(chan tgbotapi.Update)

	//create telegram worker threads, so we limit the number of parralel requests..
	for i := 0; i < telegramWorkersCount; i++ {
		go telegramWorker(workerChan)
	}

	for update := range tgUpdatesChannel {

		//chatId := update.Message.From.ID
		//litter.Dump(update)

		if update.Message != nil { // If we got a message
			//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			DBG("[" + update.Message.From.UserName + "] " + update.Message.Text)

			workerChan <- update
		}
	}
}

func telegramWorker(workerChan chan tgbotapi.Update) {
	for update := range workerChan {
		chatId := update.Message.From.ID
		message := update.Message.Text
		var waitingPrinted bool = false

		for {
			if !telegramChatIsActive(chatId) {
				//we are clear
				telegramChatSetActive(chatId, true)
				break
			} else {
				time.Sleep(250 * time.Millisecond)
				if !waitingPrinted {
					log.Printf("%d Telegram user waiting because user already talking with llm", chatId)
					waitingPrinted = true
				}
			}

		}
		//log.Println("Wait is over, time work", chatId)

		//Command or message to llm
		var reply string

		if strings.HasPrefix(update.Message.Text, "/") {
			reply = processCommands(message, int(chatId))
		} else {

			chatHistoryAdd(chatId, "user", message)
			respCh := make(chan string)
			go askLLMStreamedContext(respCh, chatId)

			msg0 := tgbotapi.NewChatAction(chatId, "typing")
			gBot.Send(msg0)

			timec := time.After(5 * time.Second)

		RangeLoop:
			for {
				select {
				case <-timec:
					//timeout, resend the typing shit
					msg0 := tgbotapi.NewChatAction(chatId, "typing")
					gBot.Send(msg0)
					timec = time.After(5 * time.Second)

				case out, ok := <-respCh:
					if !ok {
						break RangeLoop // Channel closed
					}
					reply += out
				}
			}

			for w := range respCh {
				//fmt.Printf("Ch: %s\n", v)
				//fmt.Print(v)
				reply += w
			}

		}

		chatHistoryAdd(chatId, "assistant", reply)

		//2. Send back reply to telegram user
		msg1 := tgbotapi.NewMessage(chatId, reply)
		msg1.ReplyParameters.MessageID = update.Message.MessageID
		gBotMutex.Lock()
		gBot.Send(msg1)
		gBotMutex.Unlock()

		telegramChatSetActive(chatId, false)

	}
}

// func processCommands(cmd string, pChatInfo *Chatformation) {
func processCommands(cmd string, chatId int) (reply string) {
	//commands, process here , do not send to the AI

	_ = chatId

	//	/status - Send back a status information
	helpMessage := `Hello. My name is Tolby. I am here to help you.`
	/*
		Tolby commands:
		  /help - Print this help information
		  /clear  - Clear the Chat history, forget all previous conversation.
	*/
	if strings.HasPrefix(cmd, "/model ") {
		model := strings.ReplaceAll(cmd, "/model ", "")
		gConfig.Model = model
		reply = "LLM Model changed to : " + model
		return
	}

	switch cmd {

	case "/help", "/start":
		reply = helpMessage
	case "/clear":
		chatHistoryClear(int64(chatId))
		reply = "Chat history cleared"
	default:
		reply = "Unknown command. Try /help"
	}

	return
}
