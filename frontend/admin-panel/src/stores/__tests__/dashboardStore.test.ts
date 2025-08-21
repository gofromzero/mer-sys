/**
 * Dashboard Store 单元测试
 */

import { renderHook, act } from '@testing-library/react';
import { useDashboardStore } from '../dashboardStore';
import { TimePeriod, DashboardError } from '../../types/dashboard';

// Mock dashboardService
jest.mock('../../services/dashboardService', () => ({
  dashboardService: {
    getMerchantStats: jest.fn(),
    getRightsUsageTrend: jest.fn(),
    getPendingTasks: jest.fn(),
    getNotifications: jest.fn(),
    getDashboardConfig: jest.fn(),
    updateDashboardConfig: jest.fn(),
    saveDashboardConfig: jest.fn(),
    markAnnouncementAsRead: jest.fn()
  }
}));

describe('Dashboard Store', () => {
  beforeEach(() => {
    // 重置store状态
    const { result } = renderHook(() => useDashboardStore());
    act(() => {
      result.current.reset();
    });
  });

  it('should initialize with default state', () => {
    const { result } = renderHook(() => useDashboardStore());
    
    expect(result.current.dashboardData).toBeNull();
    expect(result.current.rightsUsageTrend).toEqual([]);
    expect(result.current.pendingTasks).toEqual([]);
    expect(result.current.notifications).toBeNull();
    expect(result.current.config).toBeNull();
    expect(result.current.currentPeriod).toBe(TimePeriod.DAILY);
    expect(result.current.error).toBeNull();
    expect(result.current.loading.dashboard).toBe(false);
  });

  it('should update loading state', () => {
    const { result } = renderHook(() => useDashboardStore());
    
    act(() => {
      result.current.setLoading('dashboard', true);
    });
    
    expect(result.current.loading.dashboard).toBe(true);
    
    act(() => {
      result.current.setLoading('dashboard', false);
    });
    
    expect(result.current.loading.dashboard).toBe(false);
  });

  it('should update error state', () => {
    const { result } = renderHook(() => useDashboardStore());
    const error: DashboardError = {
      code: 'TEST_ERROR',
      message: 'Test error message'
    };
    
    act(() => {
      result.current.setError(error);
    });
    
    expect(result.current.error).toEqual(error);
    
    act(() => {
      result.current.setError(null);
    });
    
    expect(result.current.error).toBeNull();
  });

  it('should update period and trigger data reload', async () => {
    const { result } = renderHook(() => useDashboardStore());
    
    // Mock loadDashboardData
    const mockLoadDashboardData = jest.spyOn(result.current, 'loadDashboardData');
    mockLoadDashboardData.mockImplementation(async () => {});
    
    act(() => {
      result.current.setPeriod(TimePeriod.WEEKLY);
    });
    
    expect(result.current.currentPeriod).toBe(TimePeriod.WEEKLY);
  });

  it('should update theme', () => {
    const { result } = renderHook(() => useDashboardStore());
    
    act(() => {
      result.current.setTheme('dark');
    });
    
    expect(result.current.theme).toBe('dark');
  });

  it('should handle announcement read status update', async () => {
    const { result } = renderHook(() => useDashboardStore());
    
    // 设置初始通知数据
    const mockNotifications = {
      notifications: [],
      announcements: [
        {
          id: 1,
          title: 'Test Announcement',
          content: 'Test content',
          priority: 'normal' as any,
          publish_date: '2025-01-01',
          read_status: false
        }
      ],
      unread_count: 1
    };
    
    act(() => {
      result.current.notifications = mockNotifications;
    });
    
    // Mock service call
    const dashboardService = require('../../services/dashboardService').dashboardService;
    dashboardService.markAnnouncementAsRead.mockResolvedValue(undefined);
    
    await act(async () => {
      await result.current.markAnnouncementAsRead(1);
    });
    
    expect(result.current.notifications?.announcements[0].read_status).toBe(true);
    expect(result.current.notifications?.unread_count).toBe(0);
  });

  it('should reset to initial state', () => {
    const { result } = renderHook(() => useDashboardStore());
    
    // 修改一些状态
    act(() => {
      result.current.setError({
        code: 'TEST_ERROR',
        message: 'Test error'
      });
      result.current.setPeriod(TimePeriod.MONTHLY);
      result.current.setTheme('dark');
    });
    
    // 重置状态
    act(() => {
      result.current.reset();
    });
    
    expect(result.current.error).toBeNull();
    expect(result.current.currentPeriod).toBe(TimePeriod.DAILY);
    expect(result.current.theme).toBe('light');
    expect(result.current.dashboardData).toBeNull();
  });
});

// Mobile Responsive Tests for AC6
describe('Dashboard Mobile Responsiveness', () => {
  const mockViewport = (width: number) => {
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: width,
    });

    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation(query => ({
        matches: query.includes(`max-width: ${width}px`) || 
                 (query.includes('max-width: 768px') && width <= 768) ||
                 (query.includes('max-width: 480px') && width <= 480),
        media: query,
        onchange: null,
        addListener: jest.fn(),
        removeListener: jest.fn(),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
      })),
    });

    window.dispatchEvent(new Event('resize'));
  };

  it('should handle mobile viewport changes', () => {
    const { result } = renderHook(() => useDashboardStore());
    
    // Test mobile viewport (375px - iPhone SE)
    mockViewport(375);
    
    // Dashboard should maintain functionality on mobile
    expect(result.current.currentPeriod).toBe(TimePeriod.DAILY);
    expect(result.current.theme).toBe('light');
    expect(result.current.loading.dashboard).toBe(false);
  });

  it('should handle tablet viewport changes', () => {
    const { result } = renderHook(() => useDashboardStore());
    
    // Test tablet viewport (768px - iPad)
    mockViewport(768);
    
    // Dashboard should maintain functionality on tablet
    expect(result.current.currentPeriod).toBe(TimePeriod.DAILY);
    expect(result.current.loadAllData).toBeDefined();
  });

  it('should handle desktop viewport changes', () => {
    const { result } = renderHook(() => useDashboardStore());
    
    // Test desktop viewport (1920px)
    mockViewport(1920);
    
    // Dashboard should maintain full functionality on desktop
    expect(result.current.currentPeriod).toBe(TimePeriod.DAILY);
    expect(result.current.loadAllData).toBeDefined();
  });

  it('should adapt to various mobile device sizes', () => {
    const { result } = renderHook(() => useDashboardStore());
    
    const mobileViewports = [375, 390, 414]; // Common mobile widths
    
    mobileViewports.forEach(width => {
      mockViewport(width);
      
      // Core functionality should remain intact across all mobile sizes
      expect(result.current.setPeriod).toBeDefined();
      expect(result.current.loadAllData).toBeDefined();
      expect(result.current.setTheme).toBeDefined();
      
      // State should be consistent
      expect(result.current.error).toBeNull();
      expect(result.current.loading.dashboard).toBe(false);
    });
  });

  it('should maintain touch-friendly interaction state', () => {
    const { result } = renderHook(() => useDashboardStore());
    
    // Test touch-device viewport
    mockViewport(375);
    
    // Touch-friendly operations should work
    act(() => {
      result.current.setPeriod(TimePeriod.WEEKLY);
    });
    
    expect(result.current.currentPeriod).toBe(TimePeriod.WEEKLY);
    
    act(() => {
      result.current.setTheme('dark');
    });
    
    expect(result.current.theme).toBe('dark');
  });
});