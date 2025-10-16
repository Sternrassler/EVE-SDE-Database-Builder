-- Migration: 005_universe.sql
-- Description: Create Universe tables (Regions, Constellations, Solar Systems, Stargates, Planets)
-- Source: RIFT SDE Schema (https://sde.riftforeve.online/)
-- ADR Reference: ADR-001 (SQLite-Only), ADR-002 (Database Layer Design)

CREATE TABLE IF NOT EXISTS mapRegions (
    regionID INTEGER PRIMARY KEY,
    regionName TEXT,
    x REAL,
    y REAL,
    z REAL,
    factionID INTEGER
);

CREATE TABLE IF NOT EXISTS mapConstellations (
    constellationID INTEGER PRIMARY KEY,
    constellationName TEXT,
    regionID INTEGER,
    x REAL,
    y REAL,
    z REAL,
    factionID INTEGER
);

CREATE TABLE IF NOT EXISTS mapSolarSystems (
    solarSystemID INTEGER PRIMARY KEY,
    solarSystemName TEXT,
    regionID INTEGER,
    constellationID INTEGER,
    x REAL,
    y REAL,
    z REAL,
    security REAL,
    securityClass TEXT
);

CREATE TABLE IF NOT EXISTS mapStargates (
    stargateID INTEGER PRIMARY KEY,
    solarSystemID INTEGER,
    destinationID INTEGER
);

CREATE TABLE IF NOT EXISTS mapPlanets (
    planetID INTEGER PRIMARY KEY,
    planetName TEXT,
    solarSystemID INTEGER,
    typeID INTEGER,
    x REAL,
    y REAL,
    z REAL
);

CREATE INDEX IF NOT EXISTS idx_mapConstellations_regionID ON mapConstellations(regionID);
CREATE INDEX IF NOT EXISTS idx_mapSolarSystems_regionID ON mapSolarSystems(regionID);
CREATE INDEX IF NOT EXISTS idx_mapSolarSystems_constellationID ON mapSolarSystems(constellationID);
CREATE INDEX IF NOT EXISTS idx_mapStargates_solarSystemID ON mapStargates(solarSystemID);
CREATE INDEX IF NOT EXISTS idx_mapStargates_destinationID ON mapStargates(destinationID);
CREATE INDEX IF NOT EXISTS idx_mapPlanets_solarSystemID ON mapPlanets(solarSystemID);
CREATE INDEX IF NOT EXISTS idx_mapPlanets_typeID ON mapPlanets(typeID);
