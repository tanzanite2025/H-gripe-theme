package repository

import "encoding/json"

func encodeStorefrontURLIssueEventMetadata(metadata map[string]interface{}) (string, error) {
	if len(metadata) == 0 {
		return "{}", nil
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}
