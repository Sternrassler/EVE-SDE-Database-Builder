-- Migration: 001_inv_types.sql
-- Description: Create invTypes table (EVE SDE most frequently used table)
-- Source: RIFT SDE Schema (https://sde.riftforeve.online/)
-- ADR Reference: ADR-001 (SQLite-Only), ADR-002 (Database Layer Design)

CREATE TABLE IF NOT EXISTS invTypes (
    typeID INTEGER PRIMARY KEY,
    typeName TEXT NOT NULL,
    groupID INTEGER,
    description TEXT,
    mass REAL,
    volume REAL,
    capacity REAL,
    portionSize INTEGER,
    raceID INTEGER,
    basePrice REAL,
    published INTEGER,
    marketGroupID INTEGER,
    iconID INTEGER,
    soundID INTEGER,
    graphicID INTEGER
);

CREATE INDEX IF NOT EXISTS idx_invTypes_groupID ON invTypes(groupID);
CREATE INDEX IF NOT EXISTS idx_invTypes_marketGroupID ON invTypes(marketGroupID);
