// 资金充值页面

import React, { useEffect, useState } from 'react';
// import { toast } from 'react-hot-toast';
// 使用简单的alert替代toast，避免外部依赖
const toast = {
  success: (msg: string) => alert(`成功: ${msg}`),
  error: (msg: string) => alert(`错误: ${msg}`)
};
import { useFundActions, useFundData, useFundLoading, useFundError } from '../../stores/fundStore';
import type { DepositRequest, BatchDepositRequest } from '../../types/fund';
import fundService from '../../services/fundService';

const FundDepositPage: React.FC = () => {
  const actions = useFundActions();
  const { merchantList } = useFundData();
  const loading = useFundLoading();
  const error = useFundError();
  
  // 单笔充值表单状态
  const [singleForm, setSingleForm] = useState<DepositRequest>({
    merchant_id: 0,
    amount: 0,
    currency: 'CNY',
    description: '',
  });
  
  // 批量充值状态
  const [batchMode, setBatchMode] = useState(false);
  const [batchList, setBatchList] = useState<DepositRequest[]>([
    { merchant_id: 0, amount: 0, currency: 'CNY', description: '' }
  ]);
  
  useEffect(() => {
    actions.loadMerchantList();
  }, []);
  
  useEffect(() => {
    if (error) {
      toast.error(error);
      actions.clearError();
    }
  }, [error, actions]);
  
  // 处理单笔充值
  const handleSingleDeposit = async () => {
    const errors = fundService.validateDepositRequest(singleForm);
    if (errors.length > 0) {
      toast.error(errors[0]);
      return;
    }
    
    try {
      await actions.deposit(singleForm);
      toast.success('充值成功');
      setSingleForm({
        merchant_id: 0,
        amount: 0,
        currency: 'CNY',
        description: '',
      });
    } catch (error) {
      // Error already handled by store
    }
  };
  
  // 处理批量充值
  const handleBatchDeposit = async () => {
    const request: BatchDepositRequest = { deposits: batchList };
    const errors = fundService.validateBatchDepositRequest(request);
    if (errors.length > 0) {
      toast.error(errors[0]);
      return;
    }
    
    try {
      const results = await actions.batchDeposit(request);
      toast.success(`批量充值成功，共处理 ${results.length} 笔`);
      setBatchList([{ merchant_id: 0, amount: 0, currency: 'CNY', description: '' }]);
    } catch (error) {
      // Error already handled by store
    }
  };
  
  // 添加批量充值项
  const addBatchItem = () => {
    if (batchList.length >= 100) {
      toast.error('最多只能添加100笔充值');
      return;
    }
    setBatchList([...batchList, { merchant_id: 0, amount: 0, currency: 'CNY', description: '' }]);
  };
  
  // 移除批量充值项
  const removeBatchItem = (index: number) => {
    if (batchList.length === 1) {
      toast.error('至少保留一笔充值');
      return;
    }
    setBatchList(batchList.filter((_, i) => i !== index));
  };
  
  // 更新批量充值项
  const updateBatchItem = (index: number, field: keyof DepositRequest, value: any) => {
    const newBatchList = [...batchList];
    newBatchList[index] = { ...newBatchList[index], [field]: value };
    setBatchList(newBatchList);
  };
  
  const schema = {
    type: 'page',
    title: '资金充值',
    body: [
      {
        type: 'tabs',
        tabs: [
          {
            title: '单笔充值',
            body: [
              {
                type: 'form',
                api: {
                  method: 'post',
                  url: '/api/mock', // 这里用mock，实际提交通过handleSingleDeposit处理
                },
                controls: [
                  {
                    type: 'select',
                    name: 'merchant_id',
                    label: '目标商户',
                    placeholder: '请选择商户',
                    required: true,
                    options: merchantList.map(merchant => ({
                      label: `${merchant.name} (${merchant.code})`,
                      value: merchant.id,
                    })),
                    value: singleForm.merchant_id,
                    onChange: (value: number) => setSingleForm({ ...singleForm, merchant_id: value }),
                  },
                  {
                    type: 'number',
                    name: 'amount',
                    label: '充值金额',
                    placeholder: '请输入充值金额',
                    required: true,
                    min: 0.01,
                    max: 1000000,
                    precision: 2,
                    value: singleForm.amount,
                    onChange: (value: number) => setSingleForm({ ...singleForm, amount: value }),
                  },
                  {
                    type: 'select',
                    name: 'currency',
                    label: '货币类型',
                    required: true,
                    options: [
                      { label: '人民币 (CNY)', value: 'CNY' },
                      { label: '美元 (USD)', value: 'USD' },
                      { label: '欧元 (EUR)', value: 'EUR' },
                    ],
                    value: singleForm.currency,
                    onChange: (value: string) => setSingleForm({ ...singleForm, currency: value }),
                  },
                  {
                    type: 'textarea',
                    name: 'description',
                    label: '备注说明',
                    placeholder: '请输入充值说明（可选）',
                    maxLength: 200,
                    value: singleForm.description,
                    onChange: (value: string) => setSingleForm({ ...singleForm, description: value }),
                  },
                ],
                actions: [
                  {
                    type: 'button',
                    label: loading.deposit ? '充值中...' : '确认充值',
                    level: 'primary',
                    disabled: loading.deposit,
                    onClick: handleSingleDeposit,
                  },
                  {
                    type: 'button',
                    label: '重置',
                    onClick: () => setSingleForm({
                      merchant_id: 0,
                      amount: 0,
                      currency: 'CNY',
                      description: '',
                    }),
                  },
                ],
              },
            ],
          },
          {
            title: '批量充值',
            body: [
              {
                type: 'alert',
                level: 'info',
                body: '批量充值功能允许您一次性为多个商户进行充值，最多支持100笔充值操作。',
              },
              {
                type: 'form',
                controls: batchList.map((item, index) => ({
                  type: 'group',
                  label: `第 ${index + 1} 笔充值`,
                  controls: [
                    {
                      type: 'select',
                      name: `merchant_id_${index}`,
                      label: '目标商户',
                      placeholder: '请选择商户',
                      required: true,
                      options: merchantList.map(merchant => ({
                        label: `${merchant.name} (${merchant.code})`,
                        value: merchant.id,
                      })),
                      value: item.merchant_id,
                      onChange: (value: number) => updateBatchItem(index, 'merchant_id', value),
                    },
                    {
                      type: 'number',
                      name: `amount_${index}`,
                      label: '充值金额',
                      placeholder: '请输入充值金额',
                      required: true,
                      min: 0.01,
                      max: 1000000,
                      precision: 2,
                      value: item.amount,
                      onChange: (value: number) => updateBatchItem(index, 'amount', value),
                    },
                    {
                      type: 'select',
                      name: `currency_${index}`,
                      label: '货币类型',
                      required: true,
                      options: [
                        { label: '人民币 (CNY)', value: 'CNY' },
                        { label: '美元 (USD)', value: 'USD' },
                        { label: '欧元 (EUR)', value: 'EUR' },
                      ],
                      value: item.currency,
                      onChange: (value: string) => updateBatchItem(index, 'currency', value),
                    },
                    {
                      type: 'textarea',
                      name: `description_${index}`,
                      label: '备注说明',
                      placeholder: '请输入充值说明（可选）',
                      maxLength: 200,
                      value: item.description,
                      onChange: (value: string) => updateBatchItem(index, 'description', value),
                    },
                    {
                      type: 'button-group',
                      buttons: [
                        batchList.length < 100 && {
                          type: 'button',
                          label: '添加一笔',
                          level: 'primary',
                          size: 'sm',
                          onClick: addBatchItem,
                        },
                        batchList.length > 1 && {
                          type: 'button',
                          label: '删除此笔',
                          level: 'danger',
                          size: 'sm',
                          onClick: () => removeBatchItem(index),
                        },
                      ].filter(Boolean),
                    },
                  ],
                })),
                actions: [
                  {
                    type: 'button',
                    label: loading.deposit ? '批量充值中...' : '确认批量充值',
                    level: 'primary',
                    disabled: loading.deposit || batchList.length === 0,
                    onClick: handleBatchDeposit,
                  },
                  {
                    type: 'button',
                    label: '重置全部',
                    onClick: () => setBatchList([{ merchant_id: 0, amount: 0, currency: 'CNY', description: '' }]),
                  },
                ],
              },
            ],
          },
        ],
      },
    ],
  };
  
  return (
    <div className="fund-deposit-page">
      <div className="amis-container">
        {/* 这里应该使用amis渲染器，但为了演示直接展示结构 */}
        <div className="page-header">
          <h1>资金充值</h1>
          <p>为商户账户充值资金，支持单笔和批量操作</p>
        </div>
        
        <div className="page-content">
          <div className="tabs">
            <div className="tab-header">
              <button className={!batchMode ? 'active' : ''} onClick={() => setBatchMode(false)}>
                单笔充值
              </button>
              <button className={batchMode ? 'active' : ''} onClick={() => setBatchMode(true)}>
                批量充值
              </button>
            </div>
            
            {!batchMode ? (
              <div className="single-deposit-form">
                <form onSubmit={(e) => { e.preventDefault(); handleSingleDeposit(); }}>
                  <div className="form-group">
                    <label>目标商户 *</label>
                    <select
                      value={singleForm.merchant_id}
                      onChange={(e) => setSingleForm({ ...singleForm, merchant_id: parseInt(e.target.value) })}
                      required
                    >
                      <option value={0}>请选择商户</option>
                      {merchantList.map(merchant => (
                        <option key={merchant.id} value={merchant.id}>
                          {merchant.name} ({merchant.code})
                        </option>
                      ))}
                    </select>
                  </div>
                  
                  <div className="form-group">
                    <label>充值金额 *</label>
                    <input
                      type="number"
                      step="0.01"
                      min="0.01"
                      max="1000000"
                      value={singleForm.amount}
                      onChange={(e) => setSingleForm({ ...singleForm, amount: parseFloat(e.target.value) || 0 })}
                      placeholder="请输入充值金额"
                      required
                    />
                  </div>
                  
                  <div className="form-group">
                    <label>货币类型 *</label>
                    <select
                      value={singleForm.currency}
                      onChange={(e) => setSingleForm({ ...singleForm, currency: e.target.value })}
                      required
                    >
                      <option value="CNY">人民币 (CNY)</option>
                      <option value="USD">美元 (USD)</option>
                      <option value="EUR">欧元 (EUR)</option>
                    </select>
                  </div>
                  
                  <div className="form-group">
                    <label>备注说明</label>
                    <textarea
                      value={singleForm.description}
                      onChange={(e) => setSingleForm({ ...singleForm, description: e.target.value })}
                      placeholder="请输入充值说明（可选）"
                      maxLength={200}
                      rows={3}
                    />
                  </div>
                  
                  <div className="form-actions">
                    <button type="submit" disabled={loading.deposit} className="btn-primary">
                      {loading.deposit ? '充值中...' : '确认充值'}
                    </button>
                    <button
                      type="button"
                      onClick={() => setSingleForm({
                        merchant_id: 0,
                        amount: 0,
                        currency: 'CNY',
                        description: '',
                      })}
                      className="btn-secondary"
                    >
                      重置
                    </button>
                  </div>
                </form>
              </div>
            ) : (
              <div className="batch-deposit-form">
                <div className="alert alert-info">
                  批量充值功能允许您一次性为多个商户进行充值，最多支持100笔充值操作。
                </div>
                
                <form onSubmit={(e) => { e.preventDefault(); handleBatchDeposit(); }}>
                  {batchList.map((item, index) => (
                    <div key={index} className="batch-item">
                      <h4>第 {index + 1} 笔充值</h4>
                      
                      <div className="form-row">
                        <div className="form-group">
                          <label>目标商户 *</label>
                          <select
                            value={item.merchant_id}
                            onChange={(e) => updateBatchItem(index, 'merchant_id', parseInt(e.target.value))}
                            required
                          >
                            <option value={0}>请选择商户</option>
                            {merchantList.map(merchant => (
                              <option key={merchant.id} value={merchant.id}>
                                {merchant.name} ({merchant.code})
                              </option>
                            ))}
                          </select>
                        </div>
                        
                        <div className="form-group">
                          <label>充值金额 *</label>
                          <input
                            type="number"
                            step="0.01"
                            min="0.01"
                            max="1000000"
                            value={item.amount}
                            onChange={(e) => updateBatchItem(index, 'amount', parseFloat(e.target.value) || 0)}
                            placeholder="请输入充值金额"
                            required
                          />
                        </div>
                        
                        <div className="form-group">
                          <label>货币类型 *</label>
                          <select
                            value={item.currency}
                            onChange={(e) => updateBatchItem(index, 'currency', e.target.value)}
                            required
                          >
                            <option value="CNY">人民币 (CNY)</option>
                            <option value="USD">美元 (USD)</option>
                            <option value="EUR">欧元 (EUR)</option>
                          </select>
                        </div>
                      </div>
                      
                      <div className="form-group">
                        <label>备注说明</label>
                        <textarea
                          value={item.description}
                          onChange={(e) => updateBatchItem(index, 'description', e.target.value)}
                          placeholder="请输入充值说明（可选）"
                          maxLength={200}
                          rows={2}
                        />
                      </div>
                      
                      <div className="batch-item-actions">
                        {batchList.length < 100 && (
                          <button type="button" onClick={addBatchItem} className="btn-primary btn-sm">
                            添加一笔
                          </button>
                        )}
                        {batchList.length > 1 && (
                          <button
                            type="button"
                            onClick={() => removeBatchItem(index)}
                            className="btn-danger btn-sm"
                          >
                            删除此笔
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                  
                  <div className="form-actions">
                    <button
                      type="submit"
                      disabled={loading.deposit || batchList.length === 0}
                      className="btn-primary"
                    >
                      {loading.deposit ? '批量充值中...' : '确认批量充值'}
                    </button>
                    <button
                      type="button"
                      onClick={() => setBatchList([{ merchant_id: 0, amount: 0, currency: 'CNY', description: '' }])}
                      className="btn-secondary"
                    >
                      重置全部
                    </button>
                  </div>
                </form>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default FundDepositPage;