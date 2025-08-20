// 权益分配页面

import React, { useEffect, useState } from 'react';
// import { toast } from 'react-hot-toast';
// 使用简单的alert替代toast，避免外部依赖
const toast = {
  success: (msg: string) => alert(`成功: ${msg}`),
  error: (msg: string) => alert(`错误: ${msg}`)
};
import { useFundActions, useFundData, useFundLoading, useFundError } from '../../stores/fundStore';
import type { AllocateRequest, RightsBalance } from '../../types/fund';
import fundService from '../../services/fundService';

const FundAllocationPage: React.FC = () => {
  const actions = useFundActions();
  const { merchantList, balanceMap } = useFundData();
  const loading = useFundLoading();
  const error = useFundError();
  
  // 分配表单状态
  const [form, setForm] = useState<AllocateRequest>({
    merchant_id: 0,
    amount: 0,
    description: '',
  });
  
  // 选中商户的余额信息
  const [selectedMerchantBalance, setSelectedMerchantBalance] = useState<RightsBalance | null>(null);
  
  useEffect(() => {
    actions.loadMerchantList();
  }, []);
  
  useEffect(() => {
    if (error) {
      toast.error(error);
      actions.clearError();
    }
  }, [error, actions]);
  
  // 当选择商户时，加载其余额信息
  useEffect(() => {
    if (form.merchant_id > 0) {
      actions.getMerchantBalance(form.merchant_id).then(balance => {
        setSelectedMerchantBalance(balance);
      }).catch(() => {
        setSelectedMerchantBalance(null);
      });
    } else {
      setSelectedMerchantBalance(null);
    }
  }, [form.merchant_id, actions]);
  
  // 处理权益分配
  const handleAllocate = async () => {
    const errors = fundService.validateAllocateRequest(form);
    if (errors.length > 0) {
      toast.error(errors[0]);
      return;
    }
    
    try {
      await actions.allocate(form);
      toast.success('权益分配成功');
      setForm({
        merchant_id: 0,
        amount: 0,
        description: '',
      });
      setSelectedMerchantBalance(null);
    } catch (error) {
      // Error already handled by store
    }
  };
  
  return (
    <div className="fund-allocation-page">
      <div className="page-header">
        <h1>权益分配</h1>
        <p>为商户分配权益额度，用于支持业务运营</p>
      </div>
      
      <div className="page-content">
        <div className="allocation-form-container">
          <form onSubmit={(e) => { e.preventDefault(); handleAllocate(); }}>
            <div className="form-section">
              <h3>分配信息</h3>
              
              <div className="form-group">
                <label>目标商户 *</label>
                <select
                  value={form.merchant_id}
                  onChange={(e) => setForm({ ...form, merchant_id: parseInt(e.target.value) })}
                  required
                  disabled={loading.merchantList}
                >
                  <option value={0}>请选择商户</option>
                  {merchantList.map(merchant => (
                    <option key={merchant.id} value={merchant.id}>
                      {merchant.name} ({merchant.code})
                    </option>
                  ))}
                </select>
                {loading.merchantList && <span className="loading-text">加载商户列表中...</span>}
              </div>
              
              <div className="form-group">
                <label>分配金额 *</label>
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  max="1000000"
                  value={form.amount}
                  onChange={(e) => setForm({ ...form, amount: parseFloat(e.target.value) || 0 })}
                  placeholder="请输入分配金额"
                  required
                />
                <div className="input-help">
                  单次分配金额上限为 1,000,000
                </div>
              </div>
              
              <div className="form-group">
                <label>分配说明</label>
                <textarea
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  placeholder="请输入分配说明（可选）"
                  maxLength={200}
                  rows={3}
                />
                <div className="input-help">
                  {form.description.length}/200 字符
                </div>
              </div>
            </div>
            
            {/* 商户余额信息显示 */}
            {selectedMerchantBalance && (
              <div className="form-section">
                <h3>商户当前余额</h3>
                <div className="balance-info">
                  <div className="balance-item">
                    <span className="label">总余额:</span>
                    <span className="value">
                      {fundService.formatAmount(selectedMerchantBalance.total_balance)}
                    </span>
                  </div>
                  <div className="balance-item">
                    <span className="label">已使用:</span>
                    <span className="value">
                      {fundService.formatAmount(selectedMerchantBalance.used_balance)}
                    </span>
                  </div>
                  <div className="balance-item">
                    <span className="label">冻结余额:</span>
                    <span className="value">
                      {fundService.formatAmount(selectedMerchantBalance.frozen_balance)}
                    </span>
                  </div>
                  <div className="balance-item highlight">
                    <span className="label">可用余额:</span>
                    <span className="value">
                      {fundService.formatAmount(selectedMerchantBalance.available_balance)}
                    </span>
                  </div>
                  <div className="balance-item">
                    <span className="label">更新时间:</span>
                    <span className="value">
                      {new Date(selectedMerchantBalance.last_updated).toLocaleString()}
                    </span>
                  </div>
                </div>
                
                {/* 分配后预览 */}
                {form.amount > 0 && (
                  <div className="allocation-preview">
                    <h4>分配后预览</h4>
                    <div className="preview-item">
                      <span className="label">分配后总余额:</span>
                      <span className="value positive">
                        {fundService.formatAmount(selectedMerchantBalance.total_balance + form.amount)}
                      </span>
                      <span className="change">
                        (+{fundService.formatAmount(form.amount)})
                      </span>
                    </div>
                    <div className="preview-item">
                      <span className="label">分配后可用余额:</span>
                      <span className="value positive">
                        {fundService.formatAmount(selectedMerchantBalance.available_balance + form.amount)}
                      </span>
                      <span className="change">
                        (+{fundService.formatAmount(form.amount)})
                      </span>
                    </div>
                  </div>
                )}
                
                {loading.balance[form.merchant_id] && (
                  <div className="loading-overlay">
                    <span>更新余额信息中...</span>
                  </div>
                )}
              </div>
            )}
            
            <div className="form-actions">
              <button
                type="submit"
                disabled={loading.allocate || form.merchant_id === 0 || form.amount <= 0}
                className="btn-primary"
              >
                {loading.allocate ? '分配中...' : '确认分配'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setForm({
                    merchant_id: 0,
                    amount: 0,
                    description: '',
                  });
                  setSelectedMerchantBalance(null);
                }}
                className="btn-secondary"
              >
                重置
              </button>
            </div>
          </form>
        </div>
        
        {/* 分配指南 */}
        <div className="allocation-guide">
          <h3>分配指南</h3>
          <div className="guide-content">
            <div className="guide-item">
              <h4>权益分配说明</h4>
              <p>权益分配是为商户账户增加可用额度的操作，分配的权益可用于商户的各项业务活动。</p>
            </div>
            
            <div className="guide-item">
              <h4>分配限制</h4>
              <ul>
                <li>单次分配金额上限：1,000,000</li>
                <li>分配金额必须大于0</li>
                <li>分配后的权益立即生效</li>
              </ul>
            </div>
            
            <div className="guide-item">
              <h4>操作建议</h4>
              <ul>
                <li>分配前请确认商户信息无误</li>
                <li>建议添加详细的分配说明</li>
                <li>大额分配建议分批进行</li>
                <li>定期检查商户权益使用情况</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
      
      <style jsx>{`
        .fund-allocation-page {
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
        
        .page-content {
          display: grid;
          grid-template-columns: 1fr 300px;
          gap: 30px;
        }
        
        .allocation-form-container {
          background: white;
          border-radius: 8px;
          padding: 30px;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        }
        
        .form-section {
          margin-bottom: 30px;
          padding-bottom: 20px;
          border-bottom: 1px solid #eee;
        }
        
        .form-section:last-child {
          border-bottom: none;
          margin-bottom: 0;
        }
        
        .form-section h3 {
          margin: 0 0 20px 0;
          font-size: 18px;
          font-weight: 600;
        }
        
        .form-group {
          margin-bottom: 20px;
        }
        
        .form-group label {
          display: block;
          margin-bottom: 8px;
          font-weight: 500;
        }
        
        .form-group input,
        .form-group select,
        .form-group textarea {
          width: 100%;
          padding: 12px;
          border: 1px solid #ddd;
          border-radius: 6px;
          font-size: 14px;
        }
        
        .form-group input:focus,
        .form-group select:focus,
        .form-group textarea:focus {
          outline: none;
          border-color: #007bff;
          box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
        }
        
        .input-help {
          margin-top: 5px;
          font-size: 12px;
          color: #666;
        }
        
        .loading-text {
          font-size: 12px;
          color: #666;
          margin-left: 10px;
        }
        
        .balance-info {
          background: #f8f9fa;
          border-radius: 6px;
          padding: 20px;
          position: relative;
        }
        
        .balance-item {
          display: flex;
          justify-content: space-between;
          margin-bottom: 12px;
        }
        
        .balance-item:last-child {
          margin-bottom: 0;
        }
        
        .balance-item.highlight {
          font-weight: 600;
          color: #007bff;
        }
        
        .balance-item .label {
          color: #666;
        }
        
        .balance-item .value {
          font-weight: 500;
        }
        
        .allocation-preview {
          margin-top: 20px;
          padding: 15px;
          background: #e3f2fd;
          border-radius: 6px;
        }
        
        .allocation-preview h4 {
          margin: 0 0 15px 0;
          font-size: 16px;
          color: #1976d2;
        }
        
        .preview-item {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 8px;
        }
        
        .preview-item:last-child {
          margin-bottom: 0;
        }
        
        .preview-item .value.positive {
          color: #4caf50;
        }
        
        .preview-item .change {
          font-size: 12px;
          color: #4caf50;
          font-weight: 600;
        }
        
        .loading-overlay {
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: rgba(255, 255, 255, 0.8);
          display: flex;
          align-items: center;
          justify-content: center;
          border-radius: 6px;
        }
        
        .form-actions {
          display: flex;
          gap: 15px;
          padding-top: 20px;
        }
        
        .btn-primary, .btn-secondary {
          padding: 12px 24px;
          border-radius: 6px;
          font-size: 14px;
          font-weight: 500;
          cursor: pointer;
          border: none;
        }
        
        .btn-primary {
          background: #007bff;
          color: white;
        }
        
        .btn-primary:hover:not(:disabled) {
          background: #0056b3;
        }
        
        .btn-primary:disabled {
          background: #ccc;
          cursor: not-allowed;
        }
        
        .btn-secondary {
          background: #6c757d;
          color: white;
        }
        
        .btn-secondary:hover {
          background: #545b62;
        }
        
        .allocation-guide {
          background: white;
          border-radius: 8px;
          padding: 20px;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
          height: fit-content;
        }
        
        .allocation-guide h3 {
          margin: 0 0 20px 0;
          font-size: 16px;
          font-weight: 600;
        }
        
        .guide-item {
          margin-bottom: 20px;
        }
        
        .guide-item:last-child {
          margin-bottom: 0;
        }
        
        .guide-item h4 {
          margin: 0 0 10px 0;
          font-size: 14px;
          font-weight: 600;
          color: #333;
        }
        
        .guide-item p {
          margin: 0;
          font-size: 13px;
          color: #666;
          line-height: 1.5;
        }
        
        .guide-item ul {
          margin: 0;
          padding-left: 16px;
          font-size: 13px;
          color: #666;
        }
        
        .guide-item li {
          margin-bottom: 5px;
          line-height: 1.4;
        }
        
        @media (max-width: 768px) {
          .page-content {
            grid-template-columns: 1fr;
            gap: 20px;
          }
        }
      `}</style>
    </div>
  );
};

export default FundAllocationPage;