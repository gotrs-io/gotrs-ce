-- ============================================
-- Rollback OTRS Final Default Data
-- ============================================

-- Remove default notification events
DELETE FROM notification_event WHERE name IN (
    'Ticket::NewTicket', 'Ticket::FollowUp', 'Ticket::QueueUpdate',
    'Ticket::StateUpdate', 'Ticket::OwnerUpdate', 'Ticket::ResponsibleUpdate',
    'Ticket::PriorityUpdate', 'Ticket::CustomerUpdate', 'Ticket::PendingTimeUpdate',
    'Ticket::LockUpdate', 'Ticket::ArchiveFlagUpdate', 'Ticket::EscalationTimeStart',
    'Ticket::EscalationTimeStop', 'Ticket::EscalationTimeNotifyBefore',
    'Article::NewArticle', 'Article::ArticleSend'
);

-- Remove default link types
DELETE FROM link_type WHERE name IN (
    'Normal', 'ParentChild', 'DependsOn', 'RelevantTo',
    'AlternativeTo', 'ConnectedTo', 'DuplicateOf'
);

-- Note: article_flag table is populated dynamically, no defaults to remove

-- Optionally remove default admin user (be careful!)
-- DELETE FROM group_user WHERE user_id = (SELECT id FROM users WHERE login = 'root@localhost');
-- DELETE FROM users WHERE login = 'root@localhost';

-- Note: We don't remove groups as they might be critical