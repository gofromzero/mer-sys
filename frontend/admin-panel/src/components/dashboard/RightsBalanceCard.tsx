import React from 'react';
import { Card, Progress, Row, Col, Statistic, Alert, Tooltip, Spin, Button } from 'antd';
import {
  WalletOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  PlusOutlined
} from '@ant-design/icons';
import { RightsBalance, RightsAlert, AlertSeverity } from '@/types/dashboard';
import { formatCurrency } from '@/utils/format';

interface RightsBalanceCardProps {
  balance: RightsBalance | null;
  alerts?: RightsAlert[];
  loading?: boolean;
  className?: string;
  onRecharge?: () => void; // 充值回调
}

/**
 * 权益余额卡片组件
 * 显示权益余额、使用情况和预警信息
 */
const RightsBalanceCard: React.FC<RightsBalanceCardProps> = ({
  balance,
  alerts = [],
  loading = false,
  className,
  onRecharge
}) => {
  // 计算使用率
  const getUsageRate = () => {
    if (!balance || balance.total_balance === 0) return 0;
    return ((balance.used_balance + balance.frozen_balance) / balance.total_balance) * 100;
  };

  // 获取余额状态
  const getBalanceStatus = () => {
    if (!balance) return { status: 'normal', color: '#52c41a', text: '正常' };
    
    const { available_balance, warning_threshold, critical_threshold } = balance;
    
    if (critical_threshold && available_balance <= critical_threshold) {
      return { status: 'critical', color: '#ff4d4f', text: '严重不足' };
    }
    
    if (warning_threshold && available_balance <= warning_threshold) {
      return { status: 'warning', color: '#faad14', text: '余额不足' };
    }
    
    return { status: 'normal', color: '#52c41a', text: '充足' };
  };

  // 获取进度条颜色
  const getProgressColor = () => {
    const usageRate = getUsageRate();
    if (usageRate >= 90) return '#ff4d4f';
    if (usageRate >= 70) return '#faad14';
    return '#52c41a';
  };

  // 渲染预警信息
  const renderAlerts = () => {
    if (alerts.length === 0) return null;

    const criticalAlerts = alerts.filter(alert => alert.severity === AlertSeverity.CRITICAL);
    const warningAlerts = alerts.filter(alert => alert.severity === AlertSeverity.WARNING);

    return (
      <div style={{ marginTop: '16px' }}>
        {criticalAlerts.map(alert => (
          <Alert
            key={alert.id}
            type="error"
            size="small"
            message={alert.message}
            icon={<ExclamationCircleOutlined />}
            style={{ marginBottom: '8px' }}
            showIcon
          />
        ))}
        {warningAlerts.map(alert => (
          <Alert
            key={alert.id}
            type="warning"
            size="small"
            message={alert.message}
            icon={<WarningOutlined />}
            style={{ marginBottom: '8px' }}
            showIcon
          />
        ))}
      </div>
    );
  };

  const balanceStatus = getBalanceStatus();
  const usageRate = getUsageRate();

  return (
    <Card
      title={
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <WalletOutlined style={{ marginRight: '8px', color: '#1890ff' }} />
            权益余额
            <span style={{ 
              marginLeft: '8px', 
              fontSize: '12px', 
              color: balanceStatus.color,
              fontWeight: 'bold'
            }}>
              [{balanceStatus.text}]
            </span>
          </div>
          {onRecharge && (
            <Button 
              type="primary" 
              size="small" 
              icon={<PlusOutlined />} 
              onClick={onRecharge}
            >
              充值
            </Button>
          )}
        </div>
      }
      className={className}
      bodyStyle={{ padding: '16px' }}
    >
      <Spin spinning={loading}>
        {balance ? (
          <>
            {/* 可用余额展示 */}
            <div style={{ textAlign: 'center', marginBottom: '24px' }}>
              <Statistic
                title="可用余额"
                value={balance.available_balance}
                formatter={(value) => formatCurrency(Number(value))}
                valueStyle={{ 
                  color: balanceStatus.color,
                  fontSize: '32px',
                  fontWeight: 'bold'
                }}
                prefix={
                  balanceStatus.status === 'normal' ? 
                    <CheckCircleOutlined /> : 
                    <WarningOutlined />
                }
              />
            </div>

            {/* 使用情况进度条 */}
            <div style={{ marginBottom: '20px' }}>
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between', 
                marginBottom: '8px',
                fontSize: '12px',
                color: '#666'
              }}>
                <span>使用情况</span>
                <span>{usageRate.toFixed(1)}%</span>
              </div>
              <Progress
                percent={usageRate}
                strokeColor={getProgressColor()}
                showInfo={false}
                size="small"
              />
            </div>

            {/* 详细余额信息 */}
            <Row gutter={[12, 12]} style={{ fontSize: '12px' }}>
              <Col xs={12} sm={6}>
                <Tooltip title="账户总权益额度">
                  <div style={{ textAlign: 'center', padding: '8px' }}>
                    <div style={{ color: '#8c8c8c' }}>总余额</div>
                    <div style={{ fontWeight: 'bold', color: '#1890ff' }}>
                      {formatCurrency(balance.total_balance)}
                    </div>
                  </div>
                </Tooltip>
              </Col>
              <Col xs={12} sm={6}>
                <Tooltip title="已使用的权益金额">
                  <div style={{ textAlign: 'center', padding: '8px' }}>
                    <div style={{ color: '#8c8c8c' }}>已使用</div>
                    <div style={{ fontWeight: 'bold', color: '#f5222d' }}>
                      {formatCurrency(balance.used_balance)}
                    </div>
                  </div>
                </Tooltip>
              </Col>
              <Col xs={12} sm={6}>
                <Tooltip title="冻结中的权益金额">
                  <div style={{ textAlign: 'center', padding: '8px' }}>
                    <div style={{ color: '#8c8c8c' }}>冻结中</div>
                    <div style={{ fontWeight: 'bold', color: '#faad14' }}>
                      {formatCurrency(balance.frozen_balance)}
                    </div>
                  </div>
                </Tooltip>
              </Col>
              <Col xs={12} sm={6}>
                <Tooltip title="可以使用的权益金额">
                  <div style={{ textAlign: 'center', padding: '8px' }}>
                    <div style={{ color: '#8c8c8c' }}>可用</div>
                    <div style={{ fontWeight: 'bold', color: balanceStatus.color }}>
                      {formatCurrency(balance.available_balance)}
                    </div>
                  </div>
                </Tooltip>
              </Col>
            </Row>

            {/* 预警阈值显示 */}
            {(balance.warning_threshold || balance.critical_threshold) && (
              <div style={{ 
                marginTop: '16px', 
                padding: '12px', 
                backgroundColor: '#fafafa',
                borderRadius: '6px',
                fontSize: '12px'
              }}>
                <div style={{ marginBottom: '4px', fontWeight: 'bold', color: '#666' }}>
                  预警设置:
                </div>
                {balance.warning_threshold && (
                  <div style={{ color: '#faad14' }}>
                    • 预警阈值: {formatCurrency(balance.warning_threshold)}
                  </div>
                )}
                {balance.critical_threshold && (
                  <div style={{ color: '#ff4d4f' }}>
                    • 紧急阈值: {formatCurrency(balance.critical_threshold)}
                  </div>
                )}
              </div>
            )}

            {/* 渲染预警信息 */}
            {renderAlerts()}

            {/* 更新时间 */}
            <div style={{ 
              marginTop: '16px',
              textAlign: 'center',
              fontSize: '12px',
              color: '#8c8c8c'
            }}>
              更新时间: {new Date(balance.last_updated).toLocaleString()}
            </div>
          </>
        ) : (
          <div style={{ 
            textAlign: 'center', 
            padding: '40px 0',
            color: '#8c8c8c'
          }}>
            <WalletOutlined style={{ fontSize: '48px', marginBottom: '16px' }} />
            <div>暂无权益余额数据</div>
          </div>
        )}
      </Spin>
    </Card>
  );
};

export default RightsBalanceCard;