# [COMPLETED] Refactoring Task: Modularize bubbletea_input.go

## Status: COMPLETED - FUNCTIONALITY MOVED TO OLD DIRECTORY

## Overview
The `bubbletea_input.go` file has been moved to the OLD directory. This document is kept for historical reference. The original file contained ~1400 lines of code handling multiple responsibilities. Main command now uses ConfigureWithOptions for account selection instead.

## Current File Structure Analysis

### Current Responsibilities in bubbletea_input.go:
1. Main model struct and state management
2. Email list display and navigation
3. Command palette functionality
4. Account selector UI
5. Email search and filtering
6. File/reference autocomplete
7. Suggestion system
8. Various helper functions
9. UI rendering for multiple views

## Proposed Refactoring Structure

### 1. **email_list.go** - Email List View
**Move out:**
- Email list rendering logic (lines ~590-658)
- Email list navigation handling
- Email fetching functions (`fetchLastEmails`)
- Email list formatting and display
- "Load More" functionality

**Benefits:**
- Separate concern for email browsing
- Easier to add features like sorting, filtering
- Can implement virtual scrolling for large lists

### 2. **command_palette.go** - Command System
**Move out:**
- Command struct and related types
- Command list rendering (lines ~516-548)
- Command filtering logic (`filterCommands`)
- Command execution handling
- Default commands initialization (`getDefaultCommands`, `getEnhancedCommands`)

**Benefits:**
- Centralized command management
- Easy to add new commands
- Can implement command history, aliases

### 3. **account_selector.go** - Account Management UI
**Move out:**
- Account selector rendering (`renderAccountSelector`, lines ~665-742)
- Account switching logic
- Account list management
- Provider grouping display
- Local/session account preferences

**Benefits:**
- Clean separation of account management
- Can add features like account status, sync state
- Easier testing of account switching

### 4. **search_suggestions.go** - Search and Suggestions
**Move out:**
- Suggestion types and structs
- AI suggestion handling (`GetDefaultAISuggestions`)
- Search filtering logic
- Suggestion rendering (lines ~505-514)
- Email search functionality (`searchEmails`, `emailsToReferences`)

**Benefits:**
- Unified search interface
- Can implement advanced search features
- Better suggestion algorithms

### 5. **file_references.go** - File Reference System
**Move out:**
- Reference item struct
- File reference processing (`processFileReferences`)
- File filtering logic
- @ mention handling for files
- File path resolution

**Benefits:**
- Dedicated file handling logic
- Can add file preview, validation
- Support for different file types

### 6. **input_handler.go** - Core Input Management
**Move out:**
- Input mode enum and state
- Key event routing logic
- Focus management between inputs
- Input validation and processing

**Benefits:**
- Cleaner event handling
- Easier to add keyboard shortcuts
- Better input state management

### 7. **ui_components.go** - Shared UI Components
**Move out:**
- Style definitions (lines 14-41)
- Common UI helpers (`minInt`, `wrapText`)
- Box and border rendering
- Color schemes and themes

**Benefits:**
- Consistent styling across app
- Theme support potential
- Reusable UI components

## Implementation Plan

### Phase 1: Core Extraction (Priority: High)
1. Extract `email_list.go` - Most independent, clear boundaries
2. Extract `command_palette.go` - Well-defined functionality
3. Update `bubbletea_input.go` to use new modules

### Phase 2: UI Components (Priority: Medium)
1. Extract `ui_components.go` - Create shared styling
2. Extract `account_selector.go` - Standalone UI component
3. Standardize component interfaces

### Phase 3: Search & References (Priority: Medium)
1. Extract `search_suggestions.go` - Consolidate search logic
2. Extract `file_references.go` - File handling specifics
3. Create unified search interface

### Phase 4: Input Management (Priority: Low)
1. Extract `input_handler.go` - Complex but central
2. Refactor event routing
3. Optimize state management

## Resulting Structure

```
emailos/
├── bubbletea_input.go      (200-300 lines - core model and orchestration)
├── email_view.go            (existing - individual email display)
├── email_list.go            (new - email list management)
├── command_palette.go       (new - command system)
├── account_selector.go      (new - account UI)
├── search_suggestions.go    (new - search and suggestions)
├── file_references.go       (new - file handling)
├── input_handler.go         (new - input management)
└── ui_components.go         (new - shared UI elements)
```

## Benefits of Refactoring

1. **Maintainability**: Each file has a single, clear responsibility
2. **Testability**: Smaller units are easier to test in isolation
3. **Reusability**: Components can be reused in other contexts
4. **Performance**: Can optimize individual components
5. **Collaboration**: Multiple developers can work on different features
6. **Documentation**: Easier to document focused modules

## Migration Strategy

1. **Create new files** without removing old code
2. **Test new modules** independently
3. **Switch references** one at a time
4. **Remove old code** after verification
5. **Update imports** and dependencies

## Testing Requirements

Each new module should have:
- Unit tests for core functions
- Integration tests with main model
- UI/rendering tests where applicable
- Performance benchmarks for critical paths

## Risks and Mitigation

**Risk**: Breaking existing functionality
- **Mitigation**: Incremental migration with testing

**Risk**: Performance regression
- **Mitigation**: Benchmark before/after

**Risk**: Increased complexity
- **Mitigation**: Clear interfaces and documentation

## Success Criteria

- [ ] bubbletea_input.go reduced to <300 lines
- [ ] All functionality preserved
- [ ] No performance degradation
- [ ] Improved code coverage
- [ ] Clear module boundaries
- [ ] Documentation for each module

## Next Steps

1. Review and approve refactoring plan
2. Create feature branch for refactoring
3. Start with Phase 1 extraction
4. Add tests for extracted modules
5. Iterate through phases with testing