import { create } from 'zustand';
import type { UIState } from '../types';

interface UIActions {
  setSidebarOpen: (open: boolean) => void;
  toggleSidebar: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
  setLoading: (loading: boolean) => void;
}

type UIStore = UIState & UIActions;

export const useUIStore = create<UIStore>((set, get) => ({
  // State
  sidebarOpen: true,
  theme: 'light',
  loading: false,

  // Actions
  setSidebarOpen: (sidebarOpen) => set({ sidebarOpen }),
  toggleSidebar: () => set({ sidebarOpen: !get().sidebarOpen }),
  setTheme: (theme) => set({ theme }),
  setLoading: (loading) => set({ loading }),
}));