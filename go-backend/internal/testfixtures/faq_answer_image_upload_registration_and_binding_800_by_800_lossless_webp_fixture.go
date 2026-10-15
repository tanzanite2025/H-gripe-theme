package testfixtures

import _ "embed"

//go:embed testdata/faq-answer-image-upload-registration-and-binding-800-by-800-lossless-webp-fixture.webp
var faqAnswerImageUploadRegistrationAndBinding800By800LosslessWebPFixture []byte

func FAQAnswerImageUploadRegistrationAndBinding800By800LosslessWebPFixtureBytes() []byte {
	return append([]byte(nil), faqAnswerImageUploadRegistrationAndBinding800By800LosslessWebPFixture...)
}
