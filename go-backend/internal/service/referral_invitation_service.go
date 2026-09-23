package service

import (
	"errors"
	"net/url"
	"strings"

	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/repository"
)

const referralIdentityCreateAttempts = 8

// ReferralIdentityStore retains one stable code per account. Database unique
// constraints on user ID and code remain the authority under concurrent reads.
type ReferralIdentityStore interface {
	FindIdentityByUserID(uint) (*loyalty.ReferralIdentity, error)
	CreateIdentity(*loyalty.ReferralIdentity) error
}

type ReferralCodeGenerator func() (string, error)

// ReferralInvitationLinkProvider may return a canonical URL or a persisted short
// link that redirects to it. Providers must reuse their mapping for an identity;
// changing a provider must never change the user's referral code.
type ReferralInvitationLinkProvider interface {
	BuildReferralInvitationURL(loyalty.ReferralIdentity) (string, error)
}

type StorefrontReferralInvitationLinks struct {
	StorefrontURL string
}

func (p StorefrontReferralInvitationLinks) BuildReferralInvitationURL(identity loyalty.ReferralIdentity) (string, error) {
	if err := loyalty.ValidateReferralCode(identity.ReferralCode); err != nil {
		return "", err
	}
	base := strings.TrimRight(strings.TrimSpace(p.StorefrontURL), "/")
	if err := validateReferralInvitationURL(base); err != nil {
		return "", err
	}
	parsed, _ := url.Parse(base)
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("referral storefront URL must not contain a query or fragment")
	}
	return base + "/r/" + loyalty.NormalizeReferralCode(identity.ReferralCode), nil
}

func validateReferralInvitationURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil || parsed.Hostname() == "" || parsed.User != nil ||
		(parsed.Scheme != "https" && parsed.Scheme != "http") {
		return errors.New("referral invitation URL must be an absolute HTTP(S) URL without credentials")
	}
	return nil
}

// ReferralInvitationService owns code allocation and link presentation only.
// Eligibility, attribution, registration points and orders remain in their
// respective services. Nil generator selects the current secure code generator.
type ReferralInvitationService struct {
	store    ReferralIdentityStore
	generate ReferralCodeGenerator
	links    ReferralInvitationLinkProvider
}

func NewReferralInvitationService(store ReferralIdentityStore, generate ReferralCodeGenerator, links ReferralInvitationLinkProvider) *ReferralInvitationService {
	if generate == nil {
		generate = loyalty.GenerateReferralCode
	}
	return &ReferralInvitationService{store: store, generate: generate, links: links}
}

func (s *ReferralInvitationService) GetOrCreateReferralInvitation(userID uint) (*loyalty.ReferralIdentity, string, error) {
	if s == nil || s.store == nil || s.generate == nil || s.links == nil || userID == 0 {
		return nil, "", ErrReferralServiceUnavailable
	}
	identity, err := s.getOrCreateIdentity(userID)
	if err != nil {
		return nil, "", err
	}
	if !identity.IsActive {
		return nil, "", ErrReferralCodeNotFound
	}
	link, err := s.links.BuildReferralInvitationURL(*identity)
	if err != nil {
		return nil, "", err
	}
	if err := validateReferralInvitationURL(link); err != nil {
		return nil, "", err
	}
	return identity, link, nil
}

func (s *ReferralInvitationService) getOrCreateIdentity(userID uint) (*loyalty.ReferralIdentity, error) {
	identity, err := s.store.FindIdentityByUserID(userID)
	if err == nil {
		if identity == nil {
			return nil, ErrReferralServiceUnavailable
		}
		return identity, nil
	}
	if !repository.IsRecordNotFound(err) {
		return nil, err
	}
	for attempt := 0; attempt < referralIdentityCreateAttempts; attempt++ {
		code, err := s.generate()
		if err != nil {
			return nil, err
		}
		code = loyalty.NormalizeReferralCode(code)
		if err := loyalty.ValidateReferralCode(code); err != nil {
			return nil, err
		}
		identity = &loyalty.ReferralIdentity{UserID: userID, ReferralCode: code, IsActive: true}
		if err := s.store.CreateIdentity(identity); err == nil {
			return identity, nil
		} else if !repository.IsDuplicatedKey(err) {
			return nil, err
		}
		if existing, err := s.store.FindIdentityByUserID(userID); err == nil {
			return existing, nil
		} else if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}
	return nil, errors.New("could not allocate a unique referral code")
}
