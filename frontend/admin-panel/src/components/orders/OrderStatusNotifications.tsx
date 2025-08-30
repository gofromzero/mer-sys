import React, { useState } from 'react';
import { useOrderStatusStore } from '../../stores/orderStatusStore';

interface OrderStatusNotificationsProps {
  /** 显示模式：dropdown | panel */
  mode?: 'dropdown' | 'panel';
  /** 最大显示数量 */
  maxItems?: number;
  /** 是否显示时间 */
  showTime?: boolean;
}

const OrderStatusNotifications: React.FC<OrderStatusNotificationsProps> = ({
  mode = 'dropdown',
  maxItems = 10,
  showTime = true,
}) => {
  const {
    notifications,
    getUnreadCount,
    markNotificationRead,
    clearNotifications,
    clearAllNotifications,
    connectionStatus,
    isRealTimeEnabled,
  } = useOrderStatusStore();
  
  const [isOpen, setIsOpen] = useState(false);
  
  const unreadCount = getUnreadCount();
  const displayNotifications = notifications.slice(0, maxItems);
  
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
  
  const getStatusColor = (status: number): string => {
    const colorMap: Record<number, string> = {
      1: 'text-orange-600',
      2: 'text-blue-600',
      3: 'text-purple-600',
      4: 'text-green-600',
      5: 'text-gray-600',
    };
    return colorMap[status] || 'text-gray-600';
  };
  
  const formatTime = (date: Date): string => {
    const now = new Date();
    const diff = Math.floor((now.getTime() - date.getTime()) / 1000);
    
    if (diff < 60) return '刚刚';
    if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`;
    
    return date.toLocaleDateString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };
  
  if (mode === 'panel') {
    return (
      <div className="order-status-notifications-panel bg-white rounded-lg shadow-sm border p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-800">订单状态通知</h3>
          <div className="flex items-center space-x-2">
            <div className={`flex items-center space-x-1 text-sm ${
              connectionStatus === 'connected' ? 'text-green-600' : 
              connectionStatus === 'connecting' ? 'text-yellow-600' : 'text-red-600'
            }`}>
              <div className={`w-2 h-2 rounded-full ${
                connectionStatus === 'connected' ? 'bg-green-500' : 
                connectionStatus === 'connecting' ? 'bg-yellow-500' : 'bg-red-500'
              }`}></div>
              <span>
                {connectionStatus === 'connected' ? '实时连接' :
                 connectionStatus === 'connecting' ? '连接中' : '离线'}
              </span>
            </div>
            {notifications.length > 0 && (
              <button
                onClick={clearAllNotifications}
                className="text-sm text-red-600 hover:text-red-800 underline"
              >
                清空全部
              </button>
            )}
          </div>
        </div>
        
        {!isRealTimeEnabled && (
          <div className="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
            <p className="text-sm text-yellow-800">
              实时通知功能未启用。请在设置中开启以获得最新的订单状态更新。
            </p>
          </div>
        )}
        
        <div className="space-y-2 max-h-96 overflow-y-auto">
          {displayNotifications.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <div className="text-4xl mb-2">🔔</div>
              <p>暂无订单状态通知</p>
            </div>
          ) : (
            displayNotifications.map((notification, index) => (
              <div
                key={`${notification.orderId}-${notification.timestamp.getTime()}`}
                className={`p-3 rounded-md border transition-colors ${
                  notification.read 
                    ? 'bg-gray-50 border-gray-200' 
                    : 'bg-blue-50 border-blue-200'
                }`}
                onClick={() => !notification.read && markNotificationRead(index)}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center space-x-2 mb-1">
                      <span className="text-sm font-medium text-gray-600">
                        订单 #{notification.orderId}
                      </span>
                      {!notification.read && (
                        <span className="w-2 h-2 bg-blue-500 rounded-full"></span>
                      )}
                    </div>
                    <div className="text-sm text-gray-800 mb-1">
                      状态变更：
                      <span className="text-gray-500">{getStatusName(notification.oldStatus)}</span>
                      <span className="mx-1">→</span>
                      <span className={`font-medium ${getStatusColor(notification.newStatus)}`}>
                        {getStatusName(notification.newStatus)}
                      </span>
                    </div>
                    <div className="text-xs text-gray-500">
                      {notification.reason || '系统自动'}
                    </div>
                  </div>
                  {showTime && (
                    <div className="text-xs text-gray-400 ml-4">
                      {formatTime(notification.timestamp)}
                    </div>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    );
  }
  
  // Dropdown模式
  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2 text-gray-600 hover:text-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-500 rounded-full"
      >
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-5 5v-5zM9 3v12l-3-3m6 0l-3 3M9 3L6 6m3-3l3 3" />
        </svg>
        {unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full h-5 w-5 flex items-center justify-center">
            {unreadCount > 99 ? '99+' : unreadCount}
          </span>
        )}
      </button>
      
      {isOpen && (
        <>
          <div className="fixed inset-0 z-10" onClick={() => setIsOpen(false)}></div>
          <div className="absolute right-0 mt-2 w-80 bg-white rounded-lg shadow-xl border z-20">
            <div className="p-4 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <h3 className="font-semibold text-gray-800">订单状态通知</h3>
                <div className="flex items-center space-x-2">
                  <div className={`flex items-center space-x-1 text-xs ${
                    connectionStatus === 'connected' ? 'text-green-600' : 'text-red-600'
                  }`}>
                    <div className={`w-1.5 h-1.5 rounded-full ${
                      connectionStatus === 'connected' ? 'bg-green-500' : 'bg-red-500'
                    }`}></div>
                    <span>{connectionStatus === 'connected' ? '在线' : '离线'}</span>
                  </div>
                  {notifications.length > 0 && (
                    <button
                      onClick={clearNotifications}
                      className="text-xs text-blue-600 hover:text-blue-800 underline"
                    >
                      清除已读
                    </button>
                  )}
                </div>
              </div>
            </div>
            
            <div className="max-h-80 overflow-y-auto">
              {displayNotifications.length === 0 ? (
                <div className="p-8 text-center text-gray-500">
                  <div className="text-2xl mb-2">🔔</div>
                  <p className="text-sm">暂无新通知</p>
                </div>
              ) : (
                <div className="divide-y divide-gray-100">
                  {displayNotifications.map((notification, index) => (
                    <div
                      key={`${notification.orderId}-${notification.timestamp.getTime()}`}
                      className={`p-3 hover:bg-gray-50 cursor-pointer ${
                        !notification.read ? 'bg-blue-50' : ''
                      }`}
                      onClick={() => {
                        if (!notification.read) {
                          markNotificationRead(index);
                        }
                        setIsOpen(false);
                        // 可以添加跳转到订单详情的逻辑
                      }}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center space-x-2 mb-1">
                            <span className="text-sm font-medium text-gray-600">
                              订单 #{notification.orderId}
                            </span>
                            {!notification.read && (
                              <span className="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>
                            )}
                          </div>
                          <div className="text-sm text-gray-800">
                            <span className={`font-medium ${getStatusColor(notification.newStatus)}`}>
                              {getStatusName(notification.newStatus)}
                            </span>
                          </div>
                          <div className="text-xs text-gray-500 mt-1">
                            {notification.reason || '系统自动'}
                          </div>
                        </div>
                        <div className="text-xs text-gray-400 ml-2">
                          {formatTime(notification.timestamp)}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default OrderStatusNotifications;