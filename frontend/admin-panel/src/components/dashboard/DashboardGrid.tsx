import React, { useState, useCallback, useRef } from 'react';
import { Card, Modal, Button, Space, Tooltip } from 'antd';
import { 
  DragOutlined,
  SettingOutlined,
  EyeOutlined,
  EyeInvisibleOutlined,
  DeleteOutlined
} from '@ant-design/icons';
import { DashboardWidget, LayoutConfig, WidgetType } from '@/types/dashboard';
import { Responsive, WidthProvider } from 'react-grid-layout';
import RightsTrendChart from './RightsTrendChart';
import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';

const ResponsiveGridLayout = WidthProvider(Responsive);

interface DashboardGridProps {
  layout: LayoutConfig;
  onLayoutChange?: (newLayout: LayoutConfig) => void;
  editable?: boolean;
  className?: string;
}

interface GridItem {
  i: string;
  x: number;
  y: number;
  w: number;
  h: number;
  minW?: number;
  minH?: number;
  maxW?: number;
  maxH?: number;
  isDraggable?: boolean;
  isResizable?: boolean;
}

/**
 * 响应式仪表板网格布局组件
 * 支持拖拽、调整大小、显示/隐藏组件
 */
const DashboardGrid: React.FC<DashboardGridProps> = ({
  layout,
  onLayoutChange,
  editable = false,
  className
}) => {
  const [isDragging, setIsDragging] = useState(false);
  const [editingWidget, setEditingWidget] = useState<DashboardWidget | null>(null);
  const gridRef = useRef<any>(null);

  // 将DashboardWidget转换为react-grid-layout所需的格式
  const convertToGridLayout = useCallback((widgets: DashboardWidget[]): GridItem[] => {
    return widgets
      .filter(widget => widget.visible)
      .map(widget => ({
        i: widget.id,
        x: widget.position.x,
        y: widget.position.y,
        w: widget.size.width,
        h: widget.size.height,
        minW: 1,
        minH: 1,
        maxW: layout.columns,
        maxH: 10,
        isDraggable: editable,
        isResizable: editable
      }));
  }, [layout.columns, editable]);

  // 布局变化处理
  const handleLayoutChange = useCallback((newLayout: GridItem[]) => {
    if (!onLayoutChange || !editable) return;

    const updatedWidgets = layout.widgets.map(widget => {
      const gridItem = newLayout.find(item => item.i === widget.id);
      if (gridItem) {
        return {
          ...widget,
          position: { x: gridItem.x, y: gridItem.y },
          size: { width: gridItem.w, height: gridItem.h }
        };
      }
      return widget;
    });

    onLayoutChange({
      ...layout,
      widgets: updatedWidgets
    });
  }, [layout, onLayoutChange, editable]);

  // 切换组件可见性
  const toggleWidgetVisibility = useCallback((widgetId: string) => {
    if (!onLayoutChange) return;

    const updatedWidgets = layout.widgets.map(widget =>
      widget.id === widgetId
        ? { ...widget, visible: !widget.visible }
        : widget
    );

    onLayoutChange({
      ...layout,
      widgets: updatedWidgets
    });
  }, [layout, onLayoutChange]);

  // 删除组件
  const removeWidget = useCallback((widgetId: string) => {
    if (!onLayoutChange) return;

    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个组件吗？',
      onOk: () => {
        const updatedWidgets = layout.widgets.filter(widget => widget.id !== widgetId);
        onLayoutChange({
          ...layout,
          widgets: updatedWidgets
        });
      }
    });
  }, [layout, onLayoutChange]);

  // 编辑组件配置
  const editWidget = useCallback((widget: DashboardWidget) => {
    setEditingWidget(widget);
  }, []);

  // 渲染组件内容
  const renderWidgetContent = useCallback((widget: DashboardWidget) => {
    const baseStyle = {
      height: '100%',
      overflow: 'hidden'
    };

    // 根据组件类型渲染不同内容
    switch (widget.type) {
      case WidgetType.SALES_OVERVIEW:
        return (
          <div style={baseStyle}>
            <div style={{ padding: '16px', textAlign: 'center' }}>
              <h3>销售概览</h3>
              <p>销售数据展示区域</p>
            </div>
          </div>
        );
      
      case WidgetType.RIGHTS_BALANCE:
        return (
          <div style={baseStyle}>
            <div style={{ padding: '16px', textAlign: 'center' }}>
              <h3>权益余额</h3>
              <p>权益余额展示区域</p>
            </div>
          </div>
        );
      
      case WidgetType.RIGHTS_TREND:
        return (
          <div style={baseStyle}>
            <RightsTrendChart 
              data={[
                { date: '2025-01-15', balance: 10000, usage: 1200 },
                { date: '2025-01-16', balance: 9500, usage: 1500 },
                { date: '2025-01-17', balance: 9200, usage: 800 },
                { date: '2025-01-18', balance: 8800, usage: 1100 },
                { date: '2025-01-19', balance: 8500, usage: 950 },
                { date: '2025-01-20', balance: 8200, usage: 1300 },
                { date: '2025-01-21', balance: 7900, usage: 1000 }
              ]}
              height={widget.size.height * 120 - 60}
              loading={false}
            />
          </div>
        );
      
      case WidgetType.PENDING_TASKS:
        return (
          <div style={baseStyle}>
            <div style={{ padding: '16px', textAlign: 'center' }}>
              <h3>待处理事项</h3>
              <p>任务列表区域</p>
            </div>
          </div>
        );
      
      case WidgetType.ANNOUNCEMENTS:
        return (
          <div style={baseStyle}>
            <div style={{ padding: '16px', textAlign: 'center' }}>
              <h3>公告通知</h3>
              <p>公告列表区域</p>
            </div>
          </div>
        );
      
      default:
        return (
          <div style={{ ...baseStyle, padding: '16px', textAlign: 'center' }}>
            <p>未知组件类型</p>
          </div>
        );
    }
  }, []);

  // 渲染组件操作按钮
  const renderWidgetActions = useCallback((widget: DashboardWidget) => {
    if (!editable) return null;

    return (
      <div
        style={{
          position: 'absolute',
          top: '4px',
          right: '4px',
          zIndex: 10,
          display: isDragging ? 'none' : 'flex',
          gap: '4px'
        }}
        onClick={e => e.stopPropagation()}
      >
        <Tooltip title="配置">
          <Button
            size="small"
            icon={<SettingOutlined />}
            onClick={() => editWidget(widget)}
            style={{ opacity: 0.8 }}
          />
        </Tooltip>
        
        <Tooltip title={widget.visible ? '隐藏' : '显示'}>
          <Button
            size="small"
            icon={widget.visible ? <EyeOutlined /> : <EyeInvisibleOutlined />}
            onClick={() => toggleWidgetVisibility(widget.id)}
            style={{ opacity: 0.8 }}
          />
        </Tooltip>
        
        <Tooltip title="删除">
          <Button
            size="small"
            icon={<DeleteOutlined />}
            danger
            onClick={() => removeWidget(widget.id)}
            style={{ opacity: 0.8 }}
          />
        </Tooltip>
      </div>
    );
  }, [editable, isDragging, editWidget, toggleWidgetVisibility, removeWidget]);

  const gridLayout = convertToGridLayout(layout.widgets);
  
  // 响应式断点配置
  const breakpoints = { lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 };
  const cols = { lg: 4, md: 3, sm: 2, xs: 1, xxs: 1 };

  return (
    <div className={className}>
      <ResponsiveGridLayout
        ref={gridRef}
        className="layout"
        layouts={{ lg: gridLayout }}
        breakpoints={breakpoints}
        cols={cols}
        rowHeight={120}
        width={1200}
        margin={[16, 16]}
        containerPadding={[0, 0]}
        isDraggable={editable}
        isResizable={editable}
        onLayoutChange={handleLayoutChange}
        onDragStart={() => setIsDragging(true)}
        onDragStop={() => setIsDragging(false)}
        onResizeStart={() => setIsDragging(true)}
        onResizeStop={() => setIsDragging(false)}
        draggableHandle=".drag-handle"
        resizeHandles={['se']}
      >
        {layout.widgets
          .filter(widget => widget.visible)
          .map(widget => (
            <Card
              key={widget.id}
              size="small"
              style={{
                height: '100%',
                position: 'relative',
                cursor: editable ? 'move' : 'default'
              }}
              bodyStyle={{ 
                padding: 0,
                height: 'calc(100% - 32px)'
              }}
              title={
                editable ? (
                  <div style={{ display: 'flex', alignItems: 'center' }}>
                    <DragOutlined 
                      className="drag-handle" 
                      style={{ 
                        marginRight: '8px',
                        cursor: 'move',
                        color: '#1890ff'
                      }} 
                    />
                    {widget.type}
                  </div>
                ) : null
              }
              extra={renderWidgetActions(widget)}
            >
              {renderWidgetContent(widget)}
            </Card>
          ))
        }
      </ResponsiveGridLayout>

      {/* 组件配置Modal */}
      <Modal
        title="组件配置"
        open={!!editingWidget}
        onCancel={() => setEditingWidget(null)}
        footer={[
          <Button key="cancel" onClick={() => setEditingWidget(null)}>
            取消
          </Button>,
          <Button key="ok" type="primary" onClick={() => setEditingWidget(null)}>
            确定
          </Button>
        ]}
      >
        {editingWidget && (
          <div>
            <p><strong>组件ID:</strong> {editingWidget.id}</p>
            <p><strong>组件类型:</strong> {editingWidget.type}</p>
            <p><strong>位置:</strong> ({editingWidget.position.x}, {editingWidget.position.y})</p>
            <p><strong>大小:</strong> {editingWidget.size.width} × {editingWidget.size.height}</p>
            <p style={{ color: '#8c8c8c' }}>详细配置功能开发中...</p>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default DashboardGrid;