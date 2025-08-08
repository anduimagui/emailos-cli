import React from 'react';
import { Box, Text, useInput } from 'ink';
import SelectInput from 'ink-select-input';
import { useStore } from '../store.js';
import type { Email } from '../types.js';

interface EmailListProps {
  onSelect: (email: Email) => void;
  onAction: (action: string) => void;
}

export const EmailList: React.FC<EmailListProps> = ({ onSelect, onAction }) => {
  const { 
    filteredEmails, 
    selectedIndex, 
    selectedEmails,
    selectEmail,
    toggleEmailSelection,
    searchQuery,
    activeFilter
  } = useStore();

  useInput((input, key) => {
    if (key.upArrow) {
      const newIndex = Math.max(0, selectedIndex - 1);
      selectEmail(newIndex);
    } else if (key.downArrow) {
      const newIndex = Math.min(filteredEmails.length - 1, selectedIndex + 1);
      selectEmail(newIndex);
    } else if (key.return) {
      if (filteredEmails[selectedIndex]) {
        onSelect(filteredEmails[selectedIndex]);
      }
    } else if (input === ' ') {
      if (filteredEmails[selectedIndex]) {
        toggleEmailSelection(filteredEmails[selectedIndex].id);
      }
    } else if (input === 'a') {
      // Select all
      filteredEmails.forEach(email => {
        if (!selectedEmails.has(email.id)) {
          toggleEmailSelection(email.id);
        }
      });
    } else if (input === 'd') {
      onAction('delete');
    } else if (input === 'r') {
      onAction('reply');
    } else if (input === 'm') {
      onAction('mark-read');
    } else if (input === 'u') {
      onAction('mark-unread');
    } else if (input === 't') {
      onAction('tag');
    } else if (input === '/') {
      onAction('search');
    } else if (input === 'f') {
      onAction('filter');
    } else if (input === 'c') {
      onAction('compose');
    } else if (input === '?') {
      onAction('help');
    } else if (input === 'q') {
      onAction('quit');
    }
  });

  const formatDate = (date: Date) => {
    const now = new Date();
    const emailDate = new Date(date);
    const diffHours = (now.getTime() - emailDate.getTime()) / (1000 * 60 * 60);
    
    if (diffHours < 24) {
      return emailDate.toLocaleTimeString('en-US', { 
        hour: '2-digit', 
        minute: '2-digit' 
      });
    } else if (diffHours < 168) { // Less than a week
      return emailDate.toLocaleDateString('en-US', { 
        weekday: 'short',
        month: 'short',
        day: 'numeric'
      });
    } else {
      return emailDate.toLocaleDateString('en-US', { 
        month: 'short',
        day: 'numeric',
        year: '2-digit'
      });
    }
  };

  const truncate = (str: string, length: number) => {
    if (str.length <= length) return str.padEnd(length);
    return str.substring(0, length - 3) + '...';
  };

  return (
    <Box flexDirection="column" width="100%">
      <Box marginBottom={1} paddingX={1}>
        <Text bold color="cyan">📧 Email List</Text>
        {searchQuery && (
          <Text color="yellow"> (searching: {searchQuery})</Text>
        )}
        {Object.keys(activeFilter).length > 0 && (
          <Text color="magenta"> [filtered]</Text>
        )}
        <Text color="gray"> ({filteredEmails.length} emails)</Text>
      </Box>

      <Box flexDirection="column" borderStyle="single" paddingX={1}>
        <Box marginBottom={1}>
          <Text color="gray" dimColor>
            {'  '} 
            <Text>{truncate('From', 25)}</Text>
            {' '}
            <Text>{truncate('Subject', 40)}</Text>
            {' '}
            <Text>Date</Text>
            {' '}
            <Text>Tags</Text>
          </Text>
        </Box>

        {filteredEmails.length === 0 ? (
          <Box paddingY={2}>
            <Text color="gray">No emails found</Text>
          </Box>
        ) : (
          filteredEmails.map((email, index) => {
            const isSelected = index === selectedIndex;
            const isChecked = selectedEmails.has(email.id);
            const unreadIndicator = !email.isRead ? '●' : ' ';
            const attachmentIndicator = email.hasAttachments ? '📎' : '  ';
            
            return (
              <Box key={email.id}>
                <Text
                  backgroundColor={isSelected ? 'blue' : undefined}
                  color={isSelected ? 'white' : !email.isRead ? 'white' : 'gray'}
                >
                  {isChecked ? '☑' : '☐'} 
                  {unreadIndicator}
                  {attachmentIndicator} 
                  {truncate(email.from, 25)} 
                  {truncate(email.subject, 40)} 
                  {formatDate(email.date)} 
                  {email.tags.length > 0 && (
                    <Text color="yellow">
                      [{email.tags.slice(0, 2).join(', ')}
                      {email.tags.length > 2 && '...'}]
                    </Text>
                  )}
                </Text>
              </Box>
            );
          })
        )}
      </Box>

      <Box marginTop={1} paddingX={1}>
        <Text dimColor color="gray">
          ↑↓ Navigate | ⏎ Open | Space Select | a All | d Delete | r Reply | 
          m Read | u Unread | t Tag | / Search | f Filter | c Compose | ? Help | q Quit
        </Text>
      </Box>
    </Box>
  );
};