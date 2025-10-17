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

	IndustryBlueprintsParser = NewJSONLParser[IndustryBlueprint]("industryBlueprints", []string{
		"blueprintTypeID", "maxProductionLimit",
	})

	DogmaAttributesParser = NewJSONLParser[DogmaAttribute]("dogmaAttributes", []string{
		"attributeID", "attributeName", "description", "iconID", "defaultValue",
		"published", "displayName", "unitID", "stackable", "highIsGood",
	})

	MapSolarSystemsParser = NewJSONLParser[MapSolarSystem]("mapSolarSystems", []string{
		"solarSystemID", "solarSystemName", "regionID", "constellationID",
		"x", "y", "z", "security", "securityClass",
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
)

// RegisterParsers returns a map of all registered parsers keyed by table name.
// This function provides a central registry for all EVE SDE parsers.
func RegisterParsers() map[string]Parser {
	return map[string]Parser{
		"invTypes":             InvTypesParser,
		"invGroups":            InvGroupsParser,
		"industryBlueprints":   IndustryBlueprintsParser,
		"dogmaAttributes":      DogmaAttributesParser,
		"mapSolarSystems":      MapSolarSystemsParser,
		"dogmaEffects":         DogmaEffectsParser,
		"dogmaTypeAttributes":  DogmaTypeAttributesParser,
		"dogmaTypeEffects":     DogmaTypeEffectsParser,
		"mapRegions":           MapRegionsParser,
		"mapConstellations":    MapConstellationsParser,
	}
}
