# GOTRS SDK

Software Development Kits (SDKs) for interacting with the GOTRS API. These SDKs provide type-safe client libraries for various programming languages.

## Available SDKs

### Go SDK
- **Path**: `./go/`
- **Package**: `github.com/gotrs-io/gotrs-ce/sdk/go`
- **Go Version**: 1.21+
- **Features**: Full type safety, context support, concurrent operations

### TypeScript/JavaScript SDK  
- **Path**: `./typescript/`
- **Package**: `@gotrs/sdk`
- **Node Version**: 18+
- **Features**: TypeScript definitions, Promise-based, browser/Node.js support

### Python SDK
- **Path**: `./python/`
- **Package**: `gotrs-sdk`
- **Python Version**: 3.8+
- **Features**: Async/await support, type hints, Pydantic models

### PHP SDK
- **Path**: `./php/`
- **Package**: `gotrs/sdk`
- **PHP Version**: 8.1+
- **Features**: PSR-4 autoloading, type declarations, Guzzle HTTP client

## Quick Start

### Go
```go
import "github.com/gotrs-io/gotrs-ce/sdk/go"

client := gotrs.NewClient("https://your-gotrs-instance.com", "your-api-key")
tickets, err := client.Tickets.List(ctx, &gotrs.TicketListOptions{})
```

### TypeScript
```typescript
import { GotrsClient } from '@gotrs/sdk';

const client = new GotrsClient('https://your-gotrs-instance.com', 'your-api-key');
const tickets = await client.tickets.list();
```

### Python
```python
from gotrs_sdk import GotrsClient

client = GotrsClient('https://your-gotrs-instance.com', 'your-api-key')
tickets = await client.tickets.list()
```

### PHP
```php
use Gotrs\SDK\Client;

$client = new Client('https://your-gotrs-instance.com', 'your-api-key');
$tickets = $client->tickets()->list();
```

## Authentication

All SDKs support multiple authentication methods:

1. **API Key** (recommended for server-to-server)
2. **JWT Token** (for user-based authentication)
3. **OAuth2** (for third-party integrations)

## API Coverage

All SDKs provide complete coverage of the GOTRS API:

- ✅ Authentication & Session Management
- ✅ Ticket Management (CRUD, search, attachments)
- ✅ User Management
- ✅ Queue Management
- ✅ Dashboard & Analytics
- ✅ LDAP Integration
- ✅ Webhook Management
- ✅ Real-time Events (WebSocket/SSE)

## Error Handling

All SDKs implement consistent error handling:

- **Network errors**: Connection timeouts, DNS failures
- **HTTP errors**: 4xx/5xx status codes with detailed messages
- **API errors**: GOTRS-specific error codes and descriptions
- **Validation errors**: Client-side validation before API calls

## Rate Limiting

SDKs automatically handle rate limiting:

- Exponential backoff with jitter
- Configurable retry policies
- Rate limit header parsing
- Queue management for bulk operations

## Testing

Each SDK includes:

- Unit tests with 90%+ coverage
- Integration tests against live API
- Mock server for offline testing
- Examples and documentation

## Contributing

See individual SDK directories for language-specific contribution guidelines.

## License

MIT License - see LICENSE file for details.