package main

import (
    "github.com/go-martini/martini"
    "github.com/gorilla/websocket"
    "github.com/martini-contrib/render"

    "net"
    "sync"
    "net/http"
    "encoding/json"
    "log"
)

type Game struct {
    stage int

    players []ClientConn
}

type ClientConn struct {
    name string

    websocket *websocket.Conn
    clientIP net.Addr
}

type Packet struct {
    Action string
    Payload map[string]interface{}
}

func (p Packet) toJson() []byte {
    encode, _ := json.Marshal(p)
    return encode
}

var (
    games map[ClientConn]Game
    queue []ClientConn

    RWMutex sync.RWMutex
)

func main() {
    games = map[ClientConn]Game {}
    m := martini.Classic()

    m.Use(martini.Static("assets", martini.StaticOptions{
        Prefix: "assets",
    }))
    m.Use(render.Renderer(render.Options {
        Directory: "templates",
        Layout: "layout",
    }))
    handlers(m)
    m.Run()
}

func handlers(m *martini.ClassicMartini) {
    m.Get("/", homePage)

    m.Get("/websocket", websocketConn)
}

func homePage(ren render.Render) {
    ren.HTML(200, "home", nil)
}

func addToQueue(sockCli ClientConn) {
    RWMutex.Lock()
    queue = append(queue, sockCli)

    if len(queue) >= 2 {
        game := Game {
            stage: 0,
            players: []ClientConn {queue[0], queue[1]},
        }
        games[queue[0]] = game
        games[queue[1]] = game

        for i := 0; i != 2; i++ {
            queue[i].websocket.WriteMessage(1, Packet {
                Action: "Found Game",
                Payload: map[string]interface{} {},
            }.toJson())
        }

        queue = queue[2:]
    }

    RWMutex.Unlock()
}

func removeFromQueue(sockCli ClientConn) {
    RWMutex.Lock()
    index := getIndexOf(queue, sockCli)
    queue = append(queue[:index], queue[index + 1:]...)
    RWMutex.Unlock()
}

func websocketConn(r *http.Request, w http.ResponseWriter, ren render.Render) {
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if _, ok := err.(websocket.HandshakeError); ok || err != nil {
        return
    }
    _, raw, _ := ws.ReadMessage()
    log.Print("message: ", string(raw))

    client := ws.RemoteAddr()
    sockCli := ClientConn{string(raw), ws, client}

    addToQueue(sockCli)

    for {
        _, _, err := ws.ReadMessage()
        if err != nil {
            removeFromQueue(sockCli)
            return
        }
    }
}

func getIndexOf(conns []ClientConn, conn ClientConn) int {
    for i, c := range conns {
        if c.name == conn.name {
            return i
        }
    }
    return -1
}
