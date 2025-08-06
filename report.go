package mailos

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type TimeRange struct {
	Name        string
	Description string
	Since       time.Time
	Until       time.Time
}

func GetTimeRanges() []TimeRange {
	now := time.Now()
	location := now.Location()
	
	// Get start of today (midnight)
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	
	// Get start of yesterday
	startOfYesterday := startOfToday.AddDate(0, 0, -1)
	endOfYesterday := startOfToday.Add(-time.Second)
	
	// Get start of this week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 7 in this calculation
	}
	startOfWeek := startOfToday.AddDate(0, 0, -(weekday - 1))
	
	// Get start of last week
	startOfLastWeek := startOfWeek.AddDate(0, 0, -7)
	endOfLastWeek := startOfWeek.Add(-time.Second)
	
	// This morning (6 AM to noon today)
	thisMorning := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, location)
	noon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, location)
	
	// Yesterday morning
	yesterdayMorning := thisMorning.AddDate(0, 0, -1)
	yesterdayNoon := noon.AddDate(0, 0, -1)
	
	return []TimeRange{
		{
			Name:        "Last hour",
			Description: fmt.Sprintf("Emails from %s", now.Add(-time.Hour).Format("3:04 PM")),
			Since:       now.Add(-time.Hour),
			Until:       now,
		},
		{
			Name:        "Today",
			Description: fmt.Sprintf("All emails since midnight (%s)", startOfToday.Format("Jan 2")),
			Since:       startOfToday,
			Until:       now,
		},
		{
			Name:        "Yesterday",
			Description: fmt.Sprintf("All emails from yesterday (%s)", startOfYesterday.Format("Jan 2")),
			Since:       startOfYesterday,
			Until:       endOfYesterday,
		},
		{
			Name:        "This morning",
			Description: fmt.Sprintf("6 AM - 12 PM today"),
			Since:       thisMorning,
			Until:       noon,
		},
		{
			Name:        "Yesterday morning",
			Description: fmt.Sprintf("6 AM - 12 PM yesterday"),
			Since:       yesterdayMorning,
			Until:       yesterdayNoon,
		},
		{
			Name:        "Last 3 days",
			Description: fmt.Sprintf("Since %s", now.AddDate(0, 0, -3).Format("Jan 2")),
			Since:       now.AddDate(0, 0, -3),
			Until:       now,
		},
		{
			Name:        "This week",
			Description: fmt.Sprintf("Since Monday (%s)", startOfWeek.Format("Jan 2")),
			Since:       startOfWeek,
			Until:       now,
		},
		{
			Name:        "Last week",
			Description: fmt.Sprintf("%s - %s", startOfLastWeek.Format("Jan 2"), endOfLastWeek.Format("Jan 2")),
			Since:       startOfLastWeek,
			Until:       endOfLastWeek,
		},
		{
			Name:        "Last 30 days",
			Description: "Past month",
			Since:       now.AddDate(0, 0, -30),
			Until:       now,
		},
	}
}

type timeRangeItem struct {
	timeRange TimeRange
}

func (i timeRangeItem) Title() string       { return i.timeRange.Name }
func (i timeRangeItem) Description() string { return i.timeRange.Description }
func (i timeRangeItem) FilterValue() string { return i.timeRange.Name }

type timeRangeModel struct {
	list     list.Model
	selected *TimeRange
	quitting bool
}

func (m timeRangeModel) Init() tea.Cmd {
	return nil
}

func (m timeRangeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(timeRangeItem); ok {
				m.selected = &i.timeRange
			}
			m.quitting = true
			return m, tea.Quit
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m timeRangeModel) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

func SelectTimeRange() (*TimeRange, error) {
	ranges := GetTimeRanges()
	items := make([]list.Item, len(ranges))
	for i, r := range ranges {
		items[i] = timeRangeItem{timeRange: r}
	}

	const defaultWidth = 60
	const defaultHeight = 15

	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, defaultHeight)
	l.Title = "Select time range for email report"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = l.Styles.Title.Bold(true).Foreground(list.DefaultStyles().Title.GetForeground())

	m := timeRangeModel{list: l}
	p := tea.NewProgram(m)
	
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run time range selector: %v", err)
	}

	if model, ok := finalModel.(timeRangeModel); ok && model.selected != nil {
		return model.selected, nil
	}

	return nil, fmt.Errorf("no time range selected")
}

func GenerateEmailReport(emails []*Email, timeRange TimeRange) string {
	var report strings.Builder
	
	report.WriteString("📧 Email Report\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n\n")
	
	report.WriteString(fmt.Sprintf("Time Range: %s\n", timeRange.Name))
	report.WriteString(fmt.Sprintf("Period: %s to %s\n", 
		timeRange.Since.Format("Jan 2, 2006 3:04 PM"),
		timeRange.Until.Format("Jan 2, 2006 3:04 PM")))
	report.WriteString(fmt.Sprintf("Total Emails: %d\n\n", len(emails)))
	
	if len(emails) == 0 {
		report.WriteString("No emails found in this time range.\n")
		return report.String()
	}
	
	// Group emails by sender
	senderCount := make(map[string]int)
	for _, email := range emails {
		senderCount[email.From]++
	}
	
	// Find top senders
	report.WriteString("📊 Top Senders:\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	count := 0
	for sender, num := range senderCount {
		if count >= 5 {
			break
		}
		report.WriteString(fmt.Sprintf("  • %s (%d emails)\n", sender, num))
		count++
	}
	report.WriteString("\n")
	
	// List all emails
	report.WriteString("📋 Email List:\n")
	report.WriteString("───────────────────────────────────────────────────────────────\n")
	
	for i, email := range emails {
		report.WriteString(fmt.Sprintf("\n%d. ", i+1))
		report.WriteString(fmt.Sprintf("From: %s\n", email.From))
		report.WriteString(fmt.Sprintf("   Subject: %s\n", email.Subject))
		report.WriteString(fmt.Sprintf("   Date: %s\n", email.Date.Format("Jan 2, 3:04 PM")))
		
		if len(email.Attachments) > 0 {
			report.WriteString(fmt.Sprintf("   📎 Attachments: %d\n", len(email.Attachments)))
		}
		
		// Add a preview of the body
		bodyPreview := email.Body
		if len(bodyPreview) > 100 {
			bodyPreview = bodyPreview[:100] + "..."
		}
		bodyPreview = strings.ReplaceAll(bodyPreview, "\n", " ")
		report.WriteString(fmt.Sprintf("   Preview: %s\n", bodyPreview))
	}
	
	report.WriteString("\n═══════════════════════════════════════════════════════════════\n")
	report.WriteString(fmt.Sprintf("Report generated at %s\n", time.Now().Format("Jan 2, 2006 3:04 PM")))
	
	return report.String()
}

// ParseTimeRangeString converts string names like "Last hour", "Today", etc. to TimeRange
func ParseTimeRangeString(rangeName string) (*TimeRange, error) {
	ranges := GetTimeRanges()
	
	// Normalize the input
	normalized := strings.ToLower(strings.TrimSpace(rangeName))
	
	for _, r := range ranges {
		if strings.ToLower(r.Name) == normalized {
			return &r, nil
		}
	}
	
	// Try partial matches
	for _, r := range ranges {
		if strings.Contains(strings.ToLower(r.Name), normalized) ||
		   strings.Contains(normalized, strings.ToLower(r.Name)) {
			return &r, nil
		}
	}
	
	return nil, fmt.Errorf("unknown time range: %s", rangeName)
}