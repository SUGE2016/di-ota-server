import { create } from 'zustand';
import { userAPI } from '../api';
import { parseJwtSub } from '../utils/jwt';

interface AuthState {
  token: string | null;
  isLoggedIn: boolean;
  username: string;
  roles: string[];
  authSource: 'local' | 'sso';
  hasExternalAccess: boolean;
  setAuth: (token: string) => void;
  logout: () => void;
  init: () => void;
}

async function loadUserProfile(username: string) {
  try {
    const data = await userAPI.list({ limit: 20, offset: 0, search: username });
    const user = data.users.find((item) => item.username.toLowerCase() === username.toLowerCase());
    if (user) {
      useAuthStore.setState({
        roles: user.roles,
        authSource: user.auth_source,
      });
      return;
    }
  } catch {
    /* 目录未同步时仍展示 JWT 用户名 */
  }
  useAuthStore.setState({ roles: ['admin'], authSource: 'local' });
}

function applyToken(token: string) {
  localStorage.setItem('jwt_token', token);
  const username = parseJwtSub(token) ?? 'unknown';
  useAuthStore.setState({
    token,
    isLoggedIn: true,
    username,
    roles: [],
    authSource: 'local',
    hasExternalAccess: false,
  });
  void loadUserProfile(username);
}

const useAuthStore = create<AuthState>((set) => ({
  token: null,
  isLoggedIn: false,
  username: '',
  roles: [],
  authSource: 'local',
  hasExternalAccess: false,
  setAuth: (token: string) => {
    applyToken(token);
  },
  logout: () => {
    localStorage.removeItem('jwt_token');
    set({
      token: null,
      isLoggedIn: false,
      username: '',
      roles: [],
      authSource: 'local',
      hasExternalAccess: false,
    });
  },
  init: () => {
    const token = localStorage.getItem('jwt_token');
    if (token) {
      applyToken(token);
    }
  },
}));

export default useAuthStore;
