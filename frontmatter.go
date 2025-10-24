package mailos

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type EmailFrontmatter struct {
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Priority    string
	SendAfter   *time.Time
	InReplyTo   string
	References  []string
	Attachments []string
	UseTemplate bool
	PlainText   bool
	NoSignature bool
}

func ParseFrontmatter(content string) (*EmailFrontmatter, string, error) {
	frontmatterRegex := regexp.MustCompile(`^---\s*\n(.*?)\n---\s*\n(.*)$`)
	matches := frontmatterRegex.FindStringSubmatch(strings.TrimSpace(content))
	
	if len(matches) != 3 {
		return nil, content, nil
	}
	
	frontmatterText := matches[1]
	bodyContent := matches[2]
	
	fm := &EmailFrontmatter{}
	lines := strings.Split(frontmatterText, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch strings.ToLower(key) {
		case "to":
			fm.To = parseEmailList(value)
		case "cc":
			fm.CC = parseEmailList(value)
		case "bcc":
			fm.BCC = parseEmailList(value)
		case "subject":
			fm.Subject = unquote(value)
		case "priority":
			fm.Priority = strings.ToLower(unquote(value))
		case "send_after", "sendafter":
			if t, err := parseDateTime(value); err == nil {
				fm.SendAfter = &t
			}
		case "in_reply_to", "inreplyto":
			fm.InReplyTo = unquote(value)
		case "references":
			fm.References = parseStringList(value)
		case "attachments":
			fm.Attachments = parseStringList(value)
		case "use_template", "usetemplate":
			fm.UseTemplate = parseBool(value)
		case "plain_text", "plaintext":
			fm.PlainText = parseBool(value)
		case "no_signature", "nosignature":
			fm.NoSignature = parseBool(value)
		}
	}
	
	return fm, bodyContent, nil
}

func parseEmailList(value string) []string {
	value = unquote(value)
	if value == "" {
		return nil
	}
	
	var emails []string
	parts := strings.Split(value, ",")
	for _, part := range parts {
		email := strings.TrimSpace(part)
		if email != "" {
			emails = append(emails, email)
		}
	}
	return emails
}

func parseStringList(value string) []string {
	value = unquote(value)
	if value == "" {
		return nil
	}
	
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		value = strings.Trim(value, "[]")
		var items []string
		parts := strings.Split(value, ",")
		for _, part := range parts {
			item := strings.TrimSpace(unquote(part))
			if item != "" {
				items = append(items, item)
			}
		}
		return items
	}
	
	return parseEmailList(value)
}

func parseDateTime(value string) (time.Time, error) {
	value = unquote(value)
	
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"15:04:05",
		"15:04",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			if format == "15:04:05" || format == "15:04" {
				now := time.Now()
				t = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, now.Location())
			}
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse time: %s", value)
}

func parseBool(value string) bool {
	value = strings.ToLower(unquote(value))
	return value == "true" || value == "yes" || value == "1" || value == "on"
}

func unquote(value string) string {
	value = strings.TrimSpace(value)
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return value[1 : len(value)-1]
	}
	return value
}

func (fm *EmailFrontmatter) ToEmailMessage(bodyContent string) *EmailMessage {
	msg := &EmailMessage{
		To:               fm.To,
		CC:               fm.CC,
		BCC:              fm.BCC,
		Subject:          fm.Subject,
		Body:             bodyContent,
		Attachments:      fm.Attachments,
		IncludeSignature: !fm.NoSignature,
		UseTemplate:      fm.UseTemplate,
		InReplyTo:        fm.InReplyTo,
		References:       fm.References,
	}
	
	if !fm.UseTemplate && !fm.PlainText {
		msg.BodyHTML = MarkdownToHTMLContent(bodyContent)
	} else if fm.PlainText {
		msg.BodyHTML = ""
	}
	
	return msg
}

func (fm *EmailFrontmatter) ToDraftEmail(bodyContent string) DraftEmail {
	draft := DraftEmail{
		To:          fm.To,
		CC:          fm.CC,
		BCC:         fm.BCC,
		Subject:     fm.Subject,
		Body:        bodyContent,
		Attachments: fm.Attachments,
		Priority:    fm.Priority,
		InReplyTo:   fm.InReplyTo,
		References:  fm.References,
		SendAfter:   fm.SendAfter,
	}
	
	return draft
}