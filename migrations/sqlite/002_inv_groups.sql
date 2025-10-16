-- Migration: 002_inv_groups.sql
-- Description: Create invGroups table (EVE SDE item groups categorization)
-- Source: RIFT SDE Schema (https://sde.riftforeve.online/)
-- ADR Reference: ADR-001 (SQLite-Only), ADR-002 (Database Layer Design)

CREATE TABLE IF NOT EXISTS invGroups (
    groupID INTEGER PRIMARY KEY,
    categoryID INTEGER,
    groupName TEXT NOT NULL,
    iconID INTEGER,
    useBasePrice INTEGER,
    anchored INTEGER,
    anchorable INTEGER,
    fittableNonSingleton INTEGER,
    published INTEGER
);

CREATE INDEX IF NOT EXISTS idx_invGroups_categoryID ON invGroups(categoryID);
