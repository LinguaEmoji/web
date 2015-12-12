package main

import (
    "github.com/go-martini/martini"
    "github.com/gorilla/websocket"
    "github.com/martini-contrib/render"

    "os"
    "net"
    "time"
    "sync"
    "net/http"
    "math/rand"
    "encoding/json"
)

type Config struct {
    Phrases []string `json:"phrases"`
}

func ParseConfig() {
    cfg, _ := os.Open("config.json")
    decoder := json.NewDecoder(cfg)
    decoder.Decode(&config)
}

type Game struct {
    stage int
    answer string

    players []ClientConn
}

func (game *Game) NewWord() string {
    random := rand.Intn((len(config.Phrases) - 1) - 0) + 0
    game.answer = config.Phrases[random]
    return game.answer
}

func (game Game) Opponent(sockCli ClientConn) ClientConn {
    for i := 0; i != 2; i++ {
        if game.players[i].name != sockCli.name {
            return game.players[i]
        }
    }
    return ClientConn{}
}

type ClientConn struct {
    name string
    points int

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

func NewFoundGamePacket(match string) Packet {
    return Packet {
        Action: "found_game",
        Payload: map[string]interface{} {
            "match": match,
        },
    }
}

func NewTurnPacket(payload map[string]interface{}) Packet {
    return Packet {
        Action: "turn",
        Payload: payload,
    }
}

func NewAnswerPacket(b bool, answer string, owner string) Packet {
    return Packet {
        Action: "answer",
        Payload: map[string]interface{} {
            "boolean": b,
            "answer": answer,
        },
    }
}

var (
    config Config

    games map[ClientConn]*Game
    queue []ClientConn

    RWMutex sync.RWMutex
)

func main() {
    games = map[ClientConn]*Game {}
    ParseConfig()
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
        game := &Game {
            stage: 0,
            players: []ClientConn {queue[0], queue[1]},
        }
        games[queue[0]] = game
        games[queue[1]] = game

        queue[0].websocket.WriteMessage(1, NewFoundGamePacket(queue[1].name).toJson())
        queue[0].websocket.WriteMessage(1, NewTurnPacket(map[string]interface{} {
            "turn": "your",
            "state": "give_clue",
            "word": game.NewWord(),
        }).toJson())
        queue[1].websocket.WriteMessage(1, NewFoundGamePacket(queue[0].name).toJson())
        queue[1].websocket.WriteMessage(1, NewTurnPacket(map[string]interface{} {
            "turn": "their",
            "state": "waiting_for_clue",
        }).toJson())

        queue = queue[2:]
    }

    RWMutex.Unlock()
}

func removeFromQueue(sockCli ClientConn) {
    RWMutex.Lock()
    index := getIndexOf(queue, sockCli)
    if index != -1 {
        queue = append(queue[:index], queue[index + 1:]...)
    }
    RWMutex.Unlock()
}

func endGame(sockCli ClientConn) {
    RWMutex.Lock()
    if games[sockCli] != nil {
        opponent := games[sockCli].Opponent(sockCli)
        delete(games, sockCli)
        delete(games, opponent)
    }
    RWMutex.Unlock()
}

func websocketConn(r *http.Request, w http.ResponseWriter, ren render.Render) {
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if _, ok := err.(websocket.HandshakeError); ok || err != nil {
        return
    }
    _, raw, _ := ws.ReadMessage()

    client := ws.RemoteAddr()
    sockCli := ClientConn{string(raw), 0, ws, client}

    addToQueue(sockCli)

    for {
        _, raw, err := ws.ReadMessage()
        if err != nil {
            removeFromQueue(sockCli)
            endGame(sockCli)
            return
        }

        var packet Packet
        json.Unmarshal(raw, &packet)

        switch packet.Action {
            case "submit_clue":
                sockCli.websocket.WriteMessage(1, NewTurnPacket(map[string]interface{} {
                    "turn": "their",
                    "state": "waiting_for_answers",
                }).toJson())
                games[sockCli].Opponent(sockCli).websocket.WriteMessage(1, NewTurnPacket(map[string]interface{} {
                    "turn": "your",
                    "state": "give_answer",
                    "clue": packet.Payload["clue"].(string),
                }).toJson())
                break
            case "submit_answer":
                sockCli.websocket.WriteMessage(1, NewAnswerPacket(packet.Payload["answer"].(string) != games[sockCli].answer,  packet.Payload["answer"].(string), "your").toJson())
                games[sockCli].Opponent(sockCli).websocket.WriteMessage(1, NewAnswerPacket(packet.Payload["answer"].(string) != games[sockCli].answer,  packet.Payload["answer"].(string), "their").toJson())
                time.Sleep(time.Second * 2)
                sockCli.websocket.WriteMessage(1, NewTurnPacket(map[string]interface{} {
                    "turn": "your",
                    "state": "give_clue",
                    "word": games[sockCli].NewWord(),
                }).toJson())
                sockCli.websocket.WriteMessage(1, NewTurnPacket(map[string]interface{} {
                    "turn": "their",
                    "state": "waiting_for_clue",
                }).toJson())
                break
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
