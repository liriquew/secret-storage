package socketnotifier

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/liriquew/secret_storage/server/pkg/logger/sl"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Notifier struct {
	conns []*websocket.Conn
	m     sync.Mutex
	log   *slog.Logger
}

func New(log *slog.Logger) *Notifier {
	return &Notifier{
		conns: []*websocket.Conn{},
		m:     sync.Mutex{},
		log:   log,
	}
}

func (n *Notifier) AddConn(c *gin.Context) error {
	n.m.Lock()
	defer n.m.Unlock()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return err
	}

	n.conns = append(n.conns, conn)

	return nil
}

type messagesMaker func(int) ([][]byte, error)

func (n *Notifier) Notify(msgsMaker messagesMaker) error {
	n.m.Lock()
	defer n.m.Unlock()
	messages, err := msgsMaker(len(n.conns))
	if err != nil {
		return err
	}

	for i, msg := range messages {
		err := n.conns[i].WriteMessage(1, msg)
		if err != nil {
			n.log.Error("error while notify conn", sl.Err(err))
			continue
		}
		n.log.Info("notify message", slog.String("message", string(msg)))
		n.conns[i].Close()
	}

	n.conns = nil

	return nil
}
