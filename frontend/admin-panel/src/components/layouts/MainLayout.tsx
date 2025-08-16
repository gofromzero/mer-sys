import React from 'react';
import { Outlet } from 'react-router-dom';

interface MainLayoutProps {
  children?: React.ReactNode;
}

export const MainLayout: React.FC<MainLayoutProps> = () => {
  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <h1 className="text-xl font-semibold text-gray-900">
                Mer Demo - 管理后台
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-700">欢迎使用</span>
            </div>
          </div>
        </div>
      </header>
      
      <div className="flex">
        <nav className="w-64 bg-white shadow-sm h-screen overflow-y-auto">
          <div className="p-4">
            <ul className="space-y-2">
              <li>
                <a href="/dashboard" className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded">
                  仪表板
                </a>
              </li>
              <li>
                <a href="/tenant" className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded">
                  租户管理
                </a>
              </li>
            </ul>
          </div>
        </nav>
        
        <main className="flex-1 p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
};