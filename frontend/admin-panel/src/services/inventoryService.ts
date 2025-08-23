// 库存管理服务
import { 
  InventoryAdjustRequest, 
  BatchInventoryAdjustRequest,
  InventoryReserveRequest,
  InventoryAlertRequest,
  InventoryResponse,
  InventoryRecordResponse,
  InventoryMonitoringData,
  InventoryAlert,
  InventoryRecord,
  StocktakingCreateRequest,
  StocktakingListResponse,
  StocktakingResponse,
  StocktakingBatchUpdateRequest,
  StocktakingCompleteRequest
} from '../types/product';
import { api } from './api';

export class InventoryService {
  // 库存查询
  static async getInventory(productId: number): Promise<InventoryResponse> {
    const response = await api.get(`/products/${productId}/inventory`);
    return response.data;
  }

  static async getInventoryBatch(productIds: number[]): Promise<InventoryResponse[]> {
    const response = await api.post('/inventory/batch-query', { product_ids: productIds });
    return response.data;
  }

  // 库存调整
  static async adjustInventory(request: InventoryAdjustRequest): Promise<InventoryResponse> {
    const response = await api.post(`/products/${request.product_id}/inventory/adjust`, request);
    return response.data;
  }

  static async batchAdjustInventory(request: BatchInventoryAdjustRequest): Promise<InventoryResponse[]> {
    const response = await api.post('/inventory/batch-adjust', request);
    return response.data;
  }

  // 库存预留
  static async reserveInventory(request: InventoryReserveRequest): Promise<{ reservation_id: number }> {
    const response = await api.post(`/products/${request.product_id}/inventory/reserve`, request);
    return response.data;
  }

  static async releaseReservation(reservationId: number): Promise<void> {
    await api.post(`/inventory/reservations/${reservationId}/release`);
  }

  // 库存记录
  static async getInventoryRecords(params: {
    page?: number;
    page_size?: number;
    product_id?: number;
    change_type?: string;
    start_date?: string;
    end_date?: string;
  }): Promise<InventoryRecordResponse> {
    const response = await api.get('/inventory/records', { params });
    return response.data;
  }

  // 库存预警
  static async createAlert(productId: number, request: InventoryAlertRequest): Promise<InventoryAlert> {
    const response = await api.post(`/products/${productId}/inventory/alerts`, request);
    return response.data;
  }

  static async getActiveAlerts(): Promise<InventoryAlert[]> {
    const response = await api.get('/inventory/alerts/active');
    return response.data.alerts || [];
  }

  static async updateAlert(alertId: number, request: Partial<InventoryAlertRequest>): Promise<void> {
    await api.put(`/inventory/alerts/${alertId}`, request);
  }

  static async deleteAlert(alertId: number): Promise<void> {
    await api.delete(`/inventory/alerts/${alertId}`);
  }

  static async toggleAlert(alertId: number, isActive: boolean): Promise<void> {
    await api.post(`/inventory/alerts/${alertId}/toggle`, { is_active: isActive });
  }

  static async checkProductAlerts(productId: number): Promise<void> {
    await api.post(`/products/${productId}/inventory/alerts/check`);
  }

  static async checkAllLowStockAlerts(): Promise<void> {
    await api.post('/inventory/alerts/check-low-stock');
  }

  // 库存监控
  static async getMonitoringData(): Promise<InventoryMonitoringData> {
    const response = await api.get('/inventory/monitoring');
    return response.data;
  }

  static async getInventoryHealthAnalysis(): Promise<{
    health_score: number;
    health_level: string;
    stock_distribution: {
      normal_stock: number;
      low_stock: number;
      out_of_stock: number;
      low_stock_ratio: number;
      out_of_stock_ratio: number;
    };
    alert_analysis: object;
    operation_analysis: object;
  }> {
    const response = await api.get('/inventory/health-analysis');
    return response.data;
  }

  static async getLowStockSummary(): Promise<{
    total_low_stock: number;
    critical_products: Array<{
      id: number;
      name: string;
      current_stock: number;
      reserved_stock: number;
    }>;
    warning_products: Array<{
      id: number;
      name: string;
      current_stock: number;
      reserved_stock: number;
    }>;
    critical_count: number;
    warning_count: number;
  }> {
    const response = await api.get('/inventory/low-stock-summary');
    return response.data;
  }

  // 库存价值分析
  static async getInventoryValueAnalysis(): Promise<{
    total_inventory_value: number;
    currency: string;
    calculation_time: string;
    note: string;
  }> {
    const response = await api.get('/inventory/value-analysis');
    return response.data;
  }

  // 生成库存报告
  static async generateInventoryReport(startDate: string, endDate: string): Promise<{
    report_period: {
      start_date: string;
      end_date: string;
    };
    generated_at: string;
    current_statistics: object;
    health_analysis: object;
    audit_statistics: object;
  }> {
    const response = await api.post('/inventory/reports/generate', {
      start_date: startDate,
      end_date: endDate
    });
    return response.data;
  }

  // 库存盘点相关API
  static async createStocktaking(request: StocktakingCreateRequest): Promise<{ stocktaking_id: number; status: string }> {
    const response = await api.post('/inventory/stocktaking/start', request);
    return response.data;
  }

  static async getStocktakingList(params: {
    page?: number;
    page_size?: number;
    status?: string;
  }): Promise<StocktakingListResponse> {
    const response = await api.get('/inventory/stocktaking', { params });
    return response.data;
  }

  static async getStocktaking(stocktakingId: number): Promise<StocktakingResponse> {
    const response = await api.get(`/inventory/stocktaking/${stocktakingId}`);
    return response.data;
  }

  static async updateStocktakingRecords(stocktakingId: number, request: StocktakingBatchUpdateRequest): Promise<{ processed_records: number }> {
    const response = await api.put(`/inventory/stocktaking/${stocktakingId}/records`, request);
    return response.data;
  }

  static async completeStocktaking(stocktakingId: number, request: StocktakingCompleteRequest): Promise<{ stocktaking_id: number; status: string }> {
    const response = await api.post(`/inventory/stocktaking/${stocktakingId}/complete`, request);
    return response.data;
  }

  static async cancelStocktaking(stocktakingId: number): Promise<void> {
    await api.post(`/inventory/stocktaking/${stocktakingId}/cancel`);
  }

  // 导入导出功能
  static async importInventory(file: File, options: {
    import_mode?: 'update' | 'overwrite';
    validate_only?: boolean;
  }): Promise<{
    success_count: number;
    error_count: number;
    errors?: Array<{ row: number; message: string }>;
  }> {
    const formData = new FormData();
    formData.append('inventory_file', file);
    formData.append('import_mode', options.import_mode || 'update');
    formData.append('validate_only', String(options.validate_only || false));
    
    const response = await api.post('/inventory/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    });
    return response.data;
  }

  static async exportInventory(options: {
    export_format?: 'xlsx' | 'csv';
    export_scope?: 'all' | 'in_stock' | 'low_stock' | 'out_of_stock';
  }): Promise<Blob> {
    const response = await api.post('/inventory/export', options, {
      responseType: 'blob'
    });
    return response.data;
  }

  static async downloadImportTemplate(): Promise<Blob> {
    const response = await api.get('/inventory/template/download', {
      responseType: 'blob'
    });
    return response.data;
  }
}