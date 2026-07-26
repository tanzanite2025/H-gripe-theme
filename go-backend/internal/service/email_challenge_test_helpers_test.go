package service

type recordingEmailSender struct {
	bodies []string
}

func (s *recordingEmailSender) SendEmail(_ []string, _ string, body string) error {
	s.bodies = append(s.bodies, body)
	return nil
}
