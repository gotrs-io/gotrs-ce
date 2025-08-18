import { HttpClient } from '../client';
import {
  User,
  UserCreateRequest,
  UserUpdateRequest,
  Queue,
  DashboardStats,
  LDAPUser,
  LDAPSyncResult,
  Webhook,
  WebhookDelivery,
  InternalNote,
  NoteTemplate,
  AuthLoginRequest,
  AuthLoginResponse,
} from '../types';

/**
 * Service for managing users
 */
export class UsersService {
  constructor(private client: HttpClient) {}

  async list(): Promise<User[]> {
    return this.client.get<User[]>('/api/v1/users');
  }

  async get(id: number): Promise<User> {
    return this.client.get<User>(`/api/v1/users/${id}`);
  }

  async create(data: UserCreateRequest): Promise<User> {
    return this.client.post<User>('/api/v1/users', data);
  }

  async update(id: number, data: UserUpdateRequest): Promise<User> {
    return this.client.put<User>(`/api/v1/users/${id}`, data);
  }

  async delete(id: number): Promise<void> {
    await this.client.delete(`/api/v1/users/${id}`);
  }
}

/**
 * Service for managing queues
 */
export class QueuesService {
  constructor(private client: HttpClient) {}

  async list(): Promise<Queue[]> {
    return this.client.get<Queue[]>('/api/v1/queues');
  }

  async get(id: number): Promise<Queue> {
    return this.client.get<Queue>(`/api/v1/queues/${id}`);
  }
}

/**
 * Service for dashboard operations
 */
export class DashboardService {
  constructor(private client: HttpClient) {}

  async getStats(): Promise<DashboardStats> {
    return this.client.get<DashboardStats>('/api/v1/dashboard/stats');
  }

  async getMyTickets(): Promise<any[]> {
    return this.client.get<any[]>('/api/v1/dashboard/my-tickets');
  }

  async getRecentTickets(): Promise<any[]> {
    return this.client.get<any[]>('/api/v1/dashboard/recent-tickets');
  }
}

/**
 * Service for LDAP operations
 */
export class LDAPService {
  constructor(private client: HttpClient) {}

  async getUsers(): Promise<LDAPUser[]> {
    return this.client.get<LDAPUser[]>('/api/v1/ldap/users');
  }

  async getUser(username: string): Promise<LDAPUser> {
    return this.client.get<LDAPUser>(`/api/v1/ldap/users/${username}`);
  }

  async syncUsers(): Promise<LDAPSyncResult> {
    return this.client.post<LDAPSyncResult>('/api/v1/ldap/sync');
  }

  async testConnection(): Promise<void> {
    await this.client.post('/api/v1/ldap/test');
  }

  async getSyncStatus(): Promise<Record<string, any>> {
    return this.client.get<Record<string, any>>('/api/v1/ldap/sync/status');
  }
}

/**
 * Service for webhook management
 */
export class WebhooksService {
  constructor(private client: HttpClient) {}

  async list(): Promise<Webhook[]> {
    return this.client.get<Webhook[]>('/api/v1/webhooks');
  }

  async get(id: number): Promise<Webhook> {
    return this.client.get<Webhook>(`/api/v1/webhooks/${id}`);
  }

  async create(data: Partial<Webhook>): Promise<Webhook> {
    return this.client.post<Webhook>('/api/v1/webhooks', data);
  }

  async update(id: number, data: Partial<Webhook>): Promise<Webhook> {
    return this.client.put<Webhook>(`/api/v1/webhooks/${id}`, data);
  }

  async delete(id: number): Promise<void> {
    await this.client.delete(`/api/v1/webhooks/${id}`);
  }

  async test(id: number): Promise<void> {
    await this.client.post(`/api/v1/webhooks/${id}/test`);
  }

  async getDeliveries(id: number): Promise<WebhookDelivery[]> {
    return this.client.get<WebhookDelivery[]>(`/api/v1/webhooks/${id}/deliveries`);
  }
}

/**
 * Service for internal notes
 */
export class NotesService {
  constructor(private client: HttpClient) {}

  async getNotes(ticketId: number): Promise<InternalNote[]> {
    return this.client.get<InternalNote[]>(`/api/v1/tickets/${ticketId}/notes`);
  }

  async getNote(ticketId: number, noteId: number): Promise<InternalNote> {
    return this.client.get<InternalNote>(`/api/v1/tickets/${ticketId}/notes/${noteId}`);
  }

  async createNote(ticketId: number, data: Partial<InternalNote>): Promise<InternalNote> {
    return this.client.post<InternalNote>(`/api/v1/tickets/${ticketId}/notes`, data);
  }

  async updateNote(ticketId: number, noteId: number, data: Partial<InternalNote>): Promise<InternalNote> {
    return this.client.put<InternalNote>(`/api/v1/tickets/${ticketId}/notes/${noteId}`, data);
  }

  async deleteNote(ticketId: number, noteId: number): Promise<void> {
    await this.client.delete(`/api/v1/tickets/${ticketId}/notes/${noteId}`);
  }

  async getTemplates(): Promise<NoteTemplate[]> {
    return this.client.get<NoteTemplate[]>('/api/v1/notes/templates');
  }

  async createTemplate(data: Partial<NoteTemplate>): Promise<NoteTemplate> {
    return this.client.post<NoteTemplate>('/api/v1/notes/templates', data);
  }
}

/**
 * Service for authentication
 */
export class AuthService {
  constructor(private client: HttpClient) {}

  async login(data: AuthLoginRequest): Promise<AuthLoginResponse> {
    return this.client.post<AuthLoginResponse>('/api/v1/auth/login', data);
  }

  async logout(): Promise<void> {
    await this.client.post('/api/v1/auth/logout');
  }

  async refreshToken(refreshToken: string): Promise<AuthLoginResponse> {
    return this.client.post<AuthLoginResponse>('/api/v1/auth/refresh', { refresh_token: refreshToken });
  }

  async getProfile(): Promise<User> {
    return this.client.get<User>('/api/v1/auth/profile');
  }

  async updateProfile(data: UserUpdateRequest): Promise<User> {
    return this.client.put<User>('/api/v1/auth/profile', data);
  }
}

// Re-export TicketsService
export { TicketsService } from './tickets';