"""
GOTRS Python SDK

Official Python SDK for the GOTRS ticketing system API.

Basic usage:
    >>> from gotrs_sdk import GotrsClient
    >>> client = GotrsClient.with_api_key("https://your-gotrs.com", "your-api-key")
    >>> tickets = await client.tickets.list()
    >>> print(f"Found {tickets.total_count} tickets")

Authentication:
    # API Key
    client = GotrsClient.with_api_key(base_url, api_key)
    
    # JWT Token
    client = GotrsClient.with_jwt(base_url, token, refresh_token, expires_at)
    
    # OAuth2
    client = GotrsClient.with_oauth2(base_url, access_token, refresh_token, expires_at)
    
    # Login flow
    client = GotrsClient(base_url)
    await client.login("user@example.com", "password")
"""

from .client import GotrsClient
from .exceptions import (
    GotrsError,
    ValidationError,
    NetworkError,
    TimeoutError,
    NotFoundError,
    UnauthorizedError,
    ForbiddenError,
    RateLimitError,
)
from .models import (
    Ticket,
    TicketMessage,
    User,
    Queue,
    Attachment,
    Group,
    DashboardStats,
    SearchResult,
    InternalNote,
    NoteTemplate,
    LDAPUser,
    LDAPSyncResult,
    Webhook,
    WebhookDelivery,
    TicketCreateRequest,
    TicketUpdateRequest,
    TicketListOptions,
    MessageCreateRequest,
    UserCreateRequest,
    UserUpdateRequest,
    AuthLoginRequest,
    AuthLoginResponse,
)
from .auth import APIKeyAuth, JWTAuth, OAuth2Auth

__version__ = "1.0.0"
__author__ = "GOTRS Team"
__email__ = "dev@gotrs.io"
__license__ = "MIT"

__all__ = [
    # Main client
    "GotrsClient",
    # Exceptions
    "GotrsError",
    "ValidationError",
    "NetworkError",
    "TimeoutError",
    "NotFoundError",
    "UnauthorizedError",
    "ForbiddenError", 
    "RateLimitError",
    # Models
    "Ticket",
    "TicketMessage",
    "User",
    "Queue",
    "Attachment",
    "Group",
    "DashboardStats",
    "SearchResult",
    "InternalNote",
    "NoteTemplate",
    "LDAPUser",
    "LDAPSyncResult",
    "Webhook",
    "WebhookDelivery",
    "TicketCreateRequest",
    "TicketUpdateRequest",
    "TicketListOptions",
    "MessageCreateRequest",
    "UserCreateRequest",
    "UserUpdateRequest",
    "AuthLoginRequest",
    "AuthLoginResponse",
    # Auth
    "APIKeyAuth",
    "JWTAuth",
    "OAuth2Auth",
]

# Convenience functions for error checking
def is_gotrs_error(error: Exception) -> bool:
    """Check if an exception is a GOTRS API error."""
    return isinstance(error, GotrsError)

def is_not_found_error(error: Exception) -> bool:
    """Check if an exception is a 404 Not Found error."""
    return isinstance(error, NotFoundError)

def is_unauthorized_error(error: Exception) -> bool:
    """Check if an exception is a 401 Unauthorized error."""
    return isinstance(error, UnauthorizedError)

def is_forbidden_error(error: Exception) -> bool:
    """Check if an exception is a 403 Forbidden error."""
    return isinstance(error, ForbiddenError)

def is_rate_limit_error(error: Exception) -> bool:
    """Check if an exception is a 429 Rate Limit error."""
    return isinstance(error, RateLimitError)

def is_validation_error(error: Exception) -> bool:
    """Check if an exception is a validation error."""
    return isinstance(error, ValidationError)

def is_network_error(error: Exception) -> bool:
    """Check if an exception is a network error."""
    return isinstance(error, NetworkError)

def is_timeout_error(error: Exception) -> bool:
    """Check if an exception is a timeout error."""
    return isinstance(error, TimeoutError)