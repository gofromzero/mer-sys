import React, { useState } from 'react';

export interface DateRange {
  startDate: string;
  endDate: string;
}

export interface DateRangePickerProps {
  value?: DateRange;
  onChange: (dateRange: DateRange) => void;
  className?: string;
  disabled?: boolean;
  maxRange?: number; // 最大天数范围
  quickRanges?: boolean; // 是否显示快捷选择
}

// 快捷日期选择选项
const QUICK_RANGES = [
  { label: '今天', days: 0 },
  { label: '昨天', days: 1, offset: 1 },
  { label: '最近7天', days: 7 },
  { label: '最近30天', days: 30 },
  { label: '最近90天', days: 90 },
  { label: '本月', days: 'thisMonth' as const },
  { label: '上月', days: 'lastMonth' as const },
  { label: '本季度', days: 'thisQuarter' as const },
  { label: '上季度', days: 'lastQuarter' as const },
];

export const DateRangePicker: React.FC<DateRangePickerProps> = ({
  value,
  onChange,
  className = '',
  disabled = false,
  maxRange,
  quickRanges = true,
}) => {
  const [localStartDate, setLocalStartDate] = useState(
    value?.startDate || new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
  );
  const [localEndDate, setLocalEndDate] = useState(
    value?.endDate || new Date().toISOString().split('T')[0]
  );

  const handleStartDateChange = (date: string) => {
    setLocalStartDate(date);
    
    // 验证日期范围
    const startDate = new Date(date);
    const endDate = new Date(localEndDate);
    
    if (startDate > endDate) {
      const newEndDate = date;
      setLocalEndDate(newEndDate);
      onChange({ startDate: date, endDate: newEndDate });
    } else {
      // 检查最大范围限制
      if (maxRange) {
        const diffDays = Math.ceil((endDate.getTime() - startDate.getTime()) / (1000 * 60 * 60 * 24));
        if (diffDays > maxRange) {
          const newEndDate = new Date(startDate.getTime() + maxRange * 24 * 60 * 60 * 1000)
            .toISOString().split('T')[0];
          setLocalEndDate(newEndDate);
          onChange({ startDate: date, endDate: newEndDate });
          return;
        }
      }
      onChange({ startDate: date, endDate: localEndDate });
    }
  };

  const handleEndDateChange = (date: string) => {
    setLocalEndDate(date);
    
    // 验证日期范围
    const startDate = new Date(localStartDate);
    const endDate = new Date(date);
    
    if (endDate < startDate) {
      const newStartDate = date;
      setLocalStartDate(newStartDate);
      onChange({ startDate: newStartDate, endDate: date });
    } else {
      // 检查最大范围限制
      if (maxRange) {
        const diffDays = Math.ceil((endDate.getTime() - startDate.getTime()) / (1000 * 60 * 60 * 24));
        if (diffDays > maxRange) {
          const newStartDate = new Date(endDate.getTime() - maxRange * 24 * 60 * 60 * 1000)
            .toISOString().split('T')[0];
          setLocalStartDate(newStartDate);
          onChange({ startDate: newStartDate, endDate: date });
          return;
        }
      }
      onChange({ startDate: localStartDate, endDate: date });
    }
  };

  const handleQuickRangeSelect = (range: typeof QUICK_RANGES[0]) => {
    const now = new Date();
    let startDate: Date;
    let endDate: Date;

    switch (range.days) {
      case 0: // 今天
        startDate = new Date(now);
        endDate = new Date(now);
        break;
        
      case 'thisMonth':
        startDate = new Date(now.getFullYear(), now.getMonth(), 1);
        endDate = new Date(now);
        break;
        
      case 'lastMonth':
        startDate = new Date(now.getFullYear(), now.getMonth() - 1, 1);
        endDate = new Date(now.getFullYear(), now.getMonth(), 0);
        break;
        
      case 'thisQuarter':
        const currentQuarter = Math.floor(now.getMonth() / 3);
        startDate = new Date(now.getFullYear(), currentQuarter * 3, 1);
        endDate = new Date(now);
        break;
        
      case 'lastQuarter':
        const lastQuarter = Math.floor(now.getMonth() / 3) - 1;
        const lastQuarterYear = lastQuarter < 0 ? now.getFullYear() - 1 : now.getFullYear();
        const lastQuarterMonth = lastQuarter < 0 ? 9 : lastQuarter * 3;
        startDate = new Date(lastQuarterYear, lastQuarterMonth, 1);
        endDate = new Date(lastQuarterYear, lastQuarterMonth + 3, 0);
        break;
        
      default: // 数字天数
        if (typeof range.days === 'number') {
          if (range.offset) {
            // 特殊情况：昨天
            startDate = new Date(now.getTime() - range.offset * 24 * 60 * 60 * 1000);
            endDate = new Date(now.getTime() - range.offset * 24 * 60 * 60 * 1000);
          } else {
            // 最近N天
            startDate = new Date(now.getTime() - range.days * 24 * 60 * 60 * 1000);
            endDate = new Date(now);
          }
        } else {
          return;
        }
    }

    const startDateStr = startDate.toISOString().split('T')[0];
    const endDateStr = endDate.toISOString().split('T')[0];
    
    setLocalStartDate(startDateStr);
    setLocalEndDate(endDateStr);
    onChange({ startDate: startDateStr, endDate: endDateStr });
  };

  // 计算日期范围天数
  const getDaysDiff = () => {
    const start = new Date(localStartDate);
    const end = new Date(localEndDate);
    return Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24)) + 1;
  };

  // 获取今天的日期字符串
  const today = new Date().toISOString().split('T')[0];

  return (
    <div className={`date-range-picker ${className}`}>
      {quickRanges && (
        <div className="quick-ranges mb-4">
          <label className="block text-sm font-medium text-gray-700 mb-2">
            快捷选择
          </label>
          <div className="flex flex-wrap gap-2">
            {QUICK_RANGES.map((range) => (
              <button
                key={range.label}
                type="button"
                onClick={() => handleQuickRangeSelect(range)}
                disabled={disabled}
                className="inline-flex items-center px-3 py-1 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {range.label}
              </button>
            ))}
          </div>
        </div>
      )}

      <div className="date-inputs">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          自定义日期范围
        </label>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label htmlFor="start-date" className="block text-xs font-medium text-gray-600 mb-1">
              开始日期
            </label>
            <input
              id="start-date"
              type="date"
              value={localStartDate}
              max={localEndDate}
              onChange={(e) => handleStartDateChange(e.target.value)}
              disabled={disabled}
              className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm disabled:bg-gray-100 disabled:cursor-not-allowed"
            />
          </div>
          
          <div>
            <label htmlFor="end-date" className="block text-xs font-medium text-gray-600 mb-1">
              结束日期
            </label>
            <input
              id="end-date"
              type="date"
              value={localEndDate}
              min={localStartDate}
              max={today}
              onChange={(e) => handleEndDateChange(e.target.value)}
              disabled={disabled}
              className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm disabled:bg-gray-100 disabled:cursor-not-allowed"
            />
          </div>
        </div>

        <div className="mt-2 text-sm text-gray-500">
          已选择 {getDaysDiff()} 天的数据
          {maxRange && getDaysDiff() > maxRange && (
            <span className="text-red-500 ml-2">
              （超过最大范围 {maxRange} 天）
            </span>
          )}
        </div>
      </div>
    </div>
  );
};