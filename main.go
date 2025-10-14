// package main
//
// import (
//
//	"GoChatServer/server"
//	"GoChatServer/utils"
//	"fmt"
//	"log"
//
// )
//
//	func main() {
//		port := utils.GetChatPort()
//		addr := fmt.Sprintf(":%s", port)
//
//		s := server.New(addr)
//		if err := s.Start(); err != nil {
//			log.Fatal(err)
//		}
//	}
//
// main.go
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader 설정
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 모든 origin 허용 (로컬 테스트용)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 연결된 클라이언트 관리
type Client struct {
	conn *websocket.Conn
	send chan []byte
	name string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Hub 실행 - 메시지 브로드캐스팅 담당
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("새 클라이언트 연결: %s (총 %d명)", client.name, len(h.clients))

			// 입장 메시지 브로드캐스트
			message := fmt.Sprintf("[시스템] %s님이 입장했습니다.", client.name)
			h.broadcastToAll([]byte(message))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("클라이언트 연결 해제: %s (총 %d명)", client.name, len(h.clients))

				// 퇴장 메시지 브로드캐스트
				message := fmt.Sprintf("[시스템] %s님이 퇴장했습니다.", client.name)
				h.broadcastToAll([]byte(message))
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.broadcastToAll(message)
		}
	}
}

func (h *Hub) broadcastToAll(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			// 전송 실패 시 클라이언트 정리
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// 클라이언트로부터 메시지 읽기
func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 에러: %v", err)
			}
			break
		}

		// 메시지 포맷팅
		formattedMessage := fmt.Sprintf("[%s] %s", c.name, string(message))
		log.Printf("메시지 수신: %s", formattedMessage)

		// 브로드캐스트
		hub.broadcast <- []byte(formattedMessage)
	}
}

// 클라이언트로 메시지 쓰기
func (c *Client) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("메시지 전송 실패: %v", err)
			return
		}
	}
}

// WebSocket 핸들러
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// 클라이언트 이름 가져오기
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "익명"
	}

	// WebSocket으로 업그레이드
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 업그레이드 실패: %v", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
		name: name,
	}

	hub.register <- client

	// 고루틴 시작
	go client.writePump()
	go client.readPump(hub)
}

// 서버의 로컬 IP 주소 찾기
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

// 간단한 HTML 클라이언트 제공
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlTemplate))
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WebSocket 채팅</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        
        #login-screen, #chat-screen {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        
        #login-screen {
            padding: 40px;
            text-align: center;
            max-width: 400px;
        }
        
        #login-screen h1 {
            color: #333;
            margin-bottom: 30px;
            font-size: 28px;
        }
        
        #login-screen input {
            width: 100%;
            padding: 15px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 16px;
            margin-bottom: 20px;
            transition: border-color 0.3s;
        }
        
        #login-screen input:focus {
            outline: none;
            border-color: #667eea;
        }
        
        #login-screen button {
            width: 100%;
            padding: 15px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s;
        }
        
        #login-screen button:hover {
            transform: translateY(-2px);
        }
        
        #chat-screen {
            width: 800px;
            height: 600px;
            display: none;
            flex-direction: column;
        }
        
        #chat-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        #chat-header h2 {
            font-size: 20px;
        }
        
        #status {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        
        .status-dot {
            width: 10px;
            height: 10px;
            border-radius: 50%;
            background: #4ade80;
        }
        
        .status-dot.disconnected {
            background: #ef4444;
        }
        
        #messages {
            flex: 1;
            overflow-y: auto;
            padding: 20px;
            background: #f8f9fa;
        }
        
        .message {
            margin-bottom: 15px;
            padding: 12px 16px;
            border-radius: 8px;
            background: white;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            animation: slideIn 0.3s ease-out;
        }
        
        @keyframes slideIn {
            from {
                opacity: 0;
                transform: translateY(10px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        
        .message.system {
            background: #e0e7ff;
            color: #4338ca;
            text-align: center;
            font-style: italic;
        }
        
        .message .sender {
            font-weight: 600;
            color: #667eea;
            margin-bottom: 4px;
        }
        
        .message .text {
            color: #333;
        }
        
        .message .time {
            font-size: 11px;
            color: #999;
            margin-top: 4px;
        }
        
        #input-area {
            padding: 20px;
            background: white;
            border-top: 1px solid #e0e0e0;
            display: flex;
            gap: 10px;
        }
        
        #messageInput {
            flex: 1;
            padding: 12px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 14px;
            transition: border-color 0.3s;
        }
        
        #messageInput:focus {
            outline: none;
            border-color: #667eea;
        }
        
        #sendButton {
            padding: 12px 30px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s;
        }
        
        #sendButton:hover {
            transform: translateY(-2px);
        }
        
        #sendButton:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }
    </style>
</head>
<body>
    <div id="login-screen">
        <h1>💬 WebSocket 채팅</h1>
        <input type="text" id="nameInput" placeholder="이름을 입력하세요" maxlength="20">
        <button onclick="connect()">입장하기</button>
    </div>

    <div id="chat-screen">
        <div id="chat-header">
            <h2>채팅방</h2>
            <div id="status">
                <span class="status-dot" id="statusDot"></span>
                <span id="statusText">연결됨</span>
            </div>
        </div>
        <div id="messages"></div>
        <div id="input-area">
            <input type="text" id="messageInput" placeholder="메시지를 입력하세요..." maxlength="500">
            <button id="sendButton" onclick="sendMessage()">전송</button>
        </div>
    </div>

    <script>
        let ws;
        let userName;
        
        function connect() {
            userName = document.getElementById('nameInput').value.trim();
            if (!userName) {
                alert('이름을 입력해주세요!');
                return;
            }
            
            document.getElementById('login-screen').style.display = 'none';
            document.getElementById('chat-screen').style.display = 'flex';
            
            // WebSocket 연결
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/ws?name=' + encodeURIComponent(userName);
            
            ws = new WebSocket(wsUrl);
            
            ws.onopen = function() {
                console.log('WebSocket 연결됨');
                updateStatus(true);
            };
            
            ws.onmessage = function(event) {
                displayMessage(event.data);
            };
            
            ws.onclose = function() {
                console.log('WebSocket 연결 종료');
                updateStatus(false);
            };
            
            ws.onerror = function(error) {
                console.error('WebSocket 에러:', error);
                updateStatus(false);
            };
            
            // Enter 키로 메시지 전송
            document.getElementById('messageInput').addEventListener('keypress', function(e) {
                if (e.key === 'Enter') {
                    sendMessage();
                }
            });
        }
        
        function sendMessage() {
            const input = document.getElementById('messageInput');
            const message = input.value.trim();
            
            if (message && ws.readyState === WebSocket.OPEN) {
                ws.send(message);
                input.value = '';
            }
        }
        
        function displayMessage(text) {
            const messagesDiv = document.getElementById('messages');
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message';
            
            // 시스템 메시지 감지
            if (text.startsWith('[시스템]')) {
                messageDiv.classList.add('system');
                messageDiv.innerHTML = '<div class="text">' + escapeHtml(text) + '</div>';
            } else {
                // [이름] 메시지 파싱
                const match = text.match(/^\[(.*?)\] (.*)$/);
                if (match) {
                    const sender = match[1];
                    const content = match[2];
                    messageDiv.innerHTML = 
                        '<div class="sender">' + escapeHtml(sender) + '</div>' +
                        '<div class="text">' + escapeHtml(content) + '</div>' +
                        '<div class="time">' + getCurrentTime() + '</div>';
                } else {
                    messageDiv.innerHTML = '<div class="text">' + escapeHtml(text) + '</div>';
                }
            }
            
            messagesDiv.appendChild(messageDiv);
            messagesDiv.scrollTop = messagesDiv.scrollHeight;
        }
        
        function updateStatus(connected) {
            const dot = document.getElementById('statusDot');
            const text = document.getElementById('statusText');
            const sendButton = document.getElementById('sendButton');
            
            if (connected) {
                dot.classList.remove('disconnected');
                text.textContent = '연결됨';
                sendButton.disabled = false;
            } else {
                dot.classList.add('disconnected');
                text.textContent = '연결 끊김';
                sendButton.disabled = true;
            }
        }
        
        function getCurrentTime() {
            const now = new Date();
            return now.getHours().toString().padStart(2, '0') + ':' + 
                   now.getMinutes().toString().padStart(2, '0');
        }
        
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
        
        // Enter 키로 로그인
        document.getElementById('nameInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                connect();
            }
        });
    </script>
</body>
</html>
`

func main() {
	hub := newHub()
	go hub.run()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	port := "8080"
	localIP := getLocalIP()

	fmt.Println("=====================================")
	fmt.Println("🚀 WebSocket 채팅 서버 시작!")
	fmt.Println("=====================================")
	fmt.Println()
	fmt.Println("📡 서버 주소:")
	fmt.Printf("   - 이 컴퓨터에서: http://localhost:%s\n", port)
	fmt.Printf("   - 다른 컴퓨터에서: http://%s:%s\n", localIP, port)
	fmt.Println()
	fmt.Println("💡 사용법:")
	fmt.Println("   1. 위 주소를 브라우저에서 열기")
	fmt.Println("   2. 이름 입력 후 입장")
	fmt.Println("   3. 다른 컴퓨터에서도 같은 방법으로 접속")
	fmt.Println()
	fmt.Println("🛑 종료하려면 Ctrl+C를 누르세요")
	fmt.Println("=====================================")

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("서버 시작 실패:", err)
	}
}
