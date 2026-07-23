package repository

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tanzanite/internal/domain/loyalty"
)

func newMockLoyaltyRepository(t *testing.T) (*LoyaltyRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		_ = sqlDB.Close()
		t.Fatalf("failed to create GORM DB: %v", err)
	}

	return NewLoyaltyRepository(db), mock, func() {
		_ = sqlDB.Close()
	}
}

func TestFindAllMemberLevelsUsesCanonicalOrder(t *testing.T) {
	repo, mock, cleanup := newMockLoyaltyRepository(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_levels" WHERE "member_levels"."deleted_at" IS NULL ORDER BY sort_order ASC, min_points ASC, id ASC`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "min_points", "max_points", "sort_order"}).
			AddRow(1, "Bronze", 0, 999, 10).
			AddRow(2, "Silver", 1000, 4999, 20))

	levels, err := repo.FindAllMemberLevels()
	if err != nil {
		t.Fatalf("FindAllMemberLevels returned error: %v", err)
	}
	if len(levels) != 2 || levels[0].ID != 1 || levels[1].ID != 2 {
		t.Fatalf("unexpected member levels: %#v", levels)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestFindMemberLevelByPointsUsesConfiguredRange(t *testing.T) {
	repo, mock, cleanup := newMockLoyaltyRepository(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "member_levels" WHERE (min_points <= $1 AND max_points >= $2) AND "member_levels"."deleted_at" IS NULL ORDER BY min_points DESC,"member_levels"."id" LIMIT $3`)).
		WithArgs(1500, 1500, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "min_points", "max_points", "sort_order"}).
			AddRow(2, "Silver", 1000, 4999, 20))

	level, err := repo.FindMemberLevelByPoints(1500)
	if err != nil {
		t.Fatalf("FindMemberLevelByPoints returned error: %v", err)
	}
	if level.ID != 2 || level.MinPoints != 1000 {
		t.Fatalf("unexpected member level: %#v", level)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestGetLoyaltyStatsGroupsByMemberLevelID(t *testing.T) {
	repo, mock, cleanup := newMockLoyaltyRepository(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "user_loyalty"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT member_level_id AS level_id, count(*) as count FROM "user_loyalty" GROUP BY "member_level_id"`)).
		WillReturnRows(sqlmock.NewRows([]string{"level_id", "count"}).
			AddRow(1, 2).
			AddRow(2, 1))

	stats, err := repo.GetLoyaltyStats()
	if err != nil {
		t.Fatalf("GetLoyaltyStats returned error: %v", err)
	}
	if stats["total_members"] != int64(3) {
		t.Fatalf("unexpected total_members: %#v", stats["total_members"])
	}

	distribution := reflect.ValueOf(stats["level_distribution"])
	if distribution.Kind() != reflect.Slice || distribution.Len() != 2 {
		t.Fatalf("unexpected level_distribution: %#v", stats["level_distribution"])
	}
	first := distribution.Index(0)
	if first.FieldByName("LevelID").Uint() != 1 || first.FieldByName("Count").Int() != 2 {
		t.Fatalf("unexpected first level distribution entry: %#v", first.Interface())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestApplyPointsDelta(t *testing.T) {
	userLoyalty := &loyalty.UserLoyalty{
		TotalPoints:     100,
		AvailablePoints: 40,
		UsedPoints:      60,
	}

	applyPointsDelta(userLoyalty, 25, "earn")
	if userLoyalty.TotalPoints != 125 || userLoyalty.AvailablePoints != 65 || userLoyalty.UsedPoints != 60 {
		t.Fatalf("earn delta = total %d available %d used %d", userLoyalty.TotalPoints, userLoyalty.AvailablePoints, userLoyalty.UsedPoints)
	}

	applyPointsDelta(userLoyalty, -30, "spend")
	if userLoyalty.TotalPoints != 125 || userLoyalty.AvailablePoints != 35 || userLoyalty.UsedPoints != 90 {
		t.Fatalf("spend delta = total %d available %d used %d", userLoyalty.TotalPoints, userLoyalty.AvailablePoints, userLoyalty.UsedPoints)
	}

	applyPointsDelta(userLoyalty, 20, "refund")
	if userLoyalty.TotalPoints != 125 || userLoyalty.AvailablePoints != 55 || userLoyalty.UsedPoints != 70 {
		t.Fatalf("refund delta = total %d available %d used %d", userLoyalty.TotalPoints, userLoyalty.AvailablePoints, userLoyalty.UsedPoints)
	}
}

func TestApplyPointsDeltaRefundDoesNotMakeUsedNegative(t *testing.T) {
	userLoyalty := &loyalty.UserLoyalty{
		AvailablePoints: 10,
		UsedPoints:      5,
	}

	applyPointsDelta(userLoyalty, 10, "refund")
	if userLoyalty.AvailablePoints != 20 || userLoyalty.UsedPoints != 0 {
		t.Fatalf("refund clamp = available %d used %d", userLoyalty.AvailablePoints, userLoyalty.UsedPoints)
	}
}
