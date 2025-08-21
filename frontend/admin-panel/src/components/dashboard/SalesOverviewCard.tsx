import React from 'react';
import { Card, Row, Col, Statistic, Tooltip, Spin } from 'antd';
import { 
  ShoppingCartOutlined, 
  UserOutlined, 
  DollarOutlined,
  TrendingUpOutlined,
  TrendingDownOutlined 
} from '@ant-design/icons';
import { MerchantDashboardData, TimePeriod } from '@/types/dashboard';
import { formatCurrency, formatNumber } from '@/utils/format';

interface SalesOverviewCardProps {
  data: MerchantDashboardData | null;
  loading?: boolean;
  className?: string;
}

/**
 * 销售概览卡片组件
 * 显示核心业务指标：销售额、订单数、客户数
 */
const SalesOverviewCard: React.FC<SalesOverviewCardProps> = ({
  data,
  loading = false,
  className
}) => {
  // 计算增长率（这里是示例，实际应该从后端获取对比数据）
  const calculateGrowthRate = (current: number, previous: number) => {
    if (previous === 0) return 0;
    return ((current - previous) / previous) * 100;
  };

  // 获取周期显示文本
  const getPeriodText = (period: TimePeriod) => {
    switch (period) {
      case TimePeriod.DAILY:
        return '今日';
      case TimePeriod.WEEKLY:
        return '本周';
      case TimePeriod.MONTHLY:
        return '本月';
      default:
        return '今日';
    }
  };

  // 格式化增长率显示
  const renderGrowthRate = (rate: number) => {
    const isPositive = rate > 0;
    const Icon = isPositive ? TrendingUpOutlined : TrendingDownOutlined;
    const color = isPositive ? '#52c41a' : '#ff4d4f';
    
    return (
      <span style={{ color, fontSize: '12px', marginLeft: '8px' }}>
        <Icon style={{ marginRight: '2px' }} />
        {Math.abs(rate).toFixed(1)}%
      </span>
    );
  };

  const periodText = data ? getPeriodText(data.period) : '今日';

  return (
    <Card
      title={
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <DollarOutlined style={{ marginRight: '8px', color: '#1890ff' }} />
          销售概览 - {periodText}
        </div>
      }
      className={className}
      bodyStyle={{ padding: '16px' }}
    >
      <Spin spinning={loading}>
        <Row gutter={[16, 16]}>
          {/* 总销售额 */}
          <Col xs={24} sm={8}>
            <div style={{ textAlign: 'center', padding: '16px' }}>
              <Tooltip title="当前周期内的总销售金额">
                <Statistic
                  title="总销售额"
                  value={data?.total_sales || 0}
                  formatter={(value) => formatCurrency(Number(value))}
                  prefix={<DollarOutlined style={{ color: '#52c41a' }} />}
                  suffix={data ? renderGrowthRate(8.5) : null} // 示例增长率
                  valueStyle={{ color: '#52c41a', fontSize: '20px' }}
                />
              </Tooltip>
            </div>
          </Col>

          {/* 总订单数 */}
          <Col xs={24} sm={8}>
            <div style={{ textAlign: 'center', padding: '16px' }}>
              <Tooltip title="当前周期内的订单总数">
                <Statistic
                  title="总订单数"
                  value={data?.total_orders || 0}
                  formatter={(value) => formatNumber(Number(value))}
                  prefix={<ShoppingCartOutlined style={{ color: '#1890ff' }} />}
                  suffix={data ? renderGrowthRate(12.3) : null} // 示例增长率
                  valueStyle={{ color: '#1890ff', fontSize: '20px' }}
                />
              </Tooltip>
            </div>
          </Col>

          {/* 总客户数 */}
          <Col xs={24} sm={8}>
            <div style={{ textAlign: 'center', padding: '16px' }}>
              <Tooltip title="当前周期内的活跃客户数">
                <Statistic
                  title="活跃客户"
                  value={data?.total_customers || 0}
                  formatter={(value) => formatNumber(Number(value))}
                  prefix={<UserOutlined style={{ color: '#722ed1' }} />}
                  suffix={data ? renderGrowthRate(5.8) : null} // 示例增长率
                  valueStyle={{ color: '#722ed1', fontSize: '20px' }}
                />
              </Tooltip>
            </div>
          </Col>
        </Row>

        {/* 额外信息 */}
        {data && (
          <Row style={{ marginTop: '16px', paddingTop: '16px', borderTop: '1px solid #f0f0f0' }}>
            <Col span={24}>
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between',
                fontSize: '12px',
                color: '#8c8c8c'
              }}>
                <span>
                  待处理订单: <strong>{data.pending_orders}</strong>
                </span>
                <span>
                  待核销订单: <strong>{data.pending_verifications}</strong>
                </span>
                <span>
                  数据更新时间: <strong>{new Date(data.last_updated).toLocaleString()}</strong>
                </span>
              </div>
            </Col>
          </Row>
        )}
      </Spin>
    </Card>
  );
};

export default SalesOverviewCard;