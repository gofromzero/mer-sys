import { createBrowserRouter, Navigate } from 'react-router-dom';
import { MainLayout } from '../components/layouts/MainLayout';
import { LoginPage } from '../pages/auth/LoginPage';
import { DashboardPage } from '../pages/dashboard/DashboardPage';
import TenantListPage from '../pages/tenant/TenantListPage';
import TenantRegistrationPage from '../pages/tenant/TenantRegistrationPage';
import { MerchantListPage } from '../pages/merchant/MerchantListPage';
import { MerchantRegistrationPage } from '../pages/merchant/MerchantRegistrationPage';
import { MerchantUserListPage } from '../pages/merchant-user/MerchantUserListPage';
import { MerchantUserFormPage } from '../pages/merchant-user/MerchantUserFormPage';
import { MerchantUserBatchCreatePage } from '../pages/merchant-user/MerchantUserBatchCreatePage';
import { MerchantUserStatusManagePage } from '../pages/merchant-user/MerchantUserStatusManagePage';
import { MerchantUserAuditLogPage } from '../pages/merchant-user/MerchantUserAuditLogPage';
import { 
  RightsMonitoringDashboard, 
  AlertConfigurationPage, 
  AlertListPage, 
  UsageReportPage 
} from '../pages/monitoring';
import { 
  ProductListPage, 
  CategoryManagePage,
  InventoryHistoryPage,
  InventoryMonitoringPage,
  InventoryAlertPage,
  InventoryBatchPage,
  InventoryStocktakingPage
} from '../pages/merchant/products';
import CartPage from '../pages/customer/cart/CartPage';
import OrderListPage from '../pages/customer/orders/OrderListPage';
import OrderDetailPage from '../pages/customer/orders/OrderDetailPage';
import OrderFormPage from '../pages/customer/orders/OrderFormPage';

export const router = createBrowserRouter([
  {
    path: '/auth/login',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: <MainLayout />,
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: 'dashboard',
        element: <DashboardPage />,
      },
      {
        path: 'tenant',
        children: [
          {
            index: true,
            element: <TenantListPage />,
          },
          {
            path: 'list',
            element: <TenantListPage />,
          },
          {
            path: 'register',
            element: <TenantRegistrationPage />,
          },
        ],
      },
      {
        path: 'merchant',
        children: [
          {
            index: true,
            element: <MerchantListPage />,
          },
          {
            path: 'list',
            element: <MerchantListPage />,
          },
          {
            path: 'register',
            element: <MerchantRegistrationPage />,
          },
          {
            path: 'products',
            element: <ProductListPage />,
          },
          {
            path: 'categories',
            element: <CategoryManagePage />,
          },
          {
            path: 'inventory',
            children: [
              {
                index: true,
                element: <InventoryMonitoringPage />,
              },
              {
                path: 'monitoring',
                element: <InventoryMonitoringPage />,
              },
              {
                path: 'history',
                element: <InventoryHistoryPage />,
              },
              {
                path: 'alerts',
                element: <InventoryAlertPage />,
              },
              {
                path: 'batch',
                element: <InventoryBatchPage />,
              },
              {
                path: 'stocktaking',
                element: <InventoryStocktakingPage />,
              },
            ],
          },
        ],
      },
      {
        path: 'merchant-user',
        children: [
          {
            index: true,
            element: <MerchantUserListPage />,
          },
          {
            path: 'list',
            element: <MerchantUserListPage />,
          },
          {
            path: 'create',
            element: <MerchantUserFormPage />,
          },
          {
            path: 'edit/:id',
            element: <MerchantUserFormPage />,
          },
          {
            path: 'batch-create',
            element: <MerchantUserBatchCreatePage />,
          },
          {
            path: 'status/:id',
            element: <MerchantUserStatusManagePage />,
          },
          {
            path: 'audit-log',
            element: <MerchantUserAuditLogPage />,
          },
        ],
      },
      {
        path: 'monitoring',
        children: [
          {
            index: true,
            element: <RightsMonitoringDashboard />,
          },
          {
            path: 'dashboard',
            element: <RightsMonitoringDashboard />,
          },
          {
            path: 'alerts',
            children: [
              {
                index: true,
                element: <AlertListPage />,
              },
              {
                path: 'list',
                element: <AlertListPage />,
              },
              {
                path: 'config',
                element: <AlertConfigurationPage />,
              },
            ],
          },
          {
            path: 'reports',
            element: <UsageReportPage />,
          },
        ],
      },
      {
        path: 'cart',
        element: <CartPage />,
      },
      {
        path: 'orders',
        children: [
          {
            index: true,
            element: <OrderListPage />,
          },
          {
            path: 'list',
            element: <OrderListPage />,
          },
          {
            path: 'create',
            element: <OrderFormPage />,
          },
          {
            path: ':id',
            element: <OrderDetailPage />,
          },
        ],
      },
    ],
  },
  {
    path: '*',
    element: <Navigate to="/dashboard" replace />,
  },
]);