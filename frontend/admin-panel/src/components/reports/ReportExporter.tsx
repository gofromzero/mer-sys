import React, { useState } from 'react';
import { reportService, ReportType, FileFormat, PeriodType } from '../../services/reportService';
import { useReportStore } from '../../stores/reportStore';

export interface ReportExportProps {
  reportType: ReportType;
  startDate: string;
  endDate: string;
  merchantId?: number;
  className?: string;
  disabled?: boolean;
  onExportStart?: () => void;
  onExportComplete?: (reportUuid: string) => void;
  onExportError?: (error: string) => void;
}

// 导出按钮组件
export const ReportExporter: React.FC<ReportExportProps> = ({
  reportType,
  startDate,
  endDate,
  merchantId,
  className = '',
  disabled = false,
  onExportStart,
  onExportComplete,
  onExportError,
}) => {
  const [isExporting, setIsExporting] = useState(false);
  const [exportFormat, setExportFormat] = useState<FileFormat>('excel');
  const { generateReport } = useReportStore();

  const handleExport = async () => {
    if (isExporting || disabled) return;

    try {
      setIsExporting(true);
      onExportStart?.();

      const report = await generateReport({
        report_type: reportType,
        period_type: 'custom' as PeriodType,
        start_date: startDate,
        end_date: endDate,
        file_format: exportFormat,
        merchant_id: merchantId,
      });

      onExportComplete?.(report.uuid);
    } catch (error: any) {
      const errorMessage = error.message || '导出失败';
      onExportError?.(errorMessage);
      console.error('Report export failed:', error);
    } finally {
      setIsExporting(false);
    }
  };

  const formatLabels: Record<FileFormat, string> = {
    excel: 'Excel',
    pdf: 'PDF',
    json: 'JSON',
  };

  const formatIcons: Record<FileFormat, string> = {
    excel: '📊',
    pdf: '📄',
    json: '📝',
  };

  return (
    <div className={`report-exporter ${className}`}>
      <div className="export-format-selector mb-2">
        <label className="block text-sm font-medium text-gray-700 mb-1">
          导出格式
        </label>
        <div className="flex space-x-2">
          {(['excel', 'pdf', 'json'] as FileFormat[]).map((format) => (
            <button
              key={format}
              type="button"
              onClick={() => setExportFormat(format)}
              className={`
                inline-flex items-center px-3 py-2 border text-sm font-medium rounded-md
                ${exportFormat === format
                  ? 'border-blue-500 text-blue-700 bg-blue-50'
                  : 'border-gray-300 text-gray-700 bg-white hover:bg-gray-50'
                }
                ${disabled || isExporting ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
              `}
              disabled={disabled || isExporting}
            >
              <span className="mr-1">{formatIcons[format]}</span>
              {formatLabels[format]}
            </button>
          ))}
        </div>
      </div>

      <button
        type="button"
        onClick={handleExport}
        disabled={disabled || isExporting}
        className={`
          inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md
          text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500
          ${disabled || isExporting ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
        `}
      >
        {isExporting ? (
          <>
            <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            导出中...
          </>
        ) : (
          <>
            <span className="mr-1">📤</span>
            导出{formatLabels[exportFormat]}报表
          </>
        )}
      </button>
    </div>
  );
};

// 报表下载列表组件
export const ReportDownloadList: React.FC<{
  className?: string;
  maxItems?: number;
}> = ({
  className = '',
  maxItems = 10,
}) => {
  const { 
    reports, 
    reportListLoading, 
    reportListError, 
    fetchReports, 
    downloadReport,
    deleteReport 
  } = useReportStore();
  
  const [downloadingReports, setDownloadingReports] = useState<Set<string>>(new Set());

  React.useEffect(() => {
    fetchReports({ page: 1, page_size: maxItems });
  }, [fetchReports, maxItems]);

  const handleDownload = async (uuid: string, reportName: string) => {
    if (downloadingReports.has(uuid)) return;

    try {
      setDownloadingReports(prev => new Set(prev).add(uuid));
      await downloadReport(uuid);
    } catch (error: any) {
      console.error('Download failed:', error);
      // 这里可以添加错误提示
    } finally {
      setDownloadingReports(prev => {
        const newSet = new Set(prev);
        newSet.delete(uuid);
        return newSet;
      });
    }
  };

  const handleDelete = async (id: number, reportName: string) => {
    if (confirm(`确定要删除报表 "${reportName}" 吗？`)) {
      try {
        await deleteReport(id);
      } catch (error: any) {
        console.error('Delete failed:', error);
        // 这里可以添加错误提示
      }
    }
  };

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      completed: { text: '已完成', class: 'bg-green-100 text-green-800' },
      generating: { text: '生成中', class: 'bg-yellow-100 text-yellow-800' },
      failed: { text: '失败', class: 'bg-red-100 text-red-800' },
    };
    
    const config = statusConfig[status as keyof typeof statusConfig] || 
      { text: status, class: 'bg-gray-100 text-gray-800' };
    
    return (
      <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${config.class}`}>
        {config.text}
      </span>
    );
  };

  const getReportTypeLabel = (type: string) => {
    const typeLabels = {
      financial: '财务报表',
      merchant_operation: '商户运营',
      customer_analysis: '客户分析',
    };
    return typeLabels[type as keyof typeof typeLabels] || type;
  };

  const getFileFormatIcon = (format: string) => {
    const formatIcons = {
      excel: '📊',
      pdf: '📄',
      json: '📝',
    };
    return formatIcons[format as keyof typeof formatIcons] || '📄';
  };

  if (reportListLoading) {
    return (
      <div className={`report-download-list ${className}`}>
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded mb-2"></div>
          <div className="h-4 bg-gray-200 rounded mb-2"></div>
          <div className="h-4 bg-gray-200 rounded"></div>
        </div>
      </div>
    );
  }

  if (reportListError) {
    return (
      <div className={`report-download-list ${className}`}>
        <div className="text-red-600 text-sm">
          加载报表列表失败: {reportListError}
        </div>
      </div>
    );
  }

  return (
    <div className={`report-download-list ${className}`}>
      <h3 className="text-lg font-medium text-gray-900 mb-4">最近生成的报表</h3>
      
      {reports.length === 0 ? (
        <div className="text-gray-500 text-center py-4">
          暂无报表记录
        </div>
      ) : (
        <div className="space-y-2">
          {reports.slice(0, maxItems).map((report) => (
            <div key={report.id} className="flex items-center justify-between p-3 border border-gray-200 rounded-md">
              <div className="flex-1 min-w-0">
                <div className="flex items-center space-x-2 mb-1">
                  <span className="text-sm font-medium text-gray-900">
                    {getFileFormatIcon(report.file_format)} {getReportTypeLabel(report.report_type)}
                  </span>
                  {getStatusBadge(report.status)}
                </div>
                <div className="text-sm text-gray-500">
                  {report.start_date} ~ {report.end_date}
                </div>
                <div className="text-xs text-gray-400">
                  生成时间: {new Date(report.created_at).toLocaleString()}
                </div>
              </div>
              
              <div className="flex items-center space-x-2 ml-4">
                {report.status === 'completed' && (
                  <button
                    type="button"
                    onClick={() => handleDownload(report.uuid, getReportTypeLabel(report.report_type))}
                    disabled={downloadingReports.has(report.uuid)}
                    className="inline-flex items-center px-2 py-1 border border-gray-300 shadow-sm text-xs font-medium rounded text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
                  >
                    {downloadingReports.has(report.uuid) ? (
                      <>
                        <svg className="animate-spin -ml-1 mr-1 h-3 w-3 text-gray-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        下载中
                      </>
                    ) : (
                      <>
                        <span className="mr-1">⬇️</span>
                        下载
                      </>
                    )}
                  </button>
                )}
                
                <button
                  type="button"
                  onClick={() => handleDelete(report.id, getReportTypeLabel(report.report_type))}
                  className="inline-flex items-center px-2 py-1 border border-gray-300 shadow-sm text-xs font-medium rounded text-red-700 bg-white hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                >
                  <span className="mr-1">🗑️</span>
                  删除
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};