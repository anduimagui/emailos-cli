# EmailOS Template Command Documentation

The `mailos template` command manages HTML email templates for styling outgoing emails with custom designs and branding.

## Basic Usage

```bash
mailos template
```

Opens interactive template editor to customize your email appearance.

## Quick Start

### View Existing Template
```bash
mailos template --open-browser
```
Opens your current template HTML file directly in the browser for quick visual inspection.

### Edit Template Interactively
```bash
mailos template
```
Opens the full interactive template editor with design options.

## Command-Line Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--remove` | Remove existing template | `mailos template --remove` |
| `--open-browser` | Open template HTML file in browser | `mailos template --open-browser` |

## Template Features

### Supported Elements
- Custom colors and fonts
- Logo/branding placement
- Header and footer sections
- Signature formatting
- Button styles
- Responsive design
- Dark mode support

### Template Variables
Templates can use placeholders:
- `{{BODY}}` - Email content
- `{{SIGNATURE}}` - Email signature
- `{{DATE}}` - Current date
- `{{FROM_NAME}}` - Sender name
- `{{FROM_EMAIL}}` - Sender email
- `{{PROFILE_IMAGE}}` - Profile picture

## Interactive Editor

Running `mailos template` launches the editor with options:

### 1. Quick Styles
Pre-designed templates:
- **Professional** - Clean business style
- **Modern** - Contemporary design
- **Minimal** - Simple and elegant
- **Creative** - Colorful and bold
- **Dark** - Dark theme
- **Custom** - Build from scratch

### 2. Customization Options

#### Colors
- Primary color (headers, links)
- Secondary color (accents)
- Background color
- Text color
- Border color

#### Typography
- Font family selection
- Font size adjustments
- Line height settings
- Heading styles

#### Layout
- Container width
- Padding/margins
- Border radius
- Shadow effects

#### Branding
- Logo upload/URL
- Logo position
- Company name
- Tagline

## Template Examples

### Professional Template
```html
<!DOCTYPE html>
<html>
<head>
  <style>
    .email-container {
      max-width: 600px;
      margin: 0 auto;
      font-family: Arial, sans-serif;
      color: #333;
    }
    .header {
      background: #2c3e50;
      color: white;
      padding: 20px;
      text-align: center;
    }
    .content {
      padding: 30px;
      background: #ffffff;
    }
    .footer {
      background: #ecf0f1;
      padding: 20px;
      text-align: center;
      font-size: 12px;
    }
  </style>
</head>
<body>
  <div class="email-container">
    <div class="header">
      <h1>{{FROM_NAME}}</h1>
    </div>
    <div class="content">
      {{BODY}}
    </div>
    <div class="footer">
      {{SIGNATURE}}
    </div>
  </div>
</body>
</html>
```

### Minimal Template
```html
<div style="max-width: 500px; margin: 20px auto; font-family: Georgia, serif;">
  <div style="border-bottom: 2px solid #e0e0e0; padding-bottom: 10px; margin-bottom: 20px;">
    <img src="{{PROFILE_IMAGE}}" style="width: 50px; height: 50px; border-radius: 50%; vertical-align: middle;">
    <span style="margin-left: 10px; font-size: 18px;">{{FROM_NAME}}</span>
  </div>
  <div style="line-height: 1.6;">
    {{BODY}}
  </div>
  <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #e0e0e0; font-size: 14px; color: #666;">
    {{SIGNATURE}}
  </div>
</div>
```

## Template Management

### Save Location
Templates are stored in:
- Global: `~/.email/templates/default.html`
- Local: `./.email/templates/default.html`

### View Template in Browser
```bash
mailos template --open-browser
```
Instantly opens your current template HTML file in the default browser for visual review. This is useful for:
- Quick visual inspection of template design
- Testing template appearance without sending emails
- Debugging layout issues
- Reviewing HTML structure and styling

The command automatically locates your template file and opens it using the system's default browser. If no template exists, it will show an error message.

### Apply Template
Templates are automatically applied to:
- Emails sent with `mailos send`
- Unless `--plain` flag is used
- When body contains markdown

### Remove Template
```bash
mailos template --remove
```
Reverts to plain email format.

### Export Template
```bash
cp ~/.email/template.html my-template.html
```

### Import Template
```bash
cp my-template.html ~/.email/template.html
```

## Advanced Customization

### CSS Framework Integration
Use popular frameworks:
```html
<!-- Bootstrap -->
<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">

<!-- Tailwind -->
<link href="https://cdn.tailwindcss.com" rel="stylesheet">
```

### Responsive Design
```css
@media only screen and (max-width: 600px) {
  .email-container {
    width: 100% !important;
  }
  .content {
    padding: 10px !important;
  }
}
```

### Dark Mode Support
```css
@media (prefers-color-scheme: dark) {
  .email-container {
    background: #1a1a1a;
    color: #ffffff;
  }
}
```

### Custom Fonts
```html
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600&display=swap" rel="stylesheet">
<style>
  body {
    font-family: 'Inter', sans-serif;
  }
</style>
```

## Profile Images

### Adding Profile Image
```bash
mailos configure --profile-image /path/to/image.jpg
```

### Using in Template
```html
<img src="{{PROFILE_IMAGE}}" alt="Profile" style="width: 60px; border-radius: 50%;">
```

### Image Hosting
Options for profile images:
1. Base64 encoded (embedded)
2. URL to hosted image
3. Gravatar integration
4. Local file reference

## Template Best Practices

### 1. Email Client Compatibility
- Use inline CSS
- Avoid JavaScript
- Test in multiple clients
- Use table layouts for complex designs

### 2. Accessibility
- Include alt text for images
- Use semantic HTML
- Ensure color contrast
- Test with screen readers

### 3. Performance
- Optimize images
- Minimize CSS
- Avoid external dependencies
- Keep file size under 100KB

### 4. Mobile Optimization
- Single column layout
- Large tap targets
- Readable font sizes
- Responsive images

## Testing Templates

### Browser Preview
```bash
mailos template --open-browser
```
Opens template HTML file directly in browser for immediate visual inspection.

### Email Preview
```bash
mailos send --to your@email.com --subject "Template Test" --body "Test content" --preview
```
Shows complete email content with template applied before sending.

### Test Across Clients
Test in:
- Gmail
- Outlook
- Apple Mail
- Mobile apps
- Web clients

### Validation
Check for:
- HTML validity
- CSS compatibility
- Image loading
- Link functionality

## Common Patterns

### Call-to-Action Button
```html
<a href="#" style="
  display: inline-block;
  padding: 12px 24px;
  background: #007bff;
  color: white;
  text-decoration: none;
  border-radius: 4px;
  font-weight: bold;
">Click Here</a>
```

### Social Media Links
```html
<div style="text-align: center; margin-top: 20px;">
  <a href="#" style="margin: 0 10px;">Twitter</a>
  <a href="#" style="margin: 0 10px;">LinkedIn</a>
  <a href="#" style="margin: 0 10px;">Facebook</a>
</div>
```

### Quoted Text
```html
<blockquote style="
  border-left: 4px solid #007bff;
  padding-left: 15px;
  margin: 20px 0;
  font-style: italic;
  color: #666;
">
  Quote text here
</blockquote>
```

## Troubleshooting

### Template Not Applied
- Check template file exists
- Verify not using `--plain` flag
- Ensure HTML is valid
- Check for syntax errors

### Images Not Showing
- Use absolute URLs
- Check image permissions
- Verify HTTPS for external images
- Consider base64 encoding

### Styling Issues
- Use inline CSS
- Avoid modern CSS features
- Test in target email clients
- Check specificity conflicts

### Template Too Large
- Optimize images
- Minimize CSS
- Remove unused styles
- Consider simpler design

## Examples

### Corporate Template
```bash
mailos template
# Select "Professional"
# Set company colors
# Add logo URL
# Customize fonts
```

### Personal Blog Template
```bash
mailos template
# Select "Creative"
# Choose vibrant colors
# Add profile image
# Custom signature
```

### Newsletter Template
```bash
mailos template
# Select "Modern"
# Wide layout
# Header image
# Social media links
```

## Tips

1. **Start Simple**: Begin with minimal template and add features
2. **Test Often**: Send test emails after each major change
3. **Keep Backups**: Save working templates before modifications
4. **Use Variables**: Leverage placeholders for dynamic content
5. **Mobile First**: Design for mobile, enhance for desktop