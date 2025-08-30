import React, { useEffect, useMemo } from 'react';
import { 
  ReportLayout, 
  LineChart, 
  BarChart, 
  PieChart, 
  DateRange 
} from '../../components/reports';
import { useReportStore } from '../../stores/reportStore';
import { MerchantOperationReport } from '../../services/reportService';

interface MerchantAnalyticsPageProps {
  className?: string;
}

const MerchantAnalyticsPage: React.FC<MerchantAnalyticsPageProps> = ({
  className = '',
}) => {
  const {
    merchantData,
    analyticsLoading,
    analyticsError,
    dateRange,
    setDateRange,
    fetchMerchantData,
  } = useReportStore();

  // 初始加载数据
  useEffect(() => {
    fetchMerchantData();
  }, [fetchMerchantData]);

  const handleDateRangeChange = (newDateRange: DateRange) => {
    setDateRange(newDateRange.startDate, newDateRange.endDate);
    fetchMerchantData(newDateRange.startDate, newDateRange.endDate);
  };

  const handleRefresh = () => {
    fetchMerchantData();
  };

  // 商户排行榜数据处理
  const topMerchants = useMemo(() => {
    if (!merchantData?.merchant_rankings) return [];
    return merchantData.merchant_rankings.slice(0, 10);
  }, [merchantData]);

  // 商户收入排行柱状图数据
  const merchantRankingChartData = useMemo(() => {
    if (!topMerchants.length) return null;

    return {
      xAxisData: topMerchants.map(m => m.merchant_name),
      series: [
        {
          name: '收入金额',
          value: topMerchants.map(m => m.total_revenue.amount),
          itemStyle: { color: '#1890ff' },
        },
      ],
    };
  }, [topMerchants]);

  // 客单价对比图数据
  const averageOrderValueData = useMemo(() => {
    if (!topMerchants.length) return null;

    return {
      xAxisData: topMerchants.map(m => m.merchant_name),
      series: [
        {
          name: '客单价',
          value: topMerchants.map(m => m.average_order_value.amount),
          itemStyle: { color: '#52c41a' },
        },
      ],
    };
  }, [topMerchants]);

  // 类别分析饼图数据
  const categoryAnalysisData = useMemo(() => {
    if (!merchantData?.category_analysis) return null;

    return merchantData.category_analysis
      .slice(0, 8) // 显示前8个类别
      .map((category, index) => ({
        name: category.category_name,
        value: category.revenue.amount,
        itemStyle: {
          color: [
            '#1890ff', '#52c41a', '#fadb14', '#f5222d', '#722ed1',
            '#fa8c16', '#13c2c2', '#eb2f96'
          ][index % 8],
        },
      }));
  }, [merchantData]);

  // 商户增长趋势数据（如果有的话）
  const growthTrendData = useMemo(() => {
    if (!merchantData?.performance_trends) return null;

    // 选择前5名商户的趋势数据
    const topTrendMerchants = merchantData.performance_trends.slice(0, 5);
    if (!topTrendMerchants.length || !topTrendMerchants[0].trend_data.length) return null;

    return {
      xAxisData: topTrendMerchants[0].trend_data.map(item => item.month),
      series: topTrendMerchants.map((merchant, index) => ({
        name: merchant.merchant_name,
        value: merchant.trend_data.map(item => item.revenue.amount),
        itemStyle: {
          color: ['#1890ff', '#52c41a', '#fadb14', '#f5222d', '#722ed1'][index % 5],
        },
      })),
    };
  }, [merchantData]);

  // 市场份额分析数据
  const marketShareData = useMemo(() => {
    if (!merchantData?.category_analysis) return null;

    return {
      xAxisData: merchantData.category_analysis.map(c => c.category_name),
      series: [
        {
          name: '市场份额',
          value: merchantData.category_analysis.map(c => c.market_share),
          itemStyle: { color: '#722ed1' },
        },
      ],
    };
  }, [merchantData]);

  const formatNumber = (num: number): string => {
    if (num >= 10000) {
      return (num / 10000).toFixed(1) + '万';
    }
    return num.toLocaleString();
  };

  const getGrowthBadge = (growthRate: number) => {
    if (growthRate > 0) {
      return (
        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
          <svg className="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clipRule="evenodd" />
          </svg>
          +{growthRate.toFixed(1)}%
        </span>
      );
    } else if (growthRate < 0) {
      return (
        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800">
          <svg className="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clipRule="evenodd" />
          </svg>
          {growthRate.toFixed(1)}%
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
    <div className={`merchant-analytics-page ${className}`}>
      <ReportLayout
        title="商户运营分析"
        subtitle="查看各商户业绩排名、趋势分析和类别表现"
        reportType="merchant_operation"
        loading={analyticsLoading}
        error={analyticsError}
        onDateRangeChange={handleDateRangeChange}
        onRefresh={handleRefresh}
        initialDateRange={dateRange}
        showExporter={true}
        showDownloadList={true}
      >
        {merchantData && (
          <div className="space-y-6">
            {/* 概览卡片 */}
            {merchantData.growth_metrics && (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <div className="bg-blue-50 rounded-lg p-6">
                  <div className="flex items-center">
                    <div className="flex items-center justify-center h-12 w-12 rounded-md text-blue-600">
                      <span className="text-2xl">💰</span>
                    </div>
                    <div className="ml-4">
                      <p className="text-sm font-medium text-gray-600">收入增长率</p>
                      <p className="text-2xl font-semibold text-blue-600">
                        {merchantData.growth_metrics.revenue_growth_rate.toFixed(1)}%
                      </p>
                    </div>
                  </div>
                </div>

                <div className="bg-green-50 rounded-lg p-6">
                  <div className="flex items-center">
                    <div className="flex items-center justify-center h-12 w-12 rounded-md text-green-600">
                      <span className="text-2xl">📦</span>
                    </div>
                    <div className="ml-4">
                      <p className="text-sm font-medium text-gray-600">订单增长率</p>
                      <p className="text-2xl font-semibold text-green-600">
                        {merchantData.growth_metrics.order_growth_rate.toFixed(1)}%
                      </p>
                    </div>
                  </div>
                </div>

                <div className="bg-purple-50 rounded-lg p-6">
                  <div className="flex items-center">
                    <div className="flex items-center justify-center h-12 w-12 rounded-md text-purple-600">
                      <span className="text-2xl">🏪</span>
                    </div>
                    <div className="ml-4">
                      <p className="text-sm font-medium text-gray-600">商户增长率</p>
                      <p className="text-2xl font-semibold text-purple-600">
                        {merchantData.growth_metrics.merchant_growth_rate.toFixed(1)}%
                      </p>
                    </div>
                  </div>
                </div>

                <div className="bg-orange-50 rounded-lg p-6">
                  <div className="flex items-center">
                    <div className="flex items-center justify-center h-12 w-12 rounded-md text-orange-600">
                      <span className="text-2xl">👥</span>
                    </div>
                    <div className="ml-4">
                      <p className="text-sm font-medium text-gray-600">客户增长率</p>
                      <p className="text-2xl font-semibold text-orange-600">
                        {merchantData.growth_metrics.customer_growth_rate.toFixed(1)}%
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* 商户收入排行图表 */}
            {merchantRankingChartData && (
              <div className="bg-white rounded-lg shadow p-6">
                <BarChart
                  data={merchantRankingChartData.series}
                  xAxisData={merchantRankingChartData.xAxisData}
                  title="商户收入排行榜"
                  subtitle="前10名商户收入对比"
                  yAxisUnit="元"
                  height={400}
                  horizontal={false}
                />
              </div>
            )}

            {/* 客单价和类别分析对比 */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {averageOrderValueData && (
                <div className="bg-white rounded-lg shadow p-6">
                  <BarChart
                    data={averageOrderValueData.series}
                    xAxisData={averageOrderValueData.xAxisData}
                    title="商户客单价对比"
                    subtitle="前10名商户平均客单价"
                    yAxisUnit="元"
                    height={400}
                    horizontal={true}
                  />
                </div>
              )}

              {categoryAnalysisData && (
                <div className="bg-white rounded-lg shadow p-6">
                  <PieChart
                    data={categoryAnalysisData}
                    title="类别收入分布"
                    subtitle="各商品类别收入占比"
                    height={400}
                    radius={['40%', '70%']}
                    showLabel={true}
                    showLegend={true}
                  />
                </div>
              )}
            </div>

            {/* 增长趋势和市场份额 */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {growthTrendData && (
                <div className="bg-white rounded-lg shadow p-6">
                  <LineChart
                    data={growthTrendData.series}
                    xAxisData={growthTrendData.xAxisData}
                    title="TOP5商户增长趋势"
                    subtitle="前5名商户月度收入变化"
                    yAxisUnit="元"
                    height={400}
                    smooth={true}
                    showSymbol={true}
                  />
                </div>
              )}

              {marketShareData && (
                <div className="bg-white rounded-lg shadow p-6">
                  <BarChart
                    data={marketShareData.series}
                    xAxisData={marketShareData.xAxisData}
                    title="类别市场份额"
                    subtitle="各类别在市场中的占有率"
                    yAxisUnit="%"
                    height={400}
                    horizontal={false}
                  />
                </div>
              )}
            </div>

            {/* 商户排行详细表格 */}
            {topMerchants.length > 0 && (
              <div className="bg-white rounded-lg shadow overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-200">
                  <h3 className="text-lg font-medium text-gray-900">商户业绩排行榜</h3>
                </div>
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          排名
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          商户名称
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          总收入
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          订单数
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          客户数
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          客单价
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          增长率
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {topMerchants.map((merchant) => (
                        <tr key={merchant.merchant_id} className="hover:bg-gray-50">
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                            #{merchant.rank}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                            {merchant.merchant_name}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            ¥{merchant.total_revenue.amount.toLocaleString()}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {merchant.order_count}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {merchant.customer_count}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            ¥{merchant.average_order_value.amount.toLocaleString()}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {getGrowthBadge(merchant.growth_rate)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            {/* 类别分析详细表格 */}
            {merchantData.category_analysis && merchantData.category_analysis.length > 0 && (
              <div className="bg-white rounded-lg shadow overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-200">
                  <h3 className="text-lg font-medium text-gray-900">商品类别分析</h3>
                </div>
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          类别名称
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          收入金额
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          订单数量
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          商户数量
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          市场份额
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          增长率
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {merchantData.category_analysis.map((category) => (
                        <tr key={category.category_id} className="hover:bg-gray-50">
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                            {category.category_name}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            ¥{category.revenue.amount.toLocaleString()}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {category.order_count}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {category.merchant_count}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {category.market_share.toFixed(2)}%
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {getGrowthBadge(category.growth_rate)}
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

export default MerchantAnalyticsPage;