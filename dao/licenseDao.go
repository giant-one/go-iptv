package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-iptv/dto"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var WS = NewWSClient()
var Lic dto.Lic

// =========================
// 数据结构
// =========================

type Request struct {
	Action string      `json:"a"`
	Data   interface{} `json:"d"`
}

type Response struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// =========================
// WSClient（稳定版 + 心跳阈值）
// =========================

type WSClient struct {
	url    string
	conn   *websocket.Conn
	rw     sync.RWMutex
	closed bool

	reconnectCh  chan struct{}
	maxRetry     int
	stopCh       chan struct{}
	reconnecting bool // 重连状态标记，防止重复触发

	failCount   int           // 心跳连续失败计数
	failLimit   int           // 心跳失败阈值
	backoffBase time.Duration // 指数退避基础
}

// ------------------ 创建客户端 ------------------

func NewWSClient() *WSClient {
	c := &WSClient{
		maxRetry:    3,
		reconnectCh: make(chan struct{}, 1),
		stopCh:      make(chan struct{}),
		failLimit:   3,
		backoffBase: 1 * time.Second,
	}
	go c.reconnectWorker() // 启动唯一重连协程
	return c
}

// ------------------ 启动连接 ------------------

func (c *WSClient) Start(url string) error {
	c.url = url
	if !IsRunning() {
		return fmt.Errorf("引擎未启动")
	}
	return c.doConnect()
}

// ------------------ 真正执行连接 ------------------

func (c *WSClient) doConnect() error {
	dialer := websocket.Dialer{
		HandshakeTimeout:  5 * time.Second,
		EnableCompression: true,
	}

	var conn *websocket.Conn
	var err error

	for i := 1; i <= c.maxRetry; i++ {
		conn, _, err = dialer.Dial(c.url, nil)
		if err == nil {
			c.rw.Lock()
			c.conn = conn
			c.closed = false
			c.failCount = 0
			c.rw.Unlock()

			log.Println("✅ 引擎连接成功")
			go c.heartbeat()
			return nil
		}
		time.Sleep(time.Duration(i*2) * time.Second)
	}
	return fmt.Errorf("引擎连接失败: %w", err)
}

// ================== 心跳 ==================

func (c *WSClient) heartbeat() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.rw.RLock()
			conn := c.conn
			closed := c.closed
			c.rw.RUnlock()

			if closed || conn == nil {
				return
			}

			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				c.rw.Lock()
				c.failCount++
				log.Printf("⚠️ 心跳失败 #%d", c.failCount)
				if c.failCount >= c.failLimit && !c.reconnecting {
					c.rw.Unlock()
					log.Println("⚠️ 心跳连续失败，触发重连")
					c.triggerReconnect()
				} else {
					c.rw.Unlock()
				}
			} else {
				// 成功心跳，重置计数
				c.rw.Lock()
				c.failCount = 0
				c.rw.Unlock()
			}
		case <-c.stopCh:
			return
		}
	}
}

// ================== 重连控制 ==================

func (c *WSClient) triggerReconnect() {
	c.rw.Lock()
	defer c.rw.Unlock()
	if c.reconnecting || c.closed {
		return // 已经在重连中或已关闭
	}
	c.reconnecting = true
	select {
	case c.reconnectCh <- struct{}{}:
	default:
	}
}

func (c *WSClient) reconnectWorker() {
	for range c.reconnectCh {
		log.Println("🔄 执行引擎重连...")
		c.CloseConn(false)

		backoff := c.backoffBase
		success := false
		for i := 0; i < c.maxRetry; i++ {
			if err := c.doConnect(); err != nil {
				log.Printf("❌ 引擎重连第 %d 次失败: %v", i+1, err)
				time.Sleep(backoff)
				backoff *= 2
			} else {
				success = true
				break
			}
		}

		if !success {
			log.Println("❌ 重连失败，关闭连接")
			c.CloseConn(true) // 彻底关闭
		}

		c.rw.Lock()
		c.reconnecting = false
		c.failCount = 0
		c.rw.Unlock()
	}
}

// ================== 安全关闭 ==================

func (c *WSClient) CloseConn(fullClose bool) {
	c.rw.Lock()
	defer c.rw.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	if fullClose {
		c.closed = true
		select {
		case <-c.stopCh:
		default:
			close(c.stopCh)
		}
	}
}

// ================== 连接状态 ==================

func (c *WSClient) IsOnline() bool {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.conn != nil && !c.closed
}

// ================== 发送请求 ==================

func (c *WSClient) SendWS(req Request) (Response, error) {
	return c.sendWSWithRetry(req, 2)
}

func (c *WSClient) sendWSWithRetry(req Request, retry int) (Response, error) {
	if !IsRunning() {
		return Response{}, fmt.Errorf("引擎未启动")
	}

	if !c.IsOnline() {
		if err := c.doConnect(); err != nil {
			return Response{}, fmt.Errorf("引擎未在线")
		}
	}

	c.rw.RLock()
	conn := c.conn
	c.rw.RUnlock()
	if conn == nil {
		return Response{}, errors.New("连接不存在")
	}

	if err := conn.WriteJSON(req); err != nil {
		log.Println("⚠️ 发送失败，触发重连")
		c.triggerReconnect()
		if retry > 0 {
			time.Sleep(2 * time.Second)
			return c.sendWSWithRetry(req, retry-1)
		}
		return Response{}, fmt.Errorf("发送失败: %w", err)
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("⚠️ 读取响应失败，触发重连")
		c.triggerReconnect()
		if retry > 0 {
			time.Sleep(2 * time.Second)
			return c.sendWSWithRetry(req, retry-1)
		}
		return Response{}, fmt.Errorf("读取响应失败: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(msg, &resp); err != nil {
		return Response{}, fmt.Errorf("解析响应失败: %w", err)
	}
	return resp, nil
}

// ================== 引擎状态检测 ==================

func IsRunning() bool {
	cmd := exec.Command("bash", "-c", "ps -ef | grep '/license' | grep -v grep")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return checkRun()
	}
	return strings.Contains(string(output), "license")
}

func checkRun() bool {
	req, err := http.NewRequest("GET", "http://127.0.0.1:81/", nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Go-http-client/1.1")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return strings.Contains(string(body), "ok")
}
