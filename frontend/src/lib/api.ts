import { User, Todo, AuditRecord, TodoStats, TodoStatus, TodoPriority } from './types';

const API_BASE = typeof window !== 'undefined' 
  ? (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1') 
  : (process.env.INTERNAL_API_URL ? `${process.env.INTERNAL_API_URL}/api/v1` : 'http://localhost:8080/api/v1');

function getAuthHeader(): Record<string, string> {
  if (typeof window === 'undefined') return {};
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function handleResponse<T>(res: Response): Promise<T> {
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const errorMsg = data?.error?.message || data?.message || `HTTP error ${res.status}`;
    throw new Error(errorMsg);
  }
  return data as T;
}

export const api = {
  // Auth
  async login(email: string, password: string): Promise<{ token: string; user: User }> {
    const res = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    return handleResponse(res);
  },

  async register(data: { email: string; password: string; full_name: string; role?: string }): Promise<{ token: string; user: User }> {
    const res = await fetch(`${API_BASE}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async getMe(): Promise<{ user: User }> {
    const res = await fetch(`${API_BASE}/auth/me`, {
      headers: { ...getAuthHeader() },
    });
    return handleResponse(res);
  },

  // Todos
  async listTodos(params?: { status?: string; priority?: string; search?: string; page?: number }): Promise<{ data: Todo[]; total: number }> {
    const q = new URLSearchParams();
    if (params?.status) q.append('status', params.status);
    if (params?.priority) q.append('priority', params.priority);
    if (params?.search) q.append('search', params.search);
    if (params?.page) q.append('page', params.page.toString());

    const res = await fetch(`${API_BASE}/todos?${q.toString()}`, {
      headers: { ...getAuthHeader() },
    });
    return handleResponse(res);
  },

  async getTodoStats(): Promise<{ stats: TodoStats }> {
    const res = await fetch(`${API_BASE}/todos/stats`, {
      headers: { ...getAuthHeader() },
    });
    return handleResponse(res);
  },

  async createTodo(data: { title: string; description: string; priority: TodoPriority; due_date?: string; tags: string }): Promise<Todo> {
    const res = await fetch(`${API_BASE}/todos`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...getAuthHeader() },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async updateTodo(id: string, data: { title: string; description: string; priority: TodoPriority; due_date?: string; tags: string; version: number }): Promise<Todo> {
    const res = await fetch(`${API_BASE}/todos/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', ...getAuthHeader() },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async changeStatus(id: string, status: TodoStatus, version: number): Promise<Todo> {
    const res = await fetch(`${API_BASE}/todos/${id}/status`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json', ...getAuthHeader() },
      body: JSON.stringify({ status, version }),
    });
    return handleResponse(res);
  },

  async deleteTodo(id: string): Promise<{ message: string }> {
    const res = await fetch(`${API_BASE}/todos/${id}`, {
      method: 'DELETE',
      headers: { ...getAuthHeader() },
    });
    return handleResponse(res);
  },

  // Audit
  async listAuditLogs(page = 1): Promise<{ data: AuditRecord[]; total: number }> {
    const res = await fetch(`${API_BASE}/audit?page=${page}`, {
      headers: { ...getAuthHeader() },
    });
    return handleResponse(res);
  },

  async verifyAuditIntegrity(): Promise<{ tamper_evident_integrity_valid: boolean; tenant_id: string }> {
    const res = await fetch(`${API_BASE}/audit/verify`, {
      headers: { ...getAuthHeader() },
    });
    return handleResponse(res);
  },
};
