"""Data models for the GOTRS API."""

from datetime import datetime
from typing import Any, Dict, List, Optional, Union

from pydantic import BaseModel, Field, ConfigDict


class BaseGotrsModel(BaseModel):
    """Base model for all GOTRS API models."""
    
    model_config = ConfigDict(
        use_enum_values=True,
        validate_assignment=True,
        extra="forbid",
        populate_by_name=True,
    )


class Ticket(BaseGotrsModel):
    """Represents a support ticket."""
    
    id: int
    ticket_number: str
    title: str
    description: str
    status: str
    priority: str
    type: str
    queue_id: int
    customer_id: int
    assigned_to: Optional[int] = None
    created_at: datetime
    updated_at: datetime
    closed_at: Optional[datetime] = None
    tags: Optional[List[str]] = None
    custom_fields: Optional[Dict[str, Any]] = None
    customer: Optional["User"] = None
    assigned_user: Optional["User"] = None
    queue: Optional["Queue"] = None
    messages: Optional[List["TicketMessage"]] = None
    attachments: Optional[List["Attachment"]] = None


class TicketMessage(BaseGotrsModel):
    """Represents a message in a ticket."""
    
    id: int
    ticket_id: int
    content: str
    message_type: str
    is_internal: bool
    author_id: int
    created_at: datetime
    updated_at: datetime
    author: Optional["User"] = None
    attachments: Optional[List["Attachment"]] = None
    custom_fields: Optional[Dict[str, Any]] = None


class User(BaseGotrsModel):
    """Represents a user in the system."""
    
    id: int
    email: str
    first_name: str
    last_name: str
    login: str
    title: str
    role: str
    is_active: bool
    created_at: datetime
    updated_at: datetime
    last_login_at: datetime


class Queue(BaseGotrsModel):
    """Represents a ticket queue."""
    
    id: int
    name: str
    description: str
    is_active: bool
    created_at: datetime
    updated_at: datetime


class Attachment(BaseGotrsModel):
    """Represents a file attachment."""
    
    id: int
    filename: str
    content_type: str
    size: int
    ticket_id: int
    message_id: Optional[int] = None
    uploaded_by: int
    created_at: datetime


class Group(BaseGotrsModel):
    """Represents a user group."""
    
    id: int
    name: str
    description: str
    type: str
    is_active: bool
    created_at: datetime
    updated_at: datetime


class DashboardStats(BaseGotrsModel):
    """Represents dashboard statistics."""
    
    total_tickets: int
    open_tickets: int
    closed_tickets: int
    pending_tickets: int
    overdue_tickets: int
    unassigned_tickets: int
    my_tickets: int
    tickets_by_status: Dict[str, int]
    tickets_by_priority: Dict[str, int]
    tickets_by_queue: Dict[str, int]


class SearchResult(BaseGotrsModel):
    """Represents search results."""
    
    total_count: int
    page: int
    page_size: int
    tickets: List[Ticket]


class InternalNote(BaseGotrsModel):
    """Represents an internal note."""
    
    id: int
    ticket_id: int
    content: str
    category: str
    is_important: bool
    is_pinned: bool
    tags: List[str]
    author_id: int
    author_name: str
    author_email: str
    created_at: datetime
    updated_at: datetime
    edited_at: datetime
    edited_by: int


class NoteTemplate(BaseGotrsModel):
    """Represents a note template."""
    
    id: int
    name: str
    content: str
    category: str
    tags: List[str]
    is_important: bool
    created_by: int
    created_at: datetime
    updated_at: datetime


class LDAPUser(BaseGotrsModel):
    """Represents a user from LDAP."""
    
    dn: str
    username: str
    email: str
    first_name: str
    last_name: str
    display_name: str
    phone: str
    department: str
    title: str
    manager: str
    groups: List[str]
    attributes: Dict[str, str]
    object_guid: str
    object_sid: str
    last_login: datetime
    is_active: bool


class LDAPSyncResult(BaseGotrsModel):
    """Represents the result of an LDAP sync operation."""
    
    users_found: int
    users_created: int
    users_updated: int
    users_disabled: int
    groups_found: int
    groups_created: int
    groups_updated: int
    errors: List[str]
    start_time: datetime
    end_time: datetime
    duration: str
    dry_run: bool


class Webhook(BaseGotrsModel):
    """Represents a webhook configuration."""
    
    id: int
    name: str
    url: str
    events: List[str]
    secret: Optional[str] = None
    is_active: bool
    retry_count: int
    timeout: int
    headers: Optional[Dict[str, str]] = None
    created_at: datetime
    updated_at: datetime
    last_fired_at: Optional[datetime] = None


class WebhookDelivery(BaseGotrsModel):
    """Represents a webhook delivery attempt."""
    
    id: int
    webhook_id: int
    event: str
    payload: str
    status_code: int
    response: str
    success: bool
    attempt: int
    delivered_at: datetime


# Request/Response models

class TicketCreateRequest(BaseGotrsModel):
    """Request model for creating a ticket."""
    
    title: str
    description: str
    priority: Optional[str] = "normal"
    type: Optional[str] = "incident"
    queue_id: Optional[int] = None
    customer_id: Optional[int] = None
    assigned_to: Optional[int] = None
    tags: Optional[List[str]] = None
    custom_fields: Optional[Dict[str, Any]] = None


class TicketUpdateRequest(BaseGotrsModel):
    """Request model for updating a ticket."""
    
    title: Optional[str] = None
    description: Optional[str] = None
    status: Optional[str] = None
    priority: Optional[str] = None
    type: Optional[str] = None
    queue_id: Optional[int] = None
    assigned_to: Optional[int] = None
    tags: Optional[List[str]] = None
    custom_fields: Optional[Dict[str, Any]] = None


class TicketListOptions(BaseGotrsModel):
    """Options for listing tickets."""
    
    page: Optional[int] = 1
    page_size: Optional[int] = 50
    status: Optional[List[str]] = None
    priority: Optional[List[str]] = None
    queue_id: Optional[List[int]] = None
    assigned_to: Optional[int] = None
    customer_id: Optional[int] = None
    search: Optional[str] = None
    tags: Optional[List[str]] = None
    created_after: Optional[datetime] = None
    created_before: Optional[datetime] = None
    sort_by: Optional[str] = "created_at"
    sort_order: Optional[str] = "desc"


class TicketListResponse(BaseGotrsModel):
    """Response model for listing tickets."""
    
    tickets: List[Ticket]
    total_count: int
    page: int
    page_size: int
    total_pages: int


class MessageCreateRequest(BaseGotrsModel):
    """Request model for creating a message."""
    
    content: str
    message_type: Optional[str] = "note"
    is_internal: Optional[bool] = False
    custom_fields: Optional[Dict[str, Any]] = None


class UserCreateRequest(BaseGotrsModel):
    """Request model for creating a user."""
    
    email: str
    first_name: str
    last_name: str
    login: str
    title: Optional[str] = ""
    role: Optional[str] = "user"
    password: str = Field(min_length=8)


class UserUpdateRequest(BaseGotrsModel):
    """Request model for updating a user."""
    
    email: Optional[str] = None
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    title: Optional[str] = None
    role: Optional[str] = None
    is_active: Optional[bool] = None


class AuthLoginRequest(BaseGotrsModel):
    """Request model for authentication."""
    
    email: str
    password: str


class AuthLoginResponse(BaseGotrsModel):
    """Response model for authentication."""
    
    token: str
    refresh_token: str
    expires_at: datetime
    user: User


class APIResponse(BaseGotrsModel):
    """Standard API response wrapper."""
    
    success: bool
    data: Optional[Any] = None
    error: Optional[str] = None
    message: Optional[str] = None


class ErrorResponse(BaseGotrsModel):
    """API error response."""
    
    error: str
    message: str
    code: int


# Update forward references
Ticket.model_rebuild()
TicketMessage.model_rebuild()
User.model_rebuild()
Queue.model_rebuild()
Attachment.model_rebuild()
Group.model_rebuild()
DashboardStats.model_rebuild()
SearchResult.model_rebuild()
InternalNote.model_rebuild()
NoteTemplate.model_rebuild()
LDAPUser.model_rebuild()
LDAPSyncResult.model_rebuild()
Webhook.model_rebuild()
WebhookDelivery.model_rebuild()