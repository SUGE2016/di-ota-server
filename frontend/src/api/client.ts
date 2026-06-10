import axios from 'axios';

const apiBase = `${import.meta.env.BASE_URL}api/v1`.replace(/\/{2,}/g, '/');

const api = axios.create({
  baseURL: apiBase,
  timeout: 15000,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('jwt_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('jwt_token');
      window.location.hash = '#/login';
    }
    return Promise.reject(err);
  }
);

export default api;
