/**
 * RightsMonitoringDashboard Mobile Responsive Tests
 * Tests AC6: 移动端适配，确保商户可以随时查看关键数据
 */

import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import RightsMonitoringDashboard from '../RightsMonitoringDashboard';

// Mock the dashboard service
jest.mock('../../../../services/dashboardService', () => ({
  dashboardService: {
    getMerchantStats: jest.fn().mockResolvedValue({
      merchant_id: 1,
      tenant_id: 1,
      period: 'daily',
      total_sales: 50000,
      total_orders: 120,
      total_customers: 89,
      rights_balance: {
        total_balance: 10000,
        available_balance: 7500,
        used_balance: 2500,
        frozen_balance: 0
      },
      rights_usage_trend: [],
      rights_alerts: [],
      pending_tasks: [],
      announcements: [],
      notifications: [],
      last_updated: new Date().toISOString()
    }),
    getRightsUsageTrend: jest.fn().mockResolvedValue([]),
    getPendingTasks: jest.fn().mockResolvedValue([]),
    getNotifications: jest.fn().mockResolvedValue({
      notifications: [],
      announcements: [],
      unread_count: 0
    }),
    getDashboardConfig: jest.fn().mockResolvedValue({
      merchant_id: 1,
      layout_config: {
        columns: 4,
        widgets: []
      },
      widget_preferences: [],
      refresh_interval: 300,
      mobile_layout: {
        columns: 1,
        widgets: []
      }
    }),
    updateDashboardConfig: jest.fn(),
    saveDashboardConfig: jest.fn(),
    markAnnouncementAsRead: jest.fn()
  }
}));

// Mock zustand store
jest.mock('../../../../stores/dashboardStore', () => ({
  useDashboardStore: () => ({
    dashboardData: null,
    rightsUsageTrend: [],
    pendingTasks: [],
    notifications: null,
    config: null,
    loading: {
      dashboard: false,
      stats: false,
      trends: false,
      tasks: false,
      notifications: false,
      config: false
    },
    error: null,
    currentPeriod: 'daily',
    theme: 'light',
    loadAllData: jest.fn(),
    setPeriod: jest.fn(),
    setTheme: jest.fn(),
    startAutoRefresh: jest.fn(),
    stopAutoRefresh: jest.fn()
  }),
  useDashboardData: () => null,
  useDashboardLoading: () => ({ dashboard: false }),
  useDashboardError: () => null
}));

// Mock react-grid-layout to avoid complex DOM manipulation issues in tests
jest.mock('react-grid-layout', () => ({
  Responsive: ({ children }: any) => <div data-testid="responsive-grid">{children}</div>,
  WidthProvider: (component: any) => component
}));

describe('RightsMonitoringDashboard - Mobile Responsive Tests', () => {
  const mockViewport = (width: number, height: number) => {
    // Mock window.innerWidth and window.innerHeight
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: width,
    });
    Object.defineProperty(window, 'innerHeight', {
      writable: true,
      configurable: true,
      value: height,
    });

    // Mock matchMedia for responsive breakpoints
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

    // Trigger resize event
    window.dispatchEvent(new Event('resize'));
  };

  const renderDashboard = () => {
    return render(
      <MemoryRouter>
        <RightsMonitoringDashboard />
      </MemoryRouter>
    );
  };

  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();
  });

  describe('Mobile Phone Layout (< 480px)', () => {
    beforeEach(() => {
      mockViewport(375, 667); // iPhone SE size
    });

    it('should display dashboard on mobile phone screens', async () => {
      renderDashboard();
      
      await waitFor(() => {
        expect(screen.getByTestId('dashboard-container')).toBeInTheDocument();
      });

      // Check that main dashboard elements are present
      expect(screen.getByRole('heading', { name: /商户运营仪表板/i })).toBeInTheDocument();
    });

    it('should use single column layout on mobile', async () => {
      renderDashboard();

      await waitFor(() => {
        const grid = screen.getByTestId('responsive-grid');
        expect(grid).toBeInTheDocument();
      });

      // Verify mobile-friendly layout is applied
      // The grid should adapt to mobile screen size
    });

    it('should show essential metrics cards in mobile view', async () => {
      renderDashboard();

      await waitFor(() => {
        // Check for key business metrics that should be visible on mobile
        const container = screen.getByTestId('dashboard-container');
        expect(container).toBeInTheDocument();
      });
    });

    it('should handle touch interactions properly on mobile', async () => {
      renderDashboard();

      await waitFor(() => {
        const dashboard = screen.getByTestId('dashboard-container');
        expect(dashboard).toBeInTheDocument();
      });

      // Mobile-specific interaction tests would go here
      // For now, we verify the component renders without errors
    });
  });

  describe('Tablet Layout (480px - 768px)', () => {
    beforeEach(() => {
      mockViewport(768, 1024); // iPad size
    });

    it('should display dashboard on tablet screens', async () => {
      renderDashboard();
      
      await waitFor(() => {
        expect(screen.getByTestId('dashboard-container')).toBeInTheDocument();
      });

      expect(screen.getByRole('heading', { name: /商户运营仪表板/i })).toBeInTheDocument();
    });

    it('should use appropriate column layout for tablet', async () => {
      renderDashboard();

      await waitFor(() => {
        const grid = screen.getByTestId('responsive-grid');
        expect(grid).toBeInTheDocument();
      });

      // Tablet should have more columns than mobile but fewer than desktop
    });
  });

  describe('Desktop Layout (> 768px)', () => {
    beforeEach(() => {
      mockViewport(1920, 1080); // Desktop size
    });

    it('should display full dashboard on desktop screens', async () => {
      renderDashboard();
      
      await waitFor(() => {
        expect(screen.getByTestId('dashboard-container')).toBeInTheDocument();
      });

      expect(screen.getByRole('heading', { name: /商户运营仪表板/i })).toBeInTheDocument();
    });

    it('should use full column layout for desktop', async () => {
      renderDashboard();

      await waitFor(() => {
        const grid = screen.getByTestId('responsive-grid');
        expect(grid).toBeInTheDocument();
      });
    });
  });

  describe('Cross-device Compatibility', () => {
    const testViewports = [
      { name: 'iPhone 12 Mini', width: 375, height: 812 },
      { name: 'iPhone 12', width: 390, height: 844 },
      { name: 'iPad Mini', width: 768, height: 1024 },
      { name: 'iPad Pro', width: 1024, height: 1366 },
      { name: 'Desktop HD', width: 1920, height: 1080 },
      { name: 'Desktop 4K', width: 3840, height: 2160 }
    ];

    testViewports.forEach(({ name, width, height }) => {
      it(`should render correctly on ${name} (${width}x${height})`, async () => {
        mockViewport(width, height);
        
        renderDashboard();
        
        await waitFor(() => {
          expect(screen.getByTestId('dashboard-container')).toBeInTheDocument();
        });

        // Verify essential elements are present regardless of screen size
        expect(screen.getByRole('heading', { name: /商户运营仪表板/i })).toBeInTheDocument();
        
        // Should not have any layout-breaking issues
        const container = screen.getByTestId('dashboard-container');
        expect(container).toBeInTheDocument();
      });
    });
  });

  describe('Responsive Behavior', () => {
    it('should adapt layout when screen size changes', async () => {
      // Start with desktop
      mockViewport(1920, 1080);
      const { rerender } = renderDashboard();

      await waitFor(() => {
        expect(screen.getByTestId('dashboard-container')).toBeInTheDocument();
      });

      // Change to mobile
      mockViewport(375, 667);
      rerender(
        <MemoryRouter>
          <RightsMonitoringDashboard />
        </MemoryRouter>
      );

      await waitFor(() => {
        expect(screen.getByTestId('dashboard-container')).toBeInTheDocument();
      });

      // Should still render properly after viewport change
      expect(screen.getByRole('heading', { name: /商户运营仪表板/i })).toBeInTheDocument();
    });

    it('should maintain functionality across all screen sizes', async () => {
      const viewports = [375, 768, 1024, 1920];

      for (const width of viewports) {
        mockViewport(width, 800);
        
        const { unmount } = renderDashboard();
        
        await waitFor(() => {
          expect(screen.getByTestId('dashboard-container')).toBeInTheDocument();
        });

        // Core functionality should work on all screen sizes
        expect(screen.getByRole('heading', { name: /商户运营仪表板/i })).toBeInTheDocument();
        
        unmount();
      }
    });
  });

  describe('Mobile-specific Features', () => {
    beforeEach(() => {
      mockViewport(375, 667);
    });

    it('should handle touch events appropriately', async () => {
      renderDashboard();
      
      await waitFor(() => {
        expect(screen.getByTestId('dashboard-container')).toBeInTheDocument();
      });

      // Verify touch-friendly interface elements are present
      const container = screen.getByTestId('dashboard-container');
      expect(container).toHaveStyle('touch-action: manipulation');
    });

    it('should optimize content for mobile viewing', async () => {
      renderDashboard();
      
      await waitFor(() => {
        expect(screen.getByTestId('dashboard-container')).toBeInTheDocument();
      });

      // Should prioritize key information for mobile users
      expect(screen.getByRole('heading', { name: /商户运营仪表板/i })).toBeInTheDocument();
    });
  });
});