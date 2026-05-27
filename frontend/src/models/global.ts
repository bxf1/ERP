import { create } from 'zustand';

interface UserInfo {
  id: number;
  username: string;
  email: string;
}

interface GlobalState {
  user: UserInfo | null;
  collapsed: boolean;
  setUser: (user: UserInfo | null) => void;
  setCollapsed: (collapsed: boolean) => void;
}

export const useGlobalStore = create<GlobalState>((set) => ({
  user: null,
  collapsed: false,
  setUser: (user) => set({ user }),
  setCollapsed: (collapsed) => set({ collapsed }),
}));
