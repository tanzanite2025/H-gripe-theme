package refundreturn

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	Group = "refund_return"
	Key   = "refund_return_policy"
)

var nonAnchorIDChars = regexp.MustCompile(`[^a-z0-9_-]+`)

type Image struct {
	URL     string `json:"url"`
	Alt     string `json:"alt"`
	Caption string `json:"caption,omitempty"`
}

type Section struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Body    string   `json:"body"`
	Bullets []string `json:"bullets,omitempty"`
	Image   *Image   `json:"image,omitempty"`
}

type Policy struct {
	Title        string    `json:"title"`
	Intro        string    `json:"intro"`
	Sections     []Section `json:"sections"`
	ContactLabel string    `json:"contact_label,omitempty"`
	ContactURL   string    `json:"contact_url,omitempty"`
	UpdatedAt    string    `json:"updated_at,omitempty"`
}

func DefaultPolicy() Policy {
	return Policy{
		Title: "Refund & Return Policy",
		Intro: "How we handle returns, refunds, and exchanges to keep your experience predictable and fair.",
		Sections: []Section{
			{
				ID:    "eligibility",
				Title: "Eligibility",
				Body:  "We accept returns within 30 days of delivery for unused items in original packaging. Custom or personalized items are non-refundable unless defective.",
			},
			{
				ID:    "condition",
				Title: "Condition of Items",
				Body:  "Returned items must be unused, undamaged, and include all accessories/manuals. We reserve the right to refuse returns that do not meet these conditions.",
			},
			{
				ID:    "special-orders",
				Title: "Special Orders",
				Body:  "Non-stock or custom-configured products (special orders) are not eligible for return or refund, unless the issue is caused by our error.",
			},
			{
				ID:    "process",
				Title: "Process",
				Bullets: []string{
					"Contact our support team with your order number and issue details.",
					"We will provide return instructions and, if applicable, a return authorization.",
					"Ship the item using a trackable method; retain proof of shipment.",
				},
			},
			{
				ID:    "refund-timing",
				Title: "Refund Method & Timing",
				Body:  "Refunds are processed to the original payment method within 5-10 business days after we receive and inspect the returned item.",
			},
			{
				ID:    "shipping-costs",
				Title: "Shipping Costs",
				Body:  "Return shipping is the customer's responsibility unless the item is defective or incorrect. Original shipping fees are non-refundable unless required by law.",
			},
			{
				ID:    "restocking-fee",
				Title: "Restocking & Refurbishment Fee",
				Body:  "A restocking and refurbishment fee may be charged at 20% of the original purchase value, with a minimum of USD $100.",
			},
			{
				ID:    "other-costs",
				Title: "Other Costs",
				Body:  "Unless otherwise specified under the warranty policy, shipping fees, duties, taxes, and any additional charges are borne by the customer.",
			},
			{
				ID:    "exchanges",
				Title: "Exchanges",
				Body:  "For exchanges, please initiate a return first, then place a new order once the return is approved. This ensures availability and faster processing.",
			},
		},
		ContactLabel: "For refund or return questions, contact our support team through the contact page.",
		ContactURL:   "/company/contact",
		UpdatedAt:    time.Date(2024, time.December, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
}

func Normalize(policy Policy) (Policy, error) {
	policy.Title = strings.TrimSpace(policy.Title)
	policy.Intro = strings.TrimSpace(policy.Intro)
	policy.ContactLabel = strings.TrimSpace(policy.ContactLabel)
	policy.ContactURL = strings.TrimSpace(policy.ContactURL)

	if policy.Title == "" {
		return Policy{}, errors.New("policy title is required")
	}
	if len(policy.Sections) > 50 {
		return Policy{}, errors.New("policy cannot contain more than 50 sections")
	}

	seenIDs := make(map[string]struct{}, len(policy.Sections))
	seenAnchorIDs := make(map[string]struct{}, len(policy.Sections))
	normalizedSections := make([]Section, 0, len(policy.Sections))
	for index, section := range policy.Sections {
		section.ID = strings.TrimSpace(section.ID)
		if section.ID == "" {
			section.ID = fmt.Sprintf("section-%d", index+1)
		}
		if _, exists := seenIDs[section.ID]; exists {
			return Policy{}, fmt.Errorf("section id %q is duplicated", section.ID)
		}
		seenIDs[section.ID] = struct{}{}
		anchorID := sectionAnchorID(section.ID, index)
		if _, exists := seenAnchorIDs[anchorID]; exists {
			return Policy{}, fmt.Errorf("section id %q produces a duplicated page anchor", section.ID)
		}
		seenAnchorIDs[anchorID] = struct{}{}

		section.Title = strings.TrimSpace(section.Title)
		section.Body = strings.TrimSpace(section.Body)
		section.Bullets = normalizeLines(section.Bullets)

		if section.Image != nil {
			image := &Image{
				URL:     strings.TrimSpace(section.Image.URL),
				Alt:     strings.TrimSpace(section.Image.Alt),
				Caption: strings.TrimSpace(section.Image.Caption),
			}
			if image.URL == "" {
				section.Image = nil
			} else {
				if err := validateMediaURL(image.URL); err != nil {
					return Policy{}, fmt.Errorf("section %q image: %w", section.ID, err)
				}
				section.Image = image
			}
		}

		if section.Title == "" {
			return Policy{}, fmt.Errorf("section %q title is required", section.ID)
		}
		if section.Body == "" && len(section.Bullets) == 0 && section.Image == nil {
			return Policy{}, fmt.Errorf("section %q must contain text, bullets, or an image", section.ID)
		}

		normalizedSections = append(normalizedSections, section)
	}
	policy.Sections = normalizedSections

	if policy.ContactURL != "" {
		if err := validateContactURL(policy.ContactURL); err != nil {
			return Policy{}, err
		}
	}

	return policy, nil
}

func normalizeAnchorID(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = nonAnchorIDChars.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	return normalized
}

func sectionAnchorID(value string, index int) string {
	normalized := normalizeAnchorID(value)
	if normalized == "" {
		normalized = "section-" + strconv.Itoa(index+1)
	}
	return "refund-return-" + normalized
}

func normalizeLines(lines []string) []string {
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			normalized = append(normalized, line)
		}
	}
	return normalized
}

func validateMediaURL(value string) error {
	if strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") {
		return nil
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("image URL must be an absolute HTTP(S) URL or a first-party relative path")
	}
	if parsed.User != nil {
		return errors.New("image URL must not contain credentials")
	}
	return nil
}

func validateContactURL(value string) error {
	if (strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//")) || strings.HasPrefix(value, "mailto:") {
		return nil
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("contact URL must be a relative path, mailto URL, or absolute HTTP(S) URL")
	}
	if parsed.User != nil {
		return errors.New("contact URL must not contain credentials")
	}
	return nil
}
