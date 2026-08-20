package service

import (
	"fmt"
	"strings"

	"commerce-platform/internal/domain/user"
)

func afterSalesOperatorName(operator user.User) string {
	fullName := strings.TrimSpace(
		strings.TrimSpace(operator.FirstName) + " " + strings.TrimSpace(operator.LastName),
	)
	if fullName != "" {
		return fullName
	}
	if username := strings.TrimSpace(operator.Username); username != "" {
		return username
	}
	return strings.TrimSpace(operator.Email)
}

func afterSalesOperatorLabel(id uint, namesByID map[uint]string) string {
	switch {
	case id == 0:
		return "系统"
	case namesByID[id] != "":
		return namesByID[id]
	default:
		return fmt.Sprintf("账号 #%d", id)
	}
}
