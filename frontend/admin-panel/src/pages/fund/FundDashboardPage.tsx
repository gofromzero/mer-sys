// 资金管理仪表板页面

import React, { useEffect, useState } from 'react';
import { useFundActions, useFundData, useFundLoading } from '../../stores/fundStore';
import type { FundSummary } from '../../types/fund';
import fundService from '../../services/fundService';

const FundDashboardPage: React.FC = () => {
  const actions = useFundActions();
  const { merchantList, summaryMap, balanceMap } = useFundData();
  const loading = useFundLoading();
  
  const [selectedMerchantId, setSelectedMerchantId] = useState<number | undefined>();
  const [overallSummary, setOverallSummary] = useState<FundSummary | null>(null);
  const [merchantSummary, setMerchantSummary] = useState<FundSummary | null>(null);
  
  useEffect(() => {
    actions.loadMerchantList();
    actions.getFundSummary().then(setOverallSummary);
  }, []);
  
  useEffect(() => {
    if (selectedMerchantId) {
      actions.getFundSummary(selectedMerchantId).then(setMerchantSummary);
      actions.getMerchantBalance(selectedMerchantId);
    } else {
      setMerchantSummary(null);
    }
  }, [selectedMerchantId, actions]);
  
  const selectedMerchant = merchantList.find(m => m.id === selectedMerchantId);
  const selectedBalance = selectedMerchantId ? balanceMap[selectedMerchantId] : null;
  
  return (
    <div className="fund-dashboard-page">
      <div className="page-header">
        <h1>资金管理总览</h1>
        <p>查看系统整体资金状况和商户权益分布</p>
      </div>
      
      <div className="dashboard-content">
        {/* 总体统计卡片 */}
        <div className="summary-cards">
          <div className="summary-card">
            <div className="card-header">
              <h3>总体资金统计</h3>
              <div className="refresh-btn">
                <button
                  onClick={() => actions.getFundSummary(undefined, true).then(setOverallSummary)}
                  disabled={loading.summary[0]}
                  className="btn-refresh"
                >
                  {loading.summary[0] ? '刷新中...' : '刷新'}
                </button>
              </div>
            </div>
            
            {overallSummary ? (
              <div className="card-content">
                <div className="stat-item">
                  <span className="label">充值总额</span>
                  <span className="value positive">
                    {fundService.formatAmount(overallSummary.total_deposits)}
                  </span>
                </div>
                <div className="stat-item">
                  <span className="label">分配总额</span>
                  <span className="value positive">
                    {fundService.formatAmount(overallSummary.total_allocations)}
                  </span>
                </div>
                <div className="stat-item">
                  <span className="label">消费总额</span>
                  <span className="value negative">
                    {fundService.formatAmount(overallSummary.total_consumption)}
                  </span>
                </div>
                <div className="stat-item">
                  <span className="label">退款总额</span>
                  <span className="value">
                    {fundService.formatAmount(overallSummary.total_refunds)}
                  </span>
                </div>
                <div className="stat-item highlight">
                  <span className="label">可用余额</span>
                  <span className="value">
                    {fundService.formatAmount(overallSummary.available_balance)}
                  </span>
                </div>
              </div>
            ) : (
              <div className="card-loading">
                <span>加载统计数据中...</span>
              </div>
            )}
          </div>
        </div>
        
        {/* 商户选择和详情 */}
        <div className="merchant-section">
          <div className="merchant-selector">
            <label>选择商户查看详情:</label>
            <select
              value={selectedMerchantId || ''}
              onChange={(e) => setSelectedMerchantId(e.target.value ? parseInt(e.target.value) : undefined)}
            >
              <option value="">请选择商户</option>
              {merchantList.map(merchant => (
                <option key={merchant.id} value={merchant.id}>
                  {merchant.name} ({merchant.code})
                </option>
              ))}
            </select>
          </div>
          
          {selectedMerchant && (
            <div className="merchant-details">
              <div className="merchant-info">
                <h3>{selectedMerchant.name} ({selectedMerchant.code})</h3>
                <div className="info-actions">
                  <button
                    onClick={() => {
                      if (selectedMerchantId) {
                        actions.getFundSummary(selectedMerchantId, true).then(setMerchantSummary);
                        actions.refreshMerchantBalance(selectedMerchantId);
                      }
                    }}
                    disabled={loading.summary[selectedMerchantId || 0] || loading.balance[selectedMerchantId || 0]}
                    className="btn-refresh"
                  >
                    刷新数据
                  </button>
                </div>
              </div>
              
              <div className="merchant-stats">
                <div className="stats-grid">
                  {/* 资金统计 */}
                  <div className="stats-card">
                    <h4>资金统计</h4>
                    {merchantSummary ? (
                      <div className="stats-content">
                        <div className="stat-item">
                          <span className="label">充值总额</span>
                          <span className="value positive">
                            {fundService.formatAmount(merchantSummary.total_deposits)}
                          </span>
                        </div>
                        <div className="stat-item">
                          <span className="label">分配总额</span>
                          <span className="value positive">
                            {fundService.formatAmount(merchantSummary.total_allocations)}
                          </span>
                        </div>
                        <div className="stat-item">
                          <span className="label">消费总额</span>
                          <span className="value negative">
                            {fundService.formatAmount(merchantSummary.total_consumption)}
                          </span>
                        </div>
                        <div className="stat-item">
                          <span className="label">退款总额</span>
                          <span className="value">
                            {fundService.formatAmount(merchantSummary.total_refunds)}
                          </span>
                        </div>
                      </div>
                    ) : (
                      <div className="loading-text">加载中...</div>
                    )}
                  </div>
                  
                  {/* 余额详情 */}
                  <div className="stats-card">
                    <h4>余额详情</h4>
                    {selectedBalance ? (
                      <div className="stats-content">
                        <div className="stat-item">
                          <span className="label">总余额</span>
                          <span className="value">
                            {fundService.formatAmount(selectedBalance.total_balance)}
                          </span>
                        </div>
                        <div className="stat-item">
                          <span className="label">已使用</span>
                          <span className="value negative">
                            {fundService.formatAmount(selectedBalance.used_balance)}
                          </span>
                        </div>
                        <div className="stat-item">
                          <span className="label">冻结余额</span>
                          <span className="value warning">
                            {fundService.formatAmount(selectedBalance.frozen_balance)}
                          </span>
                        </div>
                        <div className="stat-item highlight">
                          <span className="label">可用余额</span>
                          <span className="value">
                            {fundService.formatAmount(selectedBalance.available_balance)}
                          </span>
                        </div>
                        <div className="stat-item">
                          <span className="label">更新时间</span>
                          <span className="value small">
                            {new Date(selectedBalance.last_updated).toLocaleString()}
                          </span>
                        </div>
                      </div>
                    ) : (
                      <div className="loading-text">加载中...</div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
        
        {/* 快捷操作 */}
        <div className="quick-actions">
          <h3>快捷操作</h3>
          <div className="actions-grid">
            <a href="/fund/deposit" className="action-card">
              <div className="action-icon">💰</div>
              <div className="action-content">
                <h4>资金充值</h4>
                <p>为商户账户充值资金</p>
              </div>
            </a>
            
            <a href="/fund/allocation" className="action-card">
              <div className="action-icon">📊</div>
              <div className="action-content">
                <h4>权益分配</h4>
                <p>为商户分配权益额度</p>
              </div>
            </a>
            
            <a href="/fund/transactions" className="action-card">
              <div className="action-icon">📋</div>
              <div className="action-content">
                <h4>交易记录</h4>
                <p>查看资金流转历史</p>
              </div>
            </a>
            
            <a href="/fund/freeze" className="action-card">
              <div className="action-icon">🔒</div>
              <div className="action-content">
                <h4>权益冻结</h4>
                <p>冻结或解冻商户权益</p>
              </div>
            </a>
          </div>
        </div>
      </div>
      
      <style jsx>{`
        .fund-dashboard-page {
          padding: 20px;
        }
        
        .page-header {
          margin-bottom: 30px;
        }
        
        .page-header h1 {
          margin: 0 0 10px 0;
          font-size: 24px;
          font-weight: 600;
        }
        
        .page-header p {
          margin: 0;
          color: #666;
        }
        
        .dashboard-content {
          display: flex;
          flex-direction: column;
          gap: 30px;
        }
        
        .summary-cards {
          display: grid;
          grid-template-columns: 1fr;
          gap: 20px;
        }
        
        .summary-card {
          background: white;
          border-radius: 8px;
          padding: 25px;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        }
        
        .card-header {
          display: flex;
          justify-content: between;
          align-items: center;
          margin-bottom: 20px;
        }
        
        .card-header h3 {
          margin: 0;
          font-size: 18px;
          font-weight: 600;
        }
        
        .btn-refresh {
          padding: 6px 12px;
          border: 1px solid #ddd;
          border-radius: 4px;
          background: white;
          cursor: pointer;
          font-size: 12px;
        }
        
        .btn-refresh:hover:not(:disabled) {
          background: #f8f9fa;
        }
        
        .btn-refresh:disabled {
          cursor: not-allowed;
          opacity: 0.6;
        }
        
        .card-content {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
          gap: 20px;
        }
        
        .stat-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          padding: 15px;
          background: #f8f9fa;
          border-radius: 6px;
        }
        
        .stat-item.highlight {
          background: #e3f2fd;
          border: 2px solid #2196f3;
        }
        
        .stat-item .label {
          font-size: 14px;
          color: #666;
          margin-bottom: 8px;
        }
        
        .stat-item .value {
          font-size: 18px;
          font-weight: 600;
          text-align: center;
        }
        
        .stat-item .value.positive {
          color: #4caf50;
        }
        
        .stat-item .value.negative {
          color: #f44336;
        }
        
        .stat-item .value.warning {
          color: #ff9800;
        }
        
        .stat-item .value.small {
          font-size: 12px;
          font-weight: normal;
        }
        
        .card-loading,
        .loading-text {
          text-align: center;
          padding: 40px;
          color: #666;
        }
        
        .merchant-section {
          background: white;
          border-radius: 8px;
          padding: 25px;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        }
        
        .merchant-selector {
          margin-bottom: 20px;
        }
        
        .merchant-selector label {
          display: block;
          margin-bottom: 8px;
          font-weight: 500;
        }
        
        .merchant-selector select {
          width: 300px;
          padding: 8px 12px;
          border: 1px solid #ddd;
          border-radius: 4px;
        }
        
        .merchant-details {
          border-top: 1px solid #eee;
          padding-top: 20px;
        }
        
        .merchant-info {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 20px;
        }
        
        .merchant-info h3 {
          margin: 0;
          font-size: 16px;
          font-weight: 600;
        }
        
        .stats-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
          gap: 20px;
        }
        
        .stats-card {
          background: #f8f9fa;
          border-radius: 6px;
          padding: 20px;
        }
        
        .stats-card h4 {
          margin: 0 0 15px 0;
          font-size: 14px;
          font-weight: 600;
          color: #333;
        }
        
        .stats-content {
          display: flex;
          flex-direction: column;
          gap: 10px;
        }
        
        .stats-card .stat-item {
          flex-direction: row;
          justify-content: space-between;
          background: white;
          padding: 10px 15px;
        }
        
        .stats-card .stat-item .label {
          margin-bottom: 0;
        }
        
        .stats-card .stat-item .value {
          font-size: 14px;
        }
        
        .quick-actions {
          background: white;
          border-radius: 8px;
          padding: 25px;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        }
        
        .quick-actions h3 {
          margin: 0 0 20px 0;
          font-size: 18px;
          font-weight: 600;
        }
        
        .actions-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
          gap: 20px;
        }
        
        .action-card {
          display: flex;
          align-items: center;
          padding: 20px;
          border: 1px solid #e0e0e0;
          border-radius: 8px;
          text-decoration: none;
          color: inherit;
          transition: all 0.2s;
        }
        
        .action-card:hover {
          border-color: #2196f3;
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }
        
        .action-icon {
          font-size: 32px;
          margin-right: 15px;
        }
        
        .action-content h4 {
          margin: 0 0 5px 0;
          font-size: 16px;
          font-weight: 600;
        }
        
        .action-content p {
          margin: 0;
          font-size: 14px;
          color: #666;
        }
        
        @media (max-width: 768px) {
          .card-content {
            grid-template-columns: 1fr;
          }
          
          .stats-grid {
            grid-template-columns: 1fr;
          }
          
          .actions-grid {
            grid-template-columns: 1fr;
          }
          
          .merchant-selector select {
            width: 100%;
          }
          
          .merchant-info {
            flex-direction: column;
            align-items: flex-start;
            gap: 10px;
          }
        }
      `}</style>
    </div>
  );
};

export default FundDashboardPage;