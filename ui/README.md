# MailOS React Ink UI

A modern, interactive terminal UI for MailOS built with React Ink.

## Features

- 📧 **Email List View**: Browse emails with keyboard navigation
- 🏷️ **Tagging System**: Organize emails with custom tags
- 🔍 **Search**: Quick search across email subjects, senders, and content
- 🎯 **Filters**: Filter by unread, attachments, tags, and date ranges
- ✅ **Batch Operations**: Select multiple emails for bulk actions
- ⌨️ **Keyboard Shortcuts**: Efficient navigation and actions

## Installation

From the main MailOS directory:

```bash
cd ui
npm install
npm run build
```

## Usage

### Launch with Environment Variable

```bash
MAILOS_REACT_UI=true mailos interactive
```

### Keyboard Shortcuts

#### Email List
- `↑/↓` - Navigate emails
- `Enter` - Open email detail
- `Space` - Toggle selection
- `a` - Select all
- `d` - Delete selected
- `r` - Reply to email
- `m` - Mark as read
- `u` - Mark as unread
- `t` - Tag email
- `/` - Search
- `f` - Filter
- `c` - Compose new
- `?` - Help
- `q` - Quit

#### Email Detail
- `b/ESC` - Back to list
- `r` - Reply
- `f` - Forward
- `d` - Delete
- `t` - Tag
- `m` - Mark as read
- `u` - Mark as unread

## Development

### Run in Development Mode

```bash
npm run dev
```

### Type Checking

```bash
npm run typecheck
```

### Project Structure

```
ui/
├── src/
│   ├── components/       # React components
│   │   ├── EmailList.tsx
│   │   ├── EmailDetail.tsx
│   │   ├── SearchBar.tsx
│   │   ├── TagManager.tsx
│   │   └── FilterPanel.tsx
│   ├── store.ts          # Zustand state management
│   ├── types.ts          # TypeScript types
│   ├── App.tsx           # Main app component
│   └── index.tsx         # Entry point
├── package.json
├── tsconfig.json
└── build.js              # Build script
```

## API Integration

The UI communicates with the Go backend via HTTP API:

- `GET /api/emails` - Fetch all emails
- `POST /api/emails/read` - Mark emails as read
- `POST /api/emails/delete` - Delete emails
- `POST /api/emails/send` - Send new email

The API server automatically starts when launching the React Ink UI.

## Customization

### Adding New Tags

Tags are managed in the Zustand store. Default tags include:
- important
- work
- personal
- newsletter
- spam

Users can create custom tags through the UI.

### Theming

The UI uses Ink's built-in color system. Colors can be customized in components:
- `cyan` - Headers and labels
- `yellow` - Tags and filters
- `gray` - Help text and inactive items
- `blue` - Selected items
- `red` - Errors
- `green` - Success messages

## Troubleshooting

### Build Issues

If the build fails:
1. Ensure Node.js 18+ is installed
2. Clear node_modules and reinstall: `rm -rf node_modules && npm install`
3. Check TypeScript errors: `npm run typecheck`

### UI Not Launching

1. Verify the UI is built: `ls dist/index.js`
2. Check environment variable: `echo $MAILOS_REACT_UI`
3. Rebuild if needed: `npm run build`

### API Connection Issues

The UI expects the API server on:
- `http://localhost:8080` (macOS/Linux)
- `http://localhost:8081` (Windows)

Check the API server is running when launching the UI.

## Future Enhancements

- [ ] Email composition with rich text editor
- [ ] Attachment handling
- [ ] Calendar integration
- [ ] Contact management
- [ ] Email templates
- [ ] Offline mode with sync
- [ ] Multiple account support
- [ ] Theme customization
- [ ] Plugin system