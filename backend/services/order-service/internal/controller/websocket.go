package controller

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gorilla/websocket"
)

// 确保WebSocketController实现WebSocketNotifier接口
var _ interface {
	BroadcastOrderStatusChange(ctx context.Context, order *types.Order, statusHistory *types.OrderStatusHistory)
	SendOrderStatusChangeToUser(ctx context.Context, userID, tenantID uint64, order *types.Order, statusHistory *types.OrderStatusHistory)
} = (*WebSocketController)(nil)

// WebSocketMessage WebSocket消息结构
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// OrderStatusNotification 订单状态通知消息
type OrderStatusNotification struct {
	OrderID       uint64                   `json:"order_id"`
	OrderNumber   string                   `json:"order_number"`
	FromStatus    types.OrderStatusInt     `json:"from_status"`
	ToStatus      types.OrderStatusInt     `json:"to_status"`
	Reason        string                   `json:"reason"`
	OperatorType  types.OrderStatusOperatorType `json:"operator_type"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

// WebSocketConnection WebSocket连接管理
type WebSocketConnection struct {
	conn       *websocket.Conn
	userID     uint64
	tenantID   uint64
	send       chan WebSocketMessage
	hub        *WebSocketHub
	mu         sync.RWMutex
	closed     bool
}

// WebSocketHub WebSocket连接管理中心
type WebSocketHub struct {
	connections map[uint64]map[uint64]*WebSocketConnection // [tenantID][userID]connection
	register    chan *WebSocketConnection
	unregister  chan *WebSocketConnection
	broadcast   chan WebSocketMessage
	mu          sync.RWMutex
}

// WebSocketController WebSocket控制器
type WebSocketController struct {
	hub *WebSocketHub
}

// 全局WebSocket Hub实例
var globalWebSocketHub *WebSocketHub

// init 初始化全局WebSocket Hub
func init() {
	globalWebSocketHub = &WebSocketHub{
		connections: make(map[uint64]map[uint64]*WebSocketConnection),
		register:    make(chan *WebSocketConnection),
		unregister:  make(chan *WebSocketConnection),
		broadcast:   make(chan WebSocketMessage),
	}
	
	// 启动Hub
	go globalWebSocketHub.run()
}

// NewWebSocketController 创建WebSocket控制器
func NewWebSocketController() *WebSocketController {
	return &WebSocketController{
		hub: globalWebSocketHub,
	}
}

// 配置WebSocket升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 在生产环境中应该检查Origin
		return true
	},
}

// HandleOrderStatusUpdates 处理订单状态更新WebSocket连接
func (c *WebSocketController) HandleOrderStatusUpdates(r *ghttp.Request) {
	// 从JWT token获取用户信息
	userID := r.GetCtx().Value("user_id")
	tenantID := r.GetCtx().Value("tenant_id")
	
	if userID == nil || tenantID == nil {
		r.Response.WriteStatus(401, "未授权的连接")
		return
	}
	
	uid, ok := userID.(uint64)
	if !ok {
		r.Response.WriteStatus(400, "无效的用户ID")
		return
	}
	
	tid, ok := tenantID.(uint64)
	if !ok {
		r.Response.WriteStatus(400, "无效的租户ID")
		return
	}
	
	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(r.Response.ResponseWriter, r.Request, nil)
	if err != nil {
		g.Log().Error(r.GetCtx(), "WebSocket升级失败", "error", err)
		return
	}
	
	// 创建WebSocket连接对象
	wsConn := &WebSocketConnection{
		conn:     conn,
		userID:   uid,
		tenantID: tid,
		send:     make(chan WebSocketMessage, 256),
		hub:      c.hub,
	}
	
	// 注册连接
	c.hub.register <- wsConn
	
	// 启动读写协程
	go wsConn.writePump()
	go wsConn.readPump()
	
	g.Log().Info(r.GetCtx(), "WebSocket连接已建立", 
		"user_id", uid, 
		"tenant_id", tid,
		"remote_addr", r.GetRemoteIp())
}

// run Hub主运行循环
func (h *WebSocketHub) run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			if h.connections[conn.tenantID] == nil {
				h.connections[conn.tenantID] = make(map[uint64]*WebSocketConnection)
			}
			h.connections[conn.tenantID][conn.userID] = conn
			h.mu.Unlock()
			
			g.Log().Info(context.Background(), "WebSocket连接已注册",
				"user_id", conn.userID,
				"tenant_id", conn.tenantID)
				
		case conn := <-h.unregister:
			h.mu.Lock()
			if tenantConns, exists := h.connections[conn.tenantID]; exists {
				if _, exists := tenantConns[conn.userID]; exists {
					delete(tenantConns, conn.userID)
					close(conn.send)
					
					// 如果租户下没有连接了，删除租户映射
					if len(tenantConns) == 0 {
						delete(h.connections, conn.tenantID)
					}
				}
			}
			h.mu.Unlock()
			
			g.Log().Info(context.Background(), "WebSocket连接已注销",
				"user_id", conn.userID,
				"tenant_id", conn.tenantID)
				
		case message := <-h.broadcast:
			// 广播消息到所有连接
			h.mu.RLock()
			for tenantID, tenantConns := range h.connections {
				for userID, conn := range tenantConns {
					select {
					case conn.send <- message:
					default:
						// 发送失败，关闭连接
						delete(tenantConns, userID)
						close(conn.send)
						
						g.Log().Warning(context.Background(), "WebSocket发送失败，关闭连接",
							"user_id", userID,
							"tenant_id", tenantID)
					}
				}
				
				// 清理空的租户映射
				if len(tenantConns) == 0 {
					delete(h.connections, tenantID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// readPump 读取WebSocket消息
func (c *WebSocketConnection) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	
	// 设置读取参数
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				g.Log().Error(context.Background(), "WebSocket读取错误", "error", err)
			}
			break
		}
	}
}

// writePump 发送WebSocket消息
func (c *WebSocketConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			// 发送JSON消息
			if err := c.conn.WriteJSON(message); err != nil {
				g.Log().Error(context.Background(), "WebSocket写入错误", "error", err)
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// BroadcastOrderStatusChange 广播订单状态变更通知
func (c *WebSocketController) BroadcastOrderStatusChange(ctx context.Context, order *types.Order, statusHistory *types.OrderStatusHistory) {
	notification := OrderStatusNotification{
		OrderID:      order.ID,
		OrderNumber:  order.OrderNumber,
		FromStatus:   statusHistory.FromStatus,
		ToStatus:     statusHistory.ToStatus,
		Reason:       statusHistory.Reason,
		OperatorType: statusHistory.OperatorType,
		UpdatedAt:    time.Now(),
	}
	
	message := WebSocketMessage{
		Type:      "order_status_changed",
		Timestamp: time.Now(),
		Data:      notification,
	}
	
	// 发送给特定租户的所有连接
	c.hub.mu.RLock()
	if tenantConns, exists := c.hub.connections[order.TenantID]; exists {
		for _, conn := range tenantConns {
			select {
			case conn.send <- message:
			default:
				g.Log().Warning(ctx, "WebSocket发送队列已满，跳过发送",
					"user_id", conn.userID,
					"tenant_id", conn.tenantID,
					"order_id", order.ID)
			}
		}
	}
	c.hub.mu.RUnlock()
	
	g.Log().Info(ctx, "订单状态变更WebSocket通知已发送",
		"order_id", order.ID,
		"order_number", order.OrderNumber,
		"tenant_id", order.TenantID)
}

// SendOrderStatusChangeToUser 发送订单状态变更通知给特定用户
func (c *WebSocketController) SendOrderStatusChangeToUser(ctx context.Context, userID, tenantID uint64, order *types.Order, statusHistory *types.OrderStatusHistory) {
	notification := OrderStatusNotification{
		OrderID:      order.ID,
		OrderNumber:  order.OrderNumber,
		FromStatus:   statusHistory.FromStatus,
		ToStatus:     statusHistory.ToStatus,
		Reason:       statusHistory.Reason,
		OperatorType: statusHistory.OperatorType,
		UpdatedAt:    time.Now(),
	}
	
	message := WebSocketMessage{
		Type:      "order_status_changed",
		Timestamp: time.Now(),
		Data:      notification,
	}
	
	// 发送给特定用户
	c.hub.mu.RLock()
	if tenantConns, exists := c.hub.connections[tenantID]; exists {
		if conn, exists := tenantConns[userID]; exists {
			select {
			case conn.send <- message:
				g.Log().Info(ctx, "订单状态变更WebSocket通知已发送给用户",
					"user_id", userID,
					"tenant_id", tenantID,
					"order_id", order.ID)
			default:
				g.Log().Warning(ctx, "用户WebSocket发送队列已满",
					"user_id", userID,
					"tenant_id", tenantID,
					"order_id", order.ID)
			}
		}
	}
	c.hub.mu.RUnlock()
}

// GetWebSocketHub 获取全局WebSocket Hub实例
func GetWebSocketHub() *WebSocketHub {
	return globalWebSocketHub
}

// GetActiveConnections 获取活跃连接数统计
func (h *WebSocketHub) GetActiveConnections() map[uint64]int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	stats := make(map[uint64]int)
	for tenantID, tenantConns := range h.connections {
		stats[tenantID] = len(tenantConns)
	}
	
	return stats
}