package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

var BLUEMIXUSER string
var BLUEMIXPASS string

type Client struct {
	conn     net.Conn
	ch       chan string
	nickname string
	language string
}

type RawMessage struct {
	language string
	msg      string
	nick     string
}

func check(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s", err.Error())
	}
}

func main() {
	address := flag.String("addr", "localhost:8080", "{host}:{port} to listen on")
	flag.StringVar(&BLUEMIXUSER, "user", "", "bluemix username")
	flag.StringVar(&BLUEMIXPASS, "pass", "", "bluemix password")
	flag.Parse()

	listener, err := net.Listen("tcp", *address)
	check(err)

	msgchan := make(chan RawMessage)
	languages := make(map[string]bool)
	addchan := make(chan Client)
	rmchan := make(chan Client)

	go printMessages(msgchan, addchan, rmchan, languages)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn, msgchan, addchan, rmchan, languages)
	}
}

func handleClient(conn net.Conn, msgchan chan RawMessage, addchan chan<- Client, rmchan chan<- Client, languages map[string]bool) {
	bufc := bufio.NewReader(conn)
	defer conn.Close()

	nick := promptNick(conn, bufc)
	lang := promptLang(conn, bufc)

	client := Client{
		conn:     conn,
		nickname: nick,
		language: lang,
		ch:       make(chan string),
	}
	if strings.TrimSpace(client.nickname) == "" {
		io.WriteString(conn, "Invalid username")
		return
	}
	if _, ok := languages[lang]; ok != true {
		languages[lang] = true
	}

	addchan <- client

	defer func() {
		msgchan <- RawMessage{
			language: "en",
			msg:      fmt.Sprintf("User %s has left the room.\n", client.nickname),
			nick:     "admin",
		}
		log.Printf("Connection from %v closed.\n", conn.RemoteAddr())
		rmchan <- client
	}()

	io.WriteString(conn, fmt.Sprintf("Welcome, %s!\n\n", client.nickname))
	msgchan <- RawMessage{
		language: "en",
		msg:      fmt.Sprintf("New user %s has joined the room.\n", client.nickname),
		nick:     "admin",
	}

	go client.Publish(msgchan)
	client.WriteLinesFrom(client.ch)
}

func printMessages(msgchan <-chan RawMessage, addchan <-chan Client, rmchan <-chan Client, languages map[string]bool) {
	clients := make(map[net.Conn]struct {
		ch       chan<- string
		language string
	})
	for {
		clientsByLanguage := make(map[string][]chan<- string)
		for _, client := range clients {
			clientsByLanguage[client.language] = append(clientsByLanguage[client.language], client.ch)
		}
		select {
		case msg := <-msgchan:
			for lang, _ := range languages {
				var translated string
				if lang == msg.language {
					translated = fmt.Sprintf("%s: %s", msg.nick, msg.msg)
				} else {
					tr := Translate(msg.language, lang, msg.msg, BLUEMIXUSER, BLUEMIXPASS)
					translated = fmt.Sprintf("%s: %s", msg.nick, tr)
				}
				for _, ch := range clientsByLanguage[lang] {
					go func(mch chan<- string) { mch <- "\033[1;33;40m" + translated + "\033[m\r\n" }(ch)
				}
			}
		case client := <-addchan:
			clients[client.conn] = struct {
				ch       chan<- string
				language string
			}{client.ch, client.language}
		case client := <-rmchan:
			delete(clients, client.conn)
		}
	}
}

func promptNick(conn net.Conn, bufc *bufio.Reader) string {
	io.WriteString(conn, "\033[1;30;41mWelcome to the multilingual chat!\033[0m\n")
	io.WriteString(conn, "What is your nick?")
	nick, _, _ := bufc.ReadLine()
	return string(nick)
}

func promptLang(conn net.Conn, bufc *bufio.Reader) string {
	io.WriteString(conn, "What language would you like to chat in?")
	lang, _, _ := bufc.ReadLine()
	return string(lang)
}

func (c Client) Publish(ch chan RawMessage) {
	bufc := bufio.NewReader(c.conn)
	for {
		line, err := bufc.ReadString('\n')
		if err != nil {
			break
		}
		ch <- RawMessage{language: c.language, nick: c.nickname, msg: strings.TrimSpace(string(line))}
	}
}

func (c Client) WriteLinesFrom(ch <-chan string) {
	for msg := range ch {
		_, err := io.WriteString(c.conn, msg)
		if err != nil {
			return
		}
	}
}
