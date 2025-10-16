-- Migration: 004_dogma.sql
-- Description: Create Dogma system tables (Attributes, Effects, Type Attributes/Effects)
-- Source: RIFT SDE Schema (https://sde.riftforeve.online/)
-- ADR Reference: ADR-001 (SQLite-Only), ADR-002 (Database Layer Design)

CREATE TABLE IF NOT EXISTS dogmaAttributes (
    attributeID INTEGER PRIMARY KEY,
    attributeName TEXT,
    description TEXT,
    iconID INTEGER,
    defaultValue REAL,
    published INTEGER,
    displayName TEXT,
    unitID INTEGER,
    stackable INTEGER,
    highIsGood INTEGER
);

CREATE TABLE IF NOT EXISTS dogmaEffects (
    effectID INTEGER PRIMARY KEY,
    effectName TEXT,
    effectCategory INTEGER,
    preExpression INTEGER,
    postExpression INTEGER,
    description TEXT,
    guid TEXT,
    iconID INTEGER,
    isOffensive INTEGER,
    isAssistance INTEGER,
    durationAttributeID INTEGER,
    trackingSpeedAttributeID INTEGER,
    dischargeAttributeID INTEGER,
    rangeAttributeID INTEGER,
    falloffAttributeID INTEGER,
    disallowAutoRepeat INTEGER,
    published INTEGER,
    displayName TEXT,
    isWarpSafe INTEGER,
    rangeChance INTEGER,
    electronicChance INTEGER,
    propulsionChance INTEGER,
    distribution INTEGER,
    sfxName TEXT,
    npcUsageChanceAttributeID INTEGER,
    npcActivationChanceAttributeID INTEGER,
    fittingUsageChanceAttributeID INTEGER,
    modifierInfo TEXT
);

CREATE TABLE IF NOT EXISTS dogmaTypeAttributes (
    typeID INTEGER,
    attributeID INTEGER,
    valueInt INTEGER,
    valueFloat REAL,
    PRIMARY KEY (typeID, attributeID)
);

CREATE TABLE IF NOT EXISTS dogmaTypeEffects (
    typeID INTEGER,
    effectID INTEGER,
    isDefault INTEGER,
    PRIMARY KEY (typeID, effectID)
);

CREATE INDEX IF NOT EXISTS idx_dogmaAttributes_attributeName ON dogmaAttributes(attributeName);
CREATE INDEX IF NOT EXISTS idx_dogmaEffects_effectName ON dogmaEffects(effectName);
CREATE INDEX IF NOT EXISTS idx_dogmaEffects_effectCategory ON dogmaEffects(effectCategory);
CREATE INDEX IF NOT EXISTS idx_dogmaTypeAttributes_typeID ON dogmaTypeAttributes(typeID);
CREATE INDEX IF NOT EXISTS idx_dogmaTypeAttributes_attributeID ON dogmaTypeAttributes(attributeID);
CREATE INDEX IF NOT EXISTS idx_dogmaTypeEffects_typeID ON dogmaTypeEffects(typeID);
CREATE INDEX IF NOT EXISTS idx_dogmaTypeEffects_effectID ON dogmaTypeEffects(effectID);
