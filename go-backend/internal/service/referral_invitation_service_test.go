package service

import (
	"errors"
	"testing"

	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
)

type referralInvitationLinkFunc func(loyalty.ReferralIdentity) (string, error)

func (f referralInvitationLinkFunc) BuildReferralInvitationURL(identity loyalty.ReferralIdentity) (string, error) {
	return f(identity)
}

func TestReferralInvitationProviderUpgradePreservesCodeAndBinding(t *testing.T) {
	svc, db := newReferralServiceFixture(t, true)
	referrer := createReferralTestUser(t, db, "invitation-owner@example.test", "invitation-owner")
	first, err := svc.Dashboard(referrer.ID)
	require.NoError(t, err)
	canonical := first.ShareURL
	providerCalls := 0
	svc.ConfigureReferralInvitationService(NewReferralInvitationService(svc.repo, func() (string, error) {
		t.Fatal("existing codes must survive generator upgrades")
		return "", nil
	}, referralInvitationLinkFunc(func(identity loyalty.ReferralIdentity) (string, error) {
		providerCalls++
		require.Equal(t, first.ReferralCode, identity.ReferralCode)
		// A future short-link adapter persists this alias -> canonical mapping.
		require.Contains(t, canonical, "/r/"+identity.ReferralCode)
		return "https://invite.example.test/s/stable-alias", nil
	})))
	for i := 0; i < 2; i++ {
		next, err := svc.Dashboard(referrer.ID)
		require.NoError(t, err)
		require.Equal(t, first.ReferralCode, next.ReferralCode)
		require.Equal(t, "https://invite.example.test/s/stable-alias", next.ShareURL)
	}
	require.Equal(t, 2, providerCalls)
	var identities int64
	require.NoError(t, db.Model(&loyalty.ReferralIdentity{}).Where("user_id = ?", referrer.ID).Count(&identities).Error)
	require.EqualValues(t, 1, identities)
	referee := createReferralTestUser(t, db, "invitation-friend@example.test", "invitation-friend")
	token, _, err := svc.CreateAttributionToken(first.ReferralCode, "link")
	require.NoError(t, err)
	bound, err := svc.BindFromToken(referee.ID, token, "203.0.113.1")
	require.NoError(t, err)
	require.Equal(t, referrer.ID, bound.ReferrerID)
}

func TestReferralInvitationRetriesCodeCollisionAndReusesAccountIdentity(t *testing.T) {
	_, db := newReferralServiceFixture(t, true)
	// Match the production case-insensitive code constraint in the SQLite fixture.
	require.NoError(t, db.Exec("CREATE UNIQUE INDEX invitation_test_code ON user_referral_identities(UPPER(referral_code))").Error)
	repo := repository.NewReferralRepository(db)
	owner := createReferralTestUser(t, db, "code-owner@example.test", "code-owner")
	next := createReferralTestUser(t, db, "code-next@example.test", "code-next")
	require.NoError(t, repo.CreateIdentity(&loyalty.ReferralIdentity{UserID: owner.ID, ReferralCode: "ABCD2345", IsActive: true}))
	attempts := 0
	svc := NewReferralInvitationService(repo, func() (string, error) {
		attempts++
		if attempts == 1 {
			return "ABCD2345", nil
		}
		return "WXYZ2345", nil
	}, StorefrontReferralInvitationLinks{StorefrontURL: "https://shop.test/"})
	identity, link, err := svc.GetOrCreateReferralInvitation(next.ID)
	require.NoError(t, err)
	require.Equal(t, "WXYZ2345", identity.ReferralCode)
	require.Equal(t, "https://shop.test/r/WXYZ2345", link)
	replayed, _, err := svc.GetOrCreateReferralInvitation(next.ID)
	require.NoError(t, err)
	require.Equal(t, identity.ID, replayed.ID)
	require.Equal(t, 2, attempts)
}

func TestReferralInvitationProviderFailureDoesNotReallocateCode(t *testing.T) {
	svc, db := newReferralServiceFixture(t, true)
	owner := createReferralTestUser(t, db, "provider-fail@example.test", "provider-fail")
	expected := errors.New("short-link provider unavailable")
	svc.ConfigureReferralInvitationService(NewReferralInvitationService(svc.repo, nil, referralInvitationLinkFunc(func(loyalty.ReferralIdentity) (string, error) {
		return "", expected
	})))
	_, err := svc.Dashboard(owner.ID)
	require.ErrorIs(t, err, expected)
	identity, err := svc.repo.FindIdentityByUserID(owner.ID)
	require.NoError(t, err)
	svc.ConfigureReferralInvitationService(NewReferralInvitationService(svc.repo, nil, StorefrontReferralInvitationLinks{StorefrontURL: "https://shop.test"}))
	dashboard, err := svc.Dashboard(owner.ID)
	require.NoError(t, err)
	require.Equal(t, identity.ReferralCode, dashboard.ReferralCode)
}

func TestReferralInvitationRejectsInvalidProviderURL(t *testing.T) {
	svc, db := newReferralServiceFixture(t, true)
	owner := createReferralTestUser(t, db, "invalid-link@example.test", "invalid-link")
	for _, value := range []string{"", "/r/ABCD2345", "javascript:alert(1)", "https://user:secret@shop.test/r/ABCD2345"} {
		svc.ConfigureReferralInvitationService(NewReferralInvitationService(svc.repo, nil, referralInvitationLinkFunc(func(loyalty.ReferralIdentity) (string, error) { return value, nil })))
		_, err := svc.Dashboard(owner.ID)
		require.Error(t, err)
	}
}

func TestReferralInvitationRejectsNilIdentityFromStore(t *testing.T) {
	store := referralInvitationStoreFunc{
		find: func(uint) (*loyalty.ReferralIdentity, error) { return nil, nil },
	}
	svc := NewReferralInvitationService(store, nil, StorefrontReferralInvitationLinks{StorefrontURL: "https://shop.test"})
	_, _, err := svc.GetOrCreateReferralInvitation(42)
	require.ErrorIs(t, err, ErrReferralServiceUnavailable)
}

type referralInvitationStoreFunc struct {
	find   func(uint) (*loyalty.ReferralIdentity, error)
	create func(*loyalty.ReferralIdentity) error
}

func (s referralInvitationStoreFunc) FindIdentityByUserID(userID uint) (*loyalty.ReferralIdentity, error) {
	return s.find(userID)
}

func (s referralInvitationStoreFunc) CreateIdentity(identity *loyalty.ReferralIdentity) error {
	if s.create != nil {
		return s.create(identity)
	}
	return nil
}
