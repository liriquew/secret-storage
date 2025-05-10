package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/liriquew/secret_storage/server/tests/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const StatusOK = "200 OK"

var isUnsealed bool

func TestMain(m *testing.M) {
	t := &testing.T{}

	UnsealWithMaster(t)

	if t.Failed() {
		os.Exit(1)
	}

	isUnsealed = true

	os.Exit(m.Run())
}

func UnsealWithMaster(t *testing.T) {
	ts := suite.New(t)

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/ready", ts.GetURL()), nil)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	if resp.Status == StatusOK {
		return
	}

	wsURL := strings.Replace(
		fmt.Sprintf("%s/master?threshold=%d", ts.GetURL(), 3),
		"http://", "ws://", 1,
	)

	conns := make([]*websocket.Conn, 0)
	for range 5 {
		headers := http.Header{}

		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
		if err != nil {
			t.Fatalf("WebSocket connection error: %v", err)
		}
		defer conn.Close()

		if resp.StatusCode != http.StatusSwitchingProtocols {
			t.Errorf("Expected status code %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
		}

		conns = append(conns, conn)
	}

	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/master/complete", ts.GetURL()), nil)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var parts []string

	for _, conn := range conns {
		var part struct {
			Part string `json:"part"`
		}
		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		fmt.Println(string(message))
		err = json.Unmarshal(message, &part)
		require.NoError(t, err)

		parts = append(parts, part.Part)
	}

	for i := range 3 {
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/unseal?part=%s", ts.GetURL(), parts[i]), nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, StatusOK, resp.Status)
	}

	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/unseal/complete", ts.GetURL()), nil)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, StatusOK, resp.Status)
}
