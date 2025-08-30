import { useEffect, useState, useCallback, useRef } from 'react';
import { useOrderStore } from '../stores/orderStore';

interface UseOrderStatusOptions {
  /** 是否启用实时更新 */
  enableRealTimeUpdate?: boolean;
  /** 轮询间隔（毫秒），默认30秒 */
  pollingInterval?: number;
  /** 订单ID（用于单个订单状态监听） */
  orderId?: number;
  /** 自动重新获取的条件 */
  autoRefreshCondition?: () => boolean;
}

interface OrderStatusNotification {
  orderId: number;
  newStatus: number;
  oldStatus: number;
  reason: string;
  timestamp: Date;
}

export const useOrderStatus = (options: UseOrderStatusOptions = {}) => {
  const {
    enableRealTimeUpdate = false,
    pollingInterval = 30000,
    orderId,
    autoRefreshCondition,
  } = options;

  const { getOrders, getOrder, orders, currentOrder, isLoading } = useOrderStore();
  const [notifications, setNotifications] = useState<OrderStatusNotification[]>([]);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());
  const pollingRef = useRef<NodeJS.Timeout>();
  const wsRef = useRef<WebSocket>();

  // WebSocket连接管理
  const connectWebSocket = useCallback(() => {
    if (!enableRealTimeUpdate) return;

    try {
      // 构建WebSocket URL
      const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${
        window.location.host
      }/ws/orders/status-updates`;

      wsRef.current = new WebSocket(wsUrl);

      wsRef.current.onopen = () => {
        console.log('订单状态WebSocket连接已建立');
      };

      wsRef.current.onmessage = (event) => {
        try {
          const statusUpdate = JSON.parse(event.data) as OrderStatusNotification;
          
          // 添加到通知列表
          setNotifications(prev => [
            {
              ...statusUpdate,
              timestamp: new Date(),
            },
            ...prev.slice(0, 9), // 保持最多10条通知
          ]);

          // 更新最后更新时间
          setLastUpdate(new Date());

          // 如果是当前正在查看的订单，自动刷新
          if (orderId && statusUpdate.orderId === orderId) {
            getOrder(orderId);
          } else {
            // 刷新订单列表
            if (autoRefreshCondition && autoRefreshCondition()) {
              getOrders();
            }
          }
        } catch (error) {
          console.error('解析订单状态更新消息失败:', error);
        }
      };

      wsRef.current.onerror = (error) => {
        console.error('订单状态WebSocket连接错误:', error);
      };

      wsRef.current.onclose = () => {
        console.log('订单状态WebSocket连接已关闭，5秒后重连...');
        setTimeout(connectWebSocket, 5000);
      };
    } catch (error) {
      console.error('创建订单状态WebSocket连接失败:', error);
      // 降级到轮询模式
      startPolling();
    }
  }, [enableRealTimeUpdate, orderId, getOrder, getOrders, autoRefreshCondition]);

  // 轮询机制（WebSocket不可用时的降级方案）
  const startPolling = useCallback(() => {
    if (!enableRealTimeUpdate || pollingRef.current) return;

    pollingRef.current = setInterval(() => {
      if (orderId) {
        getOrder(orderId);
      } else if (autoRefreshCondition && autoRefreshCondition()) {
        getOrders();
      }
      setLastUpdate(new Date());
    }, pollingInterval);
  }, [enableRealTimeUpdate, pollingInterval, orderId, getOrder, getOrders, autoRefreshCondition]);

  // 停止轮询
  const stopPolling = useCallback(() => {
    if (pollingRef.current) {
      clearInterval(pollingRef.current);
      pollingRef.current = undefined;
    }
  }, []);

  // 手动刷新
  const refresh = useCallback(() => {
    if (orderId) {
      getOrder(orderId);
    } else {
      getOrders();
    }
    setLastUpdate(new Date());
  }, [orderId, getOrder, getOrders]);

  // 清除通知
  const clearNotification = useCallback((index: number) => {
    setNotifications(prev => prev.filter((_, i) => i !== index));
  }, []);

  // 清除所有通知
  const clearAllNotifications = useCallback(() => {
    setNotifications([]);
  }, []);

  // 获取状态变更统计
  const getStatusChangeStats = useCallback(() => {
    const stats = notifications.reduce((acc, notification) => {
      const statusName = getStatusName(notification.newStatus);
      acc[statusName] = (acc[statusName] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);
    
    return stats;
  }, [notifications]);

  // 状态名称映射
  const getStatusName = (status: number): string => {
    const statusMap: Record<number, string> = {
      1: '待支付',
      2: '已支付',
      3: '处理中',
      4: '已完成',
      5: '已取消',
    };
    return statusMap[status] || '未知状态';
  };

  // 启动实时更新
  useEffect(() => {
    if (!enableRealTimeUpdate) return;

    // 优先使用WebSocket
    connectWebSocket();

    return () => {
      // 清理WebSocket连接
      if (wsRef.current) {
        wsRef.current.close();
      }
      stopPolling();
    };
  }, [enableRealTimeUpdate, connectWebSocket, stopPolling]);

  // 页面可见性变化处理
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.hidden) {
        // 页面隐藏时暂停更新
        stopPolling();
        if (wsRef.current) {
          wsRef.current.close();
        }
      } else {
        // 页面显示时恢复更新
        if (enableRealTimeUpdate) {
          connectWebSocket();
        }
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [enableRealTimeUpdate, connectWebSocket, stopPolling]);

  return {
    // 状态
    orders,
    currentOrder,
    isLoading,
    notifications,
    lastUpdate,
    
    // 操作
    refresh,
    clearNotification,
    clearAllNotifications,
    getStatusChangeStats,
    getStatusName,
    
    // 连接状态
    isWebSocketConnected: wsRef.current?.readyState === WebSocket.OPEN,
    isPolling: !!pollingRef.current,
  };
};