package library

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type OPCODE int

const (
	GATEWAY string = "wss://gateway.discord.gg/?v=9&encoding=json"
	API     string = "https://discord.com/api/v9"
)

// Only used for receiving, sending is handled
// by the other structures
type Payload struct {
	Op int             `json:"op"`
	D  json.RawMessage `json:"d"`
	S  int64           `json:"s"`
	T  string          `json:"t"`
}

type Hello struct {
	Op   int `json:"op"`
	Data struct {
		Heartbeat_interval time.Duration `json:"heartbeat_interval"`
	} `json:"d"`
}

type Heartbeat struct {
	Op   int    `json:"op"`
	Data *int64 `json:"d"`
}

type Identify struct {
	Op   int `json:"op"`
	Data struct {
		Token      string `json:"token"`
		Properties struct {
			Os      string `json:"os"`
			Browser string `json:"browser"`
			Device  string `json:"device"`
		} `json:"properties"`
		Intents int `json:"intents"`
	} `json:"d"`
}

type PostMessage struct {
	Content           string `json:"content"`
	Message_reference struct {
		Channel_id string `json:"channel_id"`
		Message_id string `json:"message_id"`
		Guild_id   string `json:"guild_id"`
	} `json:"message_reference"`
}

type SocketConnection struct {
	interval  time.Duration // time.Duration is the same thing as int64 (1 nanosecond)
	conn      *websocket.Conn
	sequences chan *int64
	writes    chan interface{}
	token     string
	handlers  map[string]func(*Message, *PostMessage, string)
	client    *http.Client
}

func (sConn *SocketConnection) Open(t string) {
	sConn.token = t
	sConn.handlers = make(map[string]func(*Message, *PostMessage, string))

	sConn.sequences = make(chan *int64)
	sConn.writes = make(chan interface{})
	go sConn.manageSequence()
	go sConn.manageWrites()

	var err error

	sConn.conn, _, err = websocket.DefaultDialer.Dial(GATEWAY, nil)
	if err != nil {
		panic(err)
	}
	mType, msg, err := sConn.conn.ReadMessage()
	if err != nil {
		panic(err)
	}

	payload := sConn.onReceiving(mType, msg)
	fmt.Println(payload.Op)
	if payload.Op != 10 {
		panic("what")
	}

	hello := new(Hello)
	hello.Op = 10
	if err = json.Unmarshal(payload.D, &hello.Data); err != nil {
		panic(err)
	}
	fmt.Println(hello.Data.Heartbeat_interval)

	sConn.interval = hello.Data.Heartbeat_interval
	go sConn.heartbeat()
	go sConn.listen()

	// next we send the identify packet
	identifyPacket := new(Identify)
	identifyPacket.Op = 2
	identifyPacket.Data.Token = sConn.token
	identifyPacket.Data.Intents = 1<<9 | 1<<0
	identifyPacket.Data.Properties.Os = "windows"
	identifyPacket.Data.Properties.Browser = "discord.go"
	identifyPacket.Data.Properties.Device = "discord.go"

	sConn.writes <- identifyPacket

}

func (sConn *SocketConnection) decompress(reader *io.Reader, msg []byte) error {
	*reader = bytes.NewBuffer(msg)

	zReader, err := zlib.NewReader(*reader)
	if err != nil {
		panic(err)
	}

	defer zReader.Close()

	*reader = zReader
	return err
}

func (sConn *SocketConnection) onReceiving(mType int, msg []byte) *Payload {
	var reader io.Reader
	var err error
	reader = bytes.NewBuffer(msg)

	if mType == websocket.BinaryMessage {
		err = sConn.decompress(&reader, msg)

		if err != nil {
			panic(err)
		}
	}

	payload := new(Payload)
	decoder := json.NewDecoder(reader)
	if err = decoder.Decode(&payload); err != nil {
		panic(err)
	}

	switch payload.Op {
	case 10:
		return payload
	case 1:
		// Heartbeat (FROM SERVER)
		heartbeat := new(Heartbeat)
		heartbeat.Op = 1
		heartbeat.Data = sConn.getSequence()

		sConn.writes <- heartbeat
		return payload
	case 0:
		sConn.setSequence(&payload.S)
		eventName := payload.T
		fmt.Println("The sequence is", payload.S)
		// Very manual event handling
		switch eventName {
		case "READY":
			ready := new(Ready)
			if err := json.Unmarshal(payload.D, &ready); err != nil {
				panic(err)
			}
			fmt.Println("we are ready")
		case "MESSAGE_CREATE":
			message := new(Message)
			if err := json.Unmarshal(payload.D, &message); err != nil {
				panic(err)
			}
			if message.Author.Bot {
				return payload
			}
			for _, v := range sConn.handlers {
				v(message, &PostMessage{}, sConn.token)
			}
			// reply := &PostMessage{}
			// reply.Content = fmt.Sprintf("\"%s\" 🤓☝️", message.Content)
			// reply.Message_reference.Message_id = string(message.ID)
			// reply.Message_reference.Channel_id = string(message.ChannelID)
			// if message.GuildID != nil {
			// 	reply.Message_reference.Guild_id = string(*message.GuildID)
			// }
			// if sConn.client == nil {
			// 	sConn.client = &http.Client{}
			// }
			// body, err := json.Marshal(reply)
			// if err != nil {
			// 	panic(err)
			// }
			// req, err := http.NewRequest("POST",
			// 	"https://discord.com/api/v9/channels/"+string(message.ChannelID)+"/messages", bytes.NewBuffer(body))
			// if err != nil {
			// 	panic(err)
			// }
			// req.Header.Set("Authorization", "Bot "+sConn.token)
			// req.Header.Set("Content-Type", "application/json")

			// _, err = sConn.client.Do(req)
			if err != nil {
				panic(err)
			}

			fmt.Println("Message created")
		}
	}
	return payload
}

func (sConn *SocketConnection) manageSequence() {
	var sequence *int64 = nil
	for {
		select {
		case seq := <-sConn.sequences:
			sequence = seq
		case sConn.sequences <- sequence:
		}
	}
}

func (sConn *SocketConnection) manageWrites() {
	for {
		msg := <-sConn.writes
		if err := sConn.conn.WriteJSON(msg); err != nil {
			panic(err)
		}
	}
}

func (sConn *SocketConnection) setSequence(sequence *int64) {
	sConn.sequences <- sequence
}

func (sConn *SocketConnection) getSequence() *int64 {
	return <-sConn.sequences
}

func (sConn *SocketConnection) listen() {
	for {
		mType, msg, err := sConn.conn.ReadMessage()
		if err != nil {
			panic(err)
		}
		fmt.Println("Got message from server! - ")
		sConn.onReceiving(mType, msg)
	}
}

func (sConn *SocketConnection) heartbeat() {
	ticker := time.NewTicker(sConn.interval * time.Millisecond)

	for {
		sConn.writes <- Heartbeat{Op: 1, Data: sConn.getSequence()}
		<-ticker.C

	}
}

func (sConn *SocketConnection) AddHandler(name string, f func(*Message, *PostMessage, string)) {
	sConn.handlers[name] = f
}

func (sConn *SocketConnection) Close() {
	sConn.conn.Close()
}
