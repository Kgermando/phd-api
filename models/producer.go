package models

import (
	"time"

	"gorm.io/gorm"
)

// Champs représente un champ agricole
type Champs struct {
	UUID            string  `gorm:"type:varchar(255);primary_key" json:"uuid"`
	ProducerUUID    string  `gorm:"type:varchar(255);not null" json:"producer_uuid"`
	Localisation    string  `gorm:"not null" json:"localisation"`
	Superficie      float64 `gorm:"not null" json:"superficie"`       // en hectares
	TypeRiziculture string  `gorm:"not null" json:"type_riziculture"` // pluviale, irriguee, bas-fond
	Irrigation      bool    `json:"irrigation"`
	ModeAcces       string  `json:"mode_acces"` // voiture, velo, pied
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

// Producer représente un producteur agricole
type Producer struct {
	UUID      string `gorm:"type:varchar(255);primary_key" json:"uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserUUID string `gorm:"type:varchar(255);not null" json:"user_uuid"`
	User	 User   `gorm:"foreignKey:UserUUID;constraint:OnDelete:CASCADE" json:"user"`

	// Section 1 - Informations personnelles
	Nom           string    `gorm:"not null" json:"nom"`
	Sexe          string    `gorm:"not null" json:"sexe"` // homme, femme
	DateNaissance time.Time `gorm:"not null" json:"date_naissance"`
	Telephone     string    `gorm:"not null" json:"telephone"`
	Village       string    `gorm:"not null" json:"village"`
	Groupement    string    `json:"groupement"`

	// Section 2 - Statut foncier & expérience
	StatutFoncier     string `gorm:"not null" json:"statut_foncier"` // proprietaire, exploitant, metayer, autre
	AnneesExperience  int    `gorm:"not null" json:"annees_experience"`
	MembreCooperative bool   `json:"membre_cooperative"`
	NomCooperative    string `json:"nom_cooperative"`

	// Section 3 - Champs
	Champs []Champs `gorm:"foreignKey:ProducerUUID;constraint:OnDelete:CASCADE" json:"champs"`

	// Section 4 - Pratiques environnementales
	RotationCultures      bool   `json:"rotation_cultures"`
	UtilisationCompost    bool   `json:"utilisation_compost"`
	SignesDegradation     bool   `json:"signes_degradation"`
	SourceEau             string `json:"source_eau"` // pluie, fleuve, barrage, forage
	EconomieEau           bool   `json:"economie_eau"`
	ParcelleInondable     bool   `json:"parcelle_inondable"`
	UtilisationPesticides bool   `json:"utilisation_pesticides"`
	FormationPesticides   bool   `json:"formation_pesticides"`
	PresenceArbres        bool   `json:"presence_arbres"`
	ActiviteDeforestation bool   `json:"activite_deforestation"`
	BaiseFaune            bool   `json:"baise_faune"`

	// Section 5 - Risques climatiques
	PerteSec             bool   `json:"perte_sec"`
	PerteInondation      bool   `json:"perte_inondation"`
	PerteVents           bool   `json:"perte_vents"`
	StrategiesAdaptation string `gorm:"type:text" json:"strategies_adaptation"`

	// Section 6 - Production
	VarietesCultivees string  `gorm:"type:text" json:"varietes_cultivees"`
	RendementMoyen    float64 `json:"rendement_moyen"` // en tonnes/hectare
	CampagnesParAn    int     `json:"campagnes_par_an"`

	// Section 7 - Contraintes
	ManqueEau              bool   `json:"manque_eau"`
	IntrantsCouteux        bool   `json:"intrants_couteux"`
	AccesCredit            bool   `json:"acces_credit"`
	DegradationSols        bool   `json:"degradation_sols"`
	ChangementsClimatiques bool   `json:"changements_climatiques"`
	LieuVente              string `json:"lieu_vente"`

	// Section 8 - Besoins
	BesoinsPrioritaires string `gorm:"type:text" json:"besoins_prioritaires"`

	// Section 9 - Géolocalisation
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`

	Scores []Score `gorm:"foreignKey:ProducerUUID;constraint:OnDelete:CASCADE" json:"scores"`
}
