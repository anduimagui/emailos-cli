import React, { useState } from 'react';
import { Box, Text } from 'ink';
import TextInput from 'ink-text-input';
import { useStore } from '../store.js';

interface SearchBarProps {
  onClose: () => void;
}

export const SearchBar: React.FC<SearchBarProps> = ({ onClose }) => {
  const [value, setValue] = useState('');
  const { setSearchQuery } = useStore();

  const handleSubmit = () => {
    setSearchQuery(value);
    onClose();
  };

  return (
    <Box borderStyle="single" paddingX={1}>
      <Text color="cyan">Search: </Text>
      <TextInput
        value={value}
        onChange={setValue}
        onSubmit={handleSubmit}
        placeholder="Type to search emails..."
      />
    </Box>
  );
};