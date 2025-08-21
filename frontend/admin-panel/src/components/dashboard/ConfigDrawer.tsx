import React, { useState, useEffect } from 'react';
import {
  Drawer,
  Form,
  Input,
  InputNumber,
  Switch,
  Select,
  Button,
  Space,
  Divider,
  Card,
  Row,
  Col,
  message,
  Tabs,
  Tooltip,
  Modal
} from 'antd';
import {
  SettingOutlined,
  MobileOutlined,
  DesktopOutlined,
  SaveOutlined,
  ReloadOutlined,
  QuestionCircleOutlined
} from '@ant-design/icons';
import { 
  DashboardConfig, 
  DashboardConfigRequest, 
  WidgetType,
  WidgetPreference,
  LayoutConfig,
  DashboardWidget
} from '@/types/dashboard';
import { useDashboardStore } from '@/stores/dashboardStore';

const { Option } = Select;
const { TabPane } = Tabs;

interface ConfigDrawerProps {
  open: boolean;
  onClose: () => void;
}

/**
 * 仪表板配置侧边栏
 * 支持布局配置、组件管理、响应式设置
 */
const ConfigDrawer: React.FC<ConfigDrawerProps> = ({ open, onClose }) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('layout');

  const { 
    config,
    loading: storeLoading,
    updateConfig,
    loadConfig
  } = useDashboardStore();

  // 组件初始化时加载配置
  useEffect(() => {
    if (open && !config) {
      loadConfig();
    }
  }, [open, config, loadConfig]);

  // 设置表单初始值
  useEffect(() => {
    if (config) {
      form.setFieldsValue({
        refresh_interval: config.refresh_interval,
        layout_columns: config.layout_config.columns,
        widget_preferences: config.widget_preferences
      });
    }
  }, [config, form]);

  // 保存配置
  const handleSave = async () => {
    try {
      setLoading(true);
      const values = await form.validateFields();
      
      if (!config) {
        message.error('配置数据未加载');
        return;
      }

      const configRequest: DashboardConfigRequest = {
        layout_config: {
          ...config.layout_config,
          columns: values.layout_columns
        },
        widget_preferences: values.widget_preferences,
        refresh_interval: values.refresh_interval,
        mobile_layout: config.mobile_layout
      };

      await updateConfig(configRequest);
      message.success('配置保存成功');
      onClose();
      
    } catch (error: any) {
      message.error(error.message || '保存配置失败');
    } finally {
      setLoading(false);
    }
  };

  // 重置配置
  const handleReset = () => {
    Modal.confirm({
      title: '确认重置',
      content: '确定要重置所有配置到默认值吗？此操作不可撤销。',
      okText: '确定',
      cancelText: '取消',
      onOk: () => {
        if (config) {
          const defaultConfig: DashboardConfigRequest = {
            layout_config: {
              columns: 4,
              widgets: getDefaultWidgets()
            },
            widget_preferences: getDefaultWidgetPreferences(),
            refresh_interval: 300
          };
          
          form.setFieldsValue({
            refresh_interval: defaultConfig.refresh_interval,
            layout_columns: defaultConfig.layout_config.columns,
            widget_preferences: defaultConfig.widget_preferences
          });
          
          message.success('配置已重置');
        }
      }
    });
  };

  // 获取默认组件配置
  const getDefaultWidgets = (): DashboardWidget[] => [
    {
      id: 'sales_overview',
      type: WidgetType.SALES_OVERVIEW,
      position: { x: 0, y: 0 },
      size: { width: 2, height: 1 },
      config: {},
      visible: true
    },
    {
      id: 'rights_balance',
      type: WidgetType.RIGHTS_BALANCE,
      position: { x: 2, y: 0 },
      size: { width: 2, height: 1 },
      config: {},
      visible: true
    },
    {
      id: 'rights_trend',
      type: WidgetType.RIGHTS_TREND,
      position: { x: 0, y: 1 },
      size: { width: 4, height: 2 },
      config: {},
      visible: true
    },
    {
      id: 'pending_tasks',
      type: WidgetType.PENDING_TASKS,
      position: { x: 0, y: 3 },
      size: { width: 2, height: 2 },
      config: {},
      visible: true
    },
    {
      id: 'announcements',
      type: WidgetType.ANNOUNCEMENTS,
      position: { x: 2, y: 3 },
      size: { width: 2, height: 2 },
      config: {},
      visible: true
    }
  ];

  // 获取默认组件偏好
  const getDefaultWidgetPreferences = (): WidgetPreference[] => [
    { widget_type: WidgetType.SALES_OVERVIEW, enabled: true, config: {} },
    { widget_type: WidgetType.RIGHTS_BALANCE, enabled: true, config: {} },
    { widget_type: WidgetType.RIGHTS_TREND, enabled: true, config: {} },
    { widget_type: WidgetType.PENDING_TASKS, enabled: true, config: {} },
    { widget_type: WidgetType.ANNOUNCEMENTS, enabled: true, config: {} }
  ];

  // 获取组件显示名称
  const getWidgetDisplayName = (type: WidgetType) => {
    switch (type) {
      case WidgetType.SALES_OVERVIEW:
        return '销售概览';
      case WidgetType.RIGHTS_BALANCE:
        return '权益余额';
      case WidgetType.RIGHTS_TREND:
        return '权益趋势';
      case WidgetType.PENDING_TASKS:
        return '待处理事项';
      case WidgetType.RECENT_ORDERS:
        return '近期订单';
      case WidgetType.ANNOUNCEMENTS:
        return '公告通知';
      case WidgetType.QUICK_ACTIONS:
        return '快速操作';
      default:
        return type;
    }
  };

  return (
    <Drawer
      title={
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <SettingOutlined style={{ marginRight: '8px' }} />
          仪表板配置
        </div>
      }
      placement="right"
      width={480}
      open={open}
      onClose={onClose}
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={handleReset}>
            重置
          </Button>
          <Button
            type="primary"
            icon={<SaveOutlined />}
            loading={loading}
            onClick={handleSave}
          >
            保存
          </Button>
        </Space>
      }
    >
      <Form form={form} layout="vertical">
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          {/* 基本设置 */}
          <TabPane 
            tab={
              <span>
                <DesktopOutlined />
                基本设置
              </span>
            }
            key="layout"
          >
            <Card size="small" title="刷新设置" style={{ marginBottom: '16px' }}>
              <Form.Item
                name="refresh_interval"
                label={
                  <span>
                    自动刷新间隔 (秒)
                    <Tooltip title="设置仪表板数据的自动刷新间隔，建议不少于60秒">
                      <QuestionCircleOutlined style={{ marginLeft: '4px' }} />
                    </Tooltip>
                  </span>
                }
                rules={[
                  { required: true, message: '请输入刷新间隔' },
                  { type: 'number', min: 60, max: 3600, message: '刷新间隔应在60-3600秒之间' }
                ]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  placeholder="300"
                  min={60}
                  max={3600}
                  addonAfter="秒"
                />
              </Form.Item>
            </Card>

            <Card size="small" title="布局设置">
              <Form.Item
                name="layout_columns"
                label={
                  <span>
                    网格列数
                    <Tooltip title="设置仪表板的网格列数，建议4-6列">
                      <QuestionCircleOutlined style={{ marginLeft: '4px' }} />
                    </Tooltip>
                  </span>
                }
                rules={[
                  { required: true, message: '请输入列数' },
                  { type: 'number', min: 1, max: 12, message: '列数应在1-12之间' }
                ]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  placeholder="4"
                  min={1}
                  max={12}
                />
              </Form.Item>
            </Card>
          </TabPane>

          {/* 组件管理 */}
          <TabPane
            tab={
              <span>
                <SettingOutlined />
                组件管理
              </span>
            }
            key="widgets"
          >
            <Card size="small" title="组件显示设置">
              <Form.List name="widget_preferences">
                {(fields, { add, remove }) => (
                  <>
                    {fields.map(({ key, name, ...restField }) => (
                      <Card
                        key={key}
                        size="small"
                        style={{ marginBottom: '8px' }}
                      >
                        <Row gutter={[16, 8]} align="middle">
                          <Col span={16}>
                            <Form.Item
                              {...restField}
                              name={[name, 'widget_type']}
                              noStyle
                            >
                              <Input
                                disabled
                                placeholder="组件类型"
                              />
                            </Form.Item>
                          </Col>
                          <Col span={8}>
                            <Form.Item
                              {...restField}
                              name={[name, 'enabled']}
                              valuePropName="checked"
                              noStyle
                            >
                              <Switch checkedChildren="显示" unCheckedChildren="隐藏" />
                            </Form.Item>
                          </Col>
                        </Row>
                      </Card>
                    ))}
                  </>
                )}
              </Form.List>
            </Card>
          </TabPane>

          {/* 移动端设置 */}
          <TabPane
            tab={
              <span>
                <MobileOutlined />
                移动端
              </span>
            }
            key="mobile"
          >
            <Card size="small" title="移动端布局">
              <div style={{ 
                textAlign: 'center', 
                padding: '40px 0',
                color: '#8c8c8c'
              }}>
                <MobileOutlined style={{ fontSize: '48px', marginBottom: '16px' }} />
                <p>移动端布局配置功能开发中</p>
                <p style={{ fontSize: '12px' }}>
                  将支持移动端专用的布局配置和组件适配
                </p>
              </div>
            </Card>
          </TabPane>
        </Tabs>
      </Form>
    </Drawer>
  );
};

export default ConfigDrawer;