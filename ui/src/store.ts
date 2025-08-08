import { create } from 'zustand';
import type { Email, AppState, EmailFilter } from './types.js';

interface Store extends AppState {
  setEmails: (emails: Email[]) => void;
  selectEmail: (index: number) => void;
  toggleEmailSelection: (id: string) => void;
  setSearchQuery: (query: string) => void;
  setActiveView: (view: AppState['activeView']) => void;
  applyFilter: (filter: EmailFilter) => void;
  addTag: (emailId: string, tag: string) => void;
  removeTag: (emailId: string, tag: string) => void;
  markAsRead: (emailIds: string[]) => void;
  markAsUnread: (emailIds: string[]) => void;
  deleteEmails: (emailIds: string[]) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  refreshEmails: () => Promise<void>;
}

export const useStore = create<Store>((set, get) => ({
  emails: [],
  filteredEmails: [],
  selectedIndex: 0,
  searchQuery: '',
  activeView: 'list',
  selectedEmails: new Set(),
  tags: ['important', 'work', 'personal', 'newsletter', 'spam'],
  activeFilter: {},
  loading: false,
  error: null,

  setEmails: (emails) => {
    set({ emails, filteredEmails: emails });
  },

  selectEmail: (index) => {
    set({ selectedIndex: index });
  },

  toggleEmailSelection: (id) => {
    set((state) => {
      const newSelected = new Set(state.selectedEmails);
      if (newSelected.has(id)) {
        newSelected.delete(id);
      } else {
        newSelected.add(id);
      }
      return { selectedEmails: newSelected };
    });
  },

  setSearchQuery: (query) => {
    set((state) => {
      const filtered = state.emails.filter(email => 
        email.subject.toLowerCase().includes(query.toLowerCase()) ||
        email.from.toLowerCase().includes(query.toLowerCase()) ||
        email.body.toLowerCase().includes(query.toLowerCase())
      );
      return { searchQuery: query, filteredEmails: filtered };
    });
  },

  setActiveView: (view) => {
    set({ activeView: view });
  },

  applyFilter: (filter) => {
    set((state) => {
      let filtered = [...state.emails];

      if (filter.unreadOnly) {
        filtered = filtered.filter(e => !e.isRead);
      }

      if (filter.hasAttachments) {
        filtered = filtered.filter(e => e.hasAttachments);
      }

      if (filter.tags && filter.tags.length > 0) {
        filtered = filtered.filter(e => 
          filter.tags!.some(tag => e.tags.includes(tag))
        );
      }

      if (filter.searchTerm) {
        const term = filter.searchTerm.toLowerCase();
        filtered = filtered.filter(e =>
          e.subject.toLowerCase().includes(term) ||
          e.from.toLowerCase().includes(term) ||
          e.body.toLowerCase().includes(term)
        );
      }

      if (filter.dateRange) {
        filtered = filtered.filter(e => {
          const emailDate = new Date(e.date);
          return emailDate >= filter.dateRange!.from && 
                 emailDate <= filter.dateRange!.to;
        });
      }

      return { activeFilter: filter, filteredEmails: filtered };
    });
  },

  addTag: (emailId, tag) => {
    set((state) => {
      const emails = state.emails.map(e => {
        if (e.id === emailId && !e.tags.includes(tag)) {
          return { ...e, tags: [...e.tags, tag] };
        }
        return e;
      });
      return { emails };
    });
  },

  removeTag: (emailId, tag) => {
    set((state) => {
      const emails = state.emails.map(e => {
        if (e.id === emailId) {
          return { ...e, tags: e.tags.filter(t => t !== tag) };
        }
        return e;
      });
      return { emails };
    });
  },

  markAsRead: (emailIds) => {
    set((state) => {
      const emails = state.emails.map(e => {
        if (emailIds.includes(e.id)) {
          return { ...e, isRead: true };
        }
        return e;
      });
      return { emails };
    });
  },

  markAsUnread: (emailIds) => {
    set((state) => {
      const emails = state.emails.map(e => {
        if (emailIds.includes(e.id)) {
          return { ...e, isRead: false };
        }
        return e;
      });
      return { emails };
    });
  },

  deleteEmails: (emailIds) => {
    set((state) => {
      const emails = state.emails.filter(e => !emailIds.includes(e.id));
      const filteredEmails = state.filteredEmails.filter(e => !emailIds.includes(e.id));
      return { emails, filteredEmails };
    });
  },

  setLoading: (loading) => {
    set({ loading });
  },

  setError: (error) => {
    set({ error });
  },

  refreshEmails: async () => {
    const { setLoading, setError, setEmails } = get();
    setLoading(true);
    setError(null);
    
    try {
      // Call to Go backend to fetch emails
      const response = await fetch('http://localhost:8080/api/emails');
      if (!response.ok) throw new Error('Failed to fetch emails');
      const emails = await response.json();
      setEmails(emails);
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  }
}));