package main

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestContributorListingsSorted(t *testing.T) {
	file, err := os.Open("../README.md")
	if err != nil {
		t.Fatalf("Failed to open README.md: %v", err)
	}
	defer file.Close()

	content := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	sections := []struct {
		name     string
		pattern  string
		emeritus bool
	}{
		{"Maintainers", `### Maintainers\n\n((?:- \[.*?\n)+)`, false},
		{"Approvers", `### Approvers\n\n((?:- \[.*?\n)+)`, false},
		{"Emeritus Maintainers", `### Emeritus Maintainers\n\n((?:- \[.*?\n)+)`, true},
		{"Emeritus Approvers", `### Emeritus Approvers\n\n((?:- \[.*?\n)+)`, true},
	}

	for _, section := range sections {
		t.Run(section.name, func(t *testing.T) {
			validateSection(t, content, section.name, section.pattern, section.emeritus)
		})
	}
}

func validateSection(t *testing.T, content, sectionName, pattern string, isEmeritus bool) {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(content)
	if match == nil {
		t.Fatalf("Section %s not found in README.md", sectionName)
	}

	sectionContent := strings.TrimSpace(match[1])
	lines := strings.Split(sectionContent, "\n")

	type contributor struct {
		firstName string
		line      string
	}

	var contributors []contributor
	nameRe := regexp.MustCompile(`-\s*\[([^\]]+)\]`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- [") {
			nameMatch := nameRe.FindStringSubmatch(line)
			if nameMatch != nil {
				fullName := nameMatch[1]
				// Handle special cases like "John L. Peterson (Jack)"
				if strings.Contains(fullName, "(") {
					fullName = strings.Split(fullName, "(")[0]
					fullName = strings.TrimSpace(fullName)
				}
				firstName := strings.Split(fullName, " ")[0]
				contributors = append(contributors, contributor{
					firstName: strings.ToLower(firstName),
					line:      line,
				})
			}
		}
	}

	if len(contributors) == 0 {
		t.Fatalf("No contributors found in %s section", sectionName)
	}

	// Check if sorted
	var currentOrder []string
	var expectedOrder []string
	for _, c := range contributors {
		currentOrder = append(currentOrder, c.firstName)
		expectedOrder = append(expectedOrder, c.firstName)
	}

	sort.Strings(expectedOrder)

	if !equalSlices(currentOrder, expectedOrder) {
		t.Errorf("%s section is not sorted alphabetically by first name", sectionName)
		t.Errorf("Current order: %v", currentOrder)
		t.Errorf("Expected order: %v", expectedOrder)
		
		// Show correct ordering
		sortedContributors := make([]contributor, len(contributors))
		copy(sortedContributors, contributors)
		sort.Slice(sortedContributors, func(i, j int) bool {
			return sortedContributors[i].firstName < sortedContributors[j].firstName
		})
		
		t.Errorf("Correct ordering should be:")
		for _, c := range sortedContributors {
			t.Errorf("  %s", c.line)
		}
	}

	// For emeritus sections, check no company affiliations
	if isEmeritus {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "- [") && strings.Contains(line, ", ") && !strings.HasSuffix(line, ")") {
				parts := strings.Split(line, ", ")
				if len(parts) > 1 {
					potentialCompany := strings.TrimSpace(parts[1])
					if potentialCompany != "" && !strings.HasPrefix(potentialCompany, "http") {
						t.Errorf("%s section contains company affiliation that should be removed: %s (company: %s)", 
							sectionName, line, potentialCompany)
					}
				}
			}
		}
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}