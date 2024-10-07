package main

import (
	"context"
	"log"
	"net/http"
	"net/url"

	ollamaapi "github.com/ollama/ollama/api"
)

// Run this with go, will stream the LLM output on the respCh
func askLLMStreamed(question string, respCh chan string) {
	// We can use the OpenAI client because Ollama is compatible with OpenAI's API.
	// client, err := ollamaapi.ClientFromEnvironment()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	urlOllama, err := url.Parse(gConfig.OllamaUrl)
	if err != nil {
		panic(err)
	}
	// log.Println("After parsing url...")
	// log.Println(url1)

	client := ollamaapi.NewClient(urlOllama, http.DefaultClient)

	//close reply channel when we exit
	defer close(respCh)

	systemPrompt := gConfig.BotSummary
	// log.Println("System prompt:")
	// log.Println(systemPrompt)

	messages := []ollamaapi.Message{
		ollamaapi.Message{
			Role:    "system",
			Content: systemPrompt,
		},
		ollamaapi.Message{
			Role:    "user",
			Content: question,
		},
	}

	DBG("Asking LLM: " + question)

	ctx := context.Background()
	req := &ollamaapi.ChatRequest{
		Model:    gConfig.Model,
		Messages: messages,
	}
	var response string
	var startedTalking = false

	respFunc := func(resp ollamaapi.ChatResponse) error {
		//log.Print(resp.Message.Content)

		//ch_resp <- resp.Message.Content
		//log.Printf("LLM Reply: '%s'\n", resp.Message.Content)

		if resp.Message.Content != "" {
			respCh <- resp.Message.Content

			if !startedTalking {
				DBG("LLM started talking...")
				startedTalking = true
			}
			response += resp.Message.Content
		}

		return nil
	}

	err = client.Chat(ctx, req, respFunc)

	if err != nil {
		log.Fatal(err)
	}

	//wait for the last word from the llm..
	//log.Println("Waiting for callback to send done...")

	DBG("LLM Done")
	DBG(response)

}

// Run this with go, will stream the LLM output on the respCh
func askLLMStreamedContext(respCh chan string, chatId int64) {
	// We can use the OpenAI client because Ollama is compatible with OpenAI's API.
	// client, err := ollamaapi.ClientFromEnvironment()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	urlOllama, err := url.Parse(gConfig.OllamaUrl)
	if err != nil {
		panic(err)
	}
	// log.Println("After parsing url...")
	// log.Println(url1)

	client := ollamaapi.NewClient(urlOllama, http.DefaultClient)

	//close reply channel when we exit
	defer close(respCh)

	systemPrompt := gConfig.BotSummary
	// log.Println("System prompt:")
	// log.Println(systemPrompt)

	messages := []ollamaapi.Message{
		ollamaapi.Message{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	messages = append(messages, chatHistoryGet(chatId)...)

	//log.Println("Asking LLM: ", question)

	ctx := context.Background()
	req := &ollamaapi.ChatRequest{
		Model:    gConfig.Model,
		Messages: messages,
	}
	var response string
	var startedTalking = false

	respFunc := func(resp ollamaapi.ChatResponse) error {
		//log.Print(resp.Message.Content)

		//ch_resp <- resp.Message.Content
		//log.Printf("LLM Reply: '%s'\n", resp.Message.Content)

		if resp.Message.Content != "" {
			respCh <- resp.Message.Content

			if !startedTalking {
				DBG("LLM started talking...")
				startedTalking = true
			}
			response += resp.Message.Content
		}

		return nil
	}

	err = client.Chat(ctx, req, respFunc)

	if err != nil {
		log.Fatal(err)
	}

	//wait for the last word from the llm..
	//log.Println("Waiting for callback to send done...")

	DBG("LLM Done")
	DBG(response)

}

// Blocking call until we get all the answers from the LLM
// Use the same stream interface, just wait it out
func askLLM(question string) string {
	respCh := make(chan string)

	go askLLMStreamed(question, respCh)

	var rez string

	for v := range respCh {
		//fmt.Printf("Ch: %s\n", v)
		//fmt.Print(v)
		rez = rez + v
	}
	return rez
}

func askLLMContext(chatId int64) string {
	respCh := make(chan string)

	go askLLMStreamedContext(respCh, chatId)

	var rez string

	for v := range respCh {
		//fmt.Printf("Ch: %s\n", v)
		//fmt.Print(v)
		rez = rez + v
	}
	return rez
}
