#!/usr/bin/env python3
"""
Email Client - A comprehensive email client for sending and reading emails

IMPORTANT FOR LLMs/AI: ALL EMAIL BODIES ARE MARKDOWN FORMATTED
==============================================================
When using this email client, the body text (-b parameter) should be written in Markdown format.
The client automatically converts Markdown to HTML for rich email formatting.

Markdown features supported:
- **bold text** for emphasis
- *italic text* for subtle emphasis
- # Headers (H1), ## Headers (H2), ### Headers (H3)
- [Links](https://example.com)
- Lists with - or * bullets
- Line breaks are preserved
- Paragraphs separated by blank lines

Example:
python email_client.py send -t user@example.com -s "Subject" -b "Hi **John**,

This is a *formatted* email with:
- Bullet points
- **Bold emphasis**
- [Links](https://example.com)

Best regards"

USAGE INSTRUCTIONS:
==================

1. Setup Requirements:
   - Python 3.6+
   - Required packages: pip install markdown imaplib2
   - Configure .env file with email credentials

2. Required Environment Variables (.env file):
   EMAIL_PROVIDER=gmail           # Email provider (gmail, zoho, fastmail)
   EMAIL_ACCOUNT=your@email.com   # Your email account
   EMAIL_PASSWORD=your_password   # Your email password (use app password for Gmail)
   EMAIL_SMTP_HOST=smtp.gmail.com # SMTP server host
   EMAIL_SMTP_PORT=587           # SMTP server port
   EMAIL_FROM_NAME=Your Name     # Your display name
   EMAIL_FROM_ADDRESS=from@email  # From address (optional, defaults to EMAIL_ACCOUNT)

3. Basic Usage Examples:

   # Send a simple email (body text is interpreted as Markdown)
   python email_client.py send -t recipient@email.com -s "Subject" -b "Email body"

   # Send HTML email from markdown file
   python email_client.py send -t recipient@email.com -s "Subject" -f email.md
   
   # Note: All email bodies are automatically formatted as Markdown and converted to HTML

   # Read last 10 emails
   python email_client.py read -n 10

   # Read emails from specific sender
   python email_client.py read --from sender@email.com

   # Search emails by subject
   python email_client.py read --subject "Important"

   # Read unread emails only
   python email_client.py read --unread

4. Advanced Usage:

   # Send with attachments
   python email_client.py send -t recipient@email.com -s "Files" -b "See attached" -a file1.pdf -a file2.docx

   # Send with CC and BCC
   python email_client.py send -t to@email.com -s "Subject" -b "Body" -c cc@email.com -B bcc@email.com

   # Read emails from last 7 days
   python email_client.py read --days 7

   # Mark emails as read
   python email_client.py mark-read --from sender@email.com

   # Delete emails
   python email_client.py delete --subject "Spam" --confirm

5. Programmatic Usage:

   from email_client import EmailClient
   
   # Initialize client
   client = EmailClient()
   
   # Send email
   client.send(
       to_email="recipient@email.com",
       subject="Test",
       body="Hello World",
       attachments=["file.pdf"]
   )
   
   # Read emails
   emails = client.read(limit=10, unread_only=True)
   for email in emails:
       print(f"From: {email['from']}, Subject: {email['subject']}")
"""

import os
import sys
import argparse
import smtplib
import imaplib
import ssl
import email
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from email.mime.base import MIMEBase
from email import encoders
from email.header import decode_header
from datetime import datetime, timedelta
import json
import re
from pathlib import Path
from typing import List, Dict, Optional, Tuple
import base64
import urllib.parse
import webbrowser

# Try to import markdown library
try:
    import markdown
    MARKDOWN_AVAILABLE = True
except ImportError:
    MARKDOWN_AVAILABLE = False

# Provider-specific settings
PROVIDER_SETTINGS = {
    'gmail': {
        'smtp': {'host': 'smtp.gmail.com', 'port': 587, 'use_tls': True},
        'imap': {'host': 'imap.gmail.com', 'port': 993}
    },
    'zoho': {
        'smtp': {'host': 'smtp.zoho.com', 'port': 465, 'use_ssl': True},
        'imap': {'host': 'imap.zoho.com', 'port': 993}
    },
    'fastmail': {
        'smtp': {'host': 'smtp.fastmail.com', 'port': 465, 'use_ssl': True},
        'imap': {'host': 'imap.fastmail.com', 'port': 993}
    }
}


class EmailConfig:
    """Email configuration manager"""
    
    def __init__(self, from_email: Optional[str] = None):
        """Initialize email configuration from environment variables"""
        self.load_env_file()
        
        # Required variables
        self.provider = self._get_required('EMAIL_PROVIDER').lower()
        self.account = self._get_required('EMAIL_ACCOUNT')
        self.password = self._get_required('EMAIL_PASSWORD')
        self.from_name = self._get_required('EMAIL_FROM_NAME')
        
        # Get provider settings
        if self.provider not in PROVIDER_SETTINGS:
            raise ValueError(f"Unsupported email provider: {self.provider}")
        
        provider_config = PROVIDER_SETTINGS[self.provider]
        
        # SMTP settings
        self.smtp_host = os.getenv('EMAIL_SMTP_HOST', provider_config['smtp']['host'])
        self.smtp_port = int(os.getenv('EMAIL_SMTP_PORT', provider_config['smtp']['port']))
        self.smtp_use_tls = provider_config['smtp'].get('use_tls', False)
        self.smtp_use_ssl = provider_config['smtp'].get('use_ssl', False)
        
        # IMAP settings
        self.imap_host = os.getenv('EMAIL_IMAP_HOST', provider_config['imap']['host'])
        self.imap_port = int(os.getenv('EMAIL_IMAP_PORT', provider_config['imap']['port']))
        
        # From address
        self.from_address = from_email or os.getenv('EMAIL_FROM_ADDRESS', self.account)
    
    def load_env_file(self):
        """Load .env file if it exists"""
        env_file = Path(__file__).parent.parent / '.env'
        if env_file.exists():
            with open(env_file, 'r') as f:
                for line in f:
                    line = line.strip()
                    if line and not line.startswith('#') and '=' in line:
                        key, value = line.split('=', 1)
                        if key not in os.environ:
                            os.environ[key] = value.strip('"').strip("'")
    
    def _get_required(self, key: str) -> str:
        """Get required environment variable or raise error"""
        value = os.getenv(key)
        if not value:
            raise ValueError(f"{key} not found in .env file. Please set {key} in your .env file.")
        return value


class EmailSender:
    """Email sending functionality"""
    
    def __init__(self, config: EmailConfig):
        self.config = config
        self.log_dir = Path(__file__).parent.parent / 'mail' / 'sent_emails'
        self.log_dir.mkdir(parents=True, exist_ok=True)
    
    def send(self, to_email: str, subject: str, body: str,
             cc: Optional[List[str]] = None, bcc: Optional[List[str]] = None,
             attachments: Optional[List[str]] = None, use_html: bool = True,
             include_signature: bool = True, signature_text: Optional[str] = None) -> bool:
        """Send an email"""
        try:
            # Create message
            message = MIMEMultipart("alternative" if use_html else "mixed")
            message["Subject"] = subject
            message["From"] = f"{self.config.from_name} <{self.config.from_address}>"
            message["To"] = to_email
            message["Reply-To"] = self.config.from_address
            
            if cc:
                message["Cc"] = ", ".join(cc)
            
            # Prepare body
            if use_html:
                body_html = self._markdown_to_html(body)
                html_content = self._create_html_template(body_html, include_signature, signature_text)
                
                plain_body = body
                if include_signature and signature_text:
                    plain_body += "\n\n" + signature_text
                
                message.attach(MIMEText(plain_body, "plain"))
                message.attach(MIMEText(html_content, "html"))
            else:
                if include_signature and signature_text:
                    body += "\n\n" + signature_text
                message.attach(MIMEText(body, "plain"))
            
            # Add attachments
            if attachments:
                for file_path in attachments:
                    self._attach_file(message, file_path)
            
            # Prepare recipients
            recipients = [to_email]
            if cc:
                recipients.extend(cc)
            if bcc:
                recipients.extend(bcc)
            
            # Send email
            self._send_smtp(message, recipients)
            
            # Log sent email
            self._log_sent_email(to_email, subject, body, cc, bcc, attachments)
            
            return True
            
        except Exception as e:
            print(f"Error sending email: {str(e)}")
            return False
    
    def _send_smtp(self, message: MIMEMultipart, recipients: List[str]):
        """Send email via SMTP"""
        if self.config.smtp_use_tls:
            # Use STARTTLS (typically port 587)
            with smtplib.SMTP(self.config.smtp_host, self.config.smtp_port) as server:
                server.starttls(context=ssl.create_default_context())
                server.login(self.config.account, self.config.password)
                server.sendmail(self.config.account, recipients, message.as_string())
        else:
            # Use SMTP_SSL (typically port 465)
            with smtplib.SMTP_SSL(self.config.smtp_host, self.config.smtp_port, 
                                   context=ssl.create_default_context()) as server:
                server.login(self.config.account, self.config.password)
                server.sendmail(self.config.account, recipients, message.as_string())
    
    def _attach_file(self, message: MIMEMultipart, file_path: str):
        """Attach a file to the message"""
        if os.path.isfile(file_path):
            with open(file_path, "rb") as attachment:
                part = MIMEBase("application", "octet-stream")
                part.set_payload(attachment.read())
            
            encoders.encode_base64(part)
            part.add_header(
                "Content-Disposition",
                f"attachment; filename= {os.path.basename(file_path)}",
            )
            message.attach(part)
    
    def _markdown_to_html(self, text: str) -> str:
        """Convert markdown to HTML"""
        if MARKDOWN_AVAILABLE:
            return markdown.markdown(text, extensions=['extra', 'nl2br'])
        else:
            # Basic conversion without library
            html = text
            html = re.sub(r'^### (.+)$', r'<h3>\1</h3>', html, flags=re.MULTILINE)
            html = re.sub(r'^## (.+)$', r'<h2>\1</h2>', html, flags=re.MULTILINE)
            html = re.sub(r'^# (.+)$', r'<h1>\1</h1>', html, flags=re.MULTILINE)
            html = re.sub(r'\*\*(.+?)\*\*', r'<strong>\1</strong>', html)
            html = re.sub(r'\*(.+?)\*', r'<em>\1</em>', html)
            html = re.sub(r'\[([^\]]+)\]\(([^)]+)\)', r'<a href="\2">\1</a>', html)
            html = html.replace('\n\n', '</p><p>')
            html = '<p>' + html + '</p>'
            return html
    
    def _create_html_template(self, body_html: str, include_signature: bool, 
                             signature_text: Optional[str]) -> str:
        """Create HTML email template"""
        signature_html = ""
        if include_signature and signature_text:
            signature_html = self._markdown_to_html(signature_text)
        
        return f"""<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {{
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }}
        h1, h2, h3 {{ color: #2c3e50; margin-top: 20px; margin-bottom: 10px; }}
        p {{ margin: 10px 0; }}
        a {{ color: #3498db; text-decoration: none; }}
        a:hover {{ text-decoration: underline; }}
        .signature {{ margin-top: 30px; padding-top: 20px; border-top: 1px solid #e0e0e0; color: #666; }}
    </style>
</head>
<body>
    <div class="content">{body_html}</div>
    {f'<div class="signature">{signature_html}</div>' if signature_html else ''}
</body>
</html>"""
    
    def _log_sent_email(self, to_email: str, subject: str, body: str,
                       cc: Optional[List[str]], bcc: Optional[List[str]], 
                       attachments: Optional[List[str]]):
        """Log sent email to JSON file"""
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        log_file = self.log_dir / f'email_{timestamp}.json'
        
        log_data = {
            'timestamp': datetime.now().isoformat(),
            'from': self.config.from_address,
            'to': to_email,
            'cc': cc or [],
            'bcc': bcc or [],
            'subject': subject,
            'body': body,
            'attachments': attachments or []
        }
        
        with open(log_file, 'w') as f:
            json.dump(log_data, f, indent=2)


class EmailReader:
    """Email reading functionality"""
    
    def __init__(self, config: EmailConfig):
        self.config = config
        self.imap = None
    
    def connect(self):
        """Connect to IMAP server"""
        self.imap = imaplib.IMAP4_SSL(self.config.imap_host, self.config.imap_port)
        self.imap.login(self.config.account, self.config.password)
        self.imap.select('INBOX')
    
    def disconnect(self):
        """Disconnect from IMAP server"""
        if self.imap:
            self.imap.close()
            self.imap.logout()
    
    def read(self, limit: int = 10, from_address: Optional[str] = None,
             subject_filter: Optional[str] = None, unread_only: bool = False,
             days: Optional[int] = None, include_html: bool = False) -> List[Dict]:
        """Read emails with various filters"""
        if not self.imap:
            self.connect()
        
        # Build search criteria
        criteria = []
        if from_address:
            criteria.append(f'FROM "{from_address}"')
        if subject_filter:
            criteria.append(f'SUBJECT "{subject_filter}"')
        if unread_only:
            criteria.append('UNSEEN')
        if days:
            date = (datetime.now() - timedelta(days=days)).strftime("%d-%b-%Y")
            criteria.append(f'SINCE {date}')
        
        search_string = ' '.join(criteria) if criteria else 'ALL'
        
        # Search emails
        _, message_numbers = self.imap.search(None, search_string)
        
        emails = []
        for num in reversed(message_numbers[0].split()[-limit:]):
            emails.append(self._fetch_email(num, include_html=include_html))
        
        return emails
    
    def _fetch_email(self, num: bytes, include_html: bool = False) -> Dict:
        """Fetch a single email"""
        # For Gmail, we need X-GM-MSGID for direct links
        if self.config.provider == 'gmail':
            # Fetch with Gmail extensions for message ID
            _, msg_data = self.imap.fetch(num, '(RFC822 X-GM-MSGID)')
        else:
            _, msg_data = self.imap.fetch(num, '(RFC822)')
        
        for response_part in msg_data:
            if isinstance(response_part, tuple):
                msg = email.message_from_bytes(response_part[1])
                
                # Get message ID for web link
                message_id = msg.get('Message-ID', '').strip('<>')
                
                # For Gmail, try to extract GM-MSGID from response
                gmail_msg_id = None
                if self.config.provider == 'gmail':
                    # Extract X-GM-MSGID from IMAP response
                    response_str = str(msg_data)
                    import re
                    msgid_match = re.search(r'X-GM-MSGID\s+(\d+)', response_str)
                    if msgid_match:
                        gmail_msg_id = msgid_match.group(1)
                
                # Extract email details
                email_data = {
                    'id': num.decode(),
                    'from': self._decode_header(msg['From']),
                    'to': self._decode_header(msg['To']),
                    'subject': self._decode_header(msg['Subject']),
                    'date': msg['Date'],
                    'body': self._get_email_body(msg, prefer_html=include_html),
                    'attachments': self._get_attachments(msg),
                    'message_id': message_id,
                    'gmail_msg_id': gmail_msg_id,
                    'web_link': self._get_web_link(message_id, gmail_msg_id),
                    'native_link': self._get_native_mail_link(message_id),
                    'unsubscribe_links': self._get_unsubscribe_links(msg)
                }
                
                return email_data
    
    def _decode_header(self, header: str) -> str:
        """Decode email header"""
        if not header:
            return ""
        
        decoded_parts = decode_header(header)
        result = []
        
        for part, encoding in decoded_parts:
            if isinstance(part, bytes):
                result.append(part.decode(encoding or 'utf-8', errors='ignore'))
            else:
                result.append(part)
        
        return ' '.join(result)
    
    def _get_email_body(self, msg: email.message.Message, prefer_html: bool = False) -> str:
        """Extract email body"""
        plain_body = ""
        html_body = ""
        
        if msg.is_multipart():
            for part in msg.walk():
                content_type = part.get_content_type()
                content_disposition = str(part.get("Content-Disposition"))
                
                if "attachment" not in content_disposition:
                    if content_type == "text/plain":
                        plain_body = part.get_payload(decode=True).decode('utf-8', errors='ignore')
                    elif content_type == "text/html":
                        html_body = part.get_payload(decode=True).decode('utf-8', errors='ignore')
        else:
            body = msg.get_payload(decode=True).decode('utf-8', errors='ignore')
            if msg.get_content_type() == "text/html":
                html_body = body
            else:
                plain_body = body
        
        # Return HTML if requested and available, otherwise plain text
        if prefer_html and html_body:
            return html_body
        return plain_body if plain_body else html_body
    
    def _get_attachments(self, msg: email.message.Message) -> List[str]:
        """Get list of attachment filenames"""
        attachments = []
        
        for part in msg.walk():
            if part.get_content_disposition() == 'attachment':
                filename = part.get_filename()
                if filename:
                    attachments.append(self._decode_header(filename))
        
        return attachments
    
    def _get_web_link(self, message_id: str, gmail_msg_id: str = None) -> str:
        """Generate web link to view email based on provider"""
        if self.config.provider == 'gmail' and gmail_msg_id:
            # Convert Gmail message ID to hex for direct link
            try:
                hex_id = format(int(gmail_msg_id), 'x')
                return f"https://mail.google.com/mail/u/0/#inbox/{hex_id}"
            except (ValueError, TypeError):
                pass
        
        if not message_id:
            return ""
        
        # Clean message ID for use in URLs
        clean_id = message_id.replace('<', '').replace('>', '')
        
        if self.config.provider == 'gmail':
            # Fallback to search by RFC822 message ID
            # URL encode the message ID properly
            import urllib.parse
            encoded_id = urllib.parse.quote(clean_id, safe='')
            return f"https://mail.google.com/mail/u/0/#search/rfc822msgid%3A{encoded_id}"
        elif self.config.provider == 'zoho':
            # Zoho Mail web interface
            return f"https://mail.zoho.com/zm/#search/msgid/{clean_id}"
        elif self.config.provider == 'fastmail':
            # Fastmail web interface
            return f"https://app.fastmail.com/mail/search:msgid:{clean_id}"
        else:
            # Generic webmail search
            return f"https://mail.{self.config.provider}.com/search?q={clean_id}"
    
    def _get_native_mail_link(self, message_id: str) -> str:
        """Generate link to open email in native mail app (macOS Mail.app)"""
        if not message_id:
            return ""
        
        # Clean message ID
        clean_id = message_id.strip('<>')
        
        # macOS Mail.app uses the message: URL scheme
        # Format: message://<message-id>
        return f"message://{urllib.parse.quote(clean_id, safe='@')}"
    
    def _get_unsubscribe_links(self, msg: email.message.Message) -> List[str]:
        """Extract unsubscribe links from email headers and body"""
        unsubscribe_links = []
        
        # Check List-Unsubscribe header (RFC 2369)
        list_unsubscribe = msg.get('List-Unsubscribe', '')
        if list_unsubscribe:
            # Extract URLs from List-Unsubscribe header
            # Format can be: <http://example.com/unsubscribe>, <mailto:unsubscribe@example.com>
            url_matches = re.findall(r'<(https?://[^>]+)>', list_unsubscribe)
            unsubscribe_links.extend(url_matches)
            
            # Also extract mailto links
            mailto_matches = re.findall(r'<(mailto:[^>]+)>', list_unsubscribe)
            unsubscribe_links.extend(mailto_matches)
        
        # Check List-Unsubscribe-Post header (RFC 8058)
        list_unsubscribe_post = msg.get('List-Unsubscribe-Post', '')
        
        # Search email body for unsubscribe links (check both plain and HTML)
        for prefer_html in [True, False]:
            body = self._get_email_body(msg, prefer_html=prefer_html)
            if body:
                # Look for common unsubscribe patterns in URLs
                unsubscribe_patterns = [
                    r'https?://[^\s<>"\']+unsubscribe[^\s<>"\']*',
                    r'https?://[^\s<>"\']+/unsub[^\s<>"\']*',
                    r'https?://[^\s<>"\']+/opt-out[^\s<>"\']*',
                    r'https?://[^\s<>"\']+/preferences[^\s<>"\']*',
                    r'https?://[^\s<>"\']+/email-preferences[^\s<>"\']*',
                    r'https?://[^\s<>"\']+/manage[^\s<>"\']*subscription[^\s<>"\']*',
                    r'https?://[^\s<>"\']+/email/preferences[^\s<>"\']*',
                    r'https?://[^\s<>"\']+/settings/notifications[^\s<>"\']*',
                    # Beehiiv specific patterns
                    r'https?://[^\s<>"\']*beehiiv\.com/[^\s<>"\']*unsubscribe[^\s<>"\']*',
                    r'https?://[^\s<>"\']*beehiiv\.com/[^\s<>"\']*preferences[^\s<>"\']*',
                    r'https?://[^\s<>"\']*beehiiv\.com/[^\s<>"\']*manage[^\s<>"\']*'
                ]
                
                # Also look for links with "unsubscribe" text nearby in HTML
                if prefer_html:
                    # Find all links with unsubscribe-related text
                    link_pattern = r'<a[^>]+href=["\']([^"\']+)["\'][^>]*>([^<]*)</a>'
                    for match in re.finditer(link_pattern, body, re.IGNORECASE):
                        url = match.group(1)
                        link_text = match.group(2)
                        # Check if link text contains unsubscribe-related words
                        if re.search(r'unsubscribe|opt.?out|preferences|manage|settings|stop.?receiv', link_text, re.IGNORECASE):
                            if url.startswith('http') and url not in unsubscribe_links:
                                unsubscribe_links.append(url)
                    
                    # Also look for links that appear right before/after "unsubscribe" text
                    # Pattern: text...unsubscribe <a href="...">
                    unsub_before_link = re.findall(r'unsubscribe[^<]*<a[^>]+href=["\']([^"\']+)["\']', body, re.IGNORECASE)
                    for url in unsub_before_link:
                        if url.startswith('http') and url not in unsubscribe_links:
                            unsubscribe_links.append(url)
                    
                    # Pattern: <a href="...">...unsubscribe
                    unsub_in_link = re.findall(r'<a[^>]+href=["\']([^"\']+)["\'][^>]*>[^<]*unsubscribe', body, re.IGNORECASE)
                    for url in unsub_in_link:
                        if url.startswith('http') and url not in unsubscribe_links:
                            unsubscribe_links.append(url)
                
                for pattern in unsubscribe_patterns:
                    matches = re.findall(pattern, body, re.IGNORECASE)
                    for match in matches:
                        # Clean up the URL
                        clean_url = match.strip('.,;:)]}\'"&amp;')
                        # Decode HTML entities
                        clean_url = clean_url.replace('&amp;', '&')
                        if clean_url not in unsubscribe_links and clean_url.startswith('http'):
                            unsubscribe_links.append(clean_url)
        
        # Remove duplicates while preserving order
        seen = set()
        unique_links = []
        for link in unsubscribe_links:
            if link not in seen:
                seen.add(link)
                unique_links.append(link)
        
        return unique_links
    
    def mark_as_read(self, email_ids: List[str]) -> bool:
        """Mark emails as read"""
        if not self.imap:
            self.connect()
        
        try:
            for email_id in email_ids:
                self.imap.store(email_id.encode(), '+FLAGS', '\\Seen')
            return True
        except Exception as e:
            print(f"Error marking emails as read: {str(e)}")
            return False
    
    def delete(self, email_ids: List[str]) -> bool:
        """Delete emails"""
        if not self.imap:
            self.connect()
        
        try:
            for email_id in email_ids:
                self.imap.store(email_id.encode(), '+FLAGS', '\\Deleted')
            self.imap.expunge()
            return True
        except Exception as e:
            print(f"Error deleting emails: {str(e)}")
            return False


class EmailClient:
    """Main email client combining sender and reader"""
    
    def __init__(self, from_email: Optional[str] = None):
        self.config = EmailConfig(from_email)
        self.sender = EmailSender(self.config)
        self.reader = EmailReader(self.config)
    
    def send(self, to_email: str, subject: str, body: str, **kwargs) -> bool:
        """Send an email"""
        return self.sender.send(to_email, subject, body, **kwargs)
    
    def read(self, **kwargs) -> List[Dict]:
        """Read emails"""
        return self.reader.read(**kwargs)
    
    def mark_as_read(self, email_ids: List[str]) -> bool:
        """Mark emails as read"""
        try:
            return self.reader.mark_as_read(email_ids)
        finally:
            self.reader.disconnect()
    
    def delete(self, email_ids: List[str]) -> bool:
        """Delete emails"""
        # Don't disconnect in finally block as delete needs active connection
        return self.reader.delete(email_ids)


def main():
    """Command line interface"""
    parser = argparse.ArgumentParser(
        description='Email Client - Send and read emails',
        formatter_class=argparse.RawDescriptionHelpFormatter
    )
    
    subparsers = parser.add_subparsers(dest='command', help='Commands')
    
    # Send command
    send_parser = subparsers.add_parser('send', help='Send an email')
    send_parser.add_argument('-t', '--to', required=True, help='Recipient email')
    send_parser.add_argument('-s', '--subject', required=True, help='Email subject')
    send_parser.add_argument('-b', '--body', help='Email body (Markdown formatted)')
    send_parser.add_argument('-f', '--file', help='Read body from file (Markdown formatted)')
    send_parser.add_argument('-c', '--cc', action='append', help='CC recipients')
    send_parser.add_argument('-B', '--bcc', action='append', help='BCC recipients')
    send_parser.add_argument('-a', '--attach', action='append', help='Attachments')
    send_parser.add_argument('-P', '--plain', action='store_true', help='Send as plain text')
    send_parser.add_argument('-S', '--no-signature', action='store_true', help='No signature')
    send_parser.add_argument('--signature', help='Custom signature')
    send_parser.add_argument('--from-email', help='Override from email')
    
    # Read command
    read_parser = subparsers.add_parser('read', help='Read emails')
    read_parser.add_argument('-n', '--number', type=int, default=10, help='Number of emails')
    read_parser.add_argument('--from', dest='from_address', help='Filter by sender')
    read_parser.add_argument('--subject', help='Filter by subject')
    read_parser.add_argument('--unread', action='store_true', help='Unread emails only')
    read_parser.add_argument('--days', type=int, help='Emails from last N days')
    read_parser.add_argument('--json', action='store_true', help='Output as JSON')
    read_parser.add_argument('--save-markdown', action='store_true', default=True, help='Save emails as markdown files (default: True)')
    read_parser.add_argument('--no-save-markdown', dest='save_markdown', action='store_false', help='Disable saving emails as markdown files')
    read_parser.add_argument('--output-dir', default='emails', help='Directory to save markdown files (default: emails)')
    
    # Mark as read command
    mark_parser = subparsers.add_parser('mark-read', help='Mark emails as read')
    mark_parser.add_argument('--ids', nargs='+', help='Email IDs to mark')
    mark_parser.add_argument('--from', dest='from_address', help='Mark all from sender')
    mark_parser.add_argument('--subject', help='Mark all with subject')
    
    # Delete command
    delete_parser = subparsers.add_parser('delete', help='Delete emails')
    delete_parser.add_argument('--ids', nargs='+', help='Email IDs to delete')
    delete_parser.add_argument('--from', dest='from_address', help='Delete all from sender')
    delete_parser.add_argument('--subject', help='Delete all with subject')
    delete_parser.add_argument('--confirm', action='store_true', help='Confirm deletion')
    
    # Unsubscribe command
    unsubscribe_parser = subparsers.add_parser('unsubscribe', help='Find unsubscribe links and optionally open in browser')
    unsubscribe_parser.add_argument('--from', dest='from_address', help='Find unsubscribe links from sender')
    unsubscribe_parser.add_argument('--subject', help='Find unsubscribe links by subject')
    unsubscribe_parser.add_argument('-n', '--number', type=int, default=10, help='Number of emails to check')
    unsubscribe_parser.add_argument('--open', action='store_true', help='Open the first unsubscribe link in browser')
    unsubscribe_parser.add_argument('--auto-open', action='store_true', help='Automatically open unsubscribe link without prompting')
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        return 1
    
    try:
        # Initialize client
        from_email = getattr(args, 'from_email', None)
        client = EmailClient(from_email=from_email)
        
        if args.command == 'send':
            # Get body
            if args.file:
                with open(args.file, 'r') as f:
                    body = f.read()
            else:
                body = args.body or ""
            
            # Get signature
            signature = ""
            if not args.no_signature:
                if args.signature:
                    signature = args.signature
                else:
                    signature = f"\n--\n{client.config.from_name}\n{client.config.from_address}"
            
            # Send email
            success = client.send(
                to_email=args.to,
                subject=args.subject,
                body=body,
                cc=args.cc,
                bcc=args.bcc,
                attachments=args.attach,
                use_html=not args.plain,
                include_signature=not args.no_signature,
                signature_text=signature
            )
            
            if success:
                print("Email sent successfully!")
            else:
                return 1
        
        elif args.command == 'read':
            emails = client.read(
                limit=args.number,
                from_address=args.from_address,
                subject_filter=args.subject,
                unread_only=args.unread,
                days=args.days
            )
            
            if args.save_markdown:
                # Save emails as markdown files
                output_dir = Path(args.output_dir)
                output_dir.mkdir(parents=True, exist_ok=True)
                
                for email_msg in emails:
                    # Create filename from subject and date
                    subject_clean = re.sub(r'[^\w\s-]', '', email_msg['subject']).strip()
                    subject_clean = re.sub(r'[-\s]+', '-', subject_clean)[:50]  # Limit length
                    
                    # Parse date to create formatted filename
                    try:
                        # Parse various date formats
                        date_str = email_msg['date']
                        # Remove timezone info in parentheses
                        date_str = re.sub(r'\s*\([^)]*\)\s*$', '', date_str)
                        # Try parsing
                        for fmt in ['%a, %d %b %Y %H:%M:%S %z', '%a, %d %b %Y %H:%M:%S %Z', 
                                   '%d %b %Y %H:%M:%S %z', '%a, %d %b %Y %H:%M:%S']:
                            try:
                                dt = datetime.strptime(date_str, fmt)
                                break
                            except ValueError:
                                continue
                        else:
                            # Fallback to current time if parsing fails
                            dt = datetime.now()
                        
                        date_formatted = dt.strftime('%Y-%m-%d')
                    except:
                        date_formatted = datetime.now().strftime('%Y-%m-%d')
                    
                    filename = f"{subject_clean}-{date_formatted}.md"
                    filepath = output_dir / filename
                    
                    # Create markdown content
                    markdown_content = f"""# {email_msg['subject']}

**From:** {email_msg['from']}  
**To:** {email_msg['to']}  
**Date:** {email_msg['date']}  
**ID:** {email_msg['id']}  
"""
                    
                    if email_msg.get('web_link'):
                        markdown_content += f"**View in browser:** [{email_msg['web_link']}]({email_msg['web_link']})  \n"
                    
                    if email_msg.get('native_link'):
                        markdown_content += f"**Open in Mail app:** [{email_msg['native_link']}]({email_msg['native_link']})  \n"
                    
                    if email_msg.get('attachments'):
                        markdown_content += f"**Attachments:** {', '.join(email_msg['attachments'])}  \n"
                    
                    markdown_content += f"\n---\n\n{email_msg['body']}\n"
                    
                    # Save file
                    with open(filepath, 'w', encoding='utf-8') as f:
                        f.write(markdown_content)
                    
                    print(f"Saved: {filepath}")
                
                print(f"\nSaved {len(emails)} emails to {output_dir}")
                
                # Disconnect after read
                client.reader.disconnect()
            
            elif args.json:
                # Output as JSON
                print(json.dumps(emails, indent=2))
                # Disconnect after read
                client.reader.disconnect()
            else:
                # Human-readable output
                for email_msg in emails:
                    print(f"\n{'='*60}")
                    print(f"ID: {email_msg['id']}")
                    print(f"From: {email_msg['from']}")
                    print(f"Subject: {email_msg['subject']}")
                    print(f"Date: {email_msg['date']}")
                    if email_msg.get('web_link'):
                        print(f"View in browser: {email_msg['web_link']}")
                    if email_msg.get('native_link'):
                        print(f"Open in Mail app: {email_msg['native_link']}")
                    if email_msg['attachments']:
                        print(f"Attachments: {', '.join(email_msg['attachments'])}")
                    print(f"\n{email_msg['body'][:500]}...")
                # Disconnect after read
                client.reader.disconnect()
        
        elif args.command == 'mark-read':
            if args.ids:
                success = client.mark_as_read(args.ids)
            else:
                # Find emails to mark
                emails = client.read(
                    from_address=args.from_address,
                    subject_filter=args.subject
                )
                if emails:
                    ids = [e['id'] for e in emails]
                    success = client.mark_as_read(ids)
                    print(f"Marked {len(ids)} emails as read")
                else:
                    print("No emails found matching criteria")
                    client.reader.disconnect()
        
        elif args.command == 'delete':
            if not args.confirm:
                print("Please use --confirm flag to delete emails")
                return 1
            
            if args.ids:
                success = client.delete(args.ids)
                if success:
                    print(f"Deleted {len(args.ids)} emails")
                client.reader.disconnect()
            else:
                # Find emails to delete
                emails = client.read(
                    from_address=args.from_address,
                    subject_filter=args.subject
                )
                if emails:
                    ids = [e['id'] for e in emails]
                    print(f"Found {len(ids)} emails to delete")
                    success = client.delete(ids)
                    if success:
                        print(f"Deleted {len(ids)} emails")
                    # Disconnect after delete
                    client.reader.disconnect()
                else:
                    print("No emails found matching criteria")
        
        elif args.command == 'unsubscribe':
            # Find unsubscribe links (fetch HTML version for better detection)
            emails = client.read(
                limit=args.number,
                from_address=args.from_address,
                subject_filter=args.subject,
                include_html=True
            )
            
            found_links = {}
            all_http_links = []  # Collect all HTTP links for opening
            for email_msg in emails:
                if email_msg.get('unsubscribe_links'):
                    sender = email_msg['from']
                    if sender not in found_links:
                        found_links[sender] = {
                            'links': email_msg['unsubscribe_links'],
                            'subject': email_msg['subject'],
                            'date': email_msg['date']
                        }
                        # Collect HTTP links (not mailto)
                        for link in email_msg['unsubscribe_links']:
                            if link.startswith('http'):
                                all_http_links.append(link)
            
            if found_links:
                print("Unsubscribe links found:\n")
                for sender, info in found_links.items():
                    print(f"From: {sender}")
                    print(f"Latest email: {info['subject']} ({info['date']})")
                    print("Unsubscribe links:")
                    for link in info['links']:
                        print(f"  - {link}")
                    print()
                
                # Open in browser if requested
                if (args.open or args.auto_open) and all_http_links:
                    # Prefer the first HTTP link (skip mailto links)
                    link_to_open = all_http_links[0]
                    
                    if args.auto_open:
                        print(f"\nOpening unsubscribe link in browser: {link_to_open}")
                        webbrowser.open(link_to_open)
                    elif args.open:
                        print(f"\nReady to open: {link_to_open}")
                        response = input("Open this link in your browser? (y/n): ").strip().lower()
                        if response == 'y':
                            webbrowser.open(link_to_open)
                            print("Opened in browser")
                        else:
                            print("Skipped opening link")
                elif args.open or args.auto_open:
                    print("\nNo HTTP unsubscribe links found to open")
            else:
                print("No unsubscribe links found in the specified emails")
                print("\nDEBUG: Checking first email for any links...")
                if emails:
                    first_email = emails[0]
                    print(f"Email from: {first_email['from']}")
                    print(f"Subject: {first_email['subject']}")
                    body = first_email.get('body', '')
                    print(f"Body type: {'HTML' if '<html' in body.lower() else 'Plain text'}")
                    print(f"Body length: {len(body)} characters")
                    
                    import re
                    # Look for any HTTP/HTTPS links
                    all_links = re.findall(r'https?://[^\s<>"\']+', body, re.IGNORECASE)
                    if all_links:
                        print(f"Found {len(all_links)} total links")
                        print("First 5 links:")
                        for link in all_links[:5]:
                            print(f"  - {link[:100]}...")
                    
                    # Look for unsubscribe text
                    if 'unsubscribe' in body.lower():
                        print("\n'unsubscribe' text found in body")
                        # Find context around unsubscribe
                        idx = body.lower().index('unsubscribe')
                        context = body[max(0, idx-100):min(len(body), idx+200)]
                        print(f"Context: ...{context}...")
                        
                        # Try to extract the unsubscribe link from context
                        unsub_links = re.findall(r'href=["\']([^"\']+)["\'][^>]*>[^<]*unsubscribe', context, re.IGNORECASE)
                        if unsub_links:
                            print(f"\nFound unsubscribe link in context: {unsub_links[0]}")
            
            # Disconnect after read
            client.reader.disconnect()
        
        return 0
        
    except Exception as e:
        print(f"Error: {str(e)}")
        return 1


if __name__ == '__main__':
    sys.exit(main())