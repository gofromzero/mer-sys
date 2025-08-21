import React, { useEffect, useState } from 'react';
import { 
  Row, 
  Col, 
  Card, 
  Select, 
  Button, 
  Spin, 
  message, 
  Modal,
  Tooltip,
  Space 
} from 'antd';
import {
  ReloadOutlined,
  SettingOutlined,
  FullscreenOutlined,
  FullscreenExitOutlined
} from '@ant-design/icons';
import { TimePeriod } from '@/types/dashboard';
import { useDashboardStore } from '@/stores/dashboardStore';
import { 
  SalesOverviewCard,
  RightsBalanceCard
} from '@/components/dashboard';

const { Option } = Select;

/**
 * 商户权益监控仪表板页面
 * 集成Amis框架，提供完整的仪表板功能
 */
const RightsMonitoringDashboard: React.FC = () => {
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  
  // 从Store获取状态和方法
  const {
    dashboardData,
    loading,
    error,
    currentPeriod,
    setPeriod,
    loadAllData,
    refreshDashboard,
    startAutoRefresh,
    stopAutoRefresh
  } = useDashboardStore();

  // 组件挂载时加载数据
  useEffect(() => {
    loadAllData();
    startAutoRefresh();
    
    // 组件卸载时清理
    return () => {
      stopAutoRefresh();
    };
  }, []);

  // 错误提示
  useEffect(() => {
    if (error) {
      message.error(error.message || '加载仪表板数据失败');
    }
  }, [error]);

  // 刷新数据
  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await refreshDashboard();
      message.success('数据刷新成功');
    } catch (error: any) {
      message.error(error.message || '刷新失败');
    } finally {
      setRefreshing(false);
    }
  };

  // 切换时间周期
  const handlePeriodChange = (period: TimePeriod) => {
    setPeriod(period);
    message.info(`已切换到${getPeriodLabel(period)}视图`);
  };

  // 获取周期标签
  const getPeriodLabel = (period: TimePeriod) => {
    switch (period) {
      case TimePeriod.DAILY:
        return '日';
      case TimePeriod.WEEKLY:
        return '周';
      case TimePeriod.MONTHLY:
        return '月';
      default:
        return '日';
    }
  };

  // 全屏切换
  const toggleFullscreen = () => {
    if (!isFullscreen) {
      document.documentElement.requestFullscreen?.();
    } else {
      document.exitFullscreen?.();
    }
    setIsFullscreen(!isFullscreen);
  };

  // 打开配置
  const openConfig = () => {
    Modal.info({
      title: '仪表板配置',
      content: '配置功能开发中...',
      okText: '确定'
    });
  };

  return (
    <div 
      data-testid="dashboard-container"
      style={{ 
        padding: '24px', 
        backgroundColor: '#f5f5f5', 
        minHeight: '100vh',
        touchAction: 'manipulation'
      }}
    >
      {/* 页面头部 */}
      <div style={{ 
        marginBottom: '24px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        backgroundColor: '#fff',
        padding: '16px 24px',
        borderRadius: '8px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
      }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '24px', color: '#262626' }}>
            商户运营仪表板
          </h1>
          <p style={{ margin: '4px 0 0 0', color: '#8c8c8c' }}>
            实时监控业务数据和权益使用情况
          </p>
        </div>
        
        <Space>
          {/* 时间周期选择 */}
          <Select
            value={currentPeriod}
            onChange={handlePeriodChange}
            style={{ width: 120 }}
          >
            <Option value={TimePeriod.DAILY}>今日</Option>
            <Option value={TimePeriod.WEEKLY}>本周</Option>
            <Option value={TimePeriod.MONTHLY}>本月</Option>
          </Select>

          {/* 刷新按钮 */}
          <Tooltip title="刷新数据">
            <Button 
              icon={<ReloadOutlined />} 
              loading={refreshing}
              onClick={handleRefresh}
            >
              刷新
            </Button>
          </Tooltip>

          {/* 全屏按钮 */}
          <Tooltip title={isFullscreen ? '退出全屏' : '全屏显示'}>
            <Button 
              icon={isFullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
              onClick={toggleFullscreen}
            />
          </Tooltip>

          {/* 配置按钮 */}
          <Tooltip title="仪表板配置">
            <Button 
              icon={<SettingOutlined />}
              onClick={openConfig}
            />
          </Tooltip>
        </Space>
      </div>

      {/* 主要内容区域 */}
      <Spin spinning={loading.dashboard} tip="加载仪表板数据...">
        <Row gutter={[24, 24]}>
          {/* 第一行 - 销售概览和权益余额 */}
          <Col xs={24} lg={12}>
            <SalesOverviewCard
              data={dashboardData}
              loading={loading.stats}
            />
          </Col>
          
          <Col xs={24} lg={12}>
            <RightsBalanceCard
              balance={dashboardData?.rights_balance || null}
              alerts={dashboardData?.rights_alerts || []}
              loading={loading.dashboard}
              onRecharge={() => {
                Modal.info({
                  title: '权益充值',
                  content: '充值功能开发中...',
                  okText: '确定'
                });
              }}
            />
          </Col>

          {/* 第二行 - 权益趋势图表 */}
          <Col xs={24}>
            <Card
              title="权益使用趋势"
              loading={loading.trends}
              bodyStyle={{ padding: '16px' }}
            >
              <div style={{ 
                height: '300px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#8c8c8c'
              }}>
                <div style={{ textAlign: 'center' }}>
                  <div style={{ fontSize: '48px', marginBottom: '16px' }}>📈</div>
                  <div>权益趋势图表组件开发中...</div>
                  <div style={{ fontSize: '12px', marginTop: '8px' }}>
                    将使用ECharts集成到Amis中展示趋势数据
                  </div>
                </div>
              </div>
            </Card>
          </Col>

          {/* 第三行 - 待处理事项和公告通知 */}
          <Col xs={24} lg={12}>
            <Card
              title="待处理事项"
              loading={loading.tasks}
              bodyStyle={{ padding: '16px' }}
            >
              <div style={{ 
                height: '250px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#8c8c8c'
              }}>
                <div style={{ textAlign: 'center' }}>
                  <div style={{ fontSize: '32px', marginBottom: '8px' }}>📋</div>
                  <div>待处理事项组件开发中...</div>
                </div>
              </div>
            </Card>
          </Col>

          <Col xs={24} lg={12}>
            <Card
              title="公告通知"
              loading={loading.notifications}
              bodyStyle={{ padding: '16px' }}
            >
              <div style={{ 
                height: '250px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#8c8c8c'
              }}>
                <div style={{ textAlign: 'center' }}>
                  <div style={{ fontSize: '32px', marginBottom: '8px' }}>📢</div>
                  <div>公告通知组件开发中...</div>
                </div>
              </div>
            </Card>
          </Col>
        </Row>
      </Spin>

      {/* 底部信息 */}
      <div style={{
        marginTop: '24px',
        textAlign: 'center',
        color: '#8c8c8c',
        fontSize: '12px'
      }}>
        <p>
          数据更新时间: {dashboardData ? new Date(dashboardData.last_updated).toLocaleString() : '--'}
          <span style={{ margin: '0 16px' }}>|</span>
          自动刷新: 5分钟
        </p>
      </div>
    </div>
  );
};

export default RightsMonitoringDashboard;