import React from 'react';
import { Box, Text, useInput } from 'ink';
import type { Email } from '../types.js';

interface EmailDetailProps {
  email: Email;
  onBack: () => void;
  onAction: (action: string) => void;
}

export const EmailDetail: React.FC<EmailDetailProps> = ({ email, onBack, onAction }) => {
  useInput((input, key) => {
    if (key.escape || input === 'b') {
      onBack();
    } else if (input === 'r') {
      onAction('reply');
    } else if (input === 'f') {
      onAction('forward');
    } else if (input === 'd') {
      onAction('delete');
    } else if (input === 't') {
      onAction('tag');
    } else if (input === 'm') {
      onAction('mark-read');
    } else if (input === 'u') {
      onAction('mark-unread');
    }
  });

  return (
    <Box flexDirection="column" width="100%">
      <Box borderStyle="single" paddingX={2} paddingY={1} flexDirection="column">
        <Box marginBottom={1}>
          <Text bold color="cyan">From: </Text>
          <Text>{email.from}</Text>
        </Box>
        
        <Box marginBottom={1}>
          <Text bold color="cyan">To: </Text>
          <Text>{email.to}</Text>
        </Box>
        
        <Box marginBottom={1}>
          <Text bold color="cyan">Subject: </Text>
          <Text bold>{email.subject}</Text>
        </Box>
        
        <Box marginBottom={1}>
          <Text bold color="cyan">Date: </Text>
          <Text>{new Date(email.date).toLocaleString()}</Text>
        </Box>

        {email.tags.length > 0 && (
          <Box marginBottom={1}>
            <Text bold color="cyan">Tags: </Text>
            {email.tags.map((tag, i) => (
              <Text key={tag} color="yellow">
                {tag}{i < email.tags.length - 1 ? ', ' : ''}
              </Text>
            ))}
          </Box>
        )}

        {email.hasAttachments && (
          <Box marginBottom={1}>
            <Text color="green">📎 Has attachments</Text>
          </Box>
        )}

        <Box marginTop={1} flexDirection="column">
          <Text bold color="cyan">Message:</Text>
          <Box marginTop={1} paddingLeft={2}>
            <Text wrap="wrap">{email.body}</Text>
          </Box>
        </Box>
      </Box>

      <Box marginTop={1} paddingX={1}>
        <Text dimColor color="gray">
          b/ESC Back | r Reply | f Forward | d Delete | t Tag | m Mark Read | u Mark Unread
        </Text>
      </Box>
    </Box>
  );
};