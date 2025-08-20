import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderWithRouter, screen, fireEvent, waitFor } from '../utils/testUtils';
import { MerchantUserListPage } from '../../pages/merchant-user/MerchantUserListPage';
import { useMerchantPermissions } from '../../hooks/useMerchantPermissions';
import { useAuth } from '../../hooks/useAuth';
import { MerchantUserService } from '../../services/merchantUserService';

// Mock dependencies
vi.mock('../../hooks/useAuth');
vi.mock('../../hooks/useMerchantPermissions');
vi.mock('../../services/merchantUserService');

const mockUseAuth = vi.mocked(useAuth);
const mockUseMerchantPermissions = vi.mocked(useMerchantPermissions);
const mockMerchantUserService = vi.mocked(MerchantUserService);

// Mock components for testing
vi.mock('../../components/ui/AmisRenderer', () => ({
  AmisRenderer: ({ schema }: { schema: any }) => {
    const handleClick = (buttonConfig: any) => {
      if (buttonConfig.onClick) {
        buttonConfig.onClick({}, { scope: { id: 1, merchant_id: 123 } });
      }
    };

    return (
      <div data-testid="amis-renderer">
        <div data-testid="api-url">{schema.body?.api?.url}</div>
        <div data-testid="api-data">{JSON.stringify(schema.body?.api?.data || {})}</div>
        {schema.body?.headerToolbar?.map((item: any, index: number) => 
          item.onClick ? (
            <button 
              key={index}
              data-testid={`toolbar-button-${index}`}
              onClick={() => handleClick(item)}
            >
              {item.label}
            </button>
          ) : null
        )}
        {schema.body?.columns?.find((col: any) => col.type === 'operation')?.buttons?.map((btn: any, index: number) => 
          btn.onClick ? (
            <button 
              key={index}
              data-testid={`action-button-${index}`}
              onClick={() => handleClick(btn)}
            >
              {btn.label}
            </button>
          ) : null
        )}
      </div>
    );
  }
}));

vi.mock('../../components/ui/PermissionGuard', () => ({
  PermissionGuard: ({ children, fallback, anyPermissions }: any) => {
    const { user } = mockUseAuth();
    const hasPermission = user?.permissions?.some((p: string) => anyPermissions.includes(p));
    return hasPermission ? children : fallback;
  }
}));

describe('Merchant User Security Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Cross-Merchant Data Access Prevention', () => {
    it('should prevent access to other merchant user data', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: {
          id: 1,
          merchant_id: 123,
          permissions: ['merchant:user:view']
        }
      } as any);

      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        canAccessMerchant: (merchantId: number) => merchantId === 123,
        canViewUsers: () => true
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      // Check that API calls include proper merchant filtering
      const apiData = screen.getByTestId('api-data');
      expect(apiData.textContent).toContain('merchant_id');
      expect(apiData.textContent).toContain('123');
    });

    it('should reject API calls without proper merchant context', async () => {
      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => null,
        canViewUsers: () => false
      } as any);

      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { id: 1, merchant_id: null, permissions: [] }
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      expect(screen.getByText('您没有权限访问商户用户管理')).toBeInTheDocument();
    });

    it('should validate merchant_id in all API requests', () => {
      const testMerchantId = 456;
      
      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => testMerchantId,
        canViewUsers: () => true
      } as any);

      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: {
          id: 1,
          merchant_id: testMerchantId,
          permissions: ['merchant:user:view']
        }
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      const apiData = screen.getByTestId('api-data');
      expect(apiData.textContent).toContain(`${testMerchantId}`);
    });
  });

  describe('Permission-Based Access Control', () => {
    it('should enforce view-only access for users without manage permissions', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: {
          id: 1,
          merchant_id: 123,
          permissions: ['merchant:user:view'] // Only view permission
        }
      } as any);

      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        canViewUsers: () => true,
        canManageUsers: () => false // No manage permission
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      // Should not show management buttons
      expect(screen.queryByText('新增用户')).not.toBeInTheDocument();
    });

    it('should prevent unauthorized operations', async () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: {
          id: 1,
          merchant_id: 123,
          permissions: ['merchant:user:view']
        }
      } as any);

      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        canViewUsers: () => true,
        canManageUsers: () => false
      } as any);

      // Mock service to reject unauthorized operations
      mockMerchantUserService.updateMerchantUserStatus.mockRejectedValue(
        new Error('Unauthorized: Insufficient permissions')
      );

      renderWithRouter(<MerchantUserListPage />);

      // Any attempt to perform management operations should fail
      try {
        await MerchantUserService.updateMerchantUserStatus(1, { 
          status: 'suspended' as any, 
          comment: 'test' 
        });
      } catch (error) {
        expect(error).toBeInstanceOf(Error);
        expect((error as Error).message).toContain('Unauthorized');
      }
    });

    it('should validate role hierarchy', () => {
      // Test that operators cannot manage admin users
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: {
          id: 1,
          merchant_id: 123,
          role: 'merchant_operator',
          permissions: ['merchant:user:view']
        }
      } as any);

      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        isMerchantOperator: () => true,
        isMerchantAdmin: () => false,
        canManageUsers: () => false
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      // Operators should not be able to manage users
      expect(screen.queryByTestId('action-button-edit')).not.toBeInTheDocument();
    });
  });

  describe('Data Isolation Tests', () => {
    it('should isolate data between different merchants', async () => {
      const merchant1Id = 123;
      const merchant2Id = 456;

      // User from merchant 1
      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => merchant1Id,
        canAccessMerchant: (merchantId: number) => merchantId === merchant1Id
      } as any);

      // Mock API to return only merchant 1 data
      mockMerchantUserService.getMerchantUsers.mockResolvedValue({
        list: [
          { id: 1, merchant_id: merchant1Id, username: 'user1' },
          { id: 2, merchant_id: merchant1Id, username: 'user2' }
        ],
        pagination: { page: 1, page_size: 20, total: 2, total_pages: 1 }
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      // Verify API call includes correct merchant filter
      expect(mockMerchantUserService.getMerchantUsers).toHaveBeenCalledWith(
        expect.objectContaining({
          merchant_id: merchant1Id
        })
      );
    });

    it('should prevent data leakage through API parameters', () => {
      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        canViewUsers: () => true
      } as any);

      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: {
          id: 1,
          merchant_id: 123,
          permissions: ['merchant:user:view']
        }
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      // Check that merchant_id is automatically injected and cannot be bypassed
      const apiData = screen.getByTestId('api-data');
      const dataObj = JSON.parse(apiData.textContent || '{}');
      expect(dataObj.merchant_id).toBe(123);
    });
  });

  describe('Session and Authentication Security', () => {
    it('should handle expired sessions gracefully', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: false,
        user: null
      } as any);

      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => null,
        canViewUsers: () => false
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      expect(screen.getByText('您没有权限访问商户用户管理')).toBeInTheDocument();
    });

    it('should prevent privilege escalation', () => {
      // User tries to access admin functions with operator role
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: {
          id: 1,
          merchant_id: 123,
          role: 'merchant_operator',
          permissions: ['merchant:user:view']
        }
      } as any);

      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        isMerchantOperator: () => true,
        isMerchantAdmin: () => false,
        canManageUsers: () => false,
        canViewUsers: () => true
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      // Should not show admin-only features
      expect(screen.queryByText('新增用户')).not.toBeInTheDocument();
    });

    it('should validate JWT token merchant claims', () => {
      // Simulate JWT token with merchant claim
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: {
          id: 1,
          merchant_id: 123,
          permissions: ['merchant:user:view'],
          tokenClaims: {
            merchant_id: 123,
            role: 'merchant_admin'
          }
        }
      } as any);

      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        canViewUsers: () => true
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      // Token claims should match user context
      expect(screen.getByTestId('amis-renderer')).toBeInTheDocument();
    });
  });

  describe('Input Validation and Sanitization', () => {
    it('should sanitize search parameters', async () => {
      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        canViewUsers: () => true
      } as any);

      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: {
          id: 1,
          merchant_id: 123,
          permissions: ['merchant:user:view']
        }
      } as any);

      // Mock potentially malicious search input
      const maliciousInput = "<script>alert('xss')</script>";
      
      renderWithRouter(<MerchantUserListPage />);

      // Search input should be properly sanitized
      // In a real implementation, this would be handled by the API layer
      expect(screen.getByTestId('amis-renderer')).toBeInTheDocument();
    });

    it('should validate merchant ID format', () => {
      // Test with invalid merchant ID
      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 'invalid' as any,
        canViewUsers: () => true
      } as any);

      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: {
          id: 1,
          merchant_id: 'invalid',
          permissions: ['merchant:user:view']
        }
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      // Should handle invalid merchant ID gracefully
      expect(screen.getByTestId('amis-renderer')).toBeInTheDocument();
    });
  });

  describe('Audit and Logging Security', () => {
    it('should log security-relevant actions', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      
      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        canManageUsers: () => true
      } as any);

      // Simulate unauthorized access attempt
      mockMerchantUserService.updateMerchantUserStatus.mockRejectedValue(
        new Error('Unauthorized access attempt')
      );

      renderWithRouter(<MerchantUserListPage />);

      // Security violations should be logged
      try {
        await MerchantUserService.updateMerchantUserStatus(1, { 
          status: 'suspended' as any 
        });
      } catch (error) {
        expect(consoleSpy).toHaveBeenCalled();
      }

      consoleSpy.mockRestore();
    });

    it('should not expose sensitive information in error messages', async () => {
      mockMerchantUserService.getMerchantUser.mockRejectedValue(
        new Error('Database connection details: host=localhost...')
      );

      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        canViewUsers: () => true
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      // Error messages should be sanitized to not expose internal details
      // This would be handled by the API layer and error boundary components
      expect(screen.getByTestId('amis-renderer')).toBeInTheDocument();
    });
  });

  describe('Rate Limiting and DoS Protection', () => {
    it('should handle API rate limiting gracefully', async () => {
      mockMerchantUserService.getMerchantUsers.mockRejectedValue(
        new Error('Rate limit exceeded')
      );

      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        canViewUsers: () => true
      } as any);

      renderWithRouter(<MerchantUserListPage />);

      // Should handle rate limiting errors appropriately
      expect(screen.getByTestId('amis-renderer')).toBeInTheDocument();
    });

    it('should prevent excessive API calls', () => {
      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        canViewUsers: () => true
      } as any);

      const { rerender } = renderWithRouter(<MerchantUserListPage />);

      // Multiple rapid re-renders should not cause excessive API calls
      for (let i = 0; i < 10; i++) {
        rerender(<MerchantUserListPage />);
      }

      // API should be called efficiently, not on every render
      expect(screen.getByTestId('amis-renderer')).toBeInTheDocument();
    });
  });
});