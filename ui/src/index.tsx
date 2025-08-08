#!/usr/bin/env node
import React from 'react';
import { render } from 'ink';
import { App } from './App.js';

// Parse command line arguments if needed
const args = process.argv.slice(2);

// Render the React Ink app
render(<App />);