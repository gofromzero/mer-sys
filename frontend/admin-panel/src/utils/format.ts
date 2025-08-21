/**
 * 格式化工具函数
 */

/**
 * 格式化货币显示
 * @param amount 金额
 * @param currency 货币符号，默认为 ¥
 * @returns 格式化后的货币字符串
 */
export const formatCurrency = (amount: number, currency = '¥'): string => {
  return `${currency}${amount.toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })}`;
};

/**
 * 格式化数字显示
 * @param num 数字
 * @returns 格式化后的数字字符串
 */
export const formatNumber = (num: number): string => {
  return num.toLocaleString('zh-CN');
};

/**
 * 格式化百分比显示
 * @param rate 比率 (0-1)
 * @param decimals 小数位数，默认1位
 * @returns 格式化后的百分比字符串
 */
export const formatPercent = (rate: number, decimals = 1): string => {
  return `${(rate * 100).toFixed(decimals)}%`;
};

/**
 * 格式化日期时间
 * @param date 日期字符串或Date对象
 * @returns 格式化后的日期时间字符串
 */
export const formatDateTime = (date: string | Date): string => {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleString('zh-CN');
};

/**
 * 格式化日期
 * @param date 日期字符串或Date对象
 * @returns 格式化后的日期字符串
 */
export const formatDate = (date: string | Date): string => {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleDateString('zh-CN');
};