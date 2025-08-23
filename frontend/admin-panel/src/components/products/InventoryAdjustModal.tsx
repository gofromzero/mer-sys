// 库存调整模态框组件
import React, { useState } from 'react';
import { InventoryAdjustRequest } from '../../types/product';

interface InventoryAdjustModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (request: InventoryAdjustRequest) => Promise<void>;
  productId: number;
  productName: string;
  currentStock: number;
}

const InventoryAdjustModal: React.FC<InventoryAdjustModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
  productId,
  productName,
  currentStock
}) => {
  const [formData, setFormData] = useState<{
    adjustment_type: 'increase' | 'decrease' | 'set';
    quantity: number;
    reason: string;
    reference_id?: string;
  }>({
    adjustment_type: 'increase',
    quantity: 0,
    reason: '',
    reference_id: ''
  });

  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.reason.trim()) {
      alert('请输入调整原因');
      return;
    }

    if (formData.quantity <= 0) {
      alert('请输入有效的数量');
      return;
    }

    setLoading(true);
    try {
      await onSubmit({
        product_id: productId,
        adjustment_type: formData.adjustment_type,
        quantity: formData.quantity,
        reason: formData.reason.trim(),
        reference_id: formData.reference_id?.trim() || undefined
      });
      onClose();
      // 重置表单
      setFormData({
        adjustment_type: 'increase',
        quantity: 0,
        reason: '',
        reference_id: ''
      });
    } catch (error) {
      console.error('库存调整失败:', error);
      alert('库存调整失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const getExpectedStock = () => {
    switch (formData.adjustment_type) {
      case 'increase':
        return currentStock + formData.quantity;
      case 'decrease':
        return Math.max(0, currentStock - formData.quantity);
      case 'set':
        return formData.quantity;
      default:
        return currentStock;
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-96 max-w-full">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium">库存调整</h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
            type="button"
          >
            ×
          </button>
        </div>

        <div className="mb-4 p-3 bg-gray-50 rounded">
          <div className="text-sm text-gray-600">商品名称</div>
          <div className="font-medium">{productName}</div>
          <div className="text-sm text-gray-600 mt-1">当前库存: {currentStock}</div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              调整类型
            </label>
            <select
              value={formData.adjustment_type}
              onChange={(e) => setFormData(prev => ({ 
                ...prev, 
                adjustment_type: e.target.value as 'increase' | 'decrease' | 'set'
              }))}
              className="w-full border rounded-md px-3 py-2"
              required
            >
              <option value="increase">增加库存</option>
              <option value="decrease">减少库存</option>
              <option value="set">设置库存</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              数量
            </label>
            <input
              type="number"
              min="1"
              value={formData.quantity || ''}
              onChange={(e) => setFormData(prev => ({ 
                ...prev, 
                quantity: parseInt(e.target.value) || 0
              }))}
              className="w-full border rounded-md px-3 py-2"
              required
            />
            <div className="text-sm text-gray-500 mt-1">
              调整后库存: {getExpectedStock()}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              调整原因 *
            </label>
            <textarea
              value={formData.reason}
              onChange={(e) => setFormData(prev => ({ 
                ...prev, 
                reason: e.target.value
              }))}
              className="w-full border rounded-md px-3 py-2"
              rows={3}
              placeholder="请输入库存调整的原因..."
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              参考单号 (可选)
            </label>
            <input
              type="text"
              value={formData.reference_id || ''}
              onChange={(e) => setFormData(prev => ({ 
                ...prev, 
                reference_id: e.target.value
              }))}
              className="w-full border rounded-md px-3 py-2"
              placeholder="相关的订单号、入库单号等"
            />
          </div>

          <div className="flex space-x-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
              disabled={loading}
            >
              取消
            </button>
            <button
              type="submit"
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
              disabled={loading}
            >
              {loading ? '提交中...' : '确认调整'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default InventoryAdjustModal;