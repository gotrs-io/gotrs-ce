import { HttpClient } from '../client';
import {
  Ticket,
  TicketCreateRequest,
  TicketUpdateRequest,
  TicketListOptions,
  TicketListResponse,
  TicketMessage,
  MessageCreateRequest,
  Attachment,
  SearchResult,
} from '../types';

/**
 * Service for managing tickets
 */
export class TicketsService {
  constructor(private client: HttpClient) {}

  /**
   * List tickets with optional filtering
   */
  async list(options?: TicketListOptions): Promise<TicketListResponse> {
    const params = new URLSearchParams();
    
    if (options) {
      if (options.page) params.set('page', options.page.toString());
      if (options.page_size) params.set('page_size', options.page_size.toString());
      if (options.status) params.set('status', options.status.join(','));
      if (options.priority) params.set('priority', options.priority.join(','));
      if (options.queue_id) params.set('queue_id', options.queue_id.map(String).join(','));
      if (options.assigned_to) params.set('assigned_to', options.assigned_to.toString());
      if (options.customer_id) params.set('customer_id', options.customer_id.toString());
      if (options.search) params.set('search', options.search);
      if (options.tags) params.set('tags', options.tags.join(','));
      if (options.created_after) params.set('created_after', options.created_after);
      if (options.created_before) params.set('created_before', options.created_before);
      if (options.sort_by) params.set('sort_by', options.sort_by);
      if (options.sort_order) params.set('sort_order', options.sort_order);
    }

    const url = `/api/v1/tickets${params.toString() ? '?' + params.toString() : ''}`;
    return this.client.get<TicketListResponse>(url);
  }

  /**
   * Get a specific ticket by ID
   */
  async get(id: number): Promise<Ticket> {
    return this.client.get<Ticket>(`/api/v1/tickets/${id}`);
  }

  /**
   * Get a specific ticket by ticket number
   */
  async getByNumber(ticketNumber: string): Promise<Ticket> {
    return this.client.get<Ticket>(`/api/v1/tickets/number/${ticketNumber}`);
  }

  /**
   * Create a new ticket
   */
  async create(data: TicketCreateRequest): Promise<Ticket> {
    return this.client.post<Ticket>('/api/v1/tickets', data);
  }

  /**
   * Update an existing ticket
   */
  async update(id: number, data: TicketUpdateRequest): Promise<Ticket> {
    return this.client.put<Ticket>(`/api/v1/tickets/${id}`, data);
  }

  /**
   * Delete a ticket
   */
  async delete(id: number): Promise<void> {
    await this.client.delete(`/api/v1/tickets/${id}`);
  }

  /**
   * Close a ticket
   */
  async close(id: number, reason?: string): Promise<Ticket> {
    return this.client.post<Ticket>(`/api/v1/tickets/${id}/close`, { reason });
  }

  /**
   * Reopen a closed ticket
   */
  async reopen(id: number, reason?: string): Promise<Ticket> {
    return this.client.post<Ticket>(`/api/v1/tickets/${id}/reopen`, { reason });
  }

  /**
   * Assign a ticket to a user
   */
  async assign(id: number, userId: number): Promise<Ticket> {
    return this.client.post<Ticket>(`/api/v1/tickets/${id}/assign`, { user_id: userId });
  }

  /**
   * Remove assignment from a ticket
   */
  async unassign(id: number): Promise<Ticket> {
    return this.client.post<Ticket>(`/api/v1/tickets/${id}/unassign`);
  }

  /**
   * Add a message to a ticket
   */
  async addMessage(id: number, data: MessageCreateRequest): Promise<TicketMessage> {
    return this.client.post<TicketMessage>(`/api/v1/tickets/${id}/messages`, data);
  }

  /**
   * Get all messages for a ticket
   */
  async getMessages(id: number): Promise<TicketMessage[]> {
    return this.client.get<TicketMessage[]>(`/api/v1/tickets/${id}/messages`);
  }

  /**
   * Get a specific message
   */
  async getMessage(ticketId: number, messageId: number): Promise<TicketMessage> {
    return this.client.get<TicketMessage>(`/api/v1/tickets/${ticketId}/messages/${messageId}`);
  }

  /**
   * Update a ticket message
   */
  async updateMessage(ticketId: number, messageId: number, content: string): Promise<TicketMessage> {
    return this.client.put<TicketMessage>(`/api/v1/tickets/${ticketId}/messages/${messageId}`, {
      content,
    });
  }

  /**
   * Delete a ticket message
   */
  async deleteMessage(ticketId: number, messageId: number): Promise<void> {
    await this.client.delete(`/api/v1/tickets/${ticketId}/messages/${messageId}`);
  }

  /**
   * Get all attachments for a ticket
   */
  async getAttachments(id: number): Promise<Attachment[]> {
    return this.client.get<Attachment[]>(`/api/v1/tickets/${id}/attachments`);
  }

  /**
   * Get a specific attachment
   */
  async getAttachment(ticketId: number, attachmentId: number): Promise<Attachment> {
    return this.client.get<Attachment>(`/api/v1/tickets/${ticketId}/attachments/${attachmentId}`);
  }

  /**
   * Download an attachment
   */
  async downloadAttachment(ticketId: number, attachmentId: number): Promise<ArrayBuffer> {
    const response = await this.client.getAxios().get(
      `/api/v1/tickets/${ticketId}/attachments/${attachmentId}/download`,
      { responseType: 'arraybuffer' }
    );
    return response.data;
  }

  /**
   * Upload an attachment to a ticket
   */
  async uploadAttachment(ticketId: number, file: File | Buffer, filename?: string): Promise<Attachment> {
    const formData = new FormData();
    
    if (file instanceof File) {
      formData.append('file', file);
    } else {
      const blob = new Blob([file]);
      formData.append('file', blob, filename || 'attachment');
    }

    return this.client.post<Attachment>(`/api/v1/tickets/${ticketId}/attachments`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  }

  /**
   * Delete an attachment
   */
  async deleteAttachment(ticketId: number, attachmentId: number): Promise<void> {
    await this.client.delete(`/api/v1/tickets/${ticketId}/attachments/${attachmentId}`);
  }

  /**
   * Search tickets with advanced options
   */
  async search(query: string, options?: TicketListOptions): Promise<SearchResult> {
    const params = new URLSearchParams();
    params.set('q', query);

    if (options) {
      if (options.page) params.set('page', options.page.toString());
      if (options.page_size) params.set('page_size', options.page_size.toString());
      if (options.status) params.set('status', options.status.join(','));
      if (options.priority) params.set('priority', options.priority.join(','));
      if (options.queue_id) params.set('queue_id', options.queue_id.map(String).join(','));
    }

    return this.client.get<SearchResult>(`/api/v1/tickets/search?${params.toString()}`);
  }

  /**
   * Get ticket history
   */
  async getHistory(id: number): Promise<any[]> {
    return this.client.get<any[]>(`/api/v1/tickets/${id}/history`);
  }

  /**
   * Add tags to a ticket
   */
  async addTags(id: number, tags: string[]): Promise<Ticket> {
    return this.client.post<Ticket>(`/api/v1/tickets/${id}/tags`, { tags });
  }

  /**
   * Remove tags from a ticket
   */
  async removeTags(id: number, tags: string[]): Promise<void> {
    await this.client.delete(`/api/v1/tickets/${id}/tags`, { data: { tags } });
  }

  /**
   * Set ticket priority
   */
  async setPriority(id: number, priority: string): Promise<Ticket> {
    return this.update(id, { priority });
  }

  /**
   * Set ticket status
   */
  async setStatus(id: number, status: string): Promise<Ticket> {
    return this.update(id, { status });
  }

  /**
   * Move ticket to different queue
   */
  async moveToQueue(id: number, queueId: number): Promise<Ticket> {
    return this.update(id, { queue_id: queueId });
  }
}