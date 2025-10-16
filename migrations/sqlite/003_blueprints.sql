-- Migration: 003_blueprints.sql
-- Description: Create industryBlueprints tables (complex structure with activities, materials, products)
-- Source: RIFT SDE Schema (https://sde.riftforeve.online/)
-- ADR Reference: ADR-001 (SQLite-Only), ADR-002 (Database Layer Design)

CREATE TABLE IF NOT EXISTS industryBlueprints (
    blueprintTypeID INTEGER PRIMARY KEY,
    maxProductionLimit INTEGER
);

CREATE TABLE IF NOT EXISTS industryActivities (
    blueprintTypeID INTEGER,
    activityID INTEGER,
    time INTEGER,
    PRIMARY KEY (blueprintTypeID, activityID)
);

CREATE TABLE IF NOT EXISTS industryActivityMaterials (
    blueprintTypeID INTEGER,
    activityID INTEGER,
    materialTypeID INTEGER,
    quantity INTEGER,
    PRIMARY KEY (blueprintTypeID, activityID, materialTypeID)
);

CREATE TABLE IF NOT EXISTS industryActivityProducts (
    blueprintTypeID INTEGER,
    activityID INTEGER,
    productTypeID INTEGER,
    quantity INTEGER,
    PRIMARY KEY (blueprintTypeID, activityID, productTypeID)
);

CREATE INDEX IF NOT EXISTS idx_industryActivities_blueprintTypeID ON industryActivities(blueprintTypeID);
CREATE INDEX IF NOT EXISTS idx_industryActivities_activityID ON industryActivities(activityID);
CREATE INDEX IF NOT EXISTS idx_industryActivityMaterials_blueprintTypeID ON industryActivityMaterials(blueprintTypeID);
CREATE INDEX IF NOT EXISTS idx_industryActivityMaterials_materialTypeID ON industryActivityMaterials(materialTypeID);
CREATE INDEX IF NOT EXISTS idx_industryActivityProducts_blueprintTypeID ON industryActivityProducts(blueprintTypeID);
CREATE INDEX IF NOT EXISTS idx_industryActivityProducts_productTypeID ON industryActivityProducts(productTypeID);
