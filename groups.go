package mailos

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type EmailGroup struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Emails      []string `json:"emails"`
}

type GroupConfig struct {
	Groups []EmailGroup `json:"groups"`
}

func GetGroupsConfigPath() (string, error) {
	emailDir, err := GetEmailStorageDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(emailDir, "groups.json"), nil
}

func LoadGroupsConfig() (*GroupConfig, error) {
	configPath, err := GetGroupsConfigPath()
	if err != nil {
		return nil, err
	}

	config := &GroupConfig{Groups: []EmailGroup{}}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read groups config: %v", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse groups config: %v", err)
	}

	return config, nil
}

func SaveGroupsConfig(config *GroupConfig) error {
	configPath, err := GetGroupsConfigPath()
	if err != nil {
		return err
	}

	if err := EnsureEmailDirectories(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal groups config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save groups config: %v", err)
	}

	return nil
}

func GetGroup(name string) (*EmailGroup, error) {
	config, err := LoadGroupsConfig()
	if err != nil {
		return nil, err
	}

	for _, group := range config.Groups {
		if strings.EqualFold(group.Name, name) {
			return &group, nil
		}
	}

	return nil, fmt.Errorf("group '%s' not found", name)
}

func ListGroups() error {
	config, err := LoadGroupsConfig()
	if err != nil {
		return err
	}

	if len(config.Groups) == 0 {
		fmt.Println("No groups configured.")
		fmt.Println("Use 'mailos groups --update <group_name> --emails <email1,email2>' to create a group.")
		return nil
	}

	fmt.Println("Email Groups:")
	fmt.Println("=============")
	
	sort.Slice(config.Groups, func(i, j int) bool {
		return config.Groups[i].Name < config.Groups[j].Name
	})

	for _, group := range config.Groups {
		fmt.Printf("\n%s (%d emails)\n", group.Name, len(group.Emails))
		if group.Description != "" {
			fmt.Printf("  Description: %s\n", group.Description)
		}
		fmt.Printf("  Emails: %s\n", strings.Join(group.Emails, ", "))
	}

	return nil
}

func UpdateGroup(name, description, emailsStr string) error {
	if name == "" {
		return fmt.Errorf("group name is required")
	}

	if emailsStr == "" {
		return fmt.Errorf("emails are required")
	}

	emails := parseGroupEmailList(emailsStr)
	if len(emails) == 0 {
		return fmt.Errorf("no valid emails provided")
	}

	config, err := LoadGroupsConfig()
	if err != nil {
		return err
	}

	groupIndex := -1
	for i, group := range config.Groups {
		if strings.EqualFold(group.Name, name) {
			groupIndex = i
			break
		}
	}

	newGroup := EmailGroup{
		Name:        name,
		Description: description,
		Emails:      emails,
	}

	if groupIndex >= 0 {
		config.Groups[groupIndex] = newGroup
		fmt.Printf("Updated group '%s' with %d emails\n", name, len(emails))
	} else {
		config.Groups = append(config.Groups, newGroup)
		fmt.Printf("Created group '%s' with %d emails\n", name, len(emails))
	}

	if err := SaveGroupsConfig(config); err != nil {
		return err
	}

	return nil
}

func DeleteGroup(name string) error {
	if name == "" {
		return fmt.Errorf("group name is required")
	}

	config, err := LoadGroupsConfig()
	if err != nil {
		return err
	}

	groupIndex := -1
	for i, group := range config.Groups {
		if strings.EqualFold(group.Name, name) {
			groupIndex = i
			break
		}
	}

	if groupIndex == -1 {
		return fmt.Errorf("group '%s' not found", name)
	}

	config.Groups = append(config.Groups[:groupIndex], config.Groups[groupIndex+1:]...)

	if err := SaveGroupsConfig(config); err != nil {
		return err
	}

	fmt.Printf("Deleted group '%s'\n", name)
	return nil
}

func parseGroupEmailList(emailsStr string) []string {
	emails := strings.Split(emailsStr, ",")
	var validEmails []string

	for _, email := range emails {
		email = strings.TrimSpace(email)
		if email != "" && strings.Contains(email, "@") {
			validEmails = append(validEmails, email)
		}
	}

	return validEmails
}

func GetGroupEmails(groupName string) ([]string, error) {
	group, err := GetGroup(groupName)
	if err != nil {
		return nil, err
	}
	return group.Emails, nil
}

func AddMemberToGroup(groupName, email string) error {
	if groupName == "" {
		return fmt.Errorf("group name is required")
	}
	if email == "" || !strings.Contains(email, "@") {
		return fmt.Errorf("valid email address is required")
	}

	config, err := LoadGroupsConfig()
	if err != nil {
		return err
	}

	groupIndex := -1
	for i, group := range config.Groups {
		if strings.EqualFold(group.Name, groupName) {
			groupIndex = i
			break
		}
	}

	if groupIndex == -1 {
		return fmt.Errorf("group '%s' not found", groupName)
	}

	for _, existingEmail := range config.Groups[groupIndex].Emails {
		if strings.EqualFold(existingEmail, email) {
			return fmt.Errorf("email '%s' already exists in group '%s'", email, groupName)
		}
	}

	config.Groups[groupIndex].Emails = append(config.Groups[groupIndex].Emails, email)

	if err := SaveGroupsConfig(config); err != nil {
		return err
	}

	fmt.Printf("Added '%s' to group '%s' (%d total emails)\n", email, groupName, len(config.Groups[groupIndex].Emails))
	return nil
}

func RemoveMemberFromGroup(groupName, email string) error {
	if groupName == "" {
		return fmt.Errorf("group name is required")
	}
	if email == "" {
		return fmt.Errorf("email address is required")
	}

	config, err := LoadGroupsConfig()
	if err != nil {
		return err
	}

	groupIndex := -1
	for i, group := range config.Groups {
		if strings.EqualFold(group.Name, groupName) {
			groupIndex = i
			break
		}
	}

	if groupIndex == -1 {
		return fmt.Errorf("group '%s' not found", groupName)
	}

	emailIndex := -1
	for i, existingEmail := range config.Groups[groupIndex].Emails {
		if strings.EqualFold(existingEmail, email) {
			emailIndex = i
			break
		}
	}

	if emailIndex == -1 {
		return fmt.Errorf("email '%s' not found in group '%s'", email, groupName)
	}

	config.Groups[groupIndex].Emails = append(
		config.Groups[groupIndex].Emails[:emailIndex],
		config.Groups[groupIndex].Emails[emailIndex+1:]...,
	)

	if err := SaveGroupsConfig(config); err != nil {
		return err
	}

	fmt.Printf("Removed '%s' from group '%s' (%d remaining emails)\n", email, groupName, len(config.Groups[groupIndex].Emails))
	return nil
}

func GetGroupInfo(groupName string) (*EmailGroup, error) {
	group, err := GetGroup(groupName)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func ListGroupMembers(groupName string) error {
	group, err := GetGroup(groupName)
	if err != nil {
		return err
	}

	fmt.Printf("Group: %s\n", group.Name)
	if group.Description != "" {
		fmt.Printf("Description: %s\n", group.Description)
	}
	fmt.Printf("Members (%d):\n", len(group.Emails))
	for i, email := range group.Emails {
		fmt.Printf("  %d. %s\n", i+1, email)
	}
	return nil
}

func ProcessGroupsForSending(groupNames []string, existingEmails []string) ([]string, error) {
	if len(groupNames) == 0 {
		return existingEmails, nil
	}

	allEmails := make([]string, len(existingEmails))
	copy(allEmails, existingEmails)

	for _, groupName := range groupNames {
		groupEmails, err := GetGroupEmails(groupName)
		if err != nil {
			return nil, fmt.Errorf("failed to get emails for group '%s': %v", groupName, err)
		}
		allEmails = append(allEmails, groupEmails...)
	}

	return removeDuplicateEmailStrings(allEmails), nil
}

func removeDuplicateEmailStrings(emails []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, email := range emails {
		email = strings.TrimSpace(strings.ToLower(email))
		if email != "" && !seen[email] {
			seen[email] = true
			result = append(result, email)
		}
	}

	return result
}