# Content Type Selector Enhancement

## Overview

This enhancement adds a **Content Type Selector** feature to PocketBase Lambda Functions, making it easier to serve different types of content (HTML, JSON, CSS, JavaScript, XML, etc.) without manually setting Content-Type headers.

## Features Added

### 1. Content Type Field
- Added `contentType` field to lambda function model
- Available options:
  - `auto` - Intelligent auto-detection based on response content
  - `text/plain` - Plain text
  - `text/html` - HTML pages
  - `application/json` - JSON data
  - `text/css` - CSS stylesheets  
  - `application/javascript` - JavaScript files
  - `application/xml` - XML data
  - `text/xml` - XML text

### 2. UI Enhancement
- Added content type selector dropdown in function creation/edit panel
- Located next to the timeout field for easy access
- Includes helpful tooltip explaining the feature

### 3. Backend Logic Enhancement
- Enhanced response handling to use selected content type as default
- Intelligent auto-detection when "auto" is selected
- Manual Content-Type headers in function responses still take priority

## How It Works

### Auto-Detection Logic
When `contentType` is set to "auto", the system intelligently detects content type based on:

- **HTML**: Looks for `<!DOCTYPE html>`, `<html>`, `<body>`, `<div>`, `<span>` tags
- **JSON**: Detects objects `{}` and arrays `[]` 
- **XML**: Identifies `<?xml` declarations and XML-like tags
- **CSS**: Recognizes CSS properties like `color:`, `font-`, `margin:`, `padding:`
- **JavaScript**: Finds `function`, `var`, `let`, `const`, `console.log`, `document.`
- **Plain Text**: Default fallback

### Usage Examples

#### Example 1: HTML Function
```javascript
// Set contentType to "text/html" in the UI
return {
  status: 200,
  body: `
    <!DOCTYPE html>
    <html>
      <head><title>My Page</title></head>
      <body><h1>Hello World!</h1></body>
    </html>
  `
};
// Content-Type: text/html (automatically set)
```

#### Example 2: JSON API Function  
```javascript
// Set contentType to "application/json" in the UI
return {
  status: 200,
  body: {
    message: "Success",
    data: { users: 123 }
  }
};
// Content-Type: application/json (automatically set)
```

#### Example 3: CSS Function
```javascript
// Set contentType to "text/css" in the UI
return {
  status: 200,
  body: `
    body { 
      font-family: Arial, sans-serif;
      background: #f5f5f5;
    }
    .container { max-width: 800px; }
  `
};
// Content-Type: text/css (automatically set)
```

#### Example 4: Manual Override
```javascript
// Even with contentType set in UI, manual headers take priority
return {
  status: 200,
  headers: {
    "Content-Type": "application/custom+json"  // This overrides UI setting
  },
  body: JSON.stringify({ custom: "data" })
};
```

## Migration Notes

- Existing lambda functions will work without changes
- New `contentType` field defaults to empty/null for existing functions
- When `contentType` is not set, behavior falls back to previous logic
- No breaking changes to existing function responses

## Developer Benefits

1. **Simplified Development**: No need to manually set Content-Type headers for common scenarios
2. **Better Developer Experience**: Clear UI selector instead of remembering MIME types
3. **Intelligent Defaults**: Auto-detection works for most common content types
4. **Flexibility**: Manual headers still override when needed
5. **Consistent Behavior**: Same content type logic across all function executions

## File Changes

### Backend Files
- `migrations/1751654000_lambdas_init.go` - Added contentType field to schema
- `apis/lambdas.go` - Updated API to handle contentType field
- `plugins/jsvm/lambdas.go` - Enhanced response logic and auto-detection

### Frontend Files  
- `ui/src/components/lambdas/LambdaFunctionUpsertPanel.svelte` - Added content type selector

This enhancement makes PocketBase Lambda Functions more developer-friendly while maintaining full backward compatibility.
