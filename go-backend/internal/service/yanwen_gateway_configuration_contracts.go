package service

// YanwenAPIMasterKeyEnv is the process environment variable used to decrypt
// and encrypt the Yanwen gateway credentials.
const YanwenAPIMasterKeyEnv = "YANWEN_API_MASTER_KEY"

type YanwenAPIConfigInput struct {
	Environment string `json:"environment"`
	Endpoint    string `json:"endpoint"`
	UserID      string `json:"user_id"`
	APIToken    string `json:"api_token"`
	Enabled     bool   `json:"enabled"`
}

type YanwenPingResult struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	LatencyMS int64  `json:"latency_ms"`
}
