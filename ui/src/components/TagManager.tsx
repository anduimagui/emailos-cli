import React, { useState } from 'react';
import { Box, Text, useInput } from 'ink';
import SelectInput from 'ink-select-input';
import TextInput from 'ink-text-input';
import { useStore } from '../store.js';

interface TagManagerProps {
  emailId: string;
  currentTags: string[];
  onClose: () => void;
}

export const TagManager: React.FC<TagManagerProps> = ({ emailId, currentTags, onClose }) => {
  const { tags, addTag, removeTag } = useStore();
  const [mode, setMode] = useState<'select' | 'create'>('select');
  const [newTag, setNewTag] = useState('');

  useInput((input, key) => {
    if (key.escape) {
      onClose();
    } else if (input === 'n') {
      setMode('create');
    }
  });

  const availableTags = tags.filter(tag => !currentTags.includes(tag));

  const handleSelectTag = (item: { label: string; value: string }) => {
    if (item.value === 'new') {
      setMode('create');
    } else if (item.value.startsWith('remove:')) {
      const tagToRemove = item.value.replace('remove:', '');
      removeTag(emailId, tagToRemove);
      onClose();
    } else {
      addTag(emailId, item.value);
      onClose();
    }
  };

  const handleCreateTag = () => {
    if (newTag.trim()) {
      addTag(emailId, newTag.trim());
      onClose();
    }
  };

  if (mode === 'create') {
    return (
      <Box flexDirection="column" borderStyle="single" paddingX={1} paddingY={1}>
        <Text bold color="cyan">Create New Tag</Text>
        <Box marginTop={1}>
          <Text>Tag name: </Text>
          <TextInput
            value={newTag}
            onChange={setNewTag}
            onSubmit={handleCreateTag}
            placeholder="Enter tag name..."
          />
        </Box>
        <Text dimColor color="gray" marginTop={1}>
          Press Enter to create, ESC to cancel
        </Text>
      </Box>
    );
  }

  const items = [
    ...availableTags.map(tag => ({
      label: `Add: ${tag}`,
      value: tag
    })),
    ...currentTags.map(tag => ({
      label: `Remove: ${tag} ✓`,
      value: `remove:${tag}`
    })),
    {
      label: '+ Create new tag',
      value: 'new'
    }
  ];

  return (
    <Box flexDirection="column" borderStyle="single" paddingX={1} paddingY={1}>
      <Text bold color="cyan" marginBottom={1}>Manage Tags</Text>
      <SelectInput items={items} onSelect={handleSelectTag} />
      <Text dimColor color="gray" marginTop={1}>
        ↑↓ Navigate | Enter Select | n New Tag | ESC Cancel
      </Text>
    </Box>
  );
};