import { createBrowserRouter, Navigate } from 'react-router-dom';
import { MainLayout } from '../components/layouts/MainLayout';
import { LoginPage } from '../pages/auth/LoginPage';
import { DashboardPage } from '../pages/dashboard/DashboardPage';
import TenantListPage from '../pages/tenant/TenantListPage';
import TenantRegistrationPage from '../pages/tenant/TenantRegistrationPage';

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
    ],
  },
  {
    path: '*',
    element: <Navigate to="/dashboard" replace />,
  },
]);