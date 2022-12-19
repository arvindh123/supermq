package groups

import (
	"github.com/mainflux/mainflux/auth/repo/groups/db"
	"github.com/mainflux/mainflux/auth/repo/groups/db/postgres"
	"github.com/mainflux/mainflux/internal/db/sqlxt"
)

func NewQueryFramer(db sqlxt.Database) db.QueryFramer {
	switch db.DriverName() {
	case "pgx", "postgres":
		fallthrough
	default:
		return postgres.New()
	}
}
