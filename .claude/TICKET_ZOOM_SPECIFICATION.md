# Ticket Zoom Functional Specification

## Overview
The Ticket Zoom interface provides detailed ticket view and action capabilities for agents, enabling comprehensive ticket management including replies, notes, status changes, and communication history display.

## User Stories

### Primary User Stories
- **As an agent**, I need to view complete ticket details so that I understand the full context
- **As an agent**, I need to reply to tickets so that I can communicate with customers  
- **As an agent**, I need to add internal notes so that I can share information with other agents
- **As an agent**, I need to record phone conversations so that there's a complete communication history
- **As an agent**, I need to change ticket status so that I can manage workflow progression
- **As an agent**, I need to assign tickets so that I can distribute workload
- **As an agent**, I need to view communication history so that I understand previous interactions

### Secondary User Stories  
- **As an agent**, I need to attach files to responses so that I can share documents
- **As an agent**, I need to use predefined responses so that I can work efficiently
- **As an agent**, I need to escalate tickets so that complex issues get proper attention
- **As an agent**, I need to merge tickets so that I can consolidate related issues
- **As an agent**, I need to split tickets so that I can handle multiple distinct issues

## Technical Requirements

### Database Schema Requirements

#### Core Tables (OTRS-compatible)
```sql
-- Primary ticket storage
ticket (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    ticket_number VARCHAR(50) UNIQUE NOT NULL,
    queue_id INTEGER REFERENCES queue(id),
    state_id INTEGER REFERENCES ticket_state(id),
    priority_id INTEGER REFERENCES ticket_priority(id),
    customer_user_id VARCHAR(200),
    customer_id VARCHAR(150),
    user_id INTEGER REFERENCES users(id),        -- Ticket owner
    responsible_user_id INTEGER REFERENCES users(id),  -- Assigned agent
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Communication articles
article (
    id SERIAL PRIMARY KEY,
    ticket_id INTEGER REFERENCES ticket(id) NOT NULL,
    article_sender_type_id INTEGER NOT NULL,     -- 1=agent, 2=customer, 3=system
    communication_channel_id INTEGER NOT NULL,   -- 1=email, 2=phone, 3=web, 4=note
    subject VARCHAR(3800),
    body TEXT,
    is_visible_for_customer INTEGER DEFAULT 1,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_by INTEGER REFERENCES users(id)
);

-- Article content storage
article_data_mime (
    id SERIAL PRIMARY KEY,
    article_id INTEGER REFERENCES article(id) NOT NULL,
    filename VARCHAR(250),
    content_type VARCHAR(450),
    content_size INTEGER,
    content TEXT,
    content_id VARCHAR(250)
);

-- Article attachments
article_data_mime_attachment (
    id SERIAL PRIMARY KEY,
    article_id INTEGER REFERENCES article(id) NOT NULL,
    filename VARCHAR(250) NOT NULL,
    content_type VARCHAR(450),
    content_size INTEGER,
    content TEXT,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### API Endpoints

#### Core Endpoints
```
GET    /agent/tickets/:id              - Ticket zoom view
POST   /agent/tickets/:id/reply        - Customer reply
POST   /agent/tickets/:id/note         - Internal note  
POST   /agent/tickets/:id/phone        - Phone call record
PUT    /agent/tickets/:id/status       - Change status
PUT    /agent/tickets/:id/assign       - Assign agent
PUT    /agent/tickets/:id/priority     - Change priority
GET    /agent/tickets/:id/history      - Change history
POST   /agent/tickets/:id/attachment   - Upload attachment
```

#### Request/Response Formats

**Reply/Note/Phone Requests:**
```json
{
    "subject": "Response subject",
    "body": "Message content",
    "is_visible_for_customer": 1,
    "communication_channel_id": 1,
    "attachments": [
        {
            "filename": "document.pdf",
            "content_type": "application/pdf",
            "content": "base64_encoded_content"
        }
    ]
}
```

**Status Change Requests:**
```json
{
    "state_id": 2,
    "note": "Optional status change note"
}
```

**Assignment Requests:**
```json
{
    "responsible_user_id": 5,
    "note": "Assignment reason"
}
```

### UI Components

#### Ticket Header Section
**Required Elements:**
- Ticket number display (prominent)
- Title/subject display and editing
- Customer information (name, email, company)
- Current status, priority, queue, assignment
- Creation and last update timestamps
- Action buttons toolbar

**Interactive Elements:**
- Status dropdown with instant update
- Priority dropdown with instant update  
- Assignment dropdown (filtered by queue permissions)
- Quick action buttons (Reply, Note, Phone, Close)

#### Communication History Section
**Required Elements:**
- Chronological article display (newest first or oldest first preference)
- Article type indicators (email, phone, note, system)
- Sender identification (agent name, customer, system)
- Timestamp display (relative and absolute)
- Visibility indicators (customer visible vs. internal)
- Attachment indicators and download links

**Interactive Elements:**
- Expand/collapse article bodies
- Quote article in reply
- Forward article
- Print article
- Download attachments

#### Article Composition Section
**Required Elements:**
- Subject field (pre-populated from ticket title)
- Rich text editor with formatting options
- Visibility toggle (customer visible / internal note)
- Communication channel selector (email/phone/note)
- Attachment uploader with preview
- Canned response selector
- Signature insertion

**Interactive Elements:**  
- Reply button → compose customer reply
- Note button → compose internal note
- Phone button → record phone conversation
- Template/canned response insertion
- Real-time auto-save of drafts
- File drag-and-drop upload

#### Action Toolbar
**Primary Actions:**
- Reply (default action, keyboard shortcut)
- Add Note (internal communication)
- Phone Call (record phone interaction)
- Close (if agent has permission)
- Forward (to another queue/agent)

**Secondary Actions:**
- Print ticket
- Merge with another ticket
- Split ticket (create child ticket)
- Lock/Unlock ticket
- Subscribe/Watch ticket
- Add to personal queue

### JavaScript Function Requirements

#### Core Functions
```javascript
// Article composition
function composeReply(ticketId) { /* Show reply form */ }
function composeNote(ticketId) { /* Show internal note form */ }
function composePhone(ticketId) { /* Show phone call form */ }

// Status management
function changeStatus(ticketId, stateId) { /* Update ticket status */ }
function assignAgent(ticketId, userId) { /* Assign to agent */ }
function changePriority(ticketId, priorityId) { /* Update priority */ }

// Article management
function quoteArticle(articleId) { /* Quote in new reply */ }
function forwardArticle(articleId) { /* Forward article */ }
function toggleArticle(articleId) { /* Expand/collapse */ }

// Attachment handling
function uploadAttachment(file) { /* Handle file upload */ }
function downloadAttachment(attachmentId) { /* Download file */ }
function previewAttachment(attachmentId) { /* Preview if possible */ }

// Auto-save and drafts
function autoSaveDraft() { /* Save composition draft */ }
function loadDraft(ticketId) { /* Load saved draft */ }
```

#### Integration Functions
```javascript
// Real-time updates
function subscribeToTicketUpdates(ticketId) { /* WebSocket/SSE */ }
function updateTicketHeader(ticketData) { /* Refresh header info */ }
function addNewArticle(articleData) { /* Add to communication history */ }

// Keyboard shortcuts
function initializeKeyboardShortcuts() {
    // R - Reply
    // N - Add Note  
    // P - Phone Call
    // S - Change Status
    // A - Assign
}

// Form validation
function validateReply(formData) { /* Validate reply form */ }
function validateNote(formData) { /* Validate note form */ }
function showValidationError(field, message) { /* Display error */ }
```

### Template Structure Requirements

#### Main Template Layout
```html
<!-- Ticket Zoom Container -->
<div class="ticket-zoom-container">
    <!-- Ticket Header -->
    <div class="ticket-header">
        <div class="ticket-meta">
            <!-- Ticket number, title, timestamps -->
        </div>
        <div class="customer-info">  
            <!-- Customer details, company info -->
        </div>
        <div class="ticket-status">
            <!-- Status, priority, queue, assignment -->
        </div>
        <div class="action-toolbar">
            <!-- Primary action buttons -->
        </div>
    </div>
    
    <!-- Communication History -->
    <div class="article-history">
        <!-- Article list with expand/collapse -->
    </div>
    
    <!-- Composition Area -->
    <div class="article-composer">
        <!-- Reply/Note/Phone composition forms -->
    </div>
    
    <!-- Sidebar (optional) -->
    <div class="ticket-sidebar">
        <!-- Related tickets, customer history -->
    </div>
</div>
```

#### Template Context Requirements
```go
type TicketZoomContext struct {
    Ticket              Ticket
    Articles            []Article
    Customer            CustomerUser
    Company             CustomerCompany
    AvailableStates     []TicketState
    AvailablePriorities []TicketPriority
    AssignableAgents    []User
    Permissions         AgentPermissions
    CannedResponses     []CannedResponse
    Attachments         []Attachment
}
```

## Integration Points

### Authentication & Authorization
- Agent login verification
- Queue access permissions
- Action-level permissions (reply, note, status change)
- Customer data access restrictions

### Email Integration  
- Outbound email sending for replies
- Email template processing
- MIME content handling for attachments
- Email signature insertion

### File Storage
- Attachment upload processing
- File type validation and virus scanning
- Storage backend integration (filesystem/database)
- Thumbnail generation for images

### Real-time Updates
- WebSocket/SSE for live updates
- Multi-agent collision detection
- Draft synchronization
- Status change notifications

### Search Integration
- Full-text search in articles
- Attachment content indexing
- Related ticket suggestions
- Knowledge base integration

## Testing Requirements

### Unit Test Scenarios
```go
// Handler tests
func TestTicketZoomView(t *testing.T)
func TestTicketReply(t *testing.T)  
func TestTicketNote(t *testing.T)
func TestTicketPhone(t *testing.T)
func TestStatusChange(t *testing.T)
func TestAgentAssignment(t *testing.T)

// Service tests
func TestArticleCreation(t *testing.T)
func TestAttachmentUpload(t *testing.T)
func TestPermissionChecking(t *testing.T)
func TestEmailSending(t *testing.T)
```

### Integration Test Workflows
1. **Complete Reply Workflow**
   - Load ticket zoom → compose reply → send → verify database → verify email
   
2. **Internal Note Workflow**  
   - Load ticket zoom → add note → save → verify visibility rules
   
3. **Phone Call Recording**
   - Load ticket zoom → record phone call → save → verify communication history
   
4. **Status Change Workflow**
   - Load ticket zoom → change status → verify update → check notifications
   
5. **Assignment Workflow**
   - Load ticket zoom → assign agent → verify permissions → check notifications

### Browser Testing Requirements
- ✅ All interactive elements respond to clicks
- ✅ Forms submit correctly with proper validation
- ✅ JavaScript functions execute without errors  
- ✅ Real-time updates work correctly
- ✅ File uploads work with progress indicators
- ✅ Keyboard shortcuts function properly
- ✅ Mobile/responsive layout works
- ✅ Dark mode support functional

### Performance Requirements
- Page load time < 2 seconds for tickets with 50+ articles
- Article composition auto-save < 500ms response time
- File upload with progress indication for files up to 25MB
- Real-time updates delivered within 1 second
- Search results returned within 1 second

## Implementation Notes

### Technical Constraints
- Must use existing OTRS-compatible database schema
- Pongo2 template engine for server-side rendering
- Container-first development environment
- PostgreSQL database with existing connection pool
- Gin framework for HTTP handlers

### Security Considerations
- HTML sanitization for article content
- File upload validation and virus scanning
- Cross-site scripting (XSS) prevention
- Cross-site request forgery (CSRF) protection
- Agent session management and timeout

### Performance Optimizations
- Lazy loading of article content
- Attachment thumbnail caching
- Database query optimization for large ticket histories
- Client-side caching of static resources
- Efficient real-time update mechanisms

### Accessibility Requirements
- Keyboard navigation for all functions
- Screen reader compatibility
- High contrast mode support
- Focus management for modal dialogs
- ARIA labels for interactive elements

### Internationalization
- EN/DE translation support required
- Date/time formatting per locale
- Right-to-left language preparation
- Cultural considerations for communication patterns

## Success Criteria

### Functional Success
- ✅ Agent can view complete ticket details
- ✅ Agent can reply to customers successfully  
- ✅ Agent can add internal notes
- ✅ Agent can record phone conversations
- ✅ Agent can change ticket status and assignment
- ✅ Communication history displays correctly
- ✅ File attachments work end-to-end
- ✅ Real-time updates function properly

### Technical Success
- ✅ All HTTP endpoints return appropriate status codes
- ✅ Database operations complete without errors
- ✅ JavaScript functions execute without console errors
- ✅ Templates render without Pongo2 errors
- ✅ Container integration works correctly
- ✅ Service starts and responds to health checks

### Quality Success
- ✅ Comprehensive test coverage for all workflows
- ✅ Performance requirements met under load
- ✅ Security requirements validated
- ✅ Accessibility standards compliance
- ✅ Cross-browser compatibility verified
- ✅ Mobile responsiveness confirmed

**This specification provides the foundation for TDD implementation of the Ticket Zoom interface with complete legal compliance and systematic testing requirements.**