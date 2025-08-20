import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { MerchantUserService } from '../merchantUserService';
import { api } from '../api';
import type { 
  CreateMerchantUserRequest, 
  UpdateMerchantUserRequest,
  MerchantUserStatusRequest 
} from '../../types/merchantUser';

// Mock the api module
jest.mock('../api', () => ({
  api: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn()
  }
}));

const mockApi = api as jest.Mocked<typeof api>;

describe('MerchantUserService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe('getMerchantUsers', () => {
    it('should fetch merchant users successfully', async () => {
      const mockResponse = {
        data: {
          code: 200,
          message: 'success',
          data: {
            list: [
              {
                id: 1,
                username: 'test_user',
                email: 'test@example.com',
                merchant_id: 123,
                role_type: 'merchant_operator'
              }
            ],
            pagination: {
              page: 1,
              page_size: 20,
              total: 1,
              total_pages: 1
            }
          }
        }
      };

      mockApi.get.mockResolvedValue(mockResponse);

      const result = await MerchantUserService.getMerchantUsers({
        page: 1,
        page_size: 20
      });

      expect(mockApi.get).toHaveBeenCalledWith('/api/v1/merchant-users', {
        params: { page: 1, page_size: 20 }
      });
      expect(result).toEqual(mockResponse.data.data);
    });

    it('should handle query parameters correctly', async () => {
      const mockResponse = {
        data: {
          code: 200,
          message: 'success',
          data: { list: [], pagination: { page: 1, page_size: 20, total: 0, total_pages: 0 } }
        }
      };

      mockApi.get.mockResolvedValue(mockResponse);

      await MerchantUserService.getMerchantUsers({
        merchant_id: 123,
        status: 'active',
        search: 'test'
      });

      expect(mockApi.get).toHaveBeenCalledWith('/api/v1/merchant-users', {
        params: {
          merchant_id: 123,
          status: 'active',
          search: 'test'
        }
      });
    });
  });

  describe('getMerchantUser', () => {
    it('should fetch single merchant user successfully', async () => {
      const mockUser = {
        id: 1,
        username: 'test_user',
        email: 'test@example.com',
        merchant_id: 123,
        role_type: 'merchant_admin'
      };

      const mockResponse = {
        data: {
          code: 200,
          message: 'success',
          data: mockUser
        }
      };

      mockApi.get.mockResolvedValue(mockResponse);

      const result = await MerchantUserService.getMerchantUser(1);

      expect(mockApi.get).toHaveBeenCalledWith('/api/v1/merchant-users/1');
      expect(result).toEqual(mockUser);
    });

    it('should handle user not found error', async () => {
      mockApi.get.mockRejectedValue(new Error('User not found'));

      await expect(MerchantUserService.getMerchantUser(999))
        .rejects.toThrow('User not found');
    });
  });

  describe('createMerchantUser', () => {
    it('should create merchant user successfully', async () => {
      const createRequest: CreateMerchantUserRequest = {
        username: 'new_user',
        email: 'new@example.com',
        merchant_id: 123,
        role_type: 'merchant_operator',
        permissions: ['merchant:product:view']
      };

      const mockCreatedUser = {
        id: 2,
        ...createRequest,
        status: 'pending',
        created_at: '2025-08-20T10:00:00Z'
      };

      const mockResponse = {
        data: {
          code: 201,
          message: 'User created successfully',
          data: mockCreatedUser
        }
      };

      mockApi.post.mockResolvedValue(mockResponse);

      const result = await MerchantUserService.createMerchantUser(createRequest);

      expect(mockApi.post).toHaveBeenCalledWith('/api/v1/merchant-users', createRequest);
      expect(result).toEqual(mockCreatedUser);
    });

    it('should handle validation errors', async () => {
      const invalidRequest = {
        username: '',
        email: 'invalid-email',
        merchant_id: 123,
        role_type: 'invalid_role' as any,
        permissions: []
      };

      mockApi.post.mockRejectedValue(new Error('Validation failed'));

      await expect(MerchantUserService.createMerchantUser(invalidRequest))
        .rejects.toThrow('Validation failed');
    });
  });

  describe('updateMerchantUser', () => {
    it('should update merchant user successfully', async () => {
      const updateRequest: UpdateMerchantUserRequest = {
        username: 'updated_user',
        email: 'updated@example.com',
        role_type: 'merchant_admin'
      };

      const mockUpdatedUser = {
        id: 1,
        ...updateRequest,
        merchant_id: 123,
        status: 'active',
        updated_at: '2025-08-20T10:00:00Z'
      };

      const mockResponse = {
        data: {
          code: 200,
          message: 'User updated successfully',
          data: mockUpdatedUser
        }
      };

      mockApi.put.mockResolvedValue(mockResponse);

      const result = await MerchantUserService.updateMerchantUser(1, updateRequest);

      expect(mockApi.put).toHaveBeenCalledWith('/api/v1/merchant-users/1', updateRequest);
      expect(result).toEqual(mockUpdatedUser);
    });
  });

  describe('updateMerchantUserStatus', () => {
    it('should update user status successfully', async () => {
      const statusRequest: MerchantUserStatusRequest = {
        status: 'suspended',
        comment: 'Violation of terms'
      };

      const mockUpdatedUser = {
        id: 1,
        username: 'test_user',
        status: 'suspended',
        updated_at: '2025-08-20T10:00:00Z'
      };

      const mockResponse = {
        data: {
          code: 200,
          message: 'Status updated successfully',
          data: mockUpdatedUser
        }
      };

      mockApi.put.mockResolvedValue(mockResponse);

      const result = await MerchantUserService.updateMerchantUserStatus(1, statusRequest);

      expect(mockApi.put).toHaveBeenCalledWith('/api/v1/merchant-users/1/status', statusRequest);
      expect(result).toEqual(mockUpdatedUser);
    });
  });

  describe('resetMerchantUserPassword', () => {
    it('should reset password successfully', async () => {
      mockApi.post.mockResolvedValue({ data: { code: 200, message: 'Password reset successfully' } });

      await MerchantUserService.resetMerchantUserPassword(1, { send_email: true });

      expect(mockApi.post).toHaveBeenCalledWith('/api/v1/merchant-users/1/reset-password', {
        send_email: true
      });
    });
  });

  describe('createMerchantUsersBatch', () => {
    it('should create multiple users successfully', async () => {
      const batchRequest: CreateMerchantUserRequest[] = [
        {
          username: 'user1',
          email: 'user1@example.com',
          merchant_id: 123,
          role_type: 'merchant_operator',
          permissions: ['merchant:product:view']
        },
        {
          username: 'user2',
          email: 'user2@example.com',
          merchant_id: 123,
          role_type: 'merchant_operator',
          permissions: ['merchant:product:view']
        }
      ];

      const mockCreatedUsers = batchRequest.map((user, index) => ({
        id: index + 1,
        ...user,
        status: 'pending' as const,
        created_at: '2025-08-20T10:00:00Z'
      }));

      const mockResponse = {
        data: {
          code: 201,
          message: 'Users created successfully',
          data: mockCreatedUsers
        }
      };

      mockApi.post.mockResolvedValue(mockResponse);

      const result = await MerchantUserService.createMerchantUsersBatch(batchRequest);

      expect(mockApi.post).toHaveBeenCalledWith('/api/v1/merchant-users/batch', {
        users: batchRequest
      });
      expect(result).toEqual(mockCreatedUsers);
    });

    it('should handle partial batch creation failure', async () => {
      const batchRequest: CreateMerchantUserRequest[] = [
        {
          username: 'user1',
          email: 'user1@example.com',
          merchant_id: 123,
          role_type: 'merchant_operator',
          permissions: ['merchant:product:view']
        }
      ];

      mockApi.post.mockRejectedValue(new Error('Duplicate username'));

      await expect(MerchantUserService.createMerchantUsersBatch(batchRequest))
        .rejects.toThrow('Duplicate username');
    });
  });

  describe('getMerchantUserAuditLog', () => {
    it('should fetch audit logs successfully', async () => {
      const mockAuditLogs = [
        {
          id: 1,
          user_id: 1,
          action: 'login',
          timestamp: '2025-08-20T10:00:00Z',
          ip_address: '192.168.1.1'
        }
      ];

      const mockResponse = {
        data: {
          code: 200,
          message: 'success',
          data: {
            list: mockAuditLogs,
            pagination: { page: 1, page_size: 20, total: 1, total_pages: 1 }
          }
        }
      };

      mockApi.get.mockResolvedValue(mockResponse);

      const result = await MerchantUserService.getMerchantUserAuditLog(1, {
        page: 1,
        page_size: 20
      });

      expect(mockApi.get).toHaveBeenCalledWith('/api/v1/merchant-users/1/audit-log', {
        params: { page: 1, page_size: 20 }
      });
      expect(result.list).toEqual(mockAuditLogs);
    });
  });

  describe('deleteMerchantUser', () => {
    it('should delete user successfully', async () => {
      mockApi.delete.mockResolvedValue({ data: { code: 200, message: 'User deleted successfully' } });

      await MerchantUserService.deleteMerchantUser(1);

      expect(mockApi.delete).toHaveBeenCalledWith('/api/v1/merchant-users/1');
    });

    it('should handle delete error', async () => {
      mockApi.delete.mockRejectedValue(new Error('Cannot delete active user'));

      await expect(MerchantUserService.deleteMerchantUser(1))
        .rejects.toThrow('Cannot delete active user');
    });
  });
});