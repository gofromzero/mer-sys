import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { renderHook, act } from '@testing-library/react';
import { useMerchantUserStore } from '../merchantUserStore';
import { MerchantUserService } from '../../services/merchantUserService';
import type { MerchantUser } from '../../types/merchantUser';

// Mock the MerchantUserService
jest.mock('../../services/merchantUserService', () => ({
  MerchantUserService: {
    getMerchantUsers: jest.fn(),
    getMerchantUser: jest.fn(),
    createMerchantUser: jest.fn(),
    updateMerchantUser: jest.fn(),
    updateMerchantUserStatus: jest.fn(),
    resetMerchantUserPassword: jest.fn(),
    deleteMerchantUser: jest.fn()
  }
}));

const mockMerchantUserService = MerchantUserService as jest.Mocked<typeof MerchantUserService>;

const mockUser: MerchantUser = {
  id: 1,
  uuid: 'test-uuid',
  username: 'test_user',
  email: 'test@example.com',
  tenant_id: 1,
  merchant_id: 123,
  status: 'active',
  role_type: 'merchant_operator',
  permissions: ['merchant:product:view'],
  created_at: '2025-08-20T10:00:00Z',
  updated_at: '2025-08-20T10:00:00Z'
};

describe('useMerchantUserStore', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  afterEach(() => {
    // Reset store state after each test
    const { result } = renderHook(() => useMerchantUserStore());
    act(() => {
      result.current.reset();
    });
  });

  describe('initial state', () => {
    it('should have correct initial state', () => {
      const { result } = renderHook(() => useMerchantUserStore());

      expect(result.current.merchantUsers).toEqual([]);
      expect(result.current.currentMerchantUser).toBeNull();
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBeNull();
      expect(result.current.pagination).toEqual({
        page: 1,
        page_size: 20,
        total: 0,
        total_pages: 0
      });
      expect(result.current.queryParams).toEqual({
        page: 1,
        page_size: 20
      });
    });
  });

  describe('setters', () => {
    it('should update merchantUsers', () => {
      const { result } = renderHook(() => useMerchantUserStore());

      act(() => {
        result.current.setMerchantUsers([mockUser]);
      });

      expect(result.current.merchantUsers).toEqual([mockUser]);
    });

    it('should update currentMerchantUser', () => {
      const { result } = renderHook(() => useMerchantUserStore());

      act(() => {
        result.current.setCurrentMerchantUser(mockUser);
      });

      expect(result.current.currentMerchantUser).toEqual(mockUser);
    });

    it('should update loading state', () => {
      const { result } = renderHook(() => useMerchantUserStore());

      act(() => {
        result.current.setLoading(true);
      });

      expect(result.current.loading).toBe(true);
    });

    it('should update error state', () => {
      const { result } = renderHook(() => useMerchantUserStore());

      act(() => {
        result.current.setError('Test error');
      });

      expect(result.current.error).toBe('Test error');
    });
  });

  describe('fetchMerchantUsers', () => {
    it('should fetch users successfully', async () => {
      const mockResponse = {
        list: [mockUser],
        pagination: {
          page: 1,
          page_size: 20,
          total: 1,
          total_pages: 1
        }
      };

      mockMerchantUserService.getMerchantUsers.mockResolvedValue(mockResponse);

      const { result } = renderHook(() => useMerchantUserStore());

      await act(async () => {
        await result.current.fetchMerchantUsers();
      });

      expect(result.current.merchantUsers).toEqual([mockUser]);
      expect(result.current.pagination).toEqual(mockResponse.pagination);
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBeNull();
    });

    it('should handle fetch error', async () => {
      const errorMessage = 'Failed to fetch users';
      mockMerchantUserService.getMerchantUsers.mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useMerchantUserStore());

      await act(async () => {
        await result.current.fetchMerchantUsers();
      });

      expect(result.current.merchantUsers).toEqual([]);
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBe(errorMessage);
    });

    it('should merge query parameters correctly', async () => {
      const mockResponse = {
        list: [],
        pagination: { page: 1, page_size: 10, total: 0, total_pages: 0 }
      };

      mockMerchantUserService.getMerchantUsers.mockResolvedValue(mockResponse);

      const { result } = renderHook(() => useMerchantUserStore());

      await act(async () => {
        await result.current.fetchMerchantUsers({ page_size: 10, status: 'active' });
      });

      expect(mockMerchantUserService.getMerchantUsers).toHaveBeenCalledWith({
        page: 1,
        page_size: 10,
        status: 'active'
      });
    });
  });

  describe('fetchMerchantUser', () => {
    it('should fetch single user successfully', async () => {
      mockMerchantUserService.getMerchantUser.mockResolvedValue(mockUser);

      const { result } = renderHook(() => useMerchantUserStore());

      await act(async () => {
        await result.current.fetchMerchantUser(1);
      });

      expect(result.current.currentMerchantUser).toEqual(mockUser);
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBeNull();
    });

    it('should handle fetch single user error', async () => {
      const errorMessage = 'User not found';
      mockMerchantUserService.getMerchantUser.mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useMerchantUserStore());

      await act(async () => {
        await result.current.fetchMerchantUser(999);
      });

      expect(result.current.currentMerchantUser).toBeNull();
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBe(errorMessage);
    });
  });

  describe('createMerchantUser', () => {
    it('should create user successfully', async () => {
      const newUserData = {
        username: 'new_user',
        email: 'new@example.com',
        merchant_id: 123,
        role_type: 'merchant_operator' as const,
        permissions: ['merchant:product:view']
      };

      const createdUser = { ...mockUser, id: 2, ...newUserData };
      mockMerchantUserService.createMerchantUser.mockResolvedValue(createdUser);

      const { result } = renderHook(() => useMerchantUserStore());

      // Set initial users
      act(() => {
        result.current.setMerchantUsers([mockUser]);
      });

      let returnedUser;
      await act(async () => {
        returnedUser = await result.current.createMerchantUser(newUserData);
      });

      expect(returnedUser).toEqual(createdUser);
      expect(result.current.merchantUsers).toEqual([createdUser, mockUser]);
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBeNull();
    });

    it('should handle create user error', async () => {
      const errorMessage = 'Username already exists';
      mockMerchantUserService.createMerchantUser.mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useMerchantUserStore());

      await act(async () => {
        try {
          await result.current.createMerchantUser({} as any);
        } catch (error) {
          // Expected to throw
        }
      });

      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBe(errorMessage);
    });
  });

  describe('updateMerchantUser', () => {
    it('should update user successfully', async () => {
      const updateData = { username: 'updated_user' };
      const updatedUser = { ...mockUser, ...updateData };

      mockMerchantUserService.updateMerchantUser.mockResolvedValue(updatedUser);

      const { result } = renderHook(() => useMerchantUserStore());

      // Set initial users
      act(() => {
        result.current.setMerchantUsers([mockUser]);
        result.current.setCurrentMerchantUser(mockUser);
      });

      let returnedUser;
      await act(async () => {
        returnedUser = await result.current.updateMerchantUser(1, updateData);
      });

      expect(returnedUser).toEqual(updatedUser);
      expect(result.current.merchantUsers[0]).toEqual(updatedUser);
      expect(result.current.currentMerchantUser).toEqual(updatedUser);
    });

    it('should handle update user error', async () => {
      const errorMessage = 'Update failed';
      mockMerchantUserService.updateMerchantUser.mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useMerchantUserStore());

      await act(async () => {
        try {
          await result.current.updateMerchantUser(1, {});
        } catch (error) {
          // Expected to throw
        }
      });

      expect(result.current.error).toBe(errorMessage);
    });
  });

  describe('updateMerchantUserStatus', () => {
    it('should update user status successfully', async () => {
      const updatedUser = { ...mockUser, status: 'suspended' as const };
      mockMerchantUserService.updateMerchantUserStatus.mockResolvedValue(updatedUser);

      const { result } = renderHook(() => useMerchantUserStore());

      // Set initial users
      act(() => {
        result.current.setMerchantUsers([mockUser]);
      });

      await act(async () => {
        await result.current.updateMerchantUserStatus(1, 'suspended', 'Violation');
      });

      expect(result.current.merchantUsers[0]).toEqual(updatedUser);
      expect(mockMerchantUserService.updateMerchantUserStatus).toHaveBeenCalledWith(1, {
        status: 'suspended',
        comment: 'Violation'
      });
    });
  });

  describe('resetMerchantUserPassword', () => {
    it('should reset password successfully', async () => {
      mockMerchantUserService.resetMerchantUserPassword.mockResolvedValue();

      const { result } = renderHook(() => useMerchantUserStore());

      await act(async () => {
        await result.current.resetMerchantUserPassword(1, { send_email: true });
      });

      expect(mockMerchantUserService.resetMerchantUserPassword).toHaveBeenCalledWith(1, {
        send_email: true
      });
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBeNull();
    });

    it('should handle reset password error', async () => {
      const errorMessage = 'Reset failed';
      mockMerchantUserService.resetMerchantUserPassword.mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useMerchantUserStore());

      await act(async () => {
        try {
          await result.current.resetMerchantUserPassword(1, {});
        } catch (error) {
          // Expected to throw
        }
      });

      expect(result.current.error).toBe(errorMessage);
    });
  });

  describe('deleteMerchantUser', () => {
    it('should delete user successfully', async () => {
      mockMerchantUserService.deleteMerchantUser.mockResolvedValue();

      const { result } = renderHook(() => useMerchantUserStore());

      // Set initial users
      act(() => {
        result.current.setMerchantUsers([mockUser]);
        result.current.setCurrentMerchantUser(mockUser);
      });

      await act(async () => {
        await result.current.deleteMerchantUser(1);
      });

      expect(result.current.merchantUsers).toEqual([]);
      expect(result.current.currentMerchantUser).toBeNull();
    });

    it('should handle delete user error', async () => {
      const errorMessage = 'Cannot delete active user';
      mockMerchantUserService.deleteMerchantUser.mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useMerchantUserStore());

      await act(async () => {
        try {
          await result.current.deleteMerchantUser(1);
        } catch (error) {
          // Expected to throw
        }
      });

      expect(result.current.error).toBe(errorMessage);
    });
  });

  describe('utility methods', () => {
    it('should clear error', () => {
      const { result } = renderHook(() => useMerchantUserStore());

      act(() => {
        result.current.setError('Test error');
      });

      expect(result.current.error).toBe('Test error');

      act(() => {
        result.current.clearError();
      });

      expect(result.current.error).toBeNull();
    });

    it('should reset store to initial state', () => {
      const { result } = renderHook(() => useMerchantUserStore());

      // Modify state
      act(() => {
        result.current.setMerchantUsers([mockUser]);
        result.current.setCurrentMerchantUser(mockUser);
        result.current.setLoading(true);
        result.current.setError('Test error');
      });

      // Reset
      act(() => {
        result.current.reset();
      });

      // Check initial state
      expect(result.current.merchantUsers).toEqual([]);
      expect(result.current.currentMerchantUser).toBeNull();
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBeNull();
    });
  });
});