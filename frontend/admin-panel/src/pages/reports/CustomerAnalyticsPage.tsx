import React, { useEffect, useMemo } from 'react';
import { 
  ReportLayout, 
  LineChart, 
  BarChart, 
  PieChart, 
  HeatmapChart,
  DateRange 
} from '../../components/reports';
import { useReportStore } from '../../stores/reportStore';
import { CustomerBehaviorReport } from '../../services/reportService';

interface CustomerAnalyticsPageProps {
  className?: string;
}

const CustomerAnalyticsPage: React.FC<CustomerAnalyticsPageProps> = ({
  className = '',
}) => {
  const {
    customerData,
    analyticsLoading,
    analyticsError,
    dateRange,
    setDateRange,
    fetchCustomerData,
  } = useReportStore();

  // 初始加载数据
  useEffect(() => {
    fetchCustomerData();
  }, [fetchCustomerData]);

  const handleDateRangeChange = (newDateRange: DateRange) => {
    setDateRange(newDateRange.startDate, newDateRange.endDate);
    fetchCustomerData(newDateRange.startDate, newDateRange.endDate);
  };

  const handleRefresh = () => {
    fetchCustomerData();
  };

  // 用户活跃度指标
  const userActivityMetrics = useMemo(() => {
    if (!customerData?.user_activity) return [];

    return [
      {
        title: 'DAU (日活跃用户)',
        value: customerData.user_activity.daily_active_users,
        unit: '人',
        icon: '👥',
        color: 'text-blue-600',
        bgColor: 'bg-blue-50',
        trend: customerData.user_activity.dau_growth_rate || 0,
      },
      {
        title: 'WAU (周活跃用户)',
        value: customerData.user_activity.weekly_active_users,
        unit: '人',
        icon: '📅',
        color: 'text-green-600',
        bgColor: 'bg-green-50',
        trend: customerData.user_activity.wau_growth_rate || 0,
      },
      {
        title: 'MAU (月活跃用户)',
        value: customerData.user_activity.monthly_active_users,
        unit: '人',
        icon: '📆',
        color: 'text-purple-600',
        bgColor: 'bg-purple-50',
        trend: customerData.user_activity.mau_growth_rate || 0,
      },
      {
        title: '平均会话时长',
        value: Math.round(customerData.user_activity.avg_session_duration / 60),
        unit: '分钟',
        icon: '⏱️',
        color: 'text-orange-600',
        bgColor: 'bg-orange-50',
        trend: customerData.user_activity.session_duration_trend || 0,
      },
    ];
  }, [customerData]);

  // 用户活跃度趋势图数据
  const activityTrendData = useMemo(() => {
    if (!customerData?.activity_trends) return null;

    const trends = customerData.activity_trends;
    return {
      xAxisData: trends.map(item => item.date),
      series: [
        {
          name: 'DAU',
          value: trends.map(item => item.daily_active_users),
          itemStyle: { color: '#1890ff' },
        },
        {
          name: 'WAU',
          value: trends.map(item => item.weekly_active_users),
          itemStyle: { color: '#52c41a' },
        },
        {
          name: 'MAU',
          value: trends.map(item => item.monthly_active_users),
          itemStyle: { color: '#722ed1' },
        },
      ],
    };
  }, [customerData]);

  // 用户留存率数据
  const retentionData = useMemo(() => {
    if (!customerData?.retention_analysis) return null;

    const retention = customerData.retention_analysis;
    return [
      { name: '1日留存', value: retention.day_1_retention },
      { name: '7日留存', value: retention.day_7_retention },
      { name: '30日留存', value: retention.day_30_retention },
      { name: '90日留存', value: retention.day_90_retention },
    ].map((item, index) => ({
      ...item,
      itemStyle: {
        color: ['#1890ff', '#52c41a', '#fadb14', '#f5222d'][index],
      },
    }));
  }, [customerData]);

  // 用户消费行为分析
  const consumptionBehaviorData = useMemo(() => {
    if (!customerData?.consumption_behavior) return null;

    const behavior = customerData.consumption_behavior;
    return {
      xAxisData: behavior.map(item => item.customer_segment),
      series: [
        {
          name: '平均订单金额',
          value: behavior.map(item => item.avg_order_amount.amount),
          itemStyle: { color: '#1890ff' },
        },
        {
          name: '平均订单频次',
          value: behavior.map(item => item.avg_order_frequency),
          itemStyle: { color: '#52c41a' },
        },
      ],
    };
  }, [customerData]);

  // 客户RFM分析热力图数据
  const rfmAnalysisData = useMemo(() => {
    if (!customerData?.rfm_analysis) return null;

    const rfm = customerData.rfm_analysis;
    // 构造热力图数据：[recency_index, frequency_index, monetary_value]
    return {
      data: rfm.map(item => [
        item.recency_score,
        item.frequency_score,
        item.monetary_score
      ]),
      categories: {
        xAxis: ['R1', 'R2', 'R3', 'R4', 'R5'],
        yAxis: ['F1', 'F2', 'F3', 'F4', 'F5'],
      }
    };
  }, [customerData]);

  // 地域分布数据
  const regionDistributionData = useMemo(() => {
    if (!customerData?.region_distribution) return null;

    return customerData.region_distribution
      .slice(0, 10) // 显示前10个地区
      .map((region, index) => ({
        name: region.region_name,
        value: region.customer_count,
        itemStyle: {
          color: [
            '#1890ff', '#52c41a', '#fadb14', '#f5222d', '#722ed1',
            '#fa8c16', '#13c2c2', '#eb2f96', '#fa541c', '#2f54eb'
          ][index % 10],
        },
      }));
  }, [customerData]);

  const formatNumber = (num: number): string => {
    if (num >= 10000) {
      return (num / 10000).toFixed(1) + '万';
    }
    return num.toLocaleString();
  };

  const getTrendBadge = (trend: number) => {
    if (trend > 0) {
      return (
        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
          <svg className="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clipRule="evenodd" />
          </svg>
          +{trend.toFixed(1)}%
        </span>
      );
    } else if (trend < 0) {
      return (
        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800">
          <svg className="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clipRule="evenodd" />
          </svg>
          {trend.toFixed(1)}%
        </span>
      );
    } else {
      return (
        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
          0%
        </span>
      );
    }
  };

  return (
    <div className={`customer-analytics-page ${className}`}>
      <ReportLayout
        title="客户行为分析"
        subtitle="深度分析用户活跃度、留存率和消费行为模式"
        reportType="customer_behavior"
        loading={analyticsLoading}
        error={analyticsError}
        onDateRangeChange={handleDateRangeChange}
        onRefresh={handleRefresh}
        initialDateRange={dateRange}
        showExporter={true}
        showDownloadList={true}
      >
        {customerData && (
          <div className="space-y-6">
            {/* 用户活跃度指标卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {userActivityMetrics.map((metric) => (
                <div key={metric.title} className={`${metric.bgColor} rounded-lg p-6`}>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center">
                      <div className={`flex items-center justify-center h-12 w-12 rounded-md ${metric.color}`}>
                        <span className="text-2xl">{metric.icon}</span>
                      </div>
                      <div className="ml-4">
                        <p className="text-sm font-medium text-gray-600">{metric.title}</p>
                        <p className={`text-2xl font-semibold ${metric.color}`}>
                          {formatNumber(metric.value)} {metric.unit}
                        </p>
                      </div>
                    </div>
                    <div className="flex flex-col items-end">
                      {getTrendBadge(metric.trend)}
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* 用户活跃度趋势图 */}
            {activityTrendData && (
              <div className="bg-white rounded-lg shadow p-6">
                <LineChart
                  data={activityTrendData.series}
                  xAxisData={activityTrendData.xAxisData}
                  title="用户活跃度趋势"
                  subtitle="日、周、月活跃用户数量变化"
                  yAxisUnit="人"
                  height={400}
                  smooth={true}
                  showSymbol={true}
                />
              </div>
            )}

            {/* 留存分析和消费行为对比 */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {retentionData && (
                <div className="bg-white rounded-lg shadow p-6">
                  <PieChart
                    data={retentionData}
                    title="用户留存分析"
                    subtitle="不同时期用户留存率对比"
                    height={400}
                    radius={['30%', '70%']}
                    showLabel={true}
                    showLegend={true}
                  />
                </div>
              )}

              {consumptionBehaviorData && (
                <div className="bg-white rounded-lg shadow p-6">
                  <BarChart
                    data={consumptionBehaviorData.series}
                    xAxisData={consumptionBehaviorData.xAxisData}
                    title="用户消费行为分析"
                    subtitle="不同用户群体的消费特征"
                    yAxisUnit="元/次"
                    height={400}
                    horizontal={false}
                  />
                </div>
              )}
            </div>

            {/* RFM分析和地域分布 */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {rfmAnalysisData && (
                <div className="bg-white rounded-lg shadow p-6">
                  <HeatmapChart
                    data={rfmAnalysisData.data}
                    xAxisData={rfmAnalysisData.categories.xAxis}
                    yAxisData={rfmAnalysisData.categories.yAxis}
                    title="RFM客户价值分析"
                    subtitle="客户最近消费、消费频次、消费金额热力图"
                    height={400}
                  />
                </div>
              )}

              {regionDistributionData && (
                <div className="bg-white rounded-lg shadow p-6">
                  <PieChart
                    data={regionDistributionData}
                    title="用户地域分布"
                    subtitle="前10个地区的用户分布情况"
                    height={400}
                    radius={['40%', '70%']}
                    showLabel={true}
                    showLegend={true}
                  />
                </div>
              )}
            </div>

            {/* 用户行为详细表格 */}
            {customerData.consumption_behavior && customerData.consumption_behavior.length > 0 && (
              <div className="bg-white rounded-lg shadow overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-200">
                  <h3 className="text-lg font-medium text-gray-900">用户群体消费详情</h3>
                </div>
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          用户群体
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          用户数量
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          平均订单金额
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          平均订单频次
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          客户生命周期价值
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {customerData.consumption_behavior.map((segment) => (
                        <tr key={segment.customer_segment} className="hover:bg-gray-50">
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                            {segment.customer_segment}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {segment.customer_count} 人
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            ¥{segment.avg_order_amount.amount.toLocaleString()}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {segment.avg_order_frequency.toFixed(1)} 次/月
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            ¥{segment.lifetime_value.amount.toLocaleString()}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            {/* 留存率详细分析表 */}
            {customerData.retention_analysis && (
              <div className="bg-white rounded-lg shadow overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-200">
                  <h3 className="text-lg font-medium text-gray-900">用户留存率详细分析</h3>
                </div>
                <div className="p-6">
                  <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                    <div className="text-center">
                      <div className="text-3xl font-bold text-blue-600">
                        {customerData.retention_analysis.day_1_retention.toFixed(1)}%
                      </div>
                      <div className="text-sm text-gray-600">1日留存率</div>
                      <div className="text-xs text-gray-400 mt-1">
                        新用户次日回访率
                      </div>
                    </div>
                    <div className="text-center">
                      <div className="text-3xl font-bold text-green-600">
                        {customerData.retention_analysis.day_7_retention.toFixed(1)}%
                      </div>
                      <div className="text-sm text-gray-600">7日留存率</div>
                      <div className="text-xs text-gray-400 mt-1">
                        新用户一周内活跃率
                      </div>
                    </div>
                    <div className="text-center">
                      <div className="text-3xl font-bold text-yellow-600">
                        {customerData.retention_analysis.day_30_retention.toFixed(1)}%
                      </div>
                      <div className="text-sm text-gray-600">30日留存率</div>
                      <div className="text-xs text-gray-400 mt-1">
                        新用户一月内活跃率
                      </div>
                    </div>
                    <div className="text-center">
                      <div className="text-3xl font-bold text-red-600">
                        {customerData.retention_analysis.day_90_retention.toFixed(1)}%
                      </div>
                      <div className="text-sm text-gray-600">90日留存率</div>
                      <div className="text-xs text-gray-400 mt-1">
                        新用户三月内活跃率
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        )}
      </ReportLayout>
    </div>
  );
};

export default CustomerAnalyticsPage;