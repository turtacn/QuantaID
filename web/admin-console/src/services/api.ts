import axios from 'axios';
import { User } from '../types/user';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 10000,
});

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export const userAPI = {
  list: (params: { page: number; size: number; search?: string }) =>
    apiClient.get('/admin/users', { params }),
  create: (data: { username: string; email: string; role_ids: string[] }) =>
    apiClient.post('/admin/users', data),
  update: (id: string, data: Partial<User>) =>
    apiClient.patch(`/admin/users/${id}`, data),
  delete: (id: string) => apiClient.delete(`/admin/users/${id}`),
};
