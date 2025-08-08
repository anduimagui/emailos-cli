import * as esbuild from 'esbuild';
import { promises as fs } from 'fs';
import path from 'path';

async function build() {
  try {
    // Ensure dist directory exists
    await fs.mkdir('dist', { recursive: true });

    // Build the application
    await esbuild.build({
      entryPoints: ['src/index.tsx'],
      bundle: true,
      platform: 'node',
      target: 'node18',
      outfile: 'dist/index.js',
      format: 'esm',
      external: [
        'react',
        'ink',
        'ink-*',
        'zustand'
      ],
      loader: {
        '.tsx': 'tsx',
        '.ts': 'ts'
      },
      resolveExtensions: ['.tsx', '.ts', '.jsx', '.js'],
      sourcemap: true,
      minify: false,
      banner: {
        js: '#!/usr/bin/env node\n'
      }
    });

    console.log('✅ Build completed successfully!');
  } catch (error) {
    console.error('❌ Build failed:', error);
    process.exit(1);
  }
}

build();