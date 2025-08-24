import React, { useState } from 'react';
import { useOrderStore } from '../../stores/orderStore';

interface PaymentComponentProps {
  orderId: number;
  totalAmount: number;
  onSuccess?: () => void;
  onError?: (error: string) => void;
}

const PaymentComponent: React.FC<PaymentComponentProps> = ({
  orderId,
  totalAmount,
  onSuccess,
  onError,
}) => {
  const [paymentMethod, setPaymentMethod] = useState<'alipay' | 'wechat'>('alipay');
  const [returnUrl, setReturnUrl] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const { initiatePayment, checkPaymentStatus } = useOrderStore();

  const handlePayment = async () => {
    try {
      setIsLoading(true);
      await initiatePayment(orderId, paymentMethod, returnUrl || undefined);
      onSuccess?.();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '支付失败';
      onError?.(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCheckStatus = async () => {
    try {
      setIsLoading(true);
      await checkPaymentStatus(orderId);
      onSuccess?.();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '查询支付状态失败';
      onError?.(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="payment-component p-6 bg-white rounded-lg shadow-lg">
      <h3 className="text-xl font-semibold mb-4">支付订单</h3>
      
      <div className="mb-4">
        <p className="text-gray-600">订单号: #{orderId}</p>
        <p className="text-lg font-semibold text-green-600">
          支付金额: ¥{totalAmount.toFixed(2)}
        </p>
      </div>

      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          选择支付方式
        </label>
        <div className="space-y-2">
          <label className="flex items-center">
            <input
              type="radio"
              value="alipay"
              checked={paymentMethod === 'alipay'}
              onChange={(e) => setPaymentMethod(e.target.value as 'alipay')}
              className="mr-2"
            />
            <img
              src="/images/alipay-logo.png"
              alt="支付宝"
              className="w-6 h-6 mr-2"
              onError={(e) => {
                const target = e.target as HTMLImageElement;
                target.style.display = 'none';
              }}
            />
            支付宝
          </label>
          <label className="flex items-center">
            <input
              type="radio"
              value="wechat"
              checked={paymentMethod === 'wechat'}
              onChange={(e) => setPaymentMethod(e.target.value as 'wechat')}
              className="mr-2"
            />
            <img
              src="/images/wechat-logo.png"
              alt="微信支付"
              className="w-6 h-6 mr-2"
              onError={(e) => {
                const target = e.target as HTMLImageElement;
                target.style.display = 'none';
              }}
            />
            微信支付
          </label>
        </div>
      </div>

      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          支付成功返回页面 (可选)
        </label>
        <input
          type="url"
          value={returnUrl}
          onChange={(e) => setReturnUrl(e.target.value)}
          placeholder="请输入支付成功后的返回地址"
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <div className="flex space-x-3">
        <button
          onClick={handlePayment}
          disabled={isLoading}
          className={`flex-1 px-6 py-3 rounded-md font-medium ${
            isLoading
              ? 'bg-gray-400 cursor-not-allowed'
              : 'bg-blue-600 hover:bg-blue-700 focus:ring-2 focus:ring-blue-500'
          } text-white transition-colors`}
        >
          {isLoading ? '处理中...' : '立即支付'}
        </button>
        
        <button
          onClick={handleCheckStatus}
          disabled={isLoading}
          className={`px-6 py-3 rounded-md font-medium border ${
            isLoading
              ? 'bg-gray-100 border-gray-300 cursor-not-allowed'
              : 'bg-white border-gray-300 hover:bg-gray-50 focus:ring-2 focus:ring-blue-500'
          } text-gray-700 transition-colors`}
        >
          查询状态
        </button>
      </div>

      <div className="mt-4 text-sm text-gray-500">
        <p>• 请在30分钟内完成支付，否则订单将自动取消</p>
        <p>• 支付过程中请勿关闭页面或重复提交</p>
        <p>• 如遇到问题，请联系客服</p>
      </div>
    </div>
  );
};

export default PaymentComponent;