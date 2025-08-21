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
import { ProductListPage, CategoryManagePage } from '../pages/merchant/products';

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
    ],
  },
  {
    path: '*',
    element: <Navigate to="/dashboard" replace />,
  },
]);