package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	ollamaapi "github.com/ollama/ollama/api"
)

// struct to hold all chat information, history, context etc
type Chatformation struct {
	ChatId               int64
	Name                 string
	UserName             string
	Messages             []ollamaapi.Message
	LastMessageTimestamp int64
}

var gChatInformation map[int64]*Chatformation
var gChatInformationMutex sync.Mutex

func initChatHistory() {
	INF("Initializing Tolby...")
	gChatInformation = make(map[int64]*Chatformation)
}

func logChatHistory(history []ollamaapi.Message) string {
	var rez string
	//rez = fmt.Sprintf("%s[ %d | %s]\n", this.Name, this.ChatId, this.UserName)
	var role string
	for _, m := range history {
		switch m.Role {
		case "user":
			role = "U-->"
		case "assistant":
			role = "<--B"
		case "system":
			role = "===="
		default:
			role = "?"
		}

		rez += fmt.Sprintf("%s: %s\n", role, m.Content)
	}
	return rez
}

func chatHistoryGet(chatId int64) []ollamaapi.Message {
	gChatInformationMutex.Lock()
	defer gChatInformationMutex.Unlock()

	if chatInfo, ok := gChatInformation[chatId]; ok {
		DBG(logChatHistory(chatInfo.Messages))
		return chatInfo.Messages

	}

	return []ollamaapi.Message{}
}

func chatHistoryAdd(chatId int64, role string, message string) {
	gChatInformationMutex.Lock()
	defer gChatInformationMutex.Unlock()

	var pChatInfo *Chatformation

	timeNow := time.Now().Unix()

	if chatInfo, ok := gChatInformation[chatId]; ok {
		//we have it in the map
		pChatInfo = chatInfo

		if (timeNow - chatInfo.LastMessageTimestamp) > int64(gConfig.HistoryRetainMinutes*60) {
			//more time has passed then history retain..forget everything..clean slate..do a /clean basically
			DBG(strconv.FormatInt(chatId, 10) + "Clearing message history since HistoryRetainSeconds expired")
			clear(pChatInfo.Messages)
		}
	} else {
		//gChatInformation[chatId] = new(Chatformation{ChatId: chatId})
		pChatInfo = new(Chatformation)
		pChatInfo.ChatId = chatId
		gChatInformation[chatId] = pChatInfo

	}

	pChatInfo.LastMessageTimestamp = timeNow

	pChatInfo.Messages = append(pChatInfo.Messages, ollamaapi.Message{Role: role, Content: message})

}

func chatHistoryClear(chatId int64) {
	gChatInformationMutex.Lock()
	defer gChatInformationMutex.Unlock()

	if chatInfo, ok := gChatInformation[chatId]; ok {
		clear(chatInfo.Messages)
	}
}
