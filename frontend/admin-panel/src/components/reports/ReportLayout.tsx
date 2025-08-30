import React, { useState } from 'react';
import { DateRangePicker, DateRange } from './DateRangePicker';
import { ReportExporter, ReportDownloadList } from './ReportExporter';
import { ReportType } from '../../services/reportService';

export interface ReportLayoutProps {
  title: string;
  subtitle?: string;
  reportType: ReportType;
  children: React.ReactNode;
  className?: string;
  loading?: boolean;
  error?: string | null;
  onDateRangeChange: (dateRange: DateRange) => void;
  onRefresh?: () => void;
  initialDateRange?: DateRange;
  showExporter?: boolean;
  showDownloadList?: boolean;
  extraActions?: React.ReactNode;
  merchantId?: number;
}

export const ReportLayout: React.FC<ReportLayoutProps> = ({
  title,
  subtitle,
  reportType,
  children,
  className = '',
  loading = false,
  error = null,
  onDateRangeChange,
  onRefresh,
  initialDateRange,
  showExporter = true,
  showDownloadList = true,
  extraActions,
  merchantId,
}) => {
  const [currentDateRange, setCurrentDateRange] = useState<DateRange>(
    initialDateRange || {
      startDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
      endDate: new Date().toISOString().split('T')[0],
    }
  );
  
  const [showSidebar, setShowSidebar] = useState(false);
  const [exportSuccess, setExportSuccess] = useState<string | null>(null);
  const [exportError, setExportError] = useState<string | null>(null);

  const handleDateRangeChange = (dateRange: DateRange) => {
    setCurrentDateRange(dateRange);
    onDateRangeChange(dateRange);
  };

  const handleRefresh = () => {
    onRefresh?.();
    // 清除提示消息
    setExportSuccess(null);
    setExportError(null);
  };

  const handleExportStart = () => {
    setExportSuccess(null);
    setExportError(null);
  };

  const handleExportComplete = (reportUuid: string) => {
    setExportSuccess(`报表生成成功！报表ID: ${reportUuid}`);
    // 3秒后自动清除成功消息
    setTimeout(() => setExportSuccess(null), 3000);
  };

  const handleExportError = (errorMessage: string) => {
    setExportError(errorMessage);
    // 5秒后自动清除错误消息
    setTimeout(() => setExportError(null), 5000);
  };

  return (
    <div className={`report-layout ${className}`}>
      {/* 页面标题栏 */}
      <div className="report-header bg-white border-b border-gray-200 px-6 py-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">{title}</h1>
            {subtitle && (
              <p className="mt-1 text-sm text-gray-600">{subtitle}</p>
            )}
          </div>
          
          <div className="flex items-center space-x-3">
            {extraActions}
            
            {onRefresh && (
              <button
                type="button"
                onClick={handleRefresh}
                disabled={loading}
                className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
              >
                <svg 
                  className={`-ml-0.5 mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} 
                  xmlns="http://www.w3.org/2000/svg" 
                  fill="none" 
                  viewBox="0 0 24 24"
                >
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                刷新
              </button>
            )}
            
            {(showExporter || showDownloadList) && (
              <button
                type="button"
                onClick={() => setShowSidebar(!showSidebar)}
                className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                <span className="mr-2">📊</span>
                报表管理
              </button>
            )}
          </div>
        </div>

        {/* 错误或成功消息 */}
        {error && (
          <div className="mt-4 rounded-md bg-red-50 p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">
                  加载失败
                </h3>
                <div className="mt-2 text-sm text-red-700">
                  {error}
                </div>
              </div>
            </div>
          </div>
        )}

        {exportSuccess && (
          <div className="mt-4 rounded-md bg-green-50 p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-green-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <div className="text-sm text-green-700">
                  {exportSuccess}
                </div>
              </div>
            </div>
          </div>
        )}

        {exportError && (
          <div className="mt-4 rounded-md bg-red-50 p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <div className="text-sm text-red-700">
                  {exportError}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="report-content flex">
        {/* 主内容区域 */}
        <div className={`flex-1 ${showSidebar ? 'mr-80' : ''} transition-all duration-300`}>
          {/* 日期选择器 */}
          <div className="bg-white border-b border-gray-200 px-6 py-4">
            <DateRangePicker
              value={currentDateRange}
              onChange={handleDateRangeChange}
              maxRange={365}
              quickRanges={true}
              disabled={loading}
            />
          </div>

          {/* 报表内容 */}
          <div className="p-6">
            {loading ? (
              <div className="flex items-center justify-center h-64">
                <div className="text-center">
                  <svg className="animate-spin mx-auto h-12 w-12 text-gray-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  <div className="mt-2 text-sm text-gray-500">加载中...</div>
                </div>
              </div>
            ) : (
              children
            )}
          </div>
        </div>

        {/* 侧边栏 */}
        {showSidebar && (
          <div className="fixed right-0 top-0 bottom-0 w-80 bg-white border-l border-gray-200 shadow-lg z-10 overflow-y-auto">
            <div className="p-4">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-gray-900">报表管理</h3>
                <button
                  type="button"
                  onClick={() => setShowSidebar(false)}
                  className="text-gray-400 hover:text-gray-500"
                >
                  <svg className="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              {showExporter && (
                <div className="mb-6">
                  <h4 className="text-md font-medium text-gray-800 mb-3">导出报表</h4>
                  <ReportExporter
                    reportType={reportType}
                    startDate={currentDateRange.startDate}
                    endDate={currentDateRange.endDate}
                    merchantId={merchantId}
                    onExportStart={handleExportStart}
                    onExportComplete={handleExportComplete}
                    onExportError={handleExportError}
                  />
                </div>
              )}

              {showDownloadList && (
                <div>
                  <ReportDownloadList maxItems={5} />
                </div>
              )}
            </div>
          </div>
        )}

        {/* 侧边栏遮罩 */}
        {showSidebar && (
          <div
            className="fixed inset-0 bg-black bg-opacity-25 z-5"
            onClick={() => setShowSidebar(false)}
          />
        )}
      </div>
    </div>
  );
};