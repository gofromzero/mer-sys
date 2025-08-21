import React, { useEffect, useRef } from 'react';
import { RightsUsagePoint } from '../../types/dashboard';

// ECharts 类型定义
interface EChartsOption {
  title?: any;
  tooltip?: any;
  legend?: any;
  grid?: any;
  xAxis?: any;
  yAxis?: any;
  series?: any[];
  dataZoom?: any[];
  toolbox?: any;
  color?: string[];
}

interface EChartsInstance {
  setOption: (option: EChartsOption) => void;
  resize: () => void;
  dispose: () => void;
}

declare global {
  interface Window {
    echarts?: {
      init: (dom: HTMLElement, theme?: string) => EChartsInstance;
      dispose: (instance: EChartsInstance) => void;
    };
  }
}

interface RightsTrendChartProps {
  data: RightsUsagePoint[];
  height?: number;
  loading?: boolean;
}

/**
 * 权益趋势图表组件
 * 使用 ECharts 展示权益使用趋势和余额变化
 */
const RightsTrendChart: React.FC<RightsTrendChartProps> = ({ 
  data = [], 
  height = 300,
  loading = false 
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstance = useRef<EChartsInstance | null>(null);

  useEffect(() => {
    if (!chartRef.current) return;

    // 检查是否有 ECharts 库可用
    if (typeof window !== 'undefined' && window.echarts) {
      // 初始化图表
      chartInstance.current = window.echarts.init(chartRef.current, 'light');
    } else {
      // 如果 ECharts 未加载，显示占位符
      console.warn('ECharts library not loaded, showing placeholder');
    }

    return () => {
      if (chartInstance.current) {
        chartInstance.current.dispose();
      }
    };
  }, []);

  useEffect(() => {
    if (!chartInstance.current || !data.length) return;

    // 准备图表数据
    const dates = data.map(item => {
      const date = new Date(item.date);
      return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
    });
    
    const balanceData = data.map(item => item.balance);
    const usageData = data.map(item => item.usage);

    // ECharts 配置选项
    const option: EChartsOption = {
      title: {
        text: '权益使用趋势',
        left: 'center',
        textStyle: {
          fontSize: 16,
          fontWeight: 'normal',
          color: '#333'
        }
      },
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(0, 0, 0, 0.8)',
        borderColor: 'transparent',
        textStyle: {
          color: '#fff'
        },
        formatter: (params: any) => {
          let content = `<div style="margin-bottom: 4px;">${params[0].axisValue}</div>`;
          params.forEach((param: any) => {
            content += `<div style="display: flex; align-items: center; margin-bottom: 2px;">
              <span style="display: inline-block; width: 10px; height: 10px; background-color: ${param.color}; border-radius: 50%; margin-right: 6px;"></span>
              <span>${param.seriesName}: ${param.value.toLocaleString()}</span>
            </div>`;
          });
          return content;
        }
      },
      legend: {
        data: ['权益余额', '使用量'],
        bottom: 10,
        textStyle: {
          color: '#666'
        }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '15%',
        top: '20%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: dates,
        axisLine: {
          lineStyle: {
            color: '#e8e8e8'
          }
        },
        axisTick: {
          show: false
        },
        axisLabel: {
          color: '#666',
          fontSize: 12
        }
      },
      yAxis: [
        {
          type: 'value',
          name: '权益余额',
          position: 'left',
          axisLine: {
            show: false
          },
          axisTick: {
            show: false
          },
          axisLabel: {
            color: '#666',
            fontSize: 12,
            formatter: (value: number) => {
              if (value >= 10000) {
                return (value / 10000).toFixed(1) + 'w';
              }
              return value.toLocaleString();
            }
          },
          splitLine: {
            lineStyle: {
              color: '#f0f0f0'
            }
          }
        },
        {
          type: 'value',
          name: '使用量',
          position: 'right',
          axisLine: {
            show: false
          },
          axisTick: {
            show: false
          },
          axisLabel: {
            color: '#666',
            fontSize: 12,
            formatter: (value: number) => {
              if (value >= 1000) {
                return (value / 1000).toFixed(1) + 'k';
              }
              return value.toString();
            }
          },
          splitLine: {
            show: false
          }
        }
      ],
      series: [
        {
          name: '权益余额',
          type: 'line',
          yAxisIndex: 0,
          data: balanceData,
          smooth: true,
          symbol: 'circle',
          symbolSize: 6,
          lineStyle: {
            width: 3,
            color: '#1890ff'
          },
          itemStyle: {
            color: '#1890ff',
            borderWidth: 2,
            borderColor: '#fff'
          },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(24, 144, 255, 0.3)' },
                { offset: 1, color: 'rgba(24, 144, 255, 0.05)' }
              ]
            }
          }
        },
        {
          name: '使用量',
          type: 'bar',
          yAxisIndex: 1,
          data: usageData,
          itemStyle: {
            color: '#52c41a',
            borderRadius: [2, 2, 0, 0]
          },
          barWidth: '60%'
        }
      ],
      dataZoom: [
        {
          type: 'inside',
          start: 0,
          end: 100
        },
        {
          start: 0,
          end: 100,
          height: 20,
          bottom: 40,
          fillerColor: 'rgba(24, 144, 255, 0.1)',
          borderColor: 'transparent',
          backgroundColor: '#f8f9fa',
          handleStyle: {
            color: '#1890ff'
          }
        }
      ],
      toolbox: {
        feature: {
          saveAsImage: {
            title: '保存为图片',
            name: '权益趋势图'
          }
        },
        right: 20,
        top: 20
      },
      color: ['#1890ff', '#52c41a']
    };

    chartInstance.current.setOption(option);
  }, [data]);

  // 响应式调整
  useEffect(() => {
    const handleResize = () => {
      if (chartInstance.current) {
        chartInstance.current.resize();
      }
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // 如果没有 ECharts 或正在加载，显示占位符
  if (!window.echarts || loading) {
    return (
      <div 
        style={{ 
          height: `${height}px`,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: '#fafafa',
          border: '1px dashed #d9d9d9',
          color: '#999',
          fontSize: '14px'
        }}
      >
        {loading ? '加载中...' : 'ECharts 图表加载中...'}
      </div>
    );
  }

  // 如果没有数据，显示空状态
  if (!data.length) {
    return (
      <div 
        style={{ 
          height: `${height}px`,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#999',
          fontSize: '14px'
        }}
      >
        <div style={{ marginBottom: '8px' }}>📊</div>
        <div>暂无权益趋势数据</div>
      </div>
    );
  }

  return (
    <div
      ref={chartRef}
      style={{ 
        height: `${height}px`,
        width: '100%'
      }}
    />
  );
};

export default RightsTrendChart;