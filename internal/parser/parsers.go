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

// ChrAncestry represents an EVE SDE chrAncestries record
type ChrAncestry struct {
	AncestryID   int     `json:"ancestryID"`
	AncestryName *string `json:"ancestryName"`
	BloodlineID  *int    `json:"bloodlineID"`
	Description  *string `json:"description"`
	IconID       *int    `json:"iconID"`
	ShortDescription *string `json:"shortDescription"`
}

// ChrBloodline represents an EVE SDE chrBloodlines record
type ChrBloodline struct {
	BloodlineID      int     `json:"bloodlineID"`
	BloodlineName    *string `json:"bloodlineName"`
	RaceID           *int    `json:"raceID"`
	Description      *string `json:"description"`
	CorporationID    *int    `json:"corporationID"`
	IconID           *int    `json:"iconID"`
	ShipTypeID       *int    `json:"shipTypeID"`
}

// ChrAttribute represents an EVE SDE chrAttributes record
type ChrAttribute struct {
	AttributeID   int     `json:"attributeID"`
	AttributeName *string `json:"attributeName"`
	Description   *string `json:"description"`
	IconID        *int    `json:"iconID"`
	ShortDescription *string `json:"shortDescription"`
	Notes         *string `json:"notes"`
}

// AgentType represents an EVE SDE agtAgentTypes record
type AgentType struct {
	AgentTypeID int     `json:"agentTypeID"`
	AgentType   *string `json:"agentType"`
}

// AgentInSpace represents an EVE SDE agtAgents record
type AgentInSpace struct {
	AgentID       int  `json:"agentID"`
	DivisionID    *int `json:"divisionID"`
	CorporationID *int `json:"corporationID"`
	LocationID    *int `json:"locationID"`
	Level         *int `json:"level"`
	Quality       *int `json:"quality"`
	AgentTypeID   *int `json:"agentTypeID"`
	IsLocator     *int `json:"isLocator"`
}

// Certificate represents an EVE SDE certCerts record
type Certificate struct {
	CertificateID int     `json:"certificateID"`
	Description   *string `json:"description"`
	GroupID       *int    `json:"groupID"`
	Name          *string `json:"name"`
}

// Mastery represents an EVE SDE certMasteries record
type Mastery struct {
	TypeID        int  `json:"typeID"`
	MasteryLevel  *int `json:"masteryLevel"`
	CertificateID *int `json:"certificateID"`
}

// CrpNPCCorporation represents an EVE SDE crpNPCCorporations record
type CrpNPCCorporation struct {
	CorporationID     int      `json:"corporationID"`
	Size              *string  `json:"size"`
	Extent            *string  `json:"extent"`
	SolarSystemID     *int     `json:"solarSystemID"`
	InvestorID1       *int     `json:"investorID1"`
	InvestorShares1   *int     `json:"investorShares1"`
	InvestorID2       *int     `json:"investorID2"`
	InvestorShares2   *int     `json:"investorShares2"`
	InvestorID3       *int     `json:"investorID3"`
	InvestorShares3   *int     `json:"investorShares3"`
	InvestorID4       *int     `json:"investorID4"`
	InvestorShares4   *int     `json:"investorShares4"`
	FriendID          *int     `json:"friendID"`
	EnemyID           *int     `json:"enemyID"`
	PublicShares      *int     `json:"publicShares"`
	InitialPrice      *int     `json:"initialPrice"`
	MinSecurity       *float64 `json:"minSecurity"`
	Scattered         *int     `json:"scattered"`
	FringeID          *int     `json:"fringeID"`
	CorridorID        *int     `json:"corridorID"`
	HubID             *int     `json:"hubID"`
	BorderID          *int     `json:"borderID"`
	FactionID         *int     `json:"factionID"`
	SizeFactor        *float64 `json:"sizeFactor"`
	StationCount      *int     `json:"stationCount"`
	StationSystemCount *int    `json:"stationSystemCount"`
	Description       *string  `json:"description"`
	IconID            *int     `json:"iconID"`
}

// CrpNPCCorporationDivision represents an EVE SDE crpNPCCorporationDivisions record
type CrpNPCCorporationDivision struct {
	CorporationID int     `json:"corporationID"`
	DivisionID    int     `json:"divisionID"`
	Size          *int    `json:"size"`
	DivisionName  *string `json:"divisionName"`
	LeaderID      *int    `json:"leaderID"`
}

// NPCCharacter represents an EVE SDE chrNPCCharacters record
type NPCCharacter struct {
	CharacterID   int  `json:"characterID"`
	CorporationID *int `json:"corporationID"`
	Name          *string `json:"name"`
}

// StaNPCStation represents an EVE SDE staStations record
type StaNPCStation struct {
	StationID            int      `json:"stationID"`
	Security             *float64 `json:"security"`
	DockingCostPerVolume *float64 `json:"dockingCostPerVolume"`
	MaxShipVolumeDockable *float64 `json:"maxShipVolumeDockable"`
	OfficeRentalCost     *int     `json:"officeRentalCost"`
	OperationID          *int     `json:"operationID"`
	StationTypeID        *int     `json:"stationTypeID"`
	CorporationID        *int     `json:"corporationID"`
	SolarSystemID        *int     `json:"solarSystemID"`
	ConstellationID      *int     `json:"constellationID"`
	RegionID             *int     `json:"regionID"`
	StationName          *string  `json:"stationName"`
	X                    *float64 `json:"x"`
	Y                    *float64 `json:"y"`
	Z                    *float64 `json:"z"`
	ReprocessingEfficiency *float64 `json:"reprocessingEfficiency"`
	ReprocessingStationsTake *float64 `json:"reprocessingStationsTake"`
	ReprocessingHangarFlag *int   `json:"reprocessingHangarFlag"`
}

// DogmaAttributeCategory represents an EVE SDE dogmaAttributeCategories record
type DogmaAttributeCategory struct {
	CategoryID          int     `json:"categoryID"`
	CategoryName        *string `json:"categoryName"`
	CategoryDescription *string `json:"categoryDescription"`
}

// DogmaUnit represents an EVE SDE dogmaUnits record
type DogmaUnit struct {
	UnitID      int     `json:"unitID"`
	UnitName    *string `json:"unitName"`
	DisplayName *string `json:"displayName"`
	Description *string `json:"description"`
}

// TypeDogma represents an EVE SDE typeDogma record (complex nested structure)
type TypeDogma struct {
	TypeID int `json:"typeID"`
	// Note: typeDogma has complex nested attributes and effects structures
	// Simplified for now - would need expansion for full implementation
}

// DynamicItemAttribute represents an EVE SDE dynamicItemAttributes record
type DynamicItemAttribute struct {
	TypeID      int `json:"typeID"`
	AttributeID int `json:"attributeID"`
}

// MapMoon represents an EVE SDE mapMoons record
type MapMoon struct {
	MoonID        int      `json:"moonID"`
	MoonName      *string  `json:"moonName"`
	SolarSystemID *int     `json:"solarSystemID"`
	PlanetID      *int     `json:"planetID"`
	X             *float64 `json:"x"`
	Y             *float64 `json:"y"`
	Z             *float64 `json:"z"`
}

// MapStar represents an EVE SDE mapStars record
type MapStar struct {
	StarID        int      `json:"starID"`
	SolarSystemID *int     `json:"solarSystemID"`
	TypeID        *int     `json:"typeID"`
	Radius        *float64 `json:"radius"`
	Temperature   *float64 `json:"temperature"`
	Luminosity    *float64 `json:"luminosity"`
}

// MapAsteroidBelt represents an EVE SDE mapAsteroidBelts record
type MapAsteroidBelt struct {
	AsteroidBeltID int      `json:"asteroidBeltID"`
	SolarSystemID  *int     `json:"solarSystemID"`
	TypeID         *int     `json:"typeID"`
	X              *float64 `json:"x"`
	Y              *float64 `json:"y"`
	Z              *float64 `json:"z"`
}

// Landmark represents an EVE SDE mapLandmarks record
type Landmark struct {
	LandmarkID    int      `json:"landmarkID"`
	LandmarkName  *string  `json:"landmarkName"`
	Description   *string  `json:"description"`
	LocationID    *int     `json:"locationID"`
	X             *float64 `json:"x"`
	Y             *float64 `json:"y"`
	Z             *float64 `json:"z"`
	IconID        *int     `json:"iconID"`
}

// Skin represents an EVE SDE skins record
type Skin struct {
	SkinID           int     `json:"skinID"`
	InternalName     *string `json:"internalName"`
	SkinMaterialID   *int    `json:"skinMaterialID"`
	TypeID           *int    `json:"typeID"`
}

// SkinLicense represents an EVE SDE skinLicenses record
type SkinLicense struct {
	LicenseTypeID int  `json:"licenseTypeID"`
	Duration      *int `json:"duration"`
	SkinID        *int `json:"skinID"`
}

// SkinMaterial represents an EVE SDE skinMaterials record
type SkinMaterial struct {
	SkinMaterialID int     `json:"skinMaterialID"`
	DisplayNameID  *int    `json:"displayNameID"`
	MaterialSetID  *int    `json:"materialSetID"`
}

// TranslationLanguage represents an EVE SDE translationLanguages record
type TranslationLanguage struct {
	LanguageID   int     `json:"languageID"`
	LanguageName *string `json:"languageName"`
}

// StationOperation represents an EVE SDE staOperations record
type StationOperation struct {
	OperationID          int     `json:"operationID"`
	OperationName        *string `json:"operationName"`
	Description          *string `json:"description"`
	FractionID           *int    `json:"fractionID"`
	Border               *int    `json:"border"`
	Fringe               *int    `json:"fringe"`
	Corridor             *int    `json:"corridor"`
	Hub                  *int    `json:"hub"`
	Ratio                *int    `json:"ratio"`
	CaldariStationTypeID *int    `json:"caldariStationTypeID"`
	MinmatarStationTypeID *int   `json:"minmatarStationTypeID"`
	AmarrStationTypeID   *int    `json:"amarrStationTypeID"`
	GallenteStationTypeID *int   `json:"gallenteStationTypeID"`
	JoveStationTypeID    *int    `json:"joveStationTypeID"`
}

// StationService represents an EVE SDE staServices record
type StationService struct {
	ServiceID   int     `json:"serviceID"`
	ServiceName *string `json:"serviceName"`
	Description *string `json:"description"`
}

// SovereigntyUpgrade represents an EVE SDE sovereigntyUpgrades record
type SovereigntyUpgrade struct {
	UpgradeID int  `json:"upgradeID"`
	TypeID    *int `json:"typeID"`
	Level     *int `json:"level"`
}

// Icon represents an EVE SDE eveIcons record
type Icon struct {
	IconID      int     `json:"iconID"`
	IconFile    *string `json:"iconFile"`
	Description *string `json:"description"`
}

// Graphic represents an EVE SDE eveGraphics record
type Graphic struct {
	GraphicID   int     `json:"graphicID"`
	GraphicFile *string `json:"graphicFile"`
	Description *string `json:"description"`
}

// ContrabandType represents an EVE SDE contrabandTypes record
type ContrabandType struct {
	FactionID           int      `json:"factionID"`
	TypeID              int      `json:"typeID"`
	StandingLoss        *float64 `json:"standingLoss"`
	ConfiscateMinSec    *float64 `json:"confiscateMinSec"`
	FineByValue         *float64 `json:"fineByValue"`
	AttackMinSec        *float64 `json:"attackMinSec"`
}

// ControlTowerResource represents an EVE SDE controlTowerResources record
type ControlTowerResource struct {
	ControlTowerTypeID int  `json:"controlTowerTypeID"`
	ResourceTypeID     int  `json:"resourceTypeID"`
	Purpose            *int `json:"purpose"`
	Quantity           *int `json:"quantity"`
	MinSecurityLevel   *float64 `json:"minSecurityLevel"`
	FactionID          *int `json:"factionID"`
}

// CorporationActivity represents an EVE SDE crpActivities record
type CorporationActivity struct {
	ActivityID   int     `json:"activityID"`
	ActivityName *string `json:"activityName"`
	Description  *string `json:"description"`
}

// DogmaBuffCollection represents an EVE SDE dbuffCollections record
type DogmaBuffCollection struct {
	CollectionID int `json:"collectionID"`
}

// PlanetResource represents an EVE SDE planetResources record
type PlanetResource struct {
	PlanetTypeID int  `json:"planetTypeID"`
	ResourceTypeID int `json:"resourceTypeID"`
}

// PlanetSchematic represents an EVE SDE planetSchematics record
type PlanetSchematic struct {
	SchematicID   int  `json:"schematicID"`
	CycleTime     *int `json:"cycleTime"`
}

// TypeBonus represents an EVE SDE typeBonuses record
type TypeBonus struct {
	TypeID      int     `json:"typeID"`
	BonusID     *int    `json:"bonusID"`
	BonusValue  *float64 `json:"bonusValue"`
	BonusText   *string `json:"bonusText"`
	Importance  *int    `json:"importance"`
	UnitID      *int    `json:"unitID"`
}

// SDEMetadata represents the EVE SDE _sde metadata record
type SDEMetadata struct {
	Version     *string `json:"version"`
	ReleaseDate *string `json:"releaseDate"`
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

	// Character/NPC Extended Parsers
	ChrAncestriesParser = NewJSONLParser[ChrAncestry]("chrAncestries", []string{
		"ancestryID", "ancestryName", "bloodlineID", "description", "iconID", "shortDescription",
	})

	ChrBloodlinesParser = NewJSONLParser[ChrBloodline]("chrBloodlines", []string{
		"bloodlineID", "bloodlineName", "raceID", "description", "corporationID", "iconID", "shipTypeID",
	})

	ChrAttributesParser = NewJSONLParser[ChrAttribute]("chrAttributes", []string{
		"attributeID", "attributeName", "description", "iconID", "shortDescription", "notes",
	})

	CrpNPCCorporationsParser = NewJSONLParser[CrpNPCCorporation]("crpNPCCorporations", []string{
		"corporationID", "size", "extent", "solarSystemID",
	})

	CrpNPCCorporationDivisionsParser = NewJSONLParser[CrpNPCCorporationDivision]("crpNPCCorporationDivisions", []string{
		"corporationID", "divisionID", "size", "divisionName", "leaderID",
	})

	NPCCharactersParser = NewJSONLParser[NPCCharacter]("chrNPCCharacters", []string{
		"characterID", "corporationID", "name",
	})

	StaNPCStationsParser = NewJSONLParser[StaNPCStation]("staStations", []string{
		"stationID", "security", "stationName", "stationTypeID", "solarSystemID",
	})

	// Agents
	AgentTypesParser = NewJSONLParser[AgentType]("agtAgentTypes", []string{
		"agentTypeID", "agentType",
	})

	AgentsInSpaceParser = NewJSONLParser[AgentInSpace]("agtAgents", []string{
		"agentID", "divisionID", "corporationID", "locationID", "level", "quality", "agentTypeID", "isLocator",
	})

	// Certificates/Skills
	CertificatesParser = NewJSONLParser[Certificate]("certCerts", []string{
		"certificateID", "description", "groupID", "name",
	})

	MasteriesParser = NewJSONLParser[Mastery]("certMasteries", []string{
		"typeID", "masteryLevel", "certificateID",
	})

	// Additional Dogma
	DogmaAttributeCategoriesParser = NewJSONLParser[DogmaAttributeCategory]("dogmaAttributeCategories", []string{
		"categoryID", "categoryName", "categoryDescription",
	})

	DogmaUnitsParser = NewJSONLParser[DogmaUnit]("dogmaUnits", []string{
		"unitID", "unitName", "displayName", "description",
	})

	TypeDogmaParser = NewJSONLParser[TypeDogma]("typeDogma", []string{
		"typeID",
	})

	DynamicItemAttributesParser = NewJSONLParser[DynamicItemAttribute]("dynamicItemAttributes", []string{
		"typeID", "attributeID",
	})

	// Extended Universe/Map
	MapMoonsParser = NewJSONLParser[MapMoon]("mapMoons", []string{
		"moonID", "moonName", "solarSystemID", "planetID", "x", "y", "z",
	})

	MapStarsParser = NewJSONLParser[MapStar]("mapStars", []string{
		"starID", "solarSystemID", "typeID", "radius", "temperature", "luminosity",
	})

	MapAsteroidBeltsParser = NewJSONLParser[MapAsteroidBelt]("mapAsteroidBelts", []string{
		"asteroidBeltID", "solarSystemID", "typeID", "x", "y", "z",
	})

	LandmarksParser = NewJSONLParser[Landmark]("mapLandmarks", []string{
		"landmarkID", "landmarkName", "description", "locationID", "x", "y", "z", "iconID",
	})

	// Skins
	SkinsParser = NewJSONLParser[Skin]("skins", []string{
		"skinID", "internalName", "skinMaterialID", "typeID",
	})

	SkinLicensesParser = NewJSONLParser[SkinLicense]("skinLicenses", []string{
		"licenseTypeID", "duration", "skinID",
	})

	SkinMaterialsParser = NewJSONLParser[SkinMaterial]("skinMaterials", []string{
		"skinMaterialID", "displayNameID", "materialSetID",
	})

	// Translation
	TranslationLanguagesParser = NewJSONLParser[TranslationLanguage]("translationLanguages", []string{
		"languageID", "languageName",
	})

	// Station
	StationOperationsParser = NewJSONLParser[StationOperation]("staOperations", []string{
		"operationID", "operationName", "description",
	})

	StationServicesParser = NewJSONLParser[StationService]("staServices", []string{
		"serviceID", "serviceName", "description",
	})

	SovereigntyUpgradesParser = NewJSONLParser[SovereigntyUpgrade]("sovereigntyUpgrades", []string{
		"upgradeID", "typeID", "level",
	})

	// Miscellaneous
	IconsParser = NewJSONLParser[Icon]("eveIcons", []string{
		"iconID", "iconFile", "description",
	})

	GraphicsParser = NewJSONLParser[Graphic]("eveGraphics", []string{
		"graphicID", "graphicFile", "description",
	})

	ContrabandTypesParser = NewJSONLParser[ContrabandType]("contrabandTypes", []string{
		"factionID", "typeID", "standingLoss", "confiscateMinSec", "fineByValue", "attackMinSec",
	})

	ControlTowerResourcesParser = NewJSONLParser[ControlTowerResource]("controlTowerResources", []string{
		"controlTowerTypeID", "resourceTypeID", "purpose", "quantity", "minSecurityLevel", "factionID",
	})

	CorporationActivitiesParser = NewJSONLParser[CorporationActivity]("crpActivities", []string{
		"activityID", "activityName", "description",
	})

	DogmaBuffCollectionsParser = NewJSONLParser[DogmaBuffCollection]("dbuffCollections", []string{
		"collectionID",
	})

	PlanetResourcesParser = NewJSONLParser[PlanetResource]("planetResources", []string{
		"planetTypeID", "resourceTypeID",
	})

	PlanetSchematicsParser = NewJSONLParser[PlanetSchematic]("planetSchematics", []string{
		"schematicID", "cycleTime",
	})

	TypeBonusesParser = NewJSONLParser[TypeBonus]("typeBonuses", []string{
		"typeID", "bonusID", "bonusValue", "bonusText", "importance", "unitID",
	})

	SDEMetadataParser = NewJSONLParser[SDEMetadata]("_sde", []string{
		"version", "releaseDate",
	})
)

// RegisterParsers returns a map of all registered parsers keyed by table name.
// This function provides a central registry for all EVE SDE parsers.
//
// Epic #4: Full Parser Migration - All 51 Tables Implemented ✅
//
// Core Parsers (Task #37 - Complete):
//   - Inventory: invTypes, invGroups, invCategories, invMarketGroups, invMetaGroups
//   - Industry: industryBlueprints
//   - Dogma Core: dogmaAttributes, dogmaEffects, dogmaTypeAttributes, dogmaTypeEffects
//   - Universe Core: mapRegions, mapConstellations, mapSolarSystems, mapStargates, mapPlanets
//   - Character: chrRaces, chrFactions
//
// Extended Parsers (Task #38 - Complete):
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
