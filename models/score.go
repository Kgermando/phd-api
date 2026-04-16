package models

import (
	"time"

	"gorm.io/gorm"
)

// Score représente la grille de scoring d'un producteur riziculteur (total /100)
// Seuil recommandé : 60/100
type Score struct {
	UUID      string `gorm:"type:varchar(255);primary_key" json:"uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	ProducerUUID string   `gorm:"type:varchar(255);not null" json:"producer_uuid"`
	Producer     Producer `gorm:"foreignKey:ProducerUUID;constraint:OnDelete:CASCADE" json:"producer"`

	// 1. Superficie cultivée — max 10
	// Priorité aux exploitations économiquement viables
	SuperficieCultivee int `gorm:"default:0;check:superficie_cultivee >= 0 AND superficie_cultivee <= 10" json:"superficie_cultivee"`

	// 2. Expérience en riziculture — max 10
	// Nombre d'années de pratique
	ExperienceRiziculture int `gorm:"default:0;check:experience_riziculture >= 0 AND experience_riziculture <= 10" json:"experience_riziculture"`

	// 3. Statut foncier sécurisé — max 10
	// Propriétaire ou droit d'exploitation stable
	StatutFoncierSecurise int `gorm:"default:0;check:statut_foncier_securise >= 0 AND statut_foncier_securise <= 10" json:"statut_foncier_securise"`

	// 4. Accès à l'eau — max 10
	// Irrigation fiable ou aménagement hydraulique
	AccesEau int `gorm:"default:0;check:acces_eau >= 0 AND acces_eau <= 10" json:"acces_eau"`

	// 5. Respect des itinéraires techniques — max 10
	// Calendrier agricole, bonnes semences, application conseil
	RespectItinerairesTechniques int `gorm:"default:0;check:respect_itineraires_techniques >= 0 AND respect_itineraires_techniques <= 10" json:"respect_itineraires_techniques"`

	// 6. Pratiques environnementales — max 10
	// Rotation, compost, gestion durable
	PratiquesEnvironnementales int `gorm:"default:0;check:pratiques_environnementales >= 0 AND pratiques_environnementales <= 10" json:"pratiques_environnementales"`

	// 7. Vulnérabilité climatique — max 10
	// Producteurs à risque mais capables de s'adapter
	VulnerabiliteClimatique int `gorm:"default:0;check:vulnerabilite_climatique >= 0 AND vulnerabilite_climatique <= 10" json:"vulnerabilite_climatique"`

	// 8. Organisation / Coopérative — max 5
	// Membre actif d'une organisation agricole
	OrganisationCooperative int `gorm:"default:0;check:organisation_cooperative >= 0 AND organisation_cooperative <= 5" json:"organisation_cooperative"`

	// 9. Capacité de production — max 10
	// Rendement et potentiel d'amélioration
	CapaciteProduction int `gorm:"default:0;check:capacite_production >= 0 AND capacite_production <= 10" json:"capacite_production"`

	// 10. Motivation / Engagement — max 10
	// Disponibilité pour formations et respect des exigences projet
	MotivationEngagement int `gorm:"default:0;check:motivation_engagement >= 0 AND motivation_engagement <= 10" json:"motivation_engagement"`

	// 11. Inclusion sociale — max 5
	// Jeunes, femmes, groupes vulnérables
	InclusionSociale int `gorm:"default:0;check:inclusion_sociale >= 0 AND inclusion_sociale <= 5" json:"inclusion_sociale"`

	// Calculé automatiquement (total /100)
	ScoreTotal int  `gorm:"default:0" json:"score_total"`
	Recommande bool `gorm:"default:false" json:"recommande"` // true si ScoreTotal >= 60
}

// CalculateScore calcule le total et détermine si le producteur est recommandé (seuil : 60/100)
func (s *Score) CalculateScore() {
	s.ScoreTotal = s.SuperficieCultivee +
		s.ExperienceRiziculture +
		s.StatutFoncierSecurise +
		s.AccesEau +
		s.RespectItinerairesTechniques +
		s.PratiquesEnvironnementales +
		s.VulnerabiliteClimatique +
		s.OrganisationCooperative +
		s.CapaciteProduction +
		s.MotivationEngagement +
		s.InclusionSociale
	s.Recommande = s.ScoreTotal >= 60
}
