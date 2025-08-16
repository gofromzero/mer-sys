import { create } from 'zustand';
import type { AuthState, User } from '../types';

interface AuthActions {
  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  setPermissions: (permissions: string[]) => void;
  setLoading: (loading: boolean) => void;
  login: (user: User, token: string, permissions: string[]) => void;
  logout: () => void;
}

type AuthStore = AuthState & AuthActions;

export const useAuthStore = create<AuthStore>((set) => ({
  // State
  user: null,
  token: localStorage.getItem('token'),
  permissions: [],
  isLoading: false,

  // Actions
  setUser: (user) => set({ user }),
  setToken: (token) => {
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
    set({ token });
  },
  setPermissions: (permissions) => set({ permissions }),
  setLoading: (isLoading) => set({ isLoading }),
  
  login: (user, token, permissions) => {
    localStorage.setItem('token', token);
    localStorage.setItem('tenant_id', user.tenant_id.toString());
    set({ user, token, permissions });
  },
  
  logout: () => {
    localStorage.removeItem('token');
    localStorage.removeItem('tenant_id');
    set({ user: null, token: null, permissions: [] });
  },
}));