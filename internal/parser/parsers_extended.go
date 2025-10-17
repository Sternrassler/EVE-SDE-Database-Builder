// Package parser provides JSONL parser instances for EVE SDE tables.
// This file contains extended parsers beyond the core 17 tables.
package parser

// ChrAncestry represents an EVE SDE chrAncestries record
type ChrAncestry struct {
	AncestryID       int     `json:"ancestryID"`
	AncestryName     *string `json:"ancestryName"`
	BloodlineID      *int    `json:"bloodlineID"`
	Description      *string `json:"description"`
	IconID           *int    `json:"iconID"`
	ShortDescription *string `json:"shortDescription"`
}

// ChrBloodline represents an EVE SDE chrBloodlines record
type ChrBloodline struct {
	BloodlineID   int     `json:"bloodlineID"`
	BloodlineName *string `json:"bloodlineName"`
	RaceID        *int    `json:"raceID"`
	Description   *string `json:"description"`
	CorporationID *int    `json:"corporationID"`
	IconID        *int    `json:"iconID"`
	ShipTypeID    *int    `json:"shipTypeID"`
}

// ChrAttribute represents an EVE SDE chrAttributes record
type ChrAttribute struct {
	AttributeID      int     `json:"attributeID"`
	AttributeName    *string `json:"attributeName"`
	Description      *string `json:"description"`
	IconID           *int    `json:"iconID"`
	ShortDescription *string `json:"shortDescription"`
	Notes            *string `json:"notes"`
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
	CorporationID      int      `json:"corporationID"`
	Size               *string  `json:"size"`
	Extent             *string  `json:"extent"`
	SolarSystemID      *int     `json:"solarSystemID"`
	InvestorID1        *int     `json:"investorID1"`
	InvestorShares1    *int     `json:"investorShares1"`
	InvestorID2        *int     `json:"investorID2"`
	InvestorShares2    *int     `json:"investorShares2"`
	InvestorID3        *int     `json:"investorID3"`
	InvestorShares3    *int     `json:"investorShares3"`
	InvestorID4        *int     `json:"investorID4"`
	InvestorShares4    *int     `json:"investorShares4"`
	FriendID           *int     `json:"friendID"`
	EnemyID            *int     `json:"enemyID"`
	PublicShares       *int     `json:"publicShares"`
	InitialPrice       *int     `json:"initialPrice"`
	MinSecurity        *float64 `json:"minSecurity"`
	Scattered          *int     `json:"scattered"`
	FringeID           *int     `json:"fringeID"`
	CorridorID         *int     `json:"corridorID"`
	HubID              *int     `json:"hubID"`
	BorderID           *int     `json:"borderID"`
	FactionID          *int     `json:"factionID"`
	SizeFactor         *float64 `json:"sizeFactor"`
	StationCount       *int     `json:"stationCount"`
	StationSystemCount *int     `json:"stationSystemCount"`
	Description        *string  `json:"description"`
	IconID             *int     `json:"iconID"`
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
	CharacterID   int     `json:"characterID"`
	CorporationID *int    `json:"corporationID"`
	Name          *string `json:"name"`
}

// StaNPCStation represents an EVE SDE staStations record
type StaNPCStation struct {
	StationID                int      `json:"stationID"`
	Security                 *float64 `json:"security"`
	DockingCostPerVolume     *float64 `json:"dockingCostPerVolume"`
	MaxShipVolumeDockable    *float64 `json:"maxShipVolumeDockable"`
	OfficeRentalCost         *int     `json:"officeRentalCost"`
	OperationID              *int     `json:"operationID"`
	StationTypeID            *int     `json:"stationTypeID"`
	CorporationID            *int     `json:"corporationID"`
	SolarSystemID            *int     `json:"solarSystemID"`
	ConstellationID          *int     `json:"constellationID"`
	RegionID                 *int     `json:"regionID"`
	StationName              *string  `json:"stationName"`
	X                        *float64 `json:"x"`
	Y                        *float64 `json:"y"`
	Z                        *float64 `json:"z"`
	ReprocessingEfficiency   *float64 `json:"reprocessingEfficiency"`
	ReprocessingStationsTake *float64 `json:"reprocessingStationsTake"`
	ReprocessingHangarFlag   *int     `json:"reprocessingHangarFlag"`
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
	LandmarkID   int      `json:"landmarkID"`
	LandmarkName *string  `json:"landmarkName"`
	Description  *string  `json:"description"`
	LocationID   *int     `json:"locationID"`
	X            *float64 `json:"x"`
	Y            *float64 `json:"y"`
	Z            *float64 `json:"z"`
	IconID       *int     `json:"iconID"`
}

// Skin represents an EVE SDE skins record
type Skin struct {
	SkinID         int     `json:"skinID"`
	InternalName   *string `json:"internalName"`
	SkinMaterialID *int    `json:"skinMaterialID"`
	TypeID         *int    `json:"typeID"`
}

// SkinLicense represents an EVE SDE skinLicenses record
type SkinLicense struct {
	LicenseTypeID int  `json:"licenseTypeID"`
	Duration      *int `json:"duration"`
	SkinID        *int `json:"skinID"`
}

// SkinMaterial represents an EVE SDE skinMaterials record
type SkinMaterial struct {
	SkinMaterialID int  `json:"skinMaterialID"`
	DisplayNameID  *int `json:"displayNameID"`
	MaterialSetID  *int `json:"materialSetID"`
}

// TranslationLanguage represents an EVE SDE translationLanguages record
type TranslationLanguage struct {
	LanguageID   int     `json:"languageID"`
	LanguageName *string `json:"languageName"`
}

// StationOperation represents an EVE SDE staOperations record
type StationOperation struct {
	OperationID           int     `json:"operationID"`
	OperationName         *string `json:"operationName"`
	Description           *string `json:"description"`
	FractionID            *int    `json:"fractionID"`
	Border                *int    `json:"border"`
	Fringe                *int    `json:"fringe"`
	Corridor              *int    `json:"corridor"`
	Hub                   *int    `json:"hub"`
	Ratio                 *int    `json:"ratio"`
	CaldariStationTypeID  *int    `json:"caldariStationTypeID"`
	MinmatarStationTypeID *int    `json:"minmatarStationTypeID"`
	AmarrStationTypeID    *int    `json:"amarrStationTypeID"`
	GallenteStationTypeID *int    `json:"gallenteStationTypeID"`
	JoveStationTypeID     *int    `json:"joveStationTypeID"`
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
	FactionID        int      `json:"factionID"`
	TypeID           int      `json:"typeID"`
	StandingLoss     *float64 `json:"standingLoss"`
	ConfiscateMinSec *float64 `json:"confiscateMinSec"`
	FineByValue      *float64 `json:"fineByValue"`
	AttackMinSec     *float64 `json:"attackMinSec"`
}

// ControlTowerResource represents an EVE SDE controlTowerResources record
type ControlTowerResource struct {
	ControlTowerTypeID int      `json:"controlTowerTypeID"`
	ResourceTypeID     int      `json:"resourceTypeID"`
	Purpose            *int     `json:"purpose"`
	Quantity           *int     `json:"quantity"`
	MinSecurityLevel   *float64 `json:"minSecurityLevel"`
	FactionID          *int     `json:"factionID"`
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
	PlanetTypeID   int `json:"planetTypeID"`
	ResourceTypeID int `json:"resourceTypeID"`
}

// PlanetSchematic represents an EVE SDE planetSchematics record
type PlanetSchematic struct {
	SchematicID int  `json:"schematicID"`
	CycleTime   *int `json:"cycleTime"`
}

// TypeBonus represents an EVE SDE typeBonuses record
type TypeBonus struct {
	TypeID     int      `json:"typeID"`
	BonusID    *int     `json:"bonusID"`
	BonusValue *float64 `json:"bonusValue"`
	BonusText  *string  `json:"bonusText"`
	Importance *int     `json:"importance"`
	UnitID     *int     `json:"unitID"`
}

// SDEMetadata represents the EVE SDE _sde metadata record
type SDEMetadata struct {
	Version     *string `json:"version"`
	ReleaseDate *string `json:"releaseDate"`
}

// Extended parser instances for EVE SDE tables (beyond core 17 tables)
var (
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
