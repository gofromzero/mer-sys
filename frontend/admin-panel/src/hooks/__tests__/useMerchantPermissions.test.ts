import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import { renderHook } from '@testing-library/react';
import { useMerchantPermissions } from '../useMerchantPermissions';
import { useAuth } from '../useAuth';
import { MERCHANT_PERMISSIONS } from '../../types/merchantUser';

// Mock the useAuth hook
jest.mock('../useAuth', () => ({
  useAuth: jest.fn()
}));

const mockUseAuth = useAuth as jest.Mocked<typeof useAuth>;

describe('useMerchantPermissions', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('when user is not authenticated', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: null,
        isAuthenticated: false
      } as any);
    });

    it('should return null merchant context', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.merchantContext).toBeNull();
      expect(result.current.isLoggedIn).toBe(false);
      expect(result.current.isMerchantUser).toBe(false);
    });

    it('should return false for all permission checks', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.hasMerchantPermission('any_permission')).toBe(false);
      expect(result.current.isMerchantAdmin()).toBe(false);
      expect(result.current.canManageProducts()).toBe(false);
      expect(result.current.canViewProducts()).toBe(false);
    });
  });

  describe('when user is authenticated but not a merchant user', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 1,
          username: 'platform_admin',
          merchant_id: null,
          roles: [
            {
              role_type: 'platform_admin',
              permissions: ['platform:all']
            }
          ]
        },
        isAuthenticated: true
      } as any);
    });

    it('should return null merchant context', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.merchantContext).toBeNull();
      expect(result.current.isLoggedIn).toBe(true);
      expect(result.current.isMerchantUser).toBe(false);
    });
  });

  describe('when user is a merchant admin', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 1,
          username: 'merchant_admin',
          merchant_id: 123,
          roles: [
            {
              role_type: 'merchant_admin',
              permissions: [
                MERCHANT_PERMISSIONS.PRODUCT_VIEW,
                MERCHANT_PERMISSIONS.PRODUCT_CREATE,
                MERCHANT_PERMISSIONS.PRODUCT_EDIT,
                MERCHANT_PERMISSIONS.PRODUCT_DELETE,
                MERCHANT_PERMISSIONS.ORDER_VIEW,
                MERCHANT_PERMISSIONS.ORDER_PROCESS,
                MERCHANT_PERMISSIONS.USER_VIEW,
                MERCHANT_PERMISSIONS.USER_MANAGE,
                MERCHANT_PERMISSIONS.REPORT_VIEW,
                MERCHANT_PERMISSIONS.REPORT_EXPORT
              ]
            }
          ]
        },
        isAuthenticated: true
      } as any);
    });

    it('should return correct merchant context', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.merchantContext).toEqual({
        merchant_id: 123,
        role_type: 'merchant_admin',
        permissions: expect.arrayContaining([
          MERCHANT_PERMISSIONS.PRODUCT_VIEW,
          MERCHANT_PERMISSIONS.USER_MANAGE
        ])
      });
      expect(result.current.isMerchantUser).toBe(true);
    });

    it('should correctly identify as merchant admin', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.isMerchantAdmin()).toBe(true);
      expect(result.current.isMerchantOperator()).toBe(false);
      expect(result.current.hasMerchantRole('merchant_admin')).toBe(true);
    });

    it('should have all product permissions', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.canViewProducts()).toBe(true);
      expect(result.current.canManageProducts()).toBe(true);
      expect(result.current.hasMerchantPermission(MERCHANT_PERMISSIONS.PRODUCT_CREATE)).toBe(true);
      expect(result.current.hasMerchantPermission(MERCHANT_PERMISSIONS.PRODUCT_DELETE)).toBe(true);
    });

    it('should have user management permissions', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.canViewUsers()).toBe(true);
      expect(result.current.canManageUsers()).toBe(true);
    });

    it('should handle multiple permission checks', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.hasAnyMerchantPermission([
        MERCHANT_PERMISSIONS.PRODUCT_VIEW,
        MERCHANT_PERMISSIONS.ORDER_VIEW
      ])).toBe(true);

      expect(result.current.hasAllMerchantPermissions([
        MERCHANT_PERMISSIONS.PRODUCT_VIEW,
        MERCHANT_PERMISSIONS.USER_MANAGE
      ])).toBe(true);

      expect(result.current.hasAllMerchantPermissions([
        MERCHANT_PERMISSIONS.PRODUCT_VIEW,
        'non_existent_permission'
      ])).toBe(false);
    });
  });

  describe('when user is a merchant operator', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 2,
          username: 'merchant_operator',
          merchant_id: 123,
          roles: [
            {
              role_type: 'merchant_operator',
              permissions: [
                MERCHANT_PERMISSIONS.PRODUCT_VIEW,
                MERCHANT_PERMISSIONS.ORDER_VIEW,
                MERCHANT_PERMISSIONS.USER_VIEW,
                MERCHANT_PERMISSIONS.REPORT_VIEW
              ]
            }
          ]
        },
        isAuthenticated: true
      } as any);
    });

    it('should correctly identify as merchant operator', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.isMerchantOperator()).toBe(true);
      expect(result.current.isMerchantAdmin()).toBe(false);
      expect(result.current.hasMerchantRole('merchant_operator')).toBe(true);
    });

    it('should have limited product permissions', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.canViewProducts()).toBe(true);
      expect(result.current.canManageProducts()).toBe(false); // No create/edit/delete permissions
      expect(result.current.hasMerchantPermission(MERCHANT_PERMISSIONS.PRODUCT_VIEW)).toBe(true);
      expect(result.current.hasMerchantPermission(MERCHANT_PERMISSIONS.PRODUCT_CREATE)).toBe(false);
    });

    it('should not have user management permissions', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.canViewUsers()).toBe(true);
      expect(result.current.canManageUsers()).toBe(false);
    });

    it('should have limited report permissions', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.canViewReports()).toBe(true);
      expect(result.current.canExportReports()).toBe(false);
    });
  });

  describe('merchant access control', () => {
    beforeEach(() => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 1,
          username: 'merchant_user',
          merchant_id: 123,
          roles: [
            {
              role_type: 'merchant_admin',
              permissions: [MERCHANT_PERMISSIONS.PRODUCT_VIEW]
            }
          ]
        },
        isAuthenticated: true
      } as any);
    });

    it('should return current merchant ID', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.getCurrentMerchantId()).toBe(123);
    });

    it('should allow access to own merchant resources', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.canAccessMerchant(123)).toBe(true);
    });

    it('should deny access to other merchant resources', () => {
      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.canAccessMerchant(456)).toBe(false);
    });
  });

  describe('edge cases', () => {
    it('should handle user with merchant_id but no merchant roles', () => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 1,
          username: 'user',
          merchant_id: 123,
          roles: [
            {
              role_type: 'tenant_admin',
              permissions: ['tenant:all']
            }
          ]
        },
        isAuthenticated: true
      } as any);

      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.merchantContext).toBeNull();
      expect(result.current.isMerchantUser).toBe(false);
    });

    it('should handle empty permissions array', () => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 1,
          username: 'user',
          merchant_id: 123,
          roles: [
            {
              role_type: 'merchant_operator',
              permissions: []
            }
          ]
        },
        isAuthenticated: true
      } as any);

      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.merchantContext?.permissions).toEqual([]);
      expect(result.current.canViewProducts()).toBe(false);
      expect(result.current.canManageProducts()).toBe(false);
    });

    it('should handle missing permissions field', () => {
      mockUseAuth.mockReturnValue({
        user: {
          id: 1,
          username: 'user',
          merchant_id: 123,
          roles: [
            {
              role_type: 'merchant_operator'
              // permissions field missing
            } as any
          ]
        },
        isAuthenticated: true
      } as any);

      const { result } = renderHook(() => useMerchantPermissions());

      expect(result.current.merchantContext?.permissions).toEqual([]);
    });
  });
});