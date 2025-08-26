/**
 * Ticket Zoom JavaScript Functions
 * Fixes: "UNCAUGHT REFERENCEERROR: ADDNOTE IS NOT DEFINED"
 */

// Global state
let currentTicketId = null;
let isComposing = false;

/**
 * Initialize ticket zoom functionality
 */
function initTicketZoom(ticketId) {
    currentTicketId = ticketId;
    console.log('Ticket zoom initialized for ticket ID:', ticketId);
    
    // Initialize keyboard shortcuts
    initializeKeyboardShortcuts();
    
    // Initialize auto-save for drafts
    initializeAutoSave();
}

/**
 * Add internal note to ticket
 * Fixes: UNCAUGHT REFERENCEERROR: ADDNOTE IS NOT DEFINED
 */
function addNote(ticketId) {
    if (!ticketId) ticketId = currentTicketId;
    if (!ticketId) {
        console.error('No ticket ID provided for addNote');
        return;
    }
    
    console.log('Adding note to ticket:', ticketId);
    showComposer('note', ticketId);
}

/**
 * Compose customer reply
 */
function composeReply(ticketId) {
    if (!ticketId) ticketId = currentTicketId;
    if (!ticketId) {
        console.error('No ticket ID provided for composeReply');
        return;
    }
    
    console.log('Composing reply for ticket:', ticketId);
    showComposer('reply', ticketId);
}

/**
 * Record phone conversation
 */
function composePhone(ticketId) {
    if (!ticketId) ticketId = currentTicketId;
    if (!ticketId) {
        console.error('No ticket ID provided for composePhone');
        return;
    }
    
    console.log('Recording phone call for ticket:', ticketId);
    showComposer('phone', ticketId);
}

/**
 * Change ticket status
 */
function changeStatus(ticketId, stateId) {
    if (!ticketId) ticketId = currentTicketId;
    if (!ticketId || !stateId) {
        console.error('Missing ticket ID or state ID for changeStatus');
        return;
    }
    
    console.log('Changing ticket status:', ticketId, 'to state:', stateId);
    
    const data = {
        state_id: stateId,
        note: 'Status changed via ticket zoom'
    };
    
    sendRequest('PUT', `/agent/tickets/${ticketId}/status`, data)
        .then(response => {
            if (response.success) {
                showToast('Status updated successfully', 'success');
                refreshTicketHeader();
            } else {
                showToast('Failed to update status: ' + response.error, 'error');
            }
        })
        .catch(error => {
            console.error('Error changing status:', error);
            showToast('Error updating status', 'error');
        });
}

/**
 * Assign ticket to agent
 */
function assignAgent(ticketId, userId) {
    if (!ticketId) ticketId = currentTicketId;
    if (!ticketId || !userId) {
        console.error('Missing ticket ID or user ID for assignAgent');
        return;
    }
    
    console.log('Assigning ticket:', ticketId, 'to user:', userId);
    
    const data = {
        responsible_user_id: userId,
        note: 'Agent assigned via ticket zoom'
    };
    
    sendRequest('PUT', `/agent/tickets/${ticketId}/assign`, data)
        .then(response => {
            if (response.success) {
                showToast('Ticket assigned successfully', 'success');
                refreshTicketHeader();
            } else {
                showToast('Failed to assign ticket: ' + response.error, 'error');
            }
        })
        .catch(error => {
            console.error('Error assigning ticket:', error);
            showToast('Error assigning ticket', 'error');
        });
}

/**
 * Change ticket priority
 */
function changePriority(ticketId, priorityId) {
    if (!ticketId) ticketId = currentTicketId;
    if (!ticketId || !priorityId) {
        console.error('Missing ticket ID or priority ID for changePriority');
        return;
    }
    
    console.log('Changing ticket priority:', ticketId, 'to priority:', priorityId);
    
    const data = {
        priority_id: priorityId,
        note: 'Priority changed via ticket zoom'
    };
    
    sendRequest('PUT', `/agent/tickets/${ticketId}/priority`, data)
        .then(response => {
            if (response.success) {
                showToast('Priority updated successfully', 'success');
                refreshTicketHeader();
            } else {
                showToast('Failed to update priority: ' + response.error, 'error');
            }
        })
        .catch(error => {
            console.error('Error changing priority:', error);
            showToast('Error updating priority', 'error');
        });
}

/**
 * Show article composer
 */
function showComposer(type, ticketId) {
    if (isComposing) {
        console.warn('Already composing an article');
        return;
    }
    
    isComposing = true;
    
    // Hide any existing composers
    hideAllComposers();
    
    // Show the appropriate composer
    const composerId = `composer-${type}`;
    const composer = document.getElementById(composerId);
    
    if (composer) {
        composer.style.display = 'block';
        
        // Pre-populate fields
        const subjectField = composer.querySelector('[name="subject"]');
        if (subjectField) {
            const ticketTitle = document.querySelector('.ticket-title')?.textContent || '';
            const prefix = type === 'reply' ? 'Re: ' : type === 'phone' ? 'Phone: ' : 'Note: ';
            subjectField.value = prefix + ticketTitle;
        }
        
        // Focus on body field
        const bodyField = composer.querySelector('[name="body"]');
        if (bodyField) {
            bodyField.focus();
        }
        
        // Scroll to composer
        composer.scrollIntoView({ behavior: 'smooth' });
        
    } else {
        console.error('Composer not found:', composerId);
        isComposing = false;
    }
}

/**
 * Hide all composers
 */
function hideAllComposers() {
    const composers = document.querySelectorAll('[id^="composer-"]');
    composers.forEach(composer => {
        composer.style.display = 'none';
    });
    isComposing = false;
}

/**
 * Submit article (reply, note, phone)
 */
function submitArticle(type, ticketId) {
    if (!ticketId) ticketId = currentTicketId;
    if (!ticketId) {
        console.error('No ticket ID for submitArticle');
        return;
    }
    
    const composerId = `composer-${type}`;
    const composer = document.getElementById(composerId);
    if (!composer) {
        console.error('Composer not found:', composerId);
        return;
    }
    
    // Collect form data
    const formData = new FormData();
    const subjectField = composer.querySelector('[name="subject"]');
    const bodyField = composer.querySelector('[name="body"]');
    
    if (subjectField) formData.append('subject', subjectField.value);
    if (bodyField) formData.append('body', bodyField.value);
    
    // Set communication channel based on type
    const channelMap = { reply: 1, phone: 2, note: 4 };
    formData.append('communication_channel_id', channelMap[type] || 4);
    formData.append('is_visible_for_customer', type === 'reply' ? '1' : '0');
    
    // Show loading state
    const submitBtn = composer.querySelector('.submit-btn');
    if (submitBtn) {
        submitBtn.disabled = true;
        submitBtn.textContent = 'Sending...';
    }
    
    // Send request
    const endpoint = type === 'reply' ? 'reply' : type === 'phone' ? 'phone' : 'note';
    
    fetch(`/agent/tickets/${ticketId}/${endpoint}`, {
        method: 'POST',
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showToast(`${type.charAt(0).toUpperCase() + type.slice(1)} sent successfully`, 'success');
            hideAllComposers();
            refreshArticleHistory();
        } else {
            showToast(`Failed to send ${type}: ` + (data.error || 'Unknown error'), 'error');
        }
    })
    .catch(error => {
        console.error(`Error sending ${type}:`, error);
        showToast(`Error sending ${type}`, 'error');
    })
    .finally(() => {
        // Reset button state
        if (submitBtn) {
            submitBtn.disabled = false;
            submitBtn.textContent = 'Send';
        }
    });
}

/**
 * Generic AJAX request helper
 */
function sendRequest(method, url, data) {
    const options = {
        method: method,
        headers: {
            'Content-Type': 'application/json',
        }
    };
    
    if (data) {
        options.body = JSON.stringify(data);
    }
    
    return fetch(url, options).then(response => response.json());
}

/**
 * Show toast notification
 */
function showToast(message, type) {
    // Create toast element if it doesn't exist
    let toast = document.getElementById('toast-notification');
    if (!toast) {
        toast = document.createElement('div');
        toast.id = 'toast-notification';
        toast.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 15px 20px;
            border-radius: 4px;
            color: white;
            font-weight: bold;
            z-index: 9999;
            opacity: 0;
            transition: opacity 0.3s ease;
        `;
        document.body.appendChild(toast);
    }
    
    // Set message and type
    toast.textContent = message;
    toast.className = `toast-${type}`;
    
    // Set background color based on type
    const colors = {
        success: '#22c55e',
        error: '#ef4444',
        warning: '#f59e0b',
        info: '#3b82f6'
    };
    toast.style.backgroundColor = colors[type] || colors.info;
    
    // Show toast
    toast.style.opacity = '1';
    
    // Hide after 3 seconds
    setTimeout(() => {
        toast.style.opacity = '0';
    }, 3000);
}

/**
 * Refresh ticket header information
 */
function refreshTicketHeader() {
    // Reload the page for now - can be improved with AJAX later
    console.log('Refreshing ticket header');
    // For now, just reload - TODO: implement AJAX refresh
}

/**
 * Refresh article history
 */
function refreshArticleHistory() {
    // Reload the page for now - can be improved with AJAX later  
    console.log('Refreshing article history');
    setTimeout(() => {
        window.location.reload();
    }, 1000);
}

/**
 * Initialize keyboard shortcuts
 */
function initializeKeyboardShortcuts() {
    document.addEventListener('keydown', function(e) {
        // Only handle shortcuts when not in input fields
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') {
            return;
        }
        
        switch(e.key.toLowerCase()) {
            case 'r':
                if (!isComposing) {
                    e.preventDefault();
                    composeReply(currentTicketId);
                }
                break;
            case 'n':
                if (!isComposing) {
                    e.preventDefault();
                    addNote(currentTicketId);
                }
                break;
            case 'p':
                if (!isComposing) {
                    e.preventDefault();
                    composePhone(currentTicketId);
                }
                break;
            case 'escape':
                if (isComposing) {
                    e.preventDefault();
                    hideAllComposers();
                }
                break;
        }
    });
}

/**
 * Initialize auto-save functionality
 */
function initializeAutoSave() {
    // Auto-save drafts every 30 seconds
    setInterval(() => {
        if (isComposing) {
            saveDraft();
        }
    }, 30000);
}

/**
 * Save draft
 */
function saveDraft() {
    // TODO: Implement draft saving
    console.log('Auto-saving draft...');
}

// Export functions for testing
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        addNote,
        composeReply,
        composePhone,
        changeStatus,
        assignAgent,
        changePriority,
        initTicketZoom
    };
}