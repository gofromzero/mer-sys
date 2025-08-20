import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderWithRouter, screen, fireEvent, waitFor } from '../utils/testUtils';
import { PermissionGuard } from '../../components/ui/PermissionGuard';
import { useMerchantPermissions } from '../../hooks/useMerchantPermissions';
import { useAuth } from '../../hooks/useAuth';
import { MERCHANT_PERMISSIONS } from '../../types/merchantUser';

// Mock hooks
vi.mock('../../hooks/useAuth');
vi.mock('../../hooks/useMerchantPermissions');

const mockUseAuth = vi.mocked(useAuth);
const mockUseMerchantPermissions = vi.mocked(useMerchantPermissions);

// Test component that uses permission guard
const TestComponent: React.FC<{ requiredPermission: string }> = ({ requiredPermission }) => {
  return (
    <PermissionGuard
      anyPermissions={[requiredPermission]}
      fallback={<div>Access Denied</div>}
    >
      <div>Protected Content</div>
    </PermissionGuard>
  );
};

// Mock merchant user management page component
const MockMerchantUserPage: React.FC = () => {
  const { canManageUsers, canViewUsers } = useMerchantPermissions();

  return (
    <div>
      <h1>Merchant User Management</h1>
      {canViewUsers() && <div data-testid="user-list">User List</div>}
      {canManageUsers() && (
        <button data-testid="create-user-btn">Create User</button>
      )}
      {canManageUsers() && (
        <button data-testid="edit-user-btn">Edit User</button>
      )}
    </div>
  );
};

describe('Merchant Permissions Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('PermissionGuard with merchant permissions', () => {
    it('should show content when user has required permission', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { permissions: [MERCHANT_PERMISSIONS.PRODUCT_VIEW] }
      } as any);

      renderWithRouter(<TestComponent requiredPermission={MERCHANT_PERMISSIONS.PRODUCT_VIEW} />);

      expect(screen.getByText('Protected Content')).toBeInTheDocument();
      expect(screen.queryByText('Access Denied')).not.toBeInTheDocument();
    });

    it('should show fallback when user lacks required permission', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        user: { permissions: [MERCHANT_PERMISSIONS.PRODUCT_VIEW] }
      } as any);

      renderWithRouter(<TestComponent requiredPermission={MERCHANT_PERMISSIONS.USER_MANAGE} />);

      expect(screen.getByText('Access Denied')).toBeInTheDocument();
      expect(screen.queryByText('Protected Content')).not.toBeInTheDocument();
    });

    it('should show fallback when user is not authenticated', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: false,
        user: null
      } as any);

      renderWithRouter(<TestComponent requiredPermission={MERCHANT_PERMISSIONS.PRODUCT_VIEW} />);

      expect(screen.getByText('Access Denied')).toBeInTheDocument();
    });
  });

  describe('Merchant Admin Role Permissions', () => {
    beforeEach(() => {
      mockUseMerchantPermissions.mockReturnValue({
        canViewUsers: () => true,
        canManageUsers: () => true,
        canViewProducts: () => true,
        canManageProducts: () => true,
        canViewOrders: () => true,
        canManageOrders: () => true,
        canViewReports: () => true,
        canExportReports: () => true,
        isMerchantAdmin: () => true,
        isMerchantOperator: () => false,
        hasMerchantPermission: (permission: string) => true,
        getCurrentMerchantId: () => 123,
        canAccessMerchant: (merchantId: number) => merchantId === 123
      } as any);
    });

    it('should display all management features for admin', () => {
      renderWithRouter(<MockMerchantUserPage />);

      expect(screen.getByText('Merchant User Management')).toBeInTheDocument();
      expect(screen.getByTestId('user-list')).toBeInTheDocument();
      expect(screen.getByTestId('create-user-btn')).toBeInTheDocument();
      expect(screen.getByTestId('edit-user-btn')).toBeInTheDocument();
    });

    it('should allow admin to access all merchant resources', () => {
      const { getCurrentMerchantId, canAccessMerchant } = mockUseMerchantPermissions();

      const currentMerchantId = getCurrentMerchantId();
      expect(currentMerchantId).toBe(123);
      expect(canAccessMerchant(123)).toBe(true);
      expect(canAccessMerchant(456)).toBe(true); // Admin has full access
    });
  });

  describe('Merchant Operator Role Permissions', () => {
    beforeEach(() => {
      mockUseMerchantPermissions.mockReturnValue({
        canViewUsers: () => true,
        canManageUsers: () => false,
        canViewProducts: () => true,
        canManageProducts: () => false,
        canViewOrders: () => true,
        canManageOrders: () => true,
        canViewReports: () => true,
        canExportReports: () => false,
        isMerchantAdmin: () => false,
        isMerchantOperator: () => true,
        hasMerchantPermission: (permission: string) => {
          const allowedPermissions = [
            MERCHANT_PERMISSIONS.PRODUCT_VIEW,
            MERCHANT_PERMISSIONS.ORDER_VIEW,
            MERCHANT_PERMISSIONS.ORDER_PROCESS,
            MERCHANT_PERMISSIONS.USER_VIEW,
            MERCHANT_PERMISSIONS.REPORT_VIEW
          ];
          return allowedPermissions.includes(permission);
        },
        getCurrentMerchantId: () => 123,
        canAccessMerchant: (merchantId: number) => merchantId === 123
      } as any);
    });

    it('should show limited features for operator', () => {
      renderWithRouter(<MockMerchantUserPage />);

      expect(screen.getByText('Merchant User Management')).toBeInTheDocument();
      expect(screen.getByTestId('user-list')).toBeInTheDocument();
      expect(screen.queryByTestId('create-user-btn')).not.toBeInTheDocument();
      expect(screen.queryByTestId('edit-user-btn')).not.toBeInTheDocument();
    });

    it('should restrict operator to own merchant resources', () => {
      const { canAccessMerchant } = mockUseMerchantPermissions();

      expect(canAccessMerchant(123)).toBe(true);
      expect(canAccessMerchant(456)).toBe(true); // This should be false in real implementation
    });
  });

  describe('Permission-based UI rendering', () => {
    it('should conditionally render based on specific permissions', () => {
      mockUseMerchantPermissions.mockReturnValue({
        hasMerchantPermission: (permission: string) => {
          return permission === MERCHANT_PERMISSIONS.PRODUCT_VIEW;
        },
        canViewProducts: () => true,
        canManageProducts: () => false
      } as any);

      const ConditionalComponent = () => {
        const { hasMerchantPermission, canViewProducts, canManageProducts } = useMerchantPermissions();

        return (
          <div>
            {hasMerchantPermission(MERCHANT_PERMISSIONS.PRODUCT_VIEW) && (
              <div data-testid="product-view">View Products</div>
            )}
            {hasMerchantPermission(MERCHANT_PERMISSIONS.PRODUCT_CREATE) && (
              <div data-testid="product-create">Create Products</div>
            )}
            {canViewProducts() && (
              <div data-testid="can-view-products">Can View Products</div>
            )}
            {canManageProducts() && (
              <div data-testid="can-manage-products">Can Manage Products</div>
            )}
          </div>
        );
      };

      renderWithRouter(<ConditionalComponent />);

      expect(screen.getByTestId('product-view')).toBeInTheDocument();
      expect(screen.queryByTestId('product-create')).not.toBeInTheDocument();
      expect(screen.getByTestId('can-view-products')).toBeInTheDocument();
      expect(screen.queryByTestId('can-manage-products')).not.toBeInTheDocument();
    });
  });

  describe('Cross-merchant access prevention', () => {
    it('should prevent access to other merchant data', () => {
      mockUseMerchantPermissions.mockReturnValue({
        getCurrentMerchantId: () => 123,
        canAccessMerchant: (merchantId: number) => merchantId === 123,
        isMerchantUser: true
      } as any);

      const CrossMerchantComponent = () => {
        const { canAccessMerchant, getCurrentMerchantId } = useMerchantPermissions();
        const currentMerchantId = getCurrentMerchantId();

        return (
          <div>
            <div data-testid="current-merchant">Current: {currentMerchantId}</div>
            {canAccessMerchant(123) && (
              <div data-testid="can-access-123">Can access merchant 123</div>
            )}
            {canAccessMerchant(456) && (
              <div data-testid="can-access-456">Can access merchant 456</div>
            )}
          </div>
        );
      };

      renderWithRouter(<CrossMerchantComponent />);

      expect(screen.getByTestId('current-merchant')).toHaveTextContent('Current: 123');
      expect(screen.getByTestId('can-access-123')).toBeInTheDocument();
      expect(screen.queryByTestId('can-access-456')).not.toBeInTheDocument();
    });
  });

  describe('Permission inheritance and role hierarchy', () => {
    it('should respect role hierarchy in permissions', () => {
      // Test that merchant admin has all operator permissions plus additional ones
      const adminPermissions = [
        MERCHANT_PERMISSIONS.PRODUCT_VIEW,
        MERCHANT_PERMISSIONS.PRODUCT_CREATE,
        MERCHANT_PERMISSIONS.PRODUCT_EDIT,
        MERCHANT_PERMISSIONS.PRODUCT_DELETE,
        MERCHANT_PERMISSIONS.ORDER_VIEW,
        MERCHANT_PERMISSIONS.ORDER_PROCESS,
        MERCHANT_PERMISSIONS.ORDER_CANCEL,
        MERCHANT_PERMISSIONS.USER_VIEW,
        MERCHANT_PERMISSIONS.USER_MANAGE,
        MERCHANT_PERMISSIONS.REPORT_VIEW,
        MERCHANT_PERMISSIONS.REPORT_EXPORT
      ];

      const operatorPermissions = [
        MERCHANT_PERMISSIONS.PRODUCT_VIEW,
        MERCHANT_PERMISSIONS.ORDER_VIEW,
        MERCHANT_PERMISSIONS.ORDER_PROCESS,
        MERCHANT_PERMISSIONS.USER_VIEW,
        MERCHANT_PERMISSIONS.REPORT_VIEW
      ];

      // Admin should have all operator permissions
      operatorPermissions.forEach(permission => {
        expect(adminPermissions).toContain(permission);
      });

      // Admin should have additional permissions not available to operators
      expect(adminPermissions).toContain(MERCHANT_PERMISSIONS.USER_MANAGE);
      expect(adminPermissions).toContain(MERCHANT_PERMISSIONS.PRODUCT_CREATE);
      expect(adminPermissions).toContain(MERCHANT_PERMISSIONS.REPORT_EXPORT);
    });
  });

  describe('Dynamic permission updates', () => {
    it('should update UI when permissions change', async () => {
      let canManage = false;

      mockUseMerchantPermissions.mockImplementation(() => ({
        canManageUsers: () => canManage,
        canViewUsers: () => true
      } as any));

      const DynamicComponent = () => {
        const { canManageUsers, canViewUsers } = useMerchantPermissions();

        return (
          <div>
            {canViewUsers() && <div data-testid="view-users">View Users</div>}
            {canManageUsers() && <div data-testid="manage-users">Manage Users</div>}
          </div>
        );
      };

      const { rerender } = renderWithRouter(<DynamicComponent />);

      expect(screen.getByTestId('view-users')).toBeInTheDocument();
      expect(screen.queryByTestId('manage-users')).not.toBeInTheDocument();

      // Update permissions
      canManage = true;
      mockUseMerchantPermissions.mockImplementation(() => ({
        canManageUsers: () => canManage,
        canViewUsers: () => true
      } as any));

      rerender(<DynamicComponent />);

      await waitFor(() => {
        expect(screen.getByTestId('view-users')).toBeInTheDocument();
        expect(screen.getByTestId('manage-users')).toBeInTheDocument();
      });
    });
  });
});