package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"go-yzs/database"
	"go-yzs/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func useTestMySQL(t *testing.T) {
	t.Helper()
	dsn := os.Getenv("YZS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("YZS_TEST_MYSQL_DSN is not set")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	previousDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = previousDB })
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql database: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
}

func TestBuildOperatorRecordUnionUsesOneCanonicalEventSource(t *testing.T) {
	filter := operatorRecordFilter{
		UserIDs:          []uint{7, 8},
		StartTime:        "2026-08-01 00:00:00",
		EndTimeExclusive: "2026-09-01 00:00:00",
		ActionType:       "pend",
		InspectStatus:    "normal",
	}
	sql, args := buildOperatorRecordUnion(filter)

	for _, fragment := range []string{
		"t.id, COALESCE(t.trade_id, 0) AS trade_id",
		"r.submitted_at AS handled_at",
		"r.action_type",
		"COALESCE(t.inspect_status, '') AS inspect_status",
		"NOT EXISTS (SELECT 1 FROM trade_reviews r2 WHERE r2.trade_id = t.id)",
		"t.pend_status_desc IN ('已挂起', 'PENDING')",
		"r.submitted_by_id IN ?",
		"t.handled_by_id IN ?",
		"r.submitted_at < ?",
		"t.handled_at < ?",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("canonical query is missing %q\n%s", fragment, sql)
		}
	}

	wantPerSource := []interface{}{filter.UserIDs, filter.StartTime, filter.EndTimeExclusive, "pend", "normal"}
	wantArgs := append(append([]interface{}{}, wantPerSource...), wantPerSource...)
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("unexpected arguments\nwant: %#v\n got: %#v", wantArgs, args)
	}
}

func TestBuildOperatorActionUnionDoesNotMaterializeWideFields(t *testing.T) {
	sql, args := buildOperatorActionUnion(operatorRecordFilter{
		UserIDs: []uint{9}, StartTime: "2026-08-01 00:00:00",
	})
	for _, fragment := range []string{"handle_goods", "handle_remark", "operator_remark", "out_order_no", "video_duration"} {
		if strings.Contains(sql, fragment) {
			t.Fatalf("narrow action query unexpectedly contains %q", fragment)
		}
	}
	for _, fragment := range []string{"r.submitted_at AS handled_at", "t.handled_at", "action_type", "NOT EXISTS"} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("narrow action query is missing %q", fragment)
		}
	}
	if len(args) != 4 {
		t.Fatalf("expected two filters for each source, got %d arguments", len(args))
	}
}

func TestParseShanghaiDate(t *testing.T) {
	date, err := parseShanghaiDate("2026-08-26")
	if err != nil {
		t.Fatal(err)
	}
	if got := date.Format("2006-01-02 15:04:05 -0700"); got != "2026-08-26 00:00:00 +0800" {
		t.Fatalf("unexpected Shanghai date: %s", got)
	}
	if _, err := parseShanghaiDate("08/26/2026"); err == nil {
		t.Fatal("invalid date should be rejected")
	}
}

func TestParseShanghaiDateTimeAcceptsMinuteAndSecondPrecision(t *testing.T) {
	for _, value := range []string{"2026-08-26 09:30", "2026-08-26 09:30:45"} {
		parsed, err := parseShanghaiDateTime(value)
		if err != nil {
			t.Fatalf("%s should be valid: %v", value, err)
		}
		if parsed.Location().String() != "Asia/Shanghai" {
			t.Fatalf("unexpected location for %s: %s", value, parsed.Location())
		}
	}
	if _, err := parseShanghaiDateTime("2026-08-26"); err == nil {
		t.Fatal("date without time should be rejected")
	}
}

func TestCanonicalOperatorQueriesAgainstMySQL(t *testing.T) {
	useTestMySQL(t)
	var videoColumnCount int64
	if err := database.DB.Raw(`SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'trade_abnormals' AND column_name = 'video_duration'`).Scan(&videoColumnCount).Error; err != nil {
		t.Fatalf("inspect trade schema: %v", err)
	}

	filter := operatorRecordFilter{
		StartTime:        "2099-01-01 00:00:00",
		EndTimeExclusive: "2099-01-02 00:00:00",
	}
	unionSQL, args := buildOperatorRecordUnion(filter)
	// Older local databases may not have run the latest AutoMigrate yet. Keep
	// the syntax check read-only by substituting a typed placeholder column.
	if videoColumnCount == 0 {
		unionSQL = strings.ReplaceAll(unionSQL, "t.video_duration", "CAST(NULL AS SIGNED) AS video_duration")
	}
	var total int64
	if err := database.DB.Raw("SELECT COUNT(*) FROM ("+unionSQL+") AS events", args...).Scan(&total).Error; err != nil {
		t.Fatalf("query record count: %v", err)
	}
	var records []opRecord
	pageArgs := append(append([]interface{}{}, args...), 1, 0)
	if err := database.DB.Raw("SELECT * FROM ("+unionSQL+") AS events ORDER BY handled_at DESC LIMIT ? OFFSET ?", pageArgs...).Scan(&records).Error; err != nil {
		t.Fatalf("query record page: %v", err)
	}
	var summary operatorRecordSummary
	if err := database.DB.Raw(`SELECT
		COALESCE(SUM(CASE WHEN action_type = 'submit' THEN 1 ELSE 0 END), 0) AS submit_count,
		COALESCE(SUM(CASE WHEN action_type = 'pend' THEN 1 ELSE 0 END), 0) AS pend_count
		FROM (`+unionSQL+") AS events", args...).Scan(&summary).Error; err != nil {
		t.Fatalf("query summary: %v", err)
	}
	var amount float64
	if err := database.DB.Raw(`SELECT COALESCE(SUM(jt.price * jt.cnt), 0) AS amount
		FROM (`+unionSQL+`) AS events
		JOIN JSON_TABLE(
			CASE WHEN JSON_VALID(events.handle_goods) THEN events.handle_goods ELSE '[]' END,
			'$[*]' COLUMNS (price DOUBLE PATH '$.goodsPrice', cnt INT PATH '$.goodsCount')
		) jt ON TRUE WHERE events.action_type = 'submit'`, args...).Scan(&amount).Error; err != nil {
		t.Fatalf("query amount: %v", err)
	}
	var rows []dailyActionRow
	if err := database.DB.Raw(`SELECT DATE(handled_at) AS date,
		SUM(CASE WHEN action_type = 'submit' THEN 1 ELSE 0 END) AS submit_count,
		SUM(CASE WHEN action_type = 'pend' THEN 1 ELSE 0 END) AS pend_count
		FROM (`+unionSQL+`) AS events GROUP BY DATE(handled_at)`, args...).Scan(&rows).Error; err != nil {
		t.Fatalf("query daily aggregation: %v", err)
	}
}

func TestReadEndpointPerformanceAgainstMySQL(t *testing.T) {
	useTestMySQL(t)
	gin.SetMode(gin.TestMode)
	admin := &models.User{ID: 1, Role: "admin", Username: "admin"}
	tests := []struct {
		name    string
		url     string
		handler gin.HandlerFunc
		user    *models.User
	}{
		{name: "operators", url: "/stats/operators", handler: GetOperatorStats},
		{name: "daily", url: "/stats/daily", handler: GetDailyStats},
		{name: "inspectors", url: "/stats/inspectors", handler: GetInspectorStats},
		{name: "stats", url: "/stats", handler: GetStats, user: admin},
		{name: "unhandled", url: "/trades?isHandled=false&page=1&size=1", handler: ListTradeAbnormal},
		{name: "trades-by-date", url: "/trades?page=1&size=20&startDate=2026-08-24&endDate=2026-08-26", handler: ListTradeAbnormal},
		{name: "hourly", url: "/trades/hourly-stats?startDate=2026-08-24&endDate=2026-08-26", handler: GetHourlyStats},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Request = httptest.NewRequest(http.MethodGet, tt.url, nil)
			if tt.user != nil {
				context.Set("user", tt.user)
			}
			started := time.Now()
			tt.handler(context)
			elapsed := time.Since(started)
			if response.Code != http.StatusOK {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			t.Logf("%s completed in %s (%d bytes)", tt.name, elapsed, response.Body.Len())
		})
	}
}
