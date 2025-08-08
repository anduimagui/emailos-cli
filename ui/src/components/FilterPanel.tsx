import React, { useState } from 'react';
import { Box, Text, useInput } from 'ink';
import SelectInput from 'ink-select-input';
import { useStore } from '../store.js';
import type { EmailFilter } from '../types.js';

interface FilterPanelProps {
  onClose: () => void;
}

export const FilterPanel: React.FC<FilterPanelProps> = ({ onClose }) => {
  const { applyFilter, tags } = useStore();
  const [filter, setFilter] = useState<EmailFilter>({});

  const items = [
    { label: `☐ Unread only`, value: 'unread' },
    { label: `☐ Has attachments`, value: 'attachments' },
    { label: 'Date range: All time', value: 'date' },
    ...tags.map(tag => ({
      label: `☐ Tag: ${tag}`,
      value: `tag:${tag}`
    })),
    { label: '✓ Apply filters', value: 'apply' },
    { label: '✗ Clear all', value: 'clear' },
    { label: '← Cancel', value: 'cancel' }
  ];

  const handleSelect = (item: { value: string }) => {
    if (item.value === 'apply') {
      applyFilter(filter);
      onClose();
    } else if (item.value === 'clear') {
      setFilter({});
      applyFilter({});
    } else if (item.value === 'cancel') {
      onClose();
    } else if (item.value === 'unread') {
      setFilter({ ...filter, unreadOnly: !filter.unreadOnly });
    } else if (item.value === 'attachments') {
      setFilter({ ...filter, hasAttachments: !filter.hasAttachments });
    } else if (item.value === 'date') {
      // In a real app, this would open a date picker
      const lastWeek = new Date();
      lastWeek.setDate(lastWeek.getDate() - 7);
      setFilter({
        ...filter,
        dateRange: {
          from: lastWeek,
          to: new Date()
        }
      });
    } else if (item.value.startsWith('tag:')) {
      const tag = item.value.replace('tag:', '');
      const currentTags = filter.tags || [];
      if (currentTags.includes(tag)) {
        setFilter({
          ...filter,
          tags: currentTags.filter(t => t !== tag)
        });
      } else {
        setFilter({
          ...filter,
          tags: [...currentTags, tag]
        });
      }
    }
  };

  useInput((input, key) => {
    if (key.escape) {
      onClose();
    }
  });

  // Update item labels based on current filter state
  const updatedItems = items.map(item => {
    if (item.value === 'unread') {
      return { ...item, label: `${filter.unreadOnly ? '☑' : '☐'} Unread only` };
    }
    if (item.value === 'attachments') {
      return { ...item, label: `${filter.hasAttachments ? '☑' : '☐'} Has attachments` };
    }
    if (item.value === 'date' && filter.dateRange) {
      return { ...item, label: 'Date range: Last 7 days' };
    }
    if (item.value.startsWith('tag:')) {
      const tag = item.value.replace('tag:', '');
      const isSelected = filter.tags?.includes(tag);
      return { ...item, label: `${isSelected ? '☑' : '☐'} Tag: ${tag}` };
    }
    return item;
  });

  return (
    <Box flexDirection="column" borderStyle="single" paddingX={1} paddingY={1}>
      <Text bold color="cyan" marginBottom={1}>Filter Emails</Text>
      <SelectInput items={updatedItems} onSelect={handleSelect} />
      <Box marginTop={1} flexDirection="column">
        <Text dimColor color="gray">Active filters:</Text>
        {filter.unreadOnly && <Text color="yellow">• Unread only</Text>}
        {filter.hasAttachments && <Text color="yellow">• Has attachments</Text>}
        {filter.dateRange && <Text color="yellow">• Last 7 days</Text>}
        {filter.tags?.map(tag => (
          <Text key={tag} color="yellow">• Tag: {tag}</Text>
        ))}
        {Object.keys(filter).length === 0 && <Text color="gray">None</Text>}
      </Box>
      <Text dimColor color="gray" marginTop={1}>
        ↑↓ Navigate | Enter Toggle/Select | ESC Cancel
      </Text>
    </Box>
  );
};