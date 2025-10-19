## Email System Testing Protocol

When testing email system functionality, ALWAYS use the actual mailos commands directly within the Go system. Do NOT use Python scripts or external tools to query email data. 

**Required approach:**
1. Test mailos commands first (e.g., `./mailos download --id 1234`, `./mailos read`, `./mailos search`)
2. Verify functionality works through the intended CLI interface
3. Only use alternative methods if the mailos commands fail
4. always use andrew@happysoft.dev as test sending email

This ensures the actual email system functionality is being tested, not just the data files.