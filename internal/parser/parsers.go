// Package parser provides JSONL parser instances for EVE SDE tables.
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
	FactionID          int      `json:"factionID"`
	FactionName        *string  `json:"factionName"`
	Description        *string  `json:"description"`
	SolarSystemID      *int     `json:"solarSystemID"`
	CorporationID      *int     `json:"corporationID"`
	SizeFactor         *float64 `json:"sizeFactor"`
	StationCount       *int     `json:"stationCount"`
	StationSystemCount *int     `json:"stationSystemCount"`
	MilitiaCorporationID *int   `json:"militiaCorporationID"`
	IconID             *int     `json:"iconID"`
}

// Parser instances for core EVE SDE tables
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
// This function provides a central registry for all EVE SDE parsers.
//
// Core Parsers (Priority 1 - Epic #4 Task #37):
//   - invTypes, invGroups, invCategories, invMarketGroups, invMetaGroups
//   - industryBlueprints
//   - dogmaAttributes, dogmaEffects, dogmaTypeAttributes, dogmaTypeEffects
//   - mapRegions, mapConstellations, mapSolarSystems, mapStargates, mapPlanets
//   - chrRaces, chrFactions
//
// Extended Parsers (Priority 2 - Epic #4 Task #38):
//   - To be implemented: ~35 additional tables including:
//     * Character/NPC: ancestries, bloodlines, characterAttributes, npcCharacters, npcCorporations, etc.
//     * Agents: agentTypes, agentsInSpace
//     * Additional Dogma: dogmaAttributeCategories, dogmaUnits, typeDogma, dynamicItemAttributes
//     * Extended Universe: mapMoons, mapStars, mapAsteroidBelts, landmarks
//     * Certificates/Skills: certificates, masteries
//     * Skins: skins, skinLicenses, skinMaterials
//     * Translation: translationLanguages
//     * Station: stationOperations, stationServices, sovereigntyUpgrades, npcStations
//     * Miscellaneous: icons, graphics, contrabandTypes, controlTowerResources, etc.
func RegisterParsers() map[string]Parser {
	return map[string]Parser{
		// Core Inventory & Market (Priority 1)
		"invTypes":         InvTypesParser,
		"invGroups":        InvGroupsParser,
		"invCategories":    InvCategoriesParser,
		"invMarketGroups":  InvMarketGroupsParser,
		"invMetaGroups":    InvMetaGroupsParser,

		// Industry & Blueprints (Priority 1)
		"industryBlueprints": IndustryBlueprintsParser,

		// Dogma System (Priority 1)
		"dogmaAttributes":     DogmaAttributesParser,
		"dogmaEffects":        DogmaEffectsParser,
		"dogmaTypeAttributes": DogmaTypeAttributesParser,
		"dogmaTypeEffects":    DogmaTypeEffectsParser,

		// Universe/Map (Priority 1)
		"mapRegions":        MapRegionsParser,
		"mapConstellations": MapConstellationsParser,
		"mapSolarSystems":   MapSolarSystemsParser,
		"mapStargates":      MapStargatesParser,
		"mapPlanets":        MapPlanetsParser,

		// Character/Faction (Priority 1)
		"chrRaces":    ChrRacesParser,
		"chrFactions": ChrFactionsParser,

		// TODO (Epic #4, Task #38): Add remaining ~35 parsers
		// See function documentation for complete list
	}
}
