import axios, { AxiosInstance, InternalAxiosRequestConfig } from 'axios';
import { AuthResponse, ApiResponse } from '@/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor to add auth token
    this.client.interceptors.request.use(
      (config) => {
        if (typeof window !== 'undefined') {
          const token = localStorage.getItem('token');
          if (token) {
            config.headers.Authorization = `Bearer ${token}`;
          }
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor to handle errors
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Clear token and redirect to login
          if (typeof window !== 'undefined') {
            localStorage.removeItem('token');
            localStorage.removeItem('refresh_token');
            localStorage.removeItem('user');
            localStorage.removeItem('tenant');
            window.location.href = '/auth/login';
          }
        }
        return Promise.reject(error);
      }
    );
  }

  // Auth endpoints
  async register(data: {
    email: string;
    password: string;
    first_name: string;
    last_name: string;
    tenant_name: string;
    tenant_slug: string;
  }): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>('/api/v1/auth/register', data);
    this.setAuthData(response.data);
    return response.data;
  }

  async login(data: { email: string; password: string }): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>('/api/v1/auth/login', data);
    this.setAuthData(response.data);
    return response.data;
  }

  async logout(): Promise<void> {
    try {
      await this.client.post('/api/v1/auth/logout');
    } finally {
      this.clearAuthData();
    }
  }

  async refreshToken(token: string): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>('/api/v1/auth/refresh', {
      refresh_token: token,
    });
    this.setAuthData(response.data);
    return response.data;
  }

  // User endpoints
  async getCurrentUser(): Promise<{ user: any; tenant: any }> {
    const response = await this.client.get('/api/v1/auth/me');
    return response.data;
  }

  async getUsers(page = 1, limit = 20): Promise<ApiResponse> {
    const response = await this.client.get<ApiResponse>('/api/v1/users', {
      params: { page, limit },
    });
    return response.data;
  }

  // Tenant endpoints
  async getTenant(): Promise<ApiResponse> {
    const response = await this.client.get<ApiResponse>('/api/v1/tenant');
    return response.data;
  }

  async getTenantUsage(): Promise<ApiResponse> {
    const response = await this.client.get<ApiResponse>('/api/v1/tenant/usage');
    return response.data;
  }

  // Helper methods
  private setAuthData(data: AuthResponse) {
    if (typeof window !== 'undefined') {
      localStorage.setItem('token', data.token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('user', JSON.stringify(data.user));
      localStorage.setItem('tenant', JSON.stringify(data.tenant));
    }
  }

  private clearAuthData() {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('user');
      localStorage.removeItem('tenant');
    }
  }

  isAuthenticated(): boolean {
    if (typeof window === 'undefined') return false;
    return !!localStorage.getItem('token');
  }

  getUser() {
    if (typeof window === 'undefined') return null;
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  }

  getTenantFromStorage() {
    if (typeof window === 'undefined') return null;
    const tenantStr = localStorage.getItem('tenant');
    return tenantStr ? JSON.parse(tenantStr) : null;
  }
}

export const apiClient = new ApiClient();
