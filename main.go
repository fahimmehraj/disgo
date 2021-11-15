package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/frykher/disgo/library"
)

var client *http.Client = &http.Client{}

func onMessage(message *library.Message, reply *library.PostMessage, token string) {
	fmt.Println("Message:", message.Content)
	reply.Content = fmt.Sprintf("\"%s\" 🤓☝️", message.Content)
	reply.Message_reference.Message_id = string(message.ID)
	reply.Message_reference.Channel_id = string(message.ChannelID)
	if message.GuildID != nil {
		reply.Message_reference.Guild_id = string(*message.GuildID)
	}

	if client == nil {
		client = &http.Client{}
	}
	body, err := json.Marshal(reply)
	if err != nil {
		panic(err)
	}
	
	req, err := http.NewRequest("POST",
		"https://discord.com/api/v9/channels/"+string(message.ChannelID)+"/messages", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}
}

func main() {
	var socket *library.SocketConnection = &library.SocketConnection{}
	
	token := os.Getenv("DISCORD_TOKEN")
	socket.Open(token)
	socket.AddHandler("MESSAGE_CREATE", onMessage)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}


