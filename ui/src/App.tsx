import React, { useState, useEffect } from 'react';
import { Box, Text, useApp, useInput } from 'ink';
import Spinner from 'ink-spinner';
import Gradient from 'ink-gradient';
import BigText from 'ink-big-text';
import { useStore } from './store.js';
import { EmailList } from './components/EmailList.js';
import { EmailDetail } from './components/EmailDetail.js';
import { SearchBar } from './components/SearchBar.js';
import { TagManager } from './components/TagManager.js';
import { FilterPanel } from './components/FilterPanel.js';
import type { Email } from './types.js';

export const App: React.FC = () => {
  const { exit } = useApp();
  const [currentView, setCurrentView] = useState<'list' | 'detail' | 'search' | 'filter' | 'tag'>('list');
  const [selectedEmail, setSelectedEmail] = useState<Email | null>(null);
  const [tagEmailId, setTagEmailId] = useState<string | null>(null);
  
  const { 
    loading, 
    error, 
    emails,
    setEmails,
    refreshEmails,
    selectedEmails,
    deleteEmails,
    markAsRead,
    markAsUnread
  } = useStore();

  useEffect(() => {
    // Load sample data or fetch from backend
    const sampleEmails: Email[] = [
      {
        id: '1',
        from: 'alice@example.com',
        to: 'you@example.com',
        subject: 'Project Update - Q4 Planning',
        date: new Date(Date.now() - 2 * 60 * 60 * 1000),
        body: 'Hi! Here\'s the latest update on our Q4 planning. We need to review the roadmap and adjust our priorities based on recent feedback.',
        isRead: false,
        hasAttachments: true,
        tags: ['work', 'important']
      },
      {
        id: '2',
        from: 'newsletter@techblog.com',
        to: 'you@example.com',
        subject: 'Weekly Tech Digest: AI Advances',
        date: new Date(Date.now() - 24 * 60 * 60 * 1000),
        body: 'This week in tech: Major breakthroughs in AI, new frameworks released, and industry insights.',
        isRead: true,
        hasAttachments: false,
        tags: ['newsletter']
      },
      {
        id: '3',
        from: 'bob@company.com',
        to: 'you@example.com',
        subject: 'Meeting Notes - Product Review',
        date: new Date(Date.now() - 48 * 60 * 60 * 1000),
        body: 'Thanks for attending the product review. Here are the action items we discussed...',
        isRead: true,
        hasAttachments: true,
        tags: ['work']
      },
      {
        id: '4',
        from: 'support@service.com',
        to: 'you@example.com',
        subject: 'Your subscription is expiring soon',
        date: new Date(Date.now() - 72 * 60 * 60 * 1000),
        body: 'Your annual subscription will expire in 30 days. Renew now to continue enjoying our services.',
        isRead: false,
        hasAttachments: false,
        tags: []
      },
      {
        id: '5',
        from: 'team@github.com',
        to: 'you@example.com',
        subject: 'Security alert: new sign-in detected',
        date: new Date(Date.now() - 96 * 60 * 60 * 1000),
        body: 'We detected a new sign-in to your account from a new device. If this was you, you can safely ignore this email.',
        isRead: false,
        hasAttachments: false,
        tags: ['important']
      }
    ];
    
    setEmails(sampleEmails);
  }, []);

  const handleSelectEmail = (email: Email) => {
    setSelectedEmail(email);
    setCurrentView('detail');
  };

  const handleAction = (action: string) => {
    switch (action) {
      case 'search':
        setCurrentView('search');
        break;
      case 'filter':
        setCurrentView('filter');
        break;
      case 'tag':
        if (selectedEmails.size > 0) {
          const firstSelectedId = Array.from(selectedEmails)[0];
          setTagEmailId(firstSelectedId);
          setCurrentView('tag');
        } else if (selectedEmail) {
          setTagEmailId(selectedEmail.id);
          setCurrentView('tag');
        }
        break;
      case 'delete':
        if (selectedEmails.size > 0) {
          deleteEmails(Array.from(selectedEmails));
        } else if (selectedEmail) {
          deleteEmails([selectedEmail.id]);
          setCurrentView('list');
        }
        break;
      case 'mark-read':
        if (selectedEmails.size > 0) {
          markAsRead(Array.from(selectedEmails));
        } else if (selectedEmail) {
          markAsRead([selectedEmail.id]);
        }
        break;
      case 'mark-unread':
        if (selectedEmails.size > 0) {
          markAsUnread(Array.from(selectedEmails));
        } else if (selectedEmail) {
          markAsUnread([selectedEmail.id]);
        }
        break;
      case 'compose':
        // In a real app, this would open a compose view
        console.log('Compose new email');
        break;
      case 'reply':
        // In a real app, this would open a reply view
        console.log('Reply to email');
        break;
      case 'forward':
        // In a real app, this would open a forward view
        console.log('Forward email');
        break;
      case 'help':
        // In a real app, this would show help
        console.log('Show help');
        break;
      case 'quit':
        exit();
        break;
    }
  };

  const handleBackToList = () => {
    setCurrentView('list');
    setSelectedEmail(null);
  };

  if (loading) {
    return (
      <Box flexDirection="column" alignItems="center" justifyContent="center" height={20}>
        <Box marginBottom={1}>
          <Spinner type="dots" />
          <Text color="cyan"> Loading emails...</Text>
        </Box>
      </Box>
    );
  }

  if (error) {
    return (
      <Box flexDirection="column" alignItems="center" justifyContent="center" height={20}>
        <Text color="red">❌ Error: {error}</Text>
        <Text color="gray" marginTop={1}>Press 'r' to retry or 'q' to quit</Text>
      </Box>
    );
  }

  return (
    <Box flexDirection="column" height="100%">
      <Box marginBottom={1} paddingX={2}>
        <Gradient name="rainbow">
          <BigText text="MailOS" font="simple" />
        </Gradient>
        <Text color="gray">Interactive Email Client</Text>
      </Box>

      {currentView === 'list' && (
        <EmailList 
          onSelect={handleSelectEmail}
          onAction={handleAction}
        />
      )}

      {currentView === 'detail' && selectedEmail && (
        <EmailDetail 
          email={selectedEmail}
          onBack={handleBackToList}
          onAction={handleAction}
        />
      )}

      {currentView === 'search' && (
        <SearchBar onClose={() => setCurrentView('list')} />
      )}

      {currentView === 'filter' && (
        <FilterPanel onClose={() => setCurrentView('list')} />
      )}

      {currentView === 'tag' && tagEmailId && (
        <TagManager 
          emailId={tagEmailId}
          currentTags={emails.find(e => e.id === tagEmailId)?.tags || []}
          onClose={() => {
            setCurrentView('list');
            setTagEmailId(null);
          }}
        />
      )}

      <Box marginTop={1} paddingX={2} borderStyle="single" borderColor="gray">
        <Text color="cyan">
          {selectedEmails.size > 0 && `${selectedEmails.size} selected | `}
          {emails.length} total emails
        </Text>
      </Box>
    </Box>
  );
};