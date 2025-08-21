// 库存显示组件
import React from 'react';
import { InventoryInfo } from '../../types/product';

interface InventoryDisplayProps {
  inventory: InventoryInfo;
  showDetails?: boolean;
  className?: string;
}

const InventoryDisplay: React.FC<InventoryDisplayProps> = ({ 
  inventory, 
  showDetails = false,
  className = '' 
}) => {
  const availableStock = inventory.stock_quantity - inventory.reserved_quantity;
  
  const getStockStatus = (available: number) => {
    if (!inventory.track_inventory) {
      return { label: '不限库存', color: 'text-blue-600' };
    }
    
    if (available <= 0) {
      return { label: '缺货', color: 'text-red-600' };
    } else if (available <= 10) {
      return { label: '库存紧张', color: 'text-orange-600' };
    } else {
      return { label: '库存充足', color: 'text-green-600' };
    }
  };

  const status = getStockStatus(availableStock);

  if (!showDetails) {
    return (
      <div className={`flex items-center space-x-2 ${className}`}>
        <span className="font-medium">
          {inventory.track_inventory ? availableStock : '∞'}
        </span>
        <span className={`text-sm ${status.color}`}>
          {status.label}
        </span>
      </div>
    );
  }

  return (
    <div className={`space-y-1 ${className}`}>
      <div className="flex items-center justify-between">
        <span className="text-sm text-gray-600">总库存:</span>
        <span className="font-medium">{inventory.stock_quantity}</span>
      </div>
      <div className="flex items-center justify-between">
        <span className="text-sm text-gray-600">预留:</span>
        <span className="font-medium">{inventory.reserved_quantity}</span>
      </div>
      <div className="flex items-center justify-between border-t pt-1">
        <span className="text-sm text-gray-600">可用:</span>
        <div className="flex items-center space-x-2">
          <span className="font-medium">
            {inventory.track_inventory ? availableStock : '不限'}
          </span>
          <span className={`text-xs ${status.color}`}>
            {status.label}
          </span>
        </div>
      </div>
      {!inventory.track_inventory && (
        <div className="text-xs text-blue-600 text-center">
          未启用库存跟踪
        </div>
      )}
    </div>
  );
};

export default InventoryDisplay;