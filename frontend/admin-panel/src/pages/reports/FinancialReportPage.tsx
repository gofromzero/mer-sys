import React, { useEffect, useMemo } from 'react';
import { 
  ReportLayout, 
  LineChart, 
  BarChart, 
  PieChart, 
  GaugeChart,
  DateRange 
} from '../../components/reports';
import { useReportStore } from '../../stores/reportStore';
import { FinancialReportData } from '../../services/reportService';

interface FinancialReportPageProps {
  className?: string;
}

const FinancialReportPage: React.FC<FinancialReportPageProps> = ({
  className = '',
}) => {
  const {
    financialData,
    analyticsLoading,
    analyticsError,
    dateRange,
    setDateRange,
    fetchFinancialData,
  } = useReportStore();

  // 初始加载数据
  useEffect(() => {
    fetchFinancialData();
  }, [fetchFinancialData]);

  const handleDateRangeChange = (newDateRange: DateRange) => {
    setDateRange(newDateRange.startDate, newDateRange.endDate);
    fetchFinancialData(newDateRange.startDate, newDateRange.endDate);
  };

  const handleRefresh = () => {
    fetchFinancialData();
  };

  // 计算财务指标卡片数据
  const financialMetrics = useMemo(() => {
    if (!financialData) return [];

    return [
      {
        title: '总收入',
        value: financialData.total_revenue.amount,
        unit: '元',
        icon: '💰',
        color: 'text-green-600',
        bgColor: 'bg-green-50',
        change: null, // 可以添加同比变化
      },
      {
        title: '净利润',
        value: financialData.net_profit.amount,
        unit: '元',
        icon: '📈',
        color: 'text-blue-600',
        bgColor: 'bg-blue-50',
        change: null,
      },
      {
        title: '订单总数',
        value: financialData.order_count,
        unit: '笔',
        icon: '🛒',
        color: 'text-purple-600',
        bgColor: 'bg-purple-50',
        change: null,
      },
      {
        title: '活跃商户',
        value: financialData.active_merchant_count,
        unit: '个',
        icon: '🏪',
        color: 'text-orange-600',
        bgColor: 'bg-orange-50',
        change: null,
      },
      {
        title: '活跃客户',
        value: financialData.active_customer_count,
        unit: '个',
        icon: '👥',
        color: 'text-indigo-600',
        bgColor: 'bg-indigo-50',
        change: null,
      },
      {
        title: '权益余额',
        value: financialData.rights_balance,
        unit: '份',
        icon: '🎁',
        color: 'text-pink-600',
        bgColor: 'bg-pink-50',
        change: null,
      },
    ];
  }, [financialData]);

  // 月度趋势图数据
  const monthlyTrendData = useMemo(() => {
    if (!financialData?.breakdown?.monthly_trend) return null;

    const monthlyData = financialData.breakdown.monthly_trend;
    return {
      xAxisData: monthlyData.map(item => item.month),
      series: [
        {
          name: '收入',
          value: monthlyData.map(item => item.revenue.amount),
          itemStyle: { color: '#1890ff' },
        },
        {
          name: '净利润',
          value: monthlyData.map(item => item.net_profit.amount),
          itemStyle: { color: '#52c41a' },
        },
      ],
    };
  }, [financialData]);

  // 商户收入分布饼图数据
  const merchantRevenueData = useMemo(() => {
    if (!financialData?.breakdown?.revenue_by_merchant) return null;

    return financialData.breakdown.revenue_by_merchant
      .slice(0, 10) // 只显示前10名
      .map((merchant, index) => ({
        name: merchant.merchant_name,
        value: merchant.revenue.amount,
        itemStyle: {
          color: [
            '#1890ff', '#52c41a', '#fadb14', '#f5222d', '#722ed1',
            '#fa8c16', '#13c2c2', '#eb2f96', '#fa541c', '#2f54eb'
          ][index % 10],
        },
      }));
  }, [financialData]);

  // 权益使用率仪表盘数据
  const rightsUtilizationRate = useMemo(() => {
    if (!financialData) return 0;
    
    const { rights_distributed, rights_consumed } = financialData;
    if (rights_distributed === 0) return 0;
    
    return Math.round((rights_consumed / rights_distributed) * 100);
  }, [financialData]);

  // 类别收入柱状图数据
  const categoryRevenueData = useMemo(() => {
    if (!financialData?.breakdown?.revenue_by_category) return null;

    const categoryData = financialData.breakdown.revenue_by_category;
    return {
      xAxisData: categoryData.map(item => item.category_name),
      series: [
        {
          name: '收入金额',
          value: categoryData.map(item => item.revenue.amount),
          itemStyle: { color: '#1890ff' },
        },
      ],
    };
  }, [financialData]);

  const formatNumber = (num: number): string => {
    if (num >= 10000) {
      return (num / 10000).toFixed(1) + '万';
    }
    return num.toLocaleString();
  };

  return (
    <div className={`financial-report-page ${className}`}>
      <ReportLayout
        title="财务分析报表"
        subtitle="查看平台整体财务状况和趋势分析"
        reportType="financial"
        loading={analyticsLoading}
        error={analyticsError}
        onDateRangeChange={handleDateRangeChange}
        onRefresh={handleRefresh}
        initialDateRange={dateRange}
        showExporter={true}
        showDownloadList={true}
      >
        {financialData && (
          <div className="space-y-6">
            {/* 关键指标卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {financialMetrics.map((metric) => (
                <div key={metric.title} className={`${metric.bgColor} rounded-lg p-6`}>
                  <div className="flex items-center">
                    <div className={`flex items-center justify-center h-12 w-12 rounded-md ${metric.color}`}>
                      <span className="text-2xl">{metric.icon}</span>
                    </div>
                    <div className="ml-4">
                      <p className="text-sm font-medium text-gray-600">{metric.title}</p>
                      <p className={`text-2xl font-semibold ${metric.color}`}>
                        {formatNumber(metric.value)} {metric.unit}
                      </p>
                      {metric.change && (
                        <p className="text-sm text-gray-500">{metric.change}</p>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* 第一行图表：月度趋势 */}
            {monthlyTrendData && (
              <div className="bg-white rounded-lg shadow p-6">
                <LineChart
                  data={monthlyTrendData.series}
                  xAxisData={monthlyTrendData.xAxisData}
                  title="月度收入趋势"
                  subtitle="收入和净利润的月度变化"
                  yAxisUnit="元"
                  height={400}
                  smooth={true}
                  showSymbol={true}
                />
              </div>
            )}

            {/* 第二行图表：商户分布和权益使用率 */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {merchantRevenueData && (
                <div className="bg-white rounded-lg shadow p-6">
                  <PieChart
                    data={merchantRevenueData}
                    title="商户收入分布"
                    subtitle="前10名商户收入占比"
                    height={400}
                    radius={['30%', '70%']}
                    showLabel={true}
                    showLegend={true}
                  />
                </div>
              )}

              <div className="bg-white rounded-lg shadow p-6">
                <GaugeChart
                  value={rightsUtilizationRate}
                  title="权益使用率"
                  subtitle={`已消耗 ${financialData.rights_consumed} / 已分发 ${financialData.rights_distributed} 份权益`}
                  height={400}
                  max={100}
                  unit="%"
                  color={['#ff4d4f', '#faad14', '#52c41a']}
                />
              </div>
            </div>

            {/* 第三行图表：类别收入分析 */}
            {categoryRevenueData && (
              <div className="bg-white rounded-lg shadow p-6">
                <BarChart
                  data={categoryRevenueData.series}
                  xAxisData={categoryRevenueData.xAxisData}
                  title="商品类别收入分析"
                  subtitle="各商品类别的收入贡献"
                  yAxisUnit="元"
                  height={400}
                  horizontal={false}
                />
              </div>
            )}

            {/* 详细数据表格 */}
            {financialData.breakdown?.revenue_by_merchant && (
              <div className="bg-white rounded-lg shadow overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-200">
                  <h3 className="text-lg font-medium text-gray-900">商户收入详情</h3>
                </div>
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          商户名称
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          收入金额
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          订单数量
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          占比
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {financialData.breakdown.revenue_by_merchant.map((merchant) => (
                        <tr key={merchant.merchant_id}>
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                            {merchant.merchant_name}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            ¥{merchant.revenue.amount.toLocaleString()}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {merchant.order_count} 笔
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {merchant.percentage.toFixed(2)}%
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        )}
      </ReportLayout>
    </div>
  );
};

export default FinancialReportPage;