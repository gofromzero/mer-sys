import React, { useRef, useEffect } from 'react';
import * as echarts from 'echarts';
import ReactECharts from 'echarts-for-react';

// 通用图表配置类型
export interface BaseChartProps {
  className?: string;
  style?: React.CSSProperties;
  theme?: 'light' | 'dark';
  height?: number | string;
  width?: number | string;
  loading?: boolean;
  notMerge?: boolean;
  lazyUpdate?: boolean;
  onChartReady?: (chart: echarts.ECharts) => void;
  onEvents?: Record<string, (params: any) => void>;
}

// 扩展ECharts配置
export interface ChartOption extends echarts.EChartOption {
  // 可以添加自定义配置
}

// 折线图配置
export interface LineChartProps extends BaseChartProps {
  data: Array<{
    name: string;
    value: number[];
    itemStyle?: any;
    lineStyle?: any;
  }>;
  xAxisData: string[];
  title?: string;
  subtitle?: string;
  yAxisUnit?: string;
  smooth?: boolean;
  area?: boolean;
  showSymbol?: boolean;
}

// 柱状图配置
export interface BarChartProps extends BaseChartProps {
  data: Array<{
    name: string;
    value: number[];
    itemStyle?: any;
  }>;
  xAxisData: string[];
  title?: string;
  subtitle?: string;
  yAxisUnit?: string;
  horizontal?: boolean;
  stack?: boolean;
}

// 饼图配置
export interface PieChartProps extends BaseChartProps {
  data: Array<{
    name: string;
    value: number;
    itemStyle?: any;
  }>;
  title?: string;
  subtitle?: string;
  radius?: [string, string];
  center?: [string, string];
  showLabel?: boolean;
  showLegend?: boolean;
  rosetype?: 'radius' | 'area';
}

// 仪表盘配置
export interface GaugeChartProps extends BaseChartProps {
  value: number;
  max?: number;
  min?: number;
  title?: string;
  subtitle?: string;
  unit?: string;
  color?: string[];
  showDetail?: boolean;
}

// 热力图配置
export interface HeatmapChartProps extends BaseChartProps {
  data: Array<[number, number, number]>;
  xAxisData: string[];
  yAxisData: string[];
  title?: string;
  subtitle?: string;
  colorRange?: [string, string];
}

// 通用图表组件
export const BaseChart: React.FC<{ option: ChartOption } & BaseChartProps> = ({
  option,
  className = '',
  style = {},
  theme = 'light',
  height = 400,
  width = '100%',
  loading = false,
  notMerge = false,
  lazyUpdate = false,
  onChartReady,
  onEvents = {},
}) => {
  const defaultStyle = {
    height,
    width,
    ...style,
  };

  return (
    <div className={`chart-container ${className}`} style={defaultStyle}>
      <ReactECharts
        option={option}
        theme={theme}
        style={{ height: '100%', width: '100%' }}
        loading={loading}
        notMerge={notMerge}
        lazyUpdate={lazyUpdate}
        onChartReady={onChartReady}
        onEvents={onEvents}
      />
    </div>
  );
};

// 折线图组件
export const LineChart: React.FC<LineChartProps> = ({
  data,
  xAxisData,
  title,
  subtitle,
  yAxisUnit = '',
  smooth = true,
  area = false,
  showSymbol = false,
  ...chartProps
}) => {
  const option: ChartOption = {
    title: {
      text: title,
      subtext: subtitle,
      left: 'left',
      textStyle: {
        fontSize: 16,
        fontWeight: 'bold',
      },
      subtextStyle: {
        fontSize: 12,
        color: '#666',
      },
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
      },
      formatter: (params: any) => {
        let result = `${params[0].axisValueLabel}<br/>`;
        params.forEach((param: any) => {
          result += `${param.marker}${param.seriesName}: ${param.value}${yAxisUnit}<br/>`;
        });
        return result;
      },
    },
    legend: {
      top: 'bottom',
      type: 'scroll',
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: xAxisData,
      boundaryGap: false,
      axisLine: {
        lineStyle: {
          color: '#e0e6ed',
        },
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: '#8c8c8c',
      },
    },
    yAxis: {
      type: 'value',
      axisLine: {
        show: false,
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: '#8c8c8c',
        formatter: `{value}${yAxisUnit}`,
      },
      splitLine: {
        lineStyle: {
          color: '#f5f5f5',
          type: 'dashed',
        },
      },
    },
    series: data.map((series) => ({
      name: series.name,
      type: 'line',
      data: series.value,
      smooth,
      showSymbol,
      itemStyle: series.itemStyle || {
        color: '#1890ff',
      },
      lineStyle: series.lineStyle || {
        width: 2,
      },
      areaStyle: area ? {} : undefined,
    })),
  };

  return <BaseChart option={option} {...chartProps} />;
};

// 柱状图组件
export const BarChart: React.FC<BarChartProps> = ({
  data,
  xAxisData,
  title,
  subtitle,
  yAxisUnit = '',
  horizontal = false,
  stack = false,
  ...chartProps
}) => {
  const option: ChartOption = {
    title: {
      text: title,
      subtext: subtitle,
      left: 'left',
      textStyle: {
        fontSize: 16,
        fontWeight: 'bold',
      },
      subtextStyle: {
        fontSize: 12,
        color: '#666',
      },
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
      formatter: (params: any) => {
        let result = `${params[0].axisValueLabel}<br/>`;
        params.forEach((param: any) => {
          result += `${param.marker}${param.seriesName}: ${param.value}${yAxisUnit}<br/>`;
        });
        return result;
      },
    },
    legend: {
      top: 'bottom',
      type: 'scroll',
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      containLabel: true,
    },
    xAxis: {
      type: horizontal ? 'value' : 'category',
      data: horizontal ? undefined : xAxisData,
      axisLine: {
        lineStyle: {
          color: '#e0e6ed',
        },
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: '#8c8c8c',
        formatter: horizontal ? `{value}${yAxisUnit}` : undefined,
      },
      splitLine: horizontal ? {
        lineStyle: {
          color: '#f5f5f5',
          type: 'dashed',
        },
      } : undefined,
    },
    yAxis: {
      type: horizontal ? 'category' : 'value',
      data: horizontal ? xAxisData : undefined,
      axisLine: {
        show: false,
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: '#8c8c8c',
        formatter: horizontal ? undefined : `{value}${yAxisUnit}`,
      },
      splitLine: horizontal ? undefined : {
        lineStyle: {
          color: '#f5f5f5',
          type: 'dashed',
        },
      },
    },
    series: data.map((series) => ({
      name: series.name,
      type: 'bar',
      data: series.value,
      stack: stack ? 'total' : undefined,
      itemStyle: series.itemStyle || {
        color: '#1890ff',
      },
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: 'rgba(0, 0, 0, 0.5)',
        },
      },
    })),
  };

  return <BaseChart option={option} {...chartProps} />;
};

// 饼图组件
export const PieChart: React.FC<PieChartProps> = ({
  data,
  title,
  subtitle,
  radius = ['40%', '70%'],
  center = ['50%', '50%'],
  showLabel = true,
  showLegend = true,
  rosetype,
  ...chartProps
}) => {
  const option: ChartOption = {
    title: {
      text: title,
      subtext: subtitle,
      left: 'left',
      textStyle: {
        fontSize: 16,
        fontWeight: 'bold',
      },
      subtextStyle: {
        fontSize: 12,
        color: '#666',
      },
    },
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)',
    },
    legend: showLegend ? {
      top: 'bottom',
      type: 'scroll',
    } : undefined,
    series: [
      {
        name: title || '数据',
        type: 'pie',
        radius,
        center,
        data: data.map((item) => ({
          name: item.name,
          value: item.value,
          itemStyle: item.itemStyle,
        })),
        roseType: rosetype,
        label: {
          show: showLabel,
          formatter: '{b}: {d}%',
        },
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)',
          },
        },
      },
    ],
  };

  return <BaseChart option={option} {...chartProps} />;
};

// 仪表盘组件
export const GaugeChart: React.FC<GaugeChartProps> = ({
  value,
  max = 100,
  min = 0,
  title,
  subtitle,
  unit = '%',
  color = ['#FF6B6B', '#FFA500', '#32CD32'],
  showDetail = true,
  ...chartProps
}) => {
  const option: ChartOption = {
    title: {
      text: title,
      subtext: subtitle,
      left: 'center',
      textStyle: {
        fontSize: 16,
        fontWeight: 'bold',
      },
      subtextStyle: {
        fontSize: 12,
        color: '#666',
      },
    },
    series: [
      {
        name: title || '仪表盘',
        type: 'gauge',
        min,
        max,
        data: [{ value, name: title || '数值' }],
        detail: {
          show: showDetail,
          formatter: `{value}${unit}`,
          fontSize: 20,
          fontWeight: 'bold',
          color: '#333',
        },
        axisLine: {
          lineStyle: {
            width: 10,
            color: [
              [0.3, color[0]],
              [0.7, color[1]],
              [1, color[2]],
            ],
          },
        },
        pointer: {
          itemStyle: {
            color: 'auto',
          },
        },
        axisTick: {
          distance: -30,
          length: 8,
          lineStyle: {
            color: '#fff',
            width: 2,
          },
        },
        splitLine: {
          distance: -30,
          length: 30,
          lineStyle: {
            color: '#fff',
            width: 4,
          },
        },
        axisLabel: {
          color: 'auto',
          distance: 40,
          fontSize: 12,
        },
        title: {
          offsetCenter: [0, '-30%'],
          fontSize: 14,
          color: '#333',
        },
      },
    ],
  };

  return <BaseChart option={option} {...chartProps} />;
};

// 热力图组件
export const HeatmapChart: React.FC<HeatmapChartProps> = ({
  data,
  xAxisData,
  yAxisData,
  title,
  subtitle,
  colorRange = ['#E3F2FD', '#1976D2'],
  ...chartProps
}) => {
  const option: ChartOption = {
    title: {
      text: title,
      subtext: subtitle,
      left: 'left',
      textStyle: {
        fontSize: 16,
        fontWeight: 'bold',
      },
      subtextStyle: {
        fontSize: 12,
        color: '#666',
      },
    },
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        return `${yAxisData[params.data[1]]}<br/>${xAxisData[params.data[0]]}: ${params.data[2]}`;
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: '15%',
      bottom: '15%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: xAxisData,
      axisLine: {
        lineStyle: {
          color: '#e0e6ed',
        },
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: '#8c8c8c',
      },
    },
    yAxis: {
      type: 'category',
      data: yAxisData,
      axisLine: {
        lineStyle: {
          color: '#e0e6ed',
        },
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: '#8c8c8c',
      },
    },
    visualMap: {
      min: Math.min(...data.map(d => d[2])),
      max: Math.max(...data.map(d => d[2])),
      calculable: true,
      orient: 'horizontal',
      left: 'center',
      bottom: '0%',
      inRange: {
        color: colorRange,
      },
    },
    series: [
      {
        name: title || '热力数据',
        type: 'heatmap',
        data,
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowColor: 'rgba(0, 0, 0, 0.5)',
          },
        },
      },
    ],
  };

  return <BaseChart option={option} {...chartProps} />;
};