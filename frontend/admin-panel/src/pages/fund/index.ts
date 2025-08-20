// 资金管理页面导出

export { default as FundDashboardPage } from './FundDashboardPage';
export { default as FundDepositPage } from './FundDepositPage';
export { default as FundAllocationPage } from './FundAllocationPage';

// 导出页面组件映射，用于路由配置
export const fundPageComponents = {
  'fund-dashboard': () => import('./FundDashboardPage'),
  'fund-deposit': () => import('./FundDepositPage'),
  'fund-allocation': () => import('./FundAllocationPage'),
};

// 导出页面信息，用于菜单配置
export const fundPageInfo = [
  {
    path: '/fund/dashboard',
    name: 'fund-dashboard',
    title: '资金总览',
    icon: '📊',
    component: 'FundDashboardPage',
    description: '查看系统整体资金状况和商户权益分布'
  },
  {
    path: '/fund/deposit',
    name: 'fund-deposit',
    title: '资金充值',
    icon: '💰',
    component: 'FundDepositPage',
    description: '为商户账户充值资金，支持单笔和批量操作'
  },
  {
    path: '/fund/allocation',
    name: 'fund-allocation',
    title: '权益分配',
    icon: '📈',
    component: 'FundAllocationPage',
    description: '为商户分配权益额度，用于支持业务运营'
  }
];