export type Role = 'ADMIN' | 'MEMBER' | 'AUDITOR';

export interface User {
  id: string;
  tenant_id: string;
  email: string;
  full_name: string;
  role: Role;
}

export type TodoStatus = 'PENDING' | 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED';
export type TodoPriority = 'LOW' | 'MEDIUM' | 'HIGH' | 'URGENT';

export interface Todo {
  id: string;
  tenant_id: string;
  user_id: string;
  title: string;
  description: string;
  status: TodoStatus;
  priority: TodoPriority;
  due_date?: string;
  tags: string;
  is_overdue: boolean;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface AuditRecord {
  id: string;
  tenant_id: string;
  actor_id: string;
  actor_email: string;
  action: string;
  entity_type: string;
  entity_id: string;
  payload: string;
  previous_hash: string;
  record_hash: string;
  timestamp: string;
}

export interface TodoStats {
  TOTAL: number;
  PENDING: number;
  IN_PROGRESS: number;
  COMPLETED: number;
  CANCELLED: number;
}
