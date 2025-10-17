// Package parser provides JSONL parser instances for EVE SDE tables.
// This file contains core parser definitions (17 essential tables).
// Extended parsers are in parsers_extended.go.
package parser

// InvType represents an EVE SDE invTypes record
type InvType struct {
	TypeID        int      `json:"typeID"`
	TypeName      string   `json:"typeName"`
	GroupID       *int     `json:"groupID"`
	Description   *string  `json:"description"`
	Mass          *float64 `json:"mass"`
	Volume        *float64 `json:"volume"`
	Capacity      *float64 `json:"capacity"`
	PortionSize   *int     `json:"portionSize"`
	RaceID        *int     `json:"raceID"`
	BasePrice     *float64 `json:"basePrice"`
	Published     *int     `json:"published"`
	MarketGroupID *int     `json:"marketGroupID"`
	IconID        *int     `json:"iconID"`
	SoundID       *int     `json:"soundID"`
	GraphicID     *int     `json:"graphicID"`
}

// InvGroup represents an EVE SDE invGroups record
type InvGroup struct {
	GroupID              int     `json:"groupID"`
	CategoryID           *int    `json:"categoryID"`
	GroupName            string  `json:"groupName"`
	IconID               *int    `json:"iconID"`
	UseBasePrice         *int    `json:"useBasePrice"`
	Anchored             *int    `json:"anchored"`
	Anchorable           *int    `json:"anchorable"`
	FittableNonSingleton *int    `json:"fittableNonSingleton"`
	Published            *int    `json:"published"`
}

// IndustryBlueprint represents an EVE SDE industryBlueprints record
type IndustryBlueprint struct {
	BlueprintTypeID    int  `json:"blueprintTypeID"`
	MaxProductionLimit *int `json:"maxProductionLimit"`
}

// DogmaAttribute represents an EVE SDE dogmaAttributes record
type DogmaAttribute struct {
	AttributeID   int      `json:"attributeID"`
	AttributeName *string  `json:"attributeName"`
	Description   *string  `json:"description"`
	IconID        *int     `json:"iconID"`
	DefaultValue  *float64 `json:"defaultValue"`
	Published     *int     `json:"published"`
	DisplayName   *string  `json:"displayName"`
	UnitID        *int     `json:"unitID"`
	Stackable     *int     `json:"stackable"`
	HighIsGood    *int     `json:"highIsGood"`
}

// MapSolarSystem represents an EVE SDE mapSolarSystems record
type MapSolarSystem struct {
	SolarSystemID   int      `json:"solarSystemID"`
	SolarSystemName *string  `json:"solarSystemName"`
	RegionID        *int     `json:"regionID"`
	ConstellationID *int     `json:"constellationID"`
	X               *float64 `json:"x"`
	Y               *float64 `json:"y"`
	Z               *float64 `json:"z"`
	Security        *float64 `json:"security"`
	SecurityClass   *string  `json:"securityClass"`
}

// DogmaEffect represents an EVE SDE dogmaEffects record
type DogmaEffect struct {
	EffectID                      int     `json:"effectID"`
	EffectName                    *string `json:"effectName"`
	EffectCategory                *int    `json:"effectCategory"`
	PreExpression                 *int    `json:"preExpression"`
	PostExpression                *int    `json:"postExpression"`
	Description                   *string `json:"description"`
	Guid                          *string `json:"guid"`
	IconID                        *int    `json:"iconID"`
	IsOffensive                   *int    `json:"isOffensive"`
	IsAssistance                  *int    `json:"isAssistance"`
	DurationAttributeID           *int    `json:"durationAttributeID"`
	TrackingSpeedAttributeID      *int    `json:"trackingSpeedAttributeID"`
	DischargeAttributeID          *int    `json:"dischargeAttributeID"`
	RangeAttributeID              *int    `json:"rangeAttributeID"`
	FalloffAttributeID            *int    `json:"falloffAttributeID"`
	DisallowAutoRepeat            *int    `json:"disallowAutoRepeat"`
	Published                     *int    `json:"published"`
	DisplayName                   *string `json:"displayName"`
	IsWarpSafe                    *int    `json:"isWarpSafe"`
	RangeChance                   *int    `json:"rangeChance"`
	ElectronicChance              *int    `json:"electronicChance"`
	PropulsionChance              *int    `json:"propulsionChance"`
	Distribution                  *int    `json:"distribution"`
	SfxName                       *string `json:"sfxName"`
	NpcUsageChanceAttributeID     *int    `json:"npcUsageChanceAttributeID"`
	NpcActivationChanceAttributeID *int   `json:"npcActivationChanceAttributeID"`
	FittingUsageChanceAttributeID *int    `json:"fittingUsageChanceAttributeID"`
	ModifierInfo                  *string `json:"modifierInfo"`
}

// DogmaTypeAttribute represents an EVE SDE dogmaTypeAttributes record
type DogmaTypeAttribute struct {
	TypeID      int      `json:"typeID"`
	AttributeID int      `json:"attributeID"`
	ValueInt    *int     `json:"valueInt"`
	ValueFloat  *float64 `json:"valueFloat"`
}

// DogmaTypeEffect represents an EVE SDE dogmaTypeEffects record
type DogmaTypeEffect struct {
	TypeID    int  `json:"typeID"`
	EffectID  int  `json:"effectID"`
	IsDefault *int `json:"isDefault"`
}

// MapRegion represents an EVE SDE mapRegions record
type MapRegion struct {
	RegionID   int      `json:"regionID"`
	RegionName *string  `json:"regionName"`
	X          *float64 `json:"x"`
	Y          *float64 `json:"y"`
	Z          *float64 `json:"z"`
	FactionID  *int     `json:"factionID"`
}

// MapConstellation represents an EVE SDE mapConstellations record
type MapConstellation struct {
	ConstellationID   int      `json:"constellationID"`
	ConstellationName *string  `json:"constellationName"`
	RegionID          *int     `json:"regionID"`
	X                 *float64 `json:"x"`
	Y                 *float64 `json:"y"`
	Z                 *float64 `json:"z"`
	FactionID         *int     `json:"factionID"`
}

// MapStargate represents an EVE SDE mapStargates record
type MapStargate struct {
	StargateID    int  `json:"stargateID"`
	SolarSystemID *int `json:"solarSystemID"`
	DestinationID *int `json:"destinationID"`
}

// MapPlanet represents an EVE SDE mapPlanets record
type MapPlanet struct {
	PlanetID      int      `json:"planetID"`
	PlanetName    *string  `json:"planetName"`
	SolarSystemID *int     `json:"solarSystemID"`
	TypeID        *int     `json:"typeID"`
	X             *float64 `json:"x"`
	Y             *float64 `json:"y"`
	Z             *float64 `json:"z"`
}

// InvCategory represents an EVE SDE invCategories record
type InvCategory struct {
	CategoryID   int     `json:"categoryID"`
	CategoryName *string `json:"categoryName"`
	IconID       *int    `json:"iconID"`
	Published    *int    `json:"published"`
}

// InvMarketGroup represents an EVE SDE invMarketGroups record
type InvMarketGroup struct {
	MarketGroupID       int     `json:"marketGroupID"`
	ParentGroupID       *int    `json:"parentGroupID"`
	MarketGroupName     *string `json:"marketGroupName"`
	Description         *string `json:"description"`
	IconID              *int    `json:"iconID"`
	HasTypes            *int    `json:"hasTypes"`
}

// InvMetaGroup represents an EVE SDE invMetaGroups record
type InvMetaGroup struct {
	MetaGroupID   int     `json:"metaGroupID"`
	MetaGroupName *string `json:"metaGroupName"`
	IconID        *int    `json:"iconID"`
	Description   *string `json:"description"`
}

// ChrRace represents an EVE SDE chrRaces record
type ChrRace struct {
	RaceID      int     `json:"raceID"`
	RaceName    *string `json:"raceName"`
	Description *string `json:"description"`
	IconID      *int    `json:"iconID"`
}

// ChrFaction represents an EVE SDE chrFactions record
type ChrFaction struct {
	FactionID            int      `json:"factionID"`
	FactionName          *string  `json:"factionName"`
	Description          *string  `json:"description"`
	SolarSystemID        *int     `json:"solarSystemID"`
	CorporationID        *int     `json:"corporationID"`
	SizeFactor           *float64 `json:"sizeFactor"`
	StationCount         *int     `json:"stationCount"`
	StationSystemCount   *int     `json:"stationSystemCount"`
	MilitiaCorporationID *int     `json:"militiaCorporationID"`
	IconID               *int     `json:"iconID"`
}



// Core parser instances for EVE SDE tables (17 essential tables).
// Extended parsers (36 tables) are defined in parsers_extended.go.
var (
	InvTypesParser = NewJSONLParser[InvType]("invTypes", []string{
		"typeID", "typeName", "groupID", "description", "mass", "volume",
		"capacity", "portionSize", "raceID", "basePrice", "published",
		"marketGroupID", "iconID", "soundID", "graphicID",
	})

	InvGroupsParser = NewJSONLParser[InvGroup]("invGroups", []string{
		"groupID", "categoryID", "groupName", "iconID", "useBasePrice",
		"anchored", "anchorable", "fittableNonSingleton", "published",
	})

	InvCategoriesParser = NewJSONLParser[InvCategory]("invCategories", []string{
		"categoryID", "categoryName", "iconID", "published",
	})

	InvMarketGroupsParser = NewJSONLParser[InvMarketGroup]("invMarketGroups", []string{
		"marketGroupID", "parentGroupID", "marketGroupName", "description", "iconID", "hasTypes",
	})

	InvMetaGroupsParser = NewJSONLParser[InvMetaGroup]("invMetaGroups", []string{
		"metaGroupID", "metaGroupName", "iconID", "description",
	})

	IndustryBlueprintsParser = NewJSONLParser[IndustryBlueprint]("industryBlueprints", []string{
		"blueprintTypeID", "maxProductionLimit",
	})

	DogmaAttributesParser = NewJSONLParser[DogmaAttribute]("dogmaAttributes", []string{
		"attributeID", "attributeName", "description", "iconID", "defaultValue",
		"published", "displayName", "unitID", "stackable", "highIsGood",
	})

	DogmaEffectsParser = NewJSONLParser[DogmaEffect]("dogmaEffects", []string{
		"effectID", "effectName", "effectCategory", "preExpression", "postExpression",
		"description", "guid", "iconID", "isOffensive", "isAssistance",
		"durationAttributeID", "trackingSpeedAttributeID", "dischargeAttributeID",
		"rangeAttributeID", "falloffAttributeID", "disallowAutoRepeat", "published",
		"displayName", "isWarpSafe", "rangeChance", "electronicChance",
		"propulsionChance", "distribution", "sfxName", "npcUsageChanceAttributeID",
		"npcActivationChanceAttributeID", "fittingUsageChanceAttributeID", "modifierInfo",
	})

	DogmaTypeAttributesParser = NewJSONLParser[DogmaTypeAttribute]("dogmaTypeAttributes", []string{
		"typeID", "attributeID", "valueInt", "valueFloat",
	})

	DogmaTypeEffectsParser = NewJSONLParser[DogmaTypeEffect]("dogmaTypeEffects", []string{
		"typeID", "effectID", "isDefault",
	})

	MapRegionsParser = NewJSONLParser[MapRegion]("mapRegions", []string{
		"regionID", "regionName", "x", "y", "z", "factionID",
	})

	MapConstellationsParser = NewJSONLParser[MapConstellation]("mapConstellations", []string{
		"constellationID", "constellationName", "regionID", "x", "y", "z", "factionID",
	})

	MapSolarSystemsParser = NewJSONLParser[MapSolarSystem]("mapSolarSystems", []string{
		"solarSystemID", "solarSystemName", "regionID", "constellationID",
		"x", "y", "z", "security", "securityClass",
	})

	MapStargatesParser = NewJSONLParser[MapStargate]("mapStargates", []string{
		"stargateID", "solarSystemID", "destinationID",
	})

	MapPlanetsParser = NewJSONLParser[MapPlanet]("mapPlanets", []string{
		"planetID", "planetName", "solarSystemID", "typeID", "x", "y", "z",
	})

	ChrRacesParser = NewJSONLParser[ChrRace]("chrRaces", []string{
		"raceID", "raceName", "description", "iconID",
	})

	ChrFactionsParser = NewJSONLParser[ChrFaction]("chrFactions", []string{
		"factionID", "factionName", "description", "solarSystemID", "corporationID",
		"sizeFactor", "stationCount", "stationSystemCount", "militiaCorporationID", "iconID",
	})
)

// RegisterParsers returns a map of all registered parsers keyed by table name.
// This function provides a central registry for all EVE SDE parsers (53 total).
//
// Core Parsers (17 tables - defined in parsers.go):
//   - Inventory: invTypes, invGroups, invCategories, invMarketGroups, invMetaGroups
//   - Industry: industryBlueprints
//   - Dogma Core: dogmaAttributes, dogmaEffects, dogmaTypeAttributes, dogmaTypeEffects
//   - Universe Core: mapRegions, mapConstellations, mapSolarSystems, mapStargates, mapPlanets
//   - Character: chrRaces, chrFactions
//
// Extended Parsers (36 tables - defined in parsers_extended.go):
//   - Character/NPC Extended: chrAncestries, chrBloodlines, chrAttributes, chrNPCCharacters
//     crpNPCCorporations, crpNPCCorporationDivisions, staStations
//   - Agents: agtAgentTypes, agtAgents
//   - Dogma Extended: dogmaAttributeCategories, dogmaUnits, typeDogma, dynamicItemAttributes
//   - Universe Extended: mapMoons, mapStars, mapAsteroidBelts, mapLandmarks
//   - Certificates: certCerts, certMasteries
//   - Skins: skins, skinLicenses, skinMaterials
//   - Translation: translationLanguages
//   - Station: staOperations, staServices, sovereigntyUpgrades
//   - Miscellaneous: eveIcons, eveGraphics, contrabandTypes, controlTowerResources,
//     crpActivities, dbuffCollections, planetResources, planetSchematics, typeBonuses, _sde
func RegisterParsers() map[string]Parser {
	return map[string]Parser{
		// Core Inventory & Market
		"invTypes":        InvTypesParser,
		"invGroups":       InvGroupsParser,
		"invCategories":   InvCategoriesParser,
		"invMarketGroups": InvMarketGroupsParser,
		"invMetaGroups":   InvMetaGroupsParser,

		// Industry & Blueprints
		"industryBlueprints": IndustryBlueprintsParser,

		// Dogma System (Core)
		"dogmaAttributes":     DogmaAttributesParser,
		"dogmaEffects":        DogmaEffectsParser,
		"dogmaTypeAttributes": DogmaTypeAttributesParser,
		"dogmaTypeEffects":    DogmaTypeEffectsParser,

		// Dogma System (Extended)
		"dogmaAttributeCategories": DogmaAttributeCategoriesParser,
		"dogmaUnits":               DogmaUnitsParser,
		"typeDogma":                TypeDogmaParser,
		"dynamicItemAttributes":    DynamicItemAttributesParser,

		// Universe/Map (Core)
		"mapRegions":        MapRegionsParser,
		"mapConstellations": MapConstellationsParser,
		"mapSolarSystems":   MapSolarSystemsParser,
		"mapStargates":      MapStargatesParser,
		"mapPlanets":        MapPlanetsParser,

		// Universe/Map (Extended)
		"mapMoons":         MapMoonsParser,
		"mapStars":         MapStarsParser,
		"mapAsteroidBelts": MapAsteroidBeltsParser,
		"mapLandmarks":     LandmarksParser,

		// Character/Faction (Core)
		"chrRaces":    ChrRacesParser,
		"chrFactions": ChrFactionsParser,

		// Character/NPC (Extended)
		"chrAncestries":                ChrAncestriesParser,
		"chrBloodlines":                ChrBloodlinesParser,
		"chrAttributes":                ChrAttributesParser,
		"chrNPCCharacters":             NPCCharactersParser,
		"crpNPCCorporations":           CrpNPCCorporationsParser,
		"crpNPCCorporationDivisions":   CrpNPCCorporationDivisionsParser,
		"staStations":                  StaNPCStationsParser,

		// Agents
		"agtAgentTypes": AgentTypesParser,
		"agtAgents":     AgentsInSpaceParser,

		// Certificates/Skills
		"certCerts":      CertificatesParser,
		"certMasteries":  MasteriesParser,

		// Skins
		"skins":         SkinsParser,
		"skinLicenses":  SkinLicensesParser,
		"skinMaterials": SkinMaterialsParser,

		// Translation
		"translationLanguages": TranslationLanguagesParser,

		// Station
		"staOperations":       StationOperationsParser,
		"staServices":         StationServicesParser,
		"sovereigntyUpgrades": SovereigntyUpgradesParser,

		// Miscellaneous
		"eveIcons":               IconsParser,
		"eveGraphics":            GraphicsParser,
		"contrabandTypes":        ContrabandTypesParser,
		"controlTowerResources":  ControlTowerResourcesParser,
		"crpActivities":          CorporationActivitiesParser,
		"dbuffCollections":       DogmaBuffCollectionsParser,
		"planetResources":        PlanetResourcesParser,
		"planetSchematics":       PlanetSchematicsParser,
		"typeBonuses":            TypeBonusesParser,
		"_sde":                   SDEMetadataParser,
	}
}
