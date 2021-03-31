package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	//存放匹配到的对手
	match sync.Map

	//等待中的人
	waiting []*websocket.Conn
	mu      sync.RWMutex

	//html
	load  sync.Once
	chess []byte
	jq    []byte
	js    []byte
	rule  []byte
)

func handler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		WriteBufferSize: 1024,
		ReadBufferSize:  1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	msg := &Result{}
	for {

		err = conn.ReadJSON(msg)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("%s", msg.Message)

		//如果没有对手，也不在等待队列里，就加入等待队列，等待匹配
		//如果有对手，就检验操作合法性，并向双方发送信息
		if opponent, ok := match.Load(conn); !ok {
			mu.Lock()
			for _, v := range waiting {
				if v == conn {
					break
				}
			}
			waiting = append(waiting, conn)
			mu.Unlock()
			msg.Message = "尚未匹配到对手，等待中！"
			msg.Bout = false
			conn.WriteJSON(msg)
		} else {
			if msg.Xy != "" && msg.Message != "" {
				conn.WriteJSON(msg)
				msg.Message = "到你了"
				msg.Bout = true
				opponent.(*websocket.Conn).WriteJSON(msg)
			} else {
				rawMsg := msg.Message
				msg.Message = fmt.Sprintf("我说：%s", rawMsg)
				conn.WriteJSON(msg)
				msg.Message = fmt.Sprintf("对面说：%s", rawMsg)
				opponent.(*websocket.Conn).WriteJSON(msg)
			}

		}

	}

}

func makeMatch() {
	t := time.NewTicker(2 * time.Second)
	for range t.C {
		mu.Lock()
		for i := range waiting {
			if i%2 == 0 && len(waiting) > i+1 {
				match.Store(waiting[i], waiting[i+1])
				match.Store(waiting[i+1], waiting[i])
				msg := &Result{Message: "匹配成功，你先下", Bout: true, Color: "black"}
				msg2 := &Result{Message: "匹配成功，对手先下", Bout: false, Color: "white"}
				waiting[i].WriteJSON(msg)
				waiting[i+1].WriteJSON(msg2)
				log.Printf("matched")
			}
		}
		if len(waiting)%2 == 0 {
			waiting = waiting[:0]
		} else {
			waiting[0] = waiting[len(waiting)-1]
			waiting = waiting[:1]
		}
		mu.Unlock()
	}
}

func serveHtml() {
	fs := http.FileServer(http.Dir("./js"))
	http.Handle("/js/", http.StripPrefix("/js/", fs))
	http.ListenAndServe("127.0.0.1:9998", nil)
}

func loadFile() {
	var err error
	chess, err = os.ReadFile("./html/chess.html")
	if err != nil {
		panic(err)
	}
	jq, err = os.ReadFile("./js/jquery-1.12.3.min.js")
	if err != nil {
		panic(err)
	}
	js, err = os.ReadFile("./js/json2.js")
	if err != nil {
		panic(err)
	}
	rule, err = os.ReadFile("./js/rule.js")
	if err != nil {
		panic(err)
	}
}

func main() {
	loadFile()
	//go serveHtml()
	go makeMatch()
	go func() {
		fs := http.FileServer(http.Dir("./js"))
		http.Handle("/js/", http.StripPrefix("/js/", fs))
		http.Handle("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write(chess)
		}))
		http.ListenAndServe(":9998", nil)
	}()
	http.ListenAndServe(":9999", http.HandlerFunc(handler))
}
