export interface Email {
  id: string;
  from: string;
  to: string;
  subject: string;
  date: Date;
  body: string;
  isRead: boolean;
  hasAttachments: boolean;
  tags: string[];
  selected?: boolean;
}

export interface AppState {
  emails: Email[];
  filteredEmails: Email[];
  selectedIndex: number;
  searchQuery: string;
  activeView: 'list' | 'detail' | 'compose' | 'settings';
  selectedEmails: Set<string>;
  tags: string[];
  activeFilter: EmailFilter;
  loading: boolean;
  error: string | null;
}

export interface EmailFilter {
  unreadOnly?: boolean;
  hasAttachments?: boolean;
  tags?: string[];
  dateRange?: {
    from: Date;
    to: Date;
  };
  searchTerm?: string;
}

export type Action = 
  | 'read'
  | 'compose'
  | 'reply'
  | 'forward'
  | 'delete'
  | 'archive'
  | 'tag'
  | 'mark-read'
  | 'mark-unread'
  | 'search'
  | 'filter'
  | 'refresh'
  | 'settings'
  | 'help'
  | 'quit';

export interface Command {
  key: string;
  description: string;
  action: Action;
  shortcut?: string;
}