package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/frykher/disgo/library"
)

func main() {
	var socket *library.SocketConnection = &library.SocketConnection{}
	
	token := os.Getenv("DISCORD_TOKEN")
	socket.Open(token)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}


