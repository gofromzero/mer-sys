// 商品状态徽章组件
import React from 'react';
import { ProductStatus } from '../../types/product';

interface ProductStatusBadgeProps {
  status: ProductStatus;
  className?: string;
}

const ProductStatusBadge: React.FC<ProductStatusBadgeProps> = ({ status, className = '' }) => {
  const getStatusConfig = (status: ProductStatus) => {
    const configs = {
      draft: {
        label: '草稿',
        className: 'bg-gray-100 text-gray-800 border-gray-200'
      },
      active: {
        label: '已上架',
        className: 'bg-green-100 text-green-800 border-green-200'
      },
      inactive: {
        label: '已下架',
        className: 'bg-yellow-100 text-yellow-800 border-yellow-200'
      },
      deleted: {
        label: '已删除',
        className: 'bg-red-100 text-red-800 border-red-200'
      }
    };
    
    return configs[status] || configs.draft;
  };

  const config = getStatusConfig(status);

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${config.className} ${className}`}
    >
      {config.label}
    </span>
  );
};

export default ProductStatusBadge;