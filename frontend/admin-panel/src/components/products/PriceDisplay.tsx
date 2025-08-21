// 价格显示组件
import React from 'react';
import { Money } from '../../types/product';

interface PriceDisplayProps {
  price: Money;
  className?: string;
}

const PriceDisplay: React.FC<PriceDisplayProps> = ({ price, className = '' }) => {
  const formatPrice = (amount: number, currency: string) => {
    const value = (amount / 100).toFixed(2);
    const symbol = getCurrencySymbol(currency);
    return `${symbol}${value}`;
  };

  const getCurrencySymbol = (currency: string) => {
    const symbols: { [key: string]: string } = {
      CNY: '¥',
      USD: '$',
      EUR: '€',
      GBP: '£'
    };
    return symbols[currency] || currency;
  };

  return (
    <span className={`font-medium text-lg ${className}`}>
      {formatPrice(price.amount, price.currency)}
    </span>
  );
};

export default PriceDisplay;