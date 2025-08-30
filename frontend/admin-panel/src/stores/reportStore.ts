import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import {
  reportService,
  Report,
  ReportListRequest,
  ReportListResponse,
  FinancialReportData,
  MerchantOperationReport,
  CustomerAnalysisReport,
  ReportType,
  ReportStatus,
} from '../services/reportService';

interface ReportState {
  // 报表列表状态
  reports: Report[];
  reportListLoading: boolean;
  reportListError: string | null;
  reportListPagination: {
    current: number;
    pageSize: number;
    total: number;
    hasNext: boolean;
  };

  // 当前查看的报表
  currentReport: Report | null;
  currentReportLoading: boolean;

  // 分析数据状态
  financialData: FinancialReportData | null;
  merchantData: MerchantOperationReport | null;
  customerData: CustomerAnalysisReport | null;
  analyticsLoading: boolean;
  analyticsError: string | null;

  // 筛选条件
  filters: {
    reportType?: ReportType;
    status?: ReportStatus;
    startDate?: string;
    endDate?: string;
    merchantId?: number;
  };

  // 时间范围
  dateRange: {
    startDate: string;
    endDate: string;
  };

  // Actions
  setFilters: (filters: Partial<ReportState['filters']>) => void;
  setDateRange: (startDate: string, endDate: string) => void;
  
  // 报表管理
  fetchReports: (request?: Partial<ReportListRequest>) => Promise<void>;
  fetchReportById: (id: number) => Promise<void>;
  generateReport: (request: {
    report_type: ReportType;
    period_type: string;
    start_date: string;
    end_date: string;
    file_format: string;
    merchant_id?: number;
  }) => Promise<Report>;
  deleteReport: (id: number) => Promise<void>;
  downloadReport: (uuid: string) => Promise<void>;
  
  // 分析数据
  fetchFinancialData: (startDate?: string, endDate?: string, merchantId?: number) => Promise<void>;
  fetchMerchantData: (startDate?: string, endDate?: string) => Promise<void>;
  fetchCustomerData: (startDate?: string, endDate?: string) => Promise<void>;
  
  // 清理状态
  clearCurrentReport: () => void;
  clearAnalyticsData: () => void;
  clearError: () => void;
}

export const useReportStore = create<ReportState>()(
  devtools(
    (set, get) => ({
      // 初始状态
      reports: [],
      reportListLoading: false,
      reportListError: null,
      reportListPagination: {
        current: 1,
        pageSize: 20,
        total: 0,
        hasNext: false,
      },

      currentReport: null,
      currentReportLoading: false,

      financialData: null,
      merchantData: null,
      customerData: null,
      analyticsLoading: false,
      analyticsError: null,

      filters: {},
      dateRange: {
        startDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0], // 30天前
        endDate: new Date().toISOString().split('T')[0], // 今天
      },

      // Actions
      setFilters: (newFilters) => {
        set((state) => ({
          filters: { ...state.filters, ...newFilters },
        }), false, 'setFilters');
      },

      setDateRange: (startDate, endDate) => {
        set({ dateRange: { startDate, endDate } }, false, 'setDateRange');
      },

      fetchReports: async (request) => {
        set({ reportListLoading: true, reportListError: null });
        
        try {
          const { filters } = get();
          const params: ReportListRequest = {
            page: 1,
            page_size: 20,
            ...request,
            ...filters,
          };

          const response = await reportService.listReports(params);
          
          set({
            reports: response.items,
            reportListPagination: {
              current: response.page,
              pageSize: response.page_size,
              total: response.total,
              hasNext: response.has_next,
            },
            reportListLoading: false,
          });
        } catch (error: any) {
          set({
            reportListLoading: false,
            reportListError: error.message || '获取报表列表失败',
          });
        }
      },

      fetchReportById: async (id) => {
        set({ currentReportLoading: true });
        
        try {
          const report = await reportService.getReport(id);
          set({ 
            currentReport: report,
            currentReportLoading: false,
          });
        } catch (error: any) {
          set({
            currentReportLoading: false,
            reportListError: error.message || '获取报表详情失败',
          });
        }
      },

      generateReport: async (request) => {
        try {
          const report = await reportService.generateReport(request);
          
          // 刷新报表列表
          get().fetchReports();
          
          return report;
        } catch (error: any) {
          set({ reportListError: error.message || '生成报表失败' });
          throw error;
        }
      },

      deleteReport: async (id) => {
        try {
          await reportService.deleteReport(id);
          
          // 从列表中移除该报表
          set((state) => ({
            reports: state.reports.filter(report => report.id !== id),
          }));
        } catch (error: any) {
          set({ reportListError: error.message || '删除报表失败' });
          throw error;
        }
      },

      downloadReport: async (uuid) => {
        try {
          const blob = await reportService.downloadReport(uuid);
          
          // 创建下载链接
          const url = window.URL.createObjectURL(blob);
          const link = document.createElement('a');
          link.href = url;
          link.download = `report_${uuid}.xlsx`;
          document.body.appendChild(link);
          link.click();
          document.body.removeChild(link);
          window.URL.revokeObjectURL(url);
        } catch (error: any) {
          set({ reportListError: error.message || '下载报表失败' });
          throw error;
        }
      },

      fetchFinancialData: async (startDate, endDate, merchantId) => {
        set({ analyticsLoading: true, analyticsError: null });
        
        try {
          const { dateRange } = get();
          const data = await reportService.getFinancialAnalytics(
            startDate || dateRange.startDate,
            endDate || dateRange.endDate,
            merchantId
          );
          
          set({
            financialData: data,
            analyticsLoading: false,
          });
        } catch (error: any) {
          set({
            analyticsLoading: false,
            analyticsError: error.message || '获取财务数据失败',
          });
        }
      },

      fetchMerchantData: async (startDate, endDate) => {
        set({ analyticsLoading: true, analyticsError: null });
        
        try {
          const { dateRange } = get();
          const data = await reportService.getMerchantAnalytics(
            startDate || dateRange.startDate,
            endDate || dateRange.endDate
          );
          
          set({
            merchantData: data,
            analyticsLoading: false,
          });
        } catch (error: any) {
          set({
            analyticsLoading: false,
            analyticsError: error.message || '获取商户数据失败',
          });
        }
      },

      fetchCustomerData: async (startDate, endDate) => {
        set({ analyticsLoading: true, analyticsError: null });
        
        try {
          const { dateRange } = get();
          const data = await reportService.getCustomerAnalytics(
            startDate || dateRange.startDate,
            endDate || dateRange.endDate
          );
          
          set({
            customerData: data,
            analyticsLoading: false,
          });
        } catch (error: any) {
          set({
            analyticsLoading: false,
            analyticsError: error.message || '获取客户数据失败',
          });
        }
      },

      clearCurrentReport: () => {
        set({ currentReport: null });
      },

      clearAnalyticsData: () => {
        set({
          financialData: null,
          merchantData: null,
          customerData: null,
          analyticsError: null,
        });
      },

      clearError: () => {
        set({
          reportListError: null,
          analyticsError: null,
        });
      },
    }),
    {
      name: 'report-store',
    }
  )
);