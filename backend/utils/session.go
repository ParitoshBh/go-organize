package utils

import (
	"go-organizer/backend/connections"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"

	// Indirect dependency for scs session manager
	_ "github.com/mattn/go-sqlite3"
)

var sessionManager *scs.SessionManager

// InitSessionStore initializes session store
func InitSessionStore() {
	// Initialize session manager and configure it to use SQLite3 session store
	sm := scs.New()
	sm.Store = sqlite3store.New(connections.GetSqlDBConnection())
	sm.IdleTimeout = time.Minute * 30
	sm.Lifetime = time.Hour * 1

	sessionManager = sm
}

// GetSessionManager returns session manager
func GetSessionManager() *scs.SessionManager {
	return sessionManager
}
