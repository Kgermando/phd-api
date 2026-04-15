package dashboard

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf"
	"github.com/kgermando/phd-api/database"
	"github.com/kgermando/phd-api/models"
	"github.com/xuri/excelize/v2"
)

// ============ Scoring Functions ============

type ScoreResult struct {
	Total                float64 `json:"total"`
	Environmental        float64 `json:"environmental"`
	Experience           float64 `json:"experience"`
	Production           float64 `json:"production"`
	RiskManagement       float64 `json:"risk_management"`
	AccessToResources    float64 `json:"access_to_resources"`
	InstitutionalSupport float64 `json:"institutional_support"`
}

// ScoreProducer calculates a producer's score based on their data
func ScoreProducer(p models.Producer) ScoreResult {
	result := ScoreResult{}

	// Environmental Score (25 points max)
	environmentalScore := 0.0
	if p.RotationCultures {
		environmentalScore += 4
	}
	if p.UtilisationCompost {
		environmentalScore += 4
	}
	if !p.SignesDegradation {
		environmentalScore += 3
	}
	if p.EconomieEau {
		environmentalScore += 4
	}
	if p.PresenceArbres {
		environmentalScore += 5
	}
	if !p.ActiviteDeforestation {
		environmentalScore += 5
	}
	result.Environmental = math.Min(environmentalScore, 25)

	// Experience Score (20 points max)
	experienceScore := 0.0
	if p.AnneesExperience > 0 {
		yearsBonus := math.Min(float64(p.AnneesExperience)*0.5, 10)
		experienceScore += yearsBonus
	}
	if p.StatutFoncier == "proprietaire" {
		experienceScore += 5
	} else if p.StatutFoncier == "exploitant" {
		experienceScore += 3
	}
	if p.MembreCooperative {
		experienceScore += 5
	}
	result.Experience = math.Min(experienceScore, 20)

	// Production Score (15 points max)
	productionScore := 0.0
	if p.RendementMoyen > 0 {
		renderScore := (p.RendementMoyen / 5) * 10
		productionScore += math.Min(renderScore, 10)
	}
	if p.CampagnesParAn >= 2 {
		productionScore += 5
	} else if p.CampagnesParAn == 1 {
		productionScore += 3
	}
	result.Production = math.Min(productionScore, 15)

	// Risk Management Score (15 points max)
	riskScore := 0.0
	if !p.PerteSec {
		riskScore += 3
	}
	if !p.PerteInondation {
		riskScore += 3
	}
	if !p.PerteVents {
		riskScore += 3
	}
	if !p.ParcelleInondable {
		riskScore += 3
	}
	result.RiskManagement = math.Min(riskScore, 15)

	// Access to Resources Score (15 points max)
	resourceScore := 0.0
	if p.SourceEau == "fleuve" || p.SourceEau == "barrage" || p.SourceEau == "forage" {
		resourceScore += 3
	} else if p.SourceEau == "pluie" {
		resourceScore += 1
	}
	if p.AccesCredit {
		resourceScore += 5
	}
	if !p.IntrantsCouteux {
		resourceScore += 4
	}
	if !p.ManqueEau {
		resourceScore += 3
	}
	result.AccessToResources = math.Min(resourceScore, 15)

	// Institutional Support Score (10 points max)
	institutionalScore := 0.0
	if p.MembreCooperative {
		institutionalScore += 4
	}
	if p.Sexe == "femme" {
		institutionalScore += 3
	}
	result.InstitutionalSupport = math.Min(institutionalScore, 10)

	// Total Score (max 100)
	result.Total = result.Environmental + result.Experience + result.Production +
		result.RiskManagement + result.AccessToResources + result.InstitutionalSupport

	if result.Total > 100 {
		result.Total = 100
	}
	if result.Total < 0 {
		result.Total = 0
	}

	return result
}

// ============ Response Structures ============

type ZoneStat struct {
	Zone      string  `json:"zone"`
	Count     int     `json:"count"`
	Eligible  int     `json:"eligible"`
	AvgScore  float64 `json:"avg_score"`
	Producers int     `json:"producers"`
}

type BarStat struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type PieSlice struct {
	Label string `json:"label"`
	Count int    `json:"count"`
	Pct   int    `json:"pct"`
	Color string `json:"color"`
	D     string `json:"d"`
}

type LineData struct {
	Month string `json:"month"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type UserStats struct {
	UserUUID             string     `json:"user_uuid"`
	UserFullname         string     `json:"user_fullname"`
	TotalProducers       int        `json:"total_producers"`
	EligibleProducers    int        `json:"eligible_producers"`
	NonEligibleProducers int        `json:"non_eligible_producers"`
	AvgScore             float64    `json:"avg_score"`
	CompletionRate       float64    `json:"completion_rate"`
	LastSurveyDate       *time.Time `json:"last_survey_date"`
}

type DashboardStats struct {
	Total            int               `json:"total"`
	Eligible         int               `json:"eligible"`
	NonEligible      int               `json:"non_eligible"`
	AvgScore         float64           `json:"avg_score"`
	Femmes           int               `json:"femmes"`
	Hommes           int               `json:"hommes"`
	Zones            []ZoneStat        `json:"zones"`
	MaxZoneCount     int               `json:"max_zone_count"`
	Recent           []models.Producer `json:"recent"`
	StatutFoncierPie []PieSlice        `json:"statut_foncier_pie"`
	SourceEauPie     []PieSlice        `json:"source_eau_pie"`
	TypeRizPie       []PieSlice        `json:"type_riz_pie"`
	CooperativePie   []PieSlice        `json:"cooperative_pie"`
	TrancheAgeStats  []BarStat         `json:"tranche_age_stats"`
	LineData         []LineData        `json:"line_data"`
}

// ============ Helper Functions ============

func buildPieSlices(data map[string]int, labels map[string]string, colors []string) []PieSlice {
	var result []PieSlice
	total := 0
	for _, count := range data {
		total += count
	}

	if total == 0 {
		return result
	}

	colorIdx := 0
	// Sort by count descending
	type kv struct {
		key   string
		value int
	}
	var sorted []kv
	for k, v := range data {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].value > sorted[j].value
	})

	for _, item := range sorted {
		key := item.key
		count := item.value

		label := labels[key]
		if label == "" {
			label = key
		}

		pct := 0
		if total > 0 {
			pct = (count * 100) / total
		}

		color := colors[colorIdx%len(colors)]
		d := generatePieSlicePath(count, total, colorIdx, len(data))

		result = append(result, PieSlice{
			Label: label,
			Count: count,
			Pct:   pct,
			Color: color,
			D:     d,
		})
		colorIdx++
	}

	return result
}

func generatePieSlicePath(count, total, index, totalSlices int) string {
	if total == 0 {
		return ""
	}

	const CX = 60
	const CY = 60
	const R = 54

	proportion := float64(count) / float64(total)

	// Calculate cumulative angle
	cumAngle := -math.Pi / 2
	angle := proportion * 2 * math.Pi

	x1 := CX + R*math.Cos(cumAngle)
	y1 := CY + R*math.Sin(cumAngle)

	cumAngle += angle
	x2 := CX + R*math.Cos(cumAngle)
	y2 := CY + R*math.Sin(cumAngle)

	largeArc := 0
	if angle > math.Pi {
		largeArc = 1
	}

	if proportion >= 1 {
		return fmt.Sprintf("M %.0f %.0f A %.0f %.0f 0 1 1 %.2f %.2f Z",
			CX, CY-R, R, R, CX-0.01, CY-R)
	}
	if proportion == 0 {
		return ""
	}

	return fmt.Sprintf("M %.0f %.0f L %.2f %.2f A %.0f %.0f 0 %d 1 %.2f %.2f Z",
		CX, CY, x1, y1, R, R, largeArc, x2, y2)
}

// ============ Main Dashboard Endpoint ============

// GetDashboardStats returns comprehensive dashboard statistics
func GetDashboardStats(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Preload("Champs").Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	type ScoredProducer struct {
		Producer models.Producer
		Score    float64
	}

	var scored []ScoredProducer
	for _, p := range producers {
		scoreResult := ScoreProducer(p)
		scored = append(scored, ScoredProducer{
			Producer: p,
			Score:    scoreResult.Total,
		})
	}

	// Basic statistics
	total := len(scored)
	eligible := 0
	femmes := 0
	hommes := 0
	totalScore := 0.0

	for _, s := range scored {
		if s.Score >= 60 {
			eligible++
		}
		if s.Producer.Sexe == "femme" {
			femmes++
		} else if s.Producer.Sexe == "homme" {
			hommes++
		}
		totalScore += s.Score
	}

	avgScore := 0.0
	if total > 0 {
		avgScore = math.Round((totalScore/float64(total))*100) / 100
	}

	// Zone statistics
	zoneMap := make(map[string]struct {
		Count    int
		Eligible int
		ScoreSum float64
	})

	for _, s := range scored {
		zone := s.Producer.Zone
		if zone == "" {
			zone = "Inconnue"
		}
		stats := zoneMap[zone]
		stats.Count++
		if s.Score >= 60 {
			stats.Eligible++
		}
		stats.ScoreSum += s.Score
		zoneMap[zone] = stats
	}

	var zones []ZoneStat
	maxZoneCount := 1
	for zone, stats := range zoneMap {
		avgZoneScore := 0.0
		if stats.Count > 0 {
			avgZoneScore = math.Round((stats.ScoreSum/float64(stats.Count))*100) / 100
		}
		zones = append(zones, ZoneStat{
			Zone:      zone,
			Count:     stats.Count,
			Eligible:  stats.Eligible,
			AvgScore:  avgZoneScore,
			Producers: stats.Count,
		})
		if stats.Count > maxZoneCount {
			maxZoneCount = stats.Count
		}
	}

	// Sort zones by count
	sort.Slice(zones, func(i, j int) bool {
		return zones[i].Count > zones[j].Count
	})

	// Recent producers (last 6)
	var recent []models.Producer
	db.Preload("Champs").Order("date_recensement DESC").Limit(6).Find(&recent)

	// Statut foncier pie
	sfMap := make(map[string]int)
	sfLabels := map[string]string{
		"proprietaire": "Propriétaire",
		"exploitant":   "Exploitant",
		"metayer":      "Métayer",
		"autre":        "Autre",
	}
	sfColors := []string{"#1565c0", "#42a5f5", "#29b6f6", "#90caf9"}

	for _, s := range scored {
		sfMap[s.Producer.StatutFoncier]++
	}
	statutFoncierPie := buildPieSlices(sfMap, sfLabels, sfColors)

	// Source d'eau pie
	seMap := make(map[string]int)
	seLabels := map[string]string{
		"pluie":   "Pluie",
		"fleuve":  "Fleuve",
		"barrage": "Barrage",
		"forage":  "Forage",
	}
	seColors := []string{"#0277bd", "#26c6da", "#00897b", "#80deea"}

	for _, s := range scored {
		seMap[s.Producer.SourceEau]++
	}
	sourceEauPie := buildPieSlices(seMap, seLabels, seColors)

	// Type de riziculture pie
	trMap := make(map[string]int)
	trLabels := map[string]string{
		"pluviale": "Pluviale",
		"irriguee": "Irriguée",
		"bas-fond": "Bas-fond",
	}
	trColors := []string{"#2e7d32", "#558b2f", "#aed581"}

	for _, s := range scored {
		for _, c := range s.Producer.Champs {
			trMap[c.TypeRiziculture]++
		}
	}
	typeRizPie := buildPieSlices(trMap, trLabels, trColors)

	// Coopérative pie
	coopMap := make(map[string]int)
	coopLabels := map[string]string{
		"member":     "Membre",
		"non-member": "Non-membre",
	}
	coopColors := []string{"#43a047", "#e53935"}

	coopCount := 0
	for _, s := range scored {
		if s.Producer.MembreCooperative {
			coopCount++
		}
	}
	coopMap["member"] = coopCount
	coopMap["non-member"] = total - coopCount
	cooperativePie := buildPieSlices(coopMap, coopLabels, coopColors)

	// Tranche d'âge bar chart
	refYear := time.Now().Year()
	ageRanges := make(map[string]int)
	ageRanges["< 25 ans"] = 0
	ageRanges["25–35 ans"] = 0
	ageRanges["36–45 ans"] = 0
	ageRanges["46–55 ans"] = 0
	ageRanges["> 55 ans"] = 0

	for _, s := range scored {
		yearStr := s.Producer.DateNaissance.Format("2006")
		year := 1980
		if yearStr != "" {
			year, _ = strconv.Atoi(yearStr)
		}
		age := refYear - year

		if age < 25 {
			ageRanges["< 25 ans"]++
		} else if age >= 25 && age <= 35 {
			ageRanges["25–35 ans"]++
		} else if age >= 36 && age <= 45 {
			ageRanges["36–45 ans"]++
		} else if age >= 46 && age <= 55 {
			ageRanges["46–55 ans"]++
		} else {
			ageRanges["> 55 ans"]++
		}
	}

	trancheAgeStats := []BarStat{
		{Label: "< 25 ans", Count: ageRanges["< 25 ans"]},
		{Label: "25–35 ans", Count: ageRanges["25–35 ans"]},
		{Label: "36–45 ans", Count: ageRanges["36–45 ans"]},
		{Label: "46–55 ans", Count: ageRanges["46–55 ans"]},
		{Label: "> 55 ans", Count: ageRanges["> 55 ans"]},
	}

	// Line chart: Monthly evolution
	monthMap := make(map[string]int)
	for _, s := range scored {
		month := s.Producer.DateRecensement.Format("2006-01")
		monthMap[month]++
	}

	var lineData []LineData
	var monthsOrder []string
	for month := range monthMap {
		monthsOrder = append(monthsOrder, month)
	}

	// Sort by date
	sort.Strings(monthsOrder)

	monthsFR := []string{"Jan", "Fév", "Mar", "Avr", "Mai", "Juin", "Juil", "Aoû", "Sep", "Oct", "Nov", "Déc"}

	for _, month := range monthsOrder {
		if len(month) == 7 {
			year := month[:4]
			monthIdx := month[5:7]
			monthNum := 0
			if m, err := strconv.Atoi(monthIdx); err == nil && m >= 1 && m <= 12 {
				monthNum = m - 1
			}
			label := monthsFR[monthNum] + " " + year
			lineData = append(lineData, LineData{
				Month: month,
				Label: label,
				Count: monthMap[month],
			})
		}
	}

	stats := DashboardStats{
		Total:            total,
		Eligible:         eligible,
		NonEligible:      total - eligible,
		AvgScore:         avgScore,
		Femmes:           femmes,
		Hommes:           hommes,
		Zones:            zones,
		MaxZoneCount:     maxZoneCount,
		Recent:           recent,
		StatutFoncierPie: statutFoncierPie,
		SourceEauPie:     sourceEauPie,
		TypeRizPie:       typeRizPie,
		CooperativePie:   cooperativePie,
		TrancheAgeStats:  trancheAgeStats,
		LineData:         lineData,
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data":   stats,
	})
}

// ============ User Performance Endpoint ============

// GetUserPerformance returns performance metrics per user/agent
func GetUserPerformance(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	userMap := make(map[string]struct {
		TotalProducers    int
		EligibleProducers int
		TotalScore        float64
		LastSurveyDate    *time.Time
	})

	for _, p := range producers {
		scoreResult := ScoreProducer(p)
		stats := userMap[p.AgentRecenseurUUID]
		stats.TotalProducers++
		if scoreResult.Total >= 60 {
			stats.EligibleProducers++
		}
		stats.TotalScore += scoreResult.Total
		if stats.LastSurveyDate == nil || p.DateRecensement.After(*stats.LastSurveyDate) {
			stats.LastSurveyDate = &p.DateRecensement
		}
		userMap[p.AgentRecenseurUUID] = stats
	}

	// Fetch user details
	var users []models.User
	db.Find(&users)

	var result []UserStats
	for _, u := range users {
		if stats, exists := userMap[u.UUID]; exists {
			avgScore := 0.0
			completionRate := 0.0
			if stats.TotalProducers > 0 {
				avgScore = math.Round((stats.TotalScore/float64(stats.TotalProducers))*100) / 100
				completionRate = (float64(stats.EligibleProducers) / float64(stats.TotalProducers)) * 100
			}

			result = append(result, UserStats{
				UserUUID:             u.UUID,
				UserFullname:         u.Fullname,
				TotalProducers:       stats.TotalProducers,
				EligibleProducers:    stats.EligibleProducers,
				NonEligibleProducers: stats.TotalProducers - stats.EligibleProducers,
				AvgScore:             avgScore,
				CompletionRate:       completionRate,
				LastSurveyDate:       stats.LastSurveyDate,
			})
		}
	}

	// Sort by total producers desc
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalProducers > result[j].TotalProducers
	})

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data":   result,
		"total":  len(result),
	})
}

// ============ Zone Performance Endpoint ============

// GetZonePerformance returns performance metrics per geographic zone
func GetZonePerformance(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Preload("Champs").Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	type ZoneAgentStats struct {
		Zone              string
		TotalProducers    int
		EligibleProducers int
		TotalScore        float64
		AgentCounts       map[string]int
	}

	zoneMap := make(map[string]ZoneAgentStats)

	for _, p := range producers {
		scoreResult := ScoreProducer(p)
		zone := p.Zone
		if zone == "" {
			zone = "Inconnue"
		}

		stats := zoneMap[zone]
		stats.Zone = zone
		stats.TotalProducers++
		if scoreResult.Total >= 60 {
			stats.EligibleProducers++
		}
		stats.TotalScore += scoreResult.Total

		if stats.AgentCounts == nil {
			stats.AgentCounts = make(map[string]int)
		}
		stats.AgentCounts[p.AgentRecenseurUUID]++

		zoneMap[zone] = stats
	}

	// Fetch users for full names
	var users []models.User
	db.Find(&users)
	userMap := make(map[string]string)
	for _, u := range users {
		userMap[u.UUID] = u.Fullname
	}

	type ZonePerformance struct {
		Zone              string      `json:"zone"`
		TotalProducers    int         `json:"total_producers"`
		EligibleProducers int         `json:"eligible_producers"`
		AvgScore          float64     `json:"avg_score"`
		TopAgents         []UserStats `json:"top_agents"`
	}

	var result []ZonePerformance
	for _, zoneStats := range zoneMap {
		avgScore := 0.0
		if zoneStats.TotalProducers > 0 {
			avgScore = math.Round((zoneStats.TotalScore/float64(zoneStats.TotalProducers))*100) / 100
		}

		// Get top agents for this zone
		var topAgents []UserStats
		for agentUUID, count := range zoneStats.AgentCounts {
			topAgents = append(topAgents, UserStats{
				UserUUID:       agentUUID,
				UserFullname:   userMap[agentUUID],
				TotalProducers: count,
			})
		}

		result = append(result, ZonePerformance{
			Zone:              zoneStats.Zone,
			TotalProducers:    zoneStats.TotalProducers,
			EligibleProducers: zoneStats.EligibleProducers,
			AvgScore:          avgScore,
			TopAgents:         topAgents,
		})
	}

	// Sort by total producers
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalProducers > result[j].TotalProducers
	})

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data":   result,
		"total":  len(result),
	})
}

// ============ Productivity Metrics Endpoint ============

// GetProductivityMetrics returns detailed productivity metrics by zone
func GetProductivityMetrics(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Preload("Champs").Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	type ProductivityMetrics struct {
		Zone           string  `json:"zone"`
		TotalArea      float64 `json:"total_area"`
		AvgRendement   float64 `json:"avg_rendement"`
		ProducersCount int     `json:"producers_count"`
		IrrigatedArea  float64 `json:"irrigated_area"`
		RainfedArea    float64 `json:"rainfed_area"`
		ValloyArea     float64 `json:"valloy_area"`
	}

	type ZoneMetrics struct {
		Zone           string
		TotalArea      float64
		TotalRendement float64
		Producers      int
		Pluviale       float64
		Irriguee       float64
		BasFond        float64
	}

	zoneMetricsMap := make(map[string]ZoneMetrics)

	for _, p := range producers {
		zone := p.Zone
		if zone == "" {
			zone = "Inconnue"
		}

		metrics := zoneMetricsMap[zone]
		metrics.Zone = zone
		metrics.Producers++
		if p.RendementMoyen > 0 {
			metrics.TotalRendement += p.RendementMoyen
		}

		for _, c := range p.Champs {
			metrics.TotalArea += c.Superficie
			if c.TypeRiziculture == "irriguee" {
				metrics.Irriguee += c.Superficie
			} else if c.TypeRiziculture == "pluviale" {
				metrics.Pluviale += c.Superficie
			} else if c.TypeRiziculture == "bas-fond" {
				metrics.BasFond += c.Superficie
			}
		}

		zoneMetricsMap[zone] = metrics
	}

	var result []ProductivityMetrics
	for _, metrics := range zoneMetricsMap {
		avgRendement := 0.0
		if metrics.Producers > 0 {
			avgRendement = math.Round((metrics.TotalRendement/float64(metrics.Producers))*100) / 100
		}

		result = append(result, ProductivityMetrics{
			Zone:           metrics.Zone,
			TotalArea:      math.Round(metrics.TotalArea*100) / 100,
			AvgRendement:   avgRendement,
			ProducersCount: metrics.Producers,
			IrrigatedArea:  math.Round(metrics.Irriguee*100) / 100,
			RainfedArea:    math.Round(metrics.Pluviale*100) / 100,
			ValloyArea:     math.Round(metrics.BasFond*100) / 100,
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

// ============ Environmental Impact Endpoint ============

// GetEnvironmentalImpact returns environmental metrics by zone
func GetEnvironmentalImpact(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	type EnvironmentalImpact struct {
		Zone                 string  `json:"zone"`
		RotationCulturesRate float64 `json:"rotation_cultures_rate"`
		CompostUsageRate     float64 `json:"compost_usage_rate"`
		TreePresenceRate     float64 `json:"tree_presence_rate"`
		WaterSavingRate      float64 `json:"water_saving_rate"`
		DegradationSignsRate float64 `json:"degradation_signs_rate"`
		PesticidesUsageRate  float64 `json:"pesticides_usage_rate"`
		NoDeforestationRate  float64 `json:"no_deforestation_rate"`
	}

	type EnvStats struct {
		RotationCultures int
		CompostUsage     int
		TreePresence     int
		WaterSaving      int
		DegradationSigns int
		PesticidesUsage  int
		NoDeforestation  int
		Total            int
	}

	zoneEnvMap := make(map[string]EnvStats)

	for _, p := range producers {
		zone := p.Zone
		if zone == "" {
			zone = "Inconnue"
		}

		stats := zoneEnvMap[zone]
		stats.Total++

		if p.RotationCultures {
			stats.RotationCultures++
		}
		if p.UtilisationCompost {
			stats.CompostUsage++
		}
		if p.PresenceArbres {
			stats.TreePresence++
		}
		if p.EconomieEau {
			stats.WaterSaving++
		}
		if !p.SignesDegradation {
			stats.DegradationSigns++
		}
		if p.UtilisationPesticides {
			stats.PesticidesUsage++
		}
		if !p.ActiviteDeforestation {
			stats.NoDeforestation++
		}

		zoneEnvMap[zone] = stats
	}

	var result []EnvironmentalImpact
	for zone, stats := range zoneEnvMap {
		result = append(result, EnvironmentalImpact{
			Zone:                 zone,
			RotationCulturesRate: (float64(stats.RotationCultures) * 100) / float64(stats.Total),
			CompostUsageRate:     (float64(stats.CompostUsage) * 100) / float64(stats.Total),
			TreePresenceRate:     (float64(stats.TreePresence) * 100) / float64(stats.Total),
			WaterSavingRate:      (float64(stats.WaterSaving) * 100) / float64(stats.Total),
			DegradationSignsRate: (float64(stats.DegradationSigns) * 100) / float64(stats.Total),
			PesticidesUsageRate:  (float64(stats.PesticidesUsage) * 100) / float64(stats.Total),
			NoDeforestationRate:  (float64(stats.NoDeforestation) * 100) / float64(stats.Total),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

// ============ Vulnerability Index Endpoint ============

// GetVulnerabilityIndex returns risk/vulnerability metrics by zone
func GetVulnerabilityIndex(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	type VulnerabilityIndex struct {
		Zone                    string  `json:"zone"`
		ClimateChangeVulnerable float64 `json:"climate_change_vulnerable"`
		DroughtRiskRate         float64 `json:"drought_risk_rate"`
		FloodRiskRate           float64 `json:"flood_risk_rate"`
		CreditAccessRate        float64 `json:"credit_access_rate"`
		SoilDegradationRate     float64 `json:"soil_degradation_rate"`
	}

	type VulnStats struct {
		ClimateChange   int
		DroughtRisk     int
		FloodRisk       int
		CreditAccess    int
		SoilDegradation int
		Total           int
	}

	zoneVulnMap := make(map[string]VulnStats)

	for _, p := range producers {
		zone := p.Zone
		if zone == "" {
			zone = "Inconnue"
		}

		stats := zoneVulnMap[zone]
		stats.Total++

		if p.ChangementsClimatiques {
			stats.ClimateChange++
		}
		if p.PerteSec {
			stats.DroughtRisk++
		}
		if p.PerteInondation {
			stats.FloodRisk++
		}
		if p.AccesCredit {
			stats.CreditAccess++
		}
		if p.DegradationSols {
			stats.SoilDegradation++
		}

		zoneVulnMap[zone] = stats
	}

	var result []VulnerabilityIndex
	for zone, stats := range zoneVulnMap {
		result = append(result, VulnerabilityIndex{
			Zone:                    zone,
			ClimateChangeVulnerable: (float64(stats.ClimateChange) * 100) / float64(stats.Total),
			DroughtRiskRate:         (float64(stats.DroughtRisk) * 100) / float64(stats.Total),
			FloodRiskRate:           (float64(stats.FloodRisk) * 100) / float64(stats.Total),
			CreditAccessRate:        (float64(stats.CreditAccess) * 100) / float64(stats.Total),
			SoilDegradationRate:     (float64(stats.SoilDegradation) * 100) / float64(stats.Total),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

// ============ Time Series Data Endpoint ============

// GetSurveyTimeSeries returns survey completion over time (daily)
func GetSurveyTimeSeries(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	type TimeSeriesData struct {
		Date  string  `json:"date"`
		Count int     `json:"count"`
		Score float64 `json:"score"`
	}

	// Group by date
	dateMap := make(map[string]struct {
		count    int
		scoreSum float64
	})

	for _, p := range producers {
		dateStr := p.DateRecensement.Format("2006-01-02")
		scoreResult := ScoreProducer(p)

		stats := dateMap[dateStr]
		stats.count++
		stats.scoreSum += scoreResult.Total
		dateMap[dateStr] = stats
	}

	var dates []string
	for date := range dateMap {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	var result []TimeSeriesData
	for _, dateStr := range dates {
		stats := dateMap[dateStr]
		avgScore := stats.scoreSum / float64(stats.count)
		result = append(result, TimeSeriesData{
			Date:  dateStr,
			Count: stats.count,
			Score: math.Round(avgScore*100) / 100,
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

// ============ Detailed Producer Scores Endpoint ============

// GetProducerScores returns scores for all producers with breakdown
func GetProducerScores(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Preload("Champs").Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	type ProducerWithScore struct {
		Producer    models.Producer `json:"producer"`
		Score       float64         `json:"score"`
		ScoreDetail ScoreResult     `json:"score_detail"`
	}

	var result []ProducerWithScore
	for _, p := range producers {
		scoreDetail := ScoreProducer(p)
		result = append(result, ProducerWithScore{
			Producer:    p,
			Score:       scoreDetail.Total,
			ScoreDetail: scoreDetail,
		})
	}

	// Sort by score desc
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data":   result,
		"total":  len(result),
	})
}

// ============ PDF Export Endpoints ============

// ExportDashboardPDF exports dashboard statistics to PDF
func ExportDashboardPDF(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Preload("Champs").Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	// Add logo in top-left corner
	pdf.Image("assets/logo-phd.png", 10, 10, 30, 0, false, "", 0, "")
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Tableau de Bord - Rapport des Producteurs", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("Généré le: %s", time.Now().Format("02/01/2006 15:04")), "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Calculate statistics
	type ScoredProducer struct {
		Producer models.Producer
		Score    float64
	}

	var scored []ScoredProducer
	totalScore := 0.0
	eligible := 0

	for _, p := range producers {
		scoreResult := ScoreProducer(p)
		scored = append(scored, ScoredProducer{
			Producer: p,
			Score:    scoreResult.Total,
		})
		totalScore += scoreResult.Total
		if scoreResult.Total >= 60 {
			eligible++
		}
	}

	avgScore := 0.0
	if len(scored) > 0 {
		avgScore = totalScore / float64(len(scored))
	}

	// KPIs Section
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 8, "Indicateurs Clés de Performance", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)

	kpiData := [][]string{
		{"Total Producteurs", fmt.Sprintf("%d", len(scored))},
		{"Producteurs Éligibles", fmt.Sprintf("%d (%.1f%%)", eligible, float64(eligible)*100/float64(len(scored)))},
		{"Score Moyen", fmt.Sprintf("%.2f/100", avgScore)},
	}

	for _, row := range kpiData {
		pdf.CellFormat(80, 6, row[0]+":", "1", 0, "L", false, 0, "")
		pdf.CellFormat(100, 6, row[1], "1", 1, "R", false, 0, "")
	}

	pdf.Ln(5)

	// Zone statistics
	zoneMap := make(map[string]struct {
		Count    int
		Eligible int
		ScoreSum float64
	})

	for _, s := range scored {
		zone := s.Producer.Zone
		if zone == "" {
			zone = "Inconnue"
		}
		stats := zoneMap[zone]
		stats.Count++
		if s.Score >= 60 {
			stats.Eligible++
		}
		stats.ScoreSum += s.Score
		zoneMap[zone] = stats
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 8, "Statistiques par Secteur Géographique", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)

	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(60, 6, "Secteur", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 6, "Total", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 6, "Éligibles", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 6, "Score Moyen", "1", 1, "C", true, 0, "")

	pdf.SetFillColor(255, 255, 255)
	for zone, stats := range zoneMap {
		avgZoneScore := 0.0
		if stats.Count > 0 {
			avgZoneScore = stats.ScoreSum / float64(stats.Count)
		}
		pdf.CellFormat(60, 6, zone, "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("%d", stats.Count), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("%d", stats.Eligible), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%.2f", avgZoneScore), "1", 1, "C", false, 0, "")
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate PDF",
		})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=dashboard_report.pdf")
	return c.Send(buf.Bytes())
}

// ExportUserPerformancePDF exports user performance metrics to PDF
func ExportUserPerformancePDF(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	userMap := make(map[string]struct {
		TotalProducers    int
		EligibleProducers int
		TotalScore        float64
		LastSurveyDate    *time.Time
	})

	for _, p := range producers {
		scoreResult := ScoreProducer(p)
		stats := userMap[p.AgentRecenseurUUID]
		stats.TotalProducers++
		if scoreResult.Total >= 60 {
			stats.EligibleProducers++
		}
		stats.TotalScore += scoreResult.Total
		if stats.LastSurveyDate == nil || p.DateRecensement.After(*stats.LastSurveyDate) {
			stats.LastSurveyDate = &p.DateRecensement
		}
		userMap[p.AgentRecenseurUUID] = stats
	}

	var users []models.User
	db.Find(&users)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	// Add logo in top-left corner
	pdf.Image("assets/logo-phd.png", 10, 10, 30, 0, false, "", 0, "")
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Performance des Agents de Recensement", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("Généré le: %s", time.Now().Format("02/01/2006 15:04")), "", 1, "C", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(50, 6, "Agent", "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 6, "Total", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 6, "Éligibles", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 6, "Score Moyen", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 6, "Taux", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)

	for _, u := range users {
		if stats, exists := userMap[u.UUID]; exists {
			avgScore := 0.0
			completionRate := 0.0
			if stats.TotalProducers > 0 {
				avgScore = stats.TotalScore / float64(stats.TotalProducers)
				completionRate = (float64(stats.EligibleProducers) / float64(stats.TotalProducers)) * 100
			}

			pdf.CellFormat(50, 6, u.Fullname, "1", 0, "L", false, 0, "")
			pdf.CellFormat(30, 6, fmt.Sprintf("%d", stats.TotalProducers), "1", 0, "C", false, 0, "")
			pdf.CellFormat(30, 6, fmt.Sprintf("%d", stats.EligibleProducers), "1", 0, "C", false, 0, "")
			pdf.CellFormat(35, 6, fmt.Sprintf("%.2f", avgScore), "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 6, fmt.Sprintf("%.1f%%", completionRate), "1", 1, "C", false, 0, "")
		}
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate PDF",
		})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=user_performance_report.pdf")
	return c.Send(buf.Bytes())
}

// ============ Excel Export Endpoints ============

// ExportDashboardExcel exports dashboard statistics to Excel
func ExportDashboardExcel(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Preload("Champs").Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	f := excelize.NewFile()

	// Sheet 1: Dashboard Summary
	sheet1 := "Tableau de Bord"
	f.NewSheet(sheet1)
	f.SetCellValue(sheet1, "A1", "TABLEAU DE BORD - RÉSUMÉ")
	f.SetCellValue(sheet1, "A2", fmt.Sprintf("Généré le: %s", time.Now().Format("02/01/2006 15:04")))

	// KPIs
	type ScoredProducer struct {
		Producer models.Producer
		Score    float64
	}

	var scored []ScoredProducer
	totalScore := 0.0
	eligible := 0
	femmes := 0

	for _, p := range producers {
		scoreResult := ScoreProducer(p)
		scored = append(scored, ScoredProducer{
			Producer: p,
			Score:    scoreResult.Total,
		})
		totalScore += scoreResult.Total
		if scoreResult.Total >= 60 {
			eligible++
		}
		if p.Sexe == "femme" {
			femmes++
		}
	}

	avgScore := 0.0
	if len(scored) > 0 {
		avgScore = totalScore / float64(len(scored))
	}

	row := 4
	f.SetCellValue(sheet1, fmt.Sprintf("A%d", row), "Indicateur")
	f.SetCellValue(sheet1, fmt.Sprintf("B%d", row), "Valeur")

	row++
	f.SetCellValue(sheet1, fmt.Sprintf("A%d", row), "Total Producteurs")
	f.SetCellValue(sheet1, fmt.Sprintf("B%d", row), len(scored))

	row++
	f.SetCellValue(sheet1, fmt.Sprintf("A%d", row), "Producteurs Éligibles")
	f.SetCellValue(sheet1, fmt.Sprintf("B%d", row), eligible)

	row++
	f.SetCellValue(sheet1, fmt.Sprintf("A%d", row), "Producteurs Non-Éligibles")
	f.SetCellValue(sheet1, fmt.Sprintf("B%d", row), len(scored)-eligible)

	row++
	f.SetCellValue(sheet1, fmt.Sprintf("A%d", row), "Score Moyen")
	f.SetCellValue(sheet1, fmt.Sprintf("B%d", row), math.Round(avgScore*100)/100)

	row++
	f.SetCellValue(sheet1, fmt.Sprintf("A%d", row), "Femmes")
	f.SetCellValue(sheet1, fmt.Sprintf("B%d", row), femmes)

	// Sheet 2: Zone Statistics
	sheet2 := "Statistiques par Zone"
	f.NewSheet(sheet2)

	f.SetCellValue(sheet2, "A1", "STATISTIQUES PAR SECTEUR GÉOGRAPHIQUE")

	zoneMap := make(map[string]struct {
		Count    int
		Eligible int
		ScoreSum float64
	})

	for _, s := range scored {
		zone := s.Producer.Zone
		if zone == "" {
			zone = "Inconnue"
		}
		stats := zoneMap[zone]
		stats.Count++
		if s.Score >= 60 {
			stats.Eligible++
		}
		stats.ScoreSum += s.Score
		zoneMap[zone] = stats
	}

	row = 3
	f.SetCellValue(sheet2, fmt.Sprintf("A%d", row), "Secteur")
	f.SetCellValue(sheet2, fmt.Sprintf("B%d", row), "Total")
	f.SetCellValue(sheet2, fmt.Sprintf("C%d", row), "Éligibles")
	f.SetCellValue(sheet2, fmt.Sprintf("D%d", row), "Score Moyen")

	row++
	for zone, stats := range zoneMap {
		avgZoneScore := 0.0
		if stats.Count > 0 {
			avgZoneScore = math.Round((stats.ScoreSum/float64(stats.Count))*100) / 100
		}
		f.SetCellValue(sheet2, fmt.Sprintf("A%d", row), zone)
		f.SetCellValue(sheet2, fmt.Sprintf("B%d", row), stats.Count)
		f.SetCellValue(sheet2, fmt.Sprintf("C%d", row), stats.Eligible)
		f.SetCellValue(sheet2, fmt.Sprintf("D%d", row), avgZoneScore)
		row++
	}

	// Sheet 3: Producer List with Scores
	sheet3 := "Liste des Producteurs"
	f.NewSheet(sheet3)

	f.SetCellValue(sheet3, "A1", "LISTE DES PRODUCTEURS AVEC SCORES")

	row = 3
	f.SetCellValue(sheet3, fmt.Sprintf("A%d", row), "Nom")
	f.SetCellValue(sheet3, fmt.Sprintf("B%d", row), "Secteur")
	f.SetCellValue(sheet3, fmt.Sprintf("C%d", row), "Village")
	f.SetCellValue(sheet3, fmt.Sprintf("D%d", row), "Sexe")
	f.SetCellValue(sheet3, fmt.Sprintf("E%d", row), "Score Total")
	f.SetCellValue(sheet3, fmt.Sprintf("F%d", row), "Score Env.")
	f.SetCellValue(sheet3, fmt.Sprintf("G%d", row), "Score Exp.")

	row++
	for _, sp := range scored {
		scoreDetail := ScoreProducer(sp.Producer)
		f.SetCellValue(sheet3, fmt.Sprintf("A%d", row), sp.Producer.Nom)
		f.SetCellValue(sheet3, fmt.Sprintf("B%d", row), sp.Producer.Zone)
		f.SetCellValue(sheet3, fmt.Sprintf("C%d", row), sp.Producer.Village)
		f.SetCellValue(sheet3, fmt.Sprintf("D%d", row), sp.Producer.Sexe)
		f.SetCellValue(sheet3, fmt.Sprintf("E%d", row), sp.Score)
		f.SetCellValue(sheet3, fmt.Sprintf("F%d", row), math.Round(scoreDetail.Environmental*100)/100)
		f.SetCellValue(sheet3, fmt.Sprintf("G%d", row), math.Round(scoreDetail.Experience*100)/100)
		row++
	}

	f.DeleteSheet("Sheet")

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate Excel file",
		})
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=dashboard_report.xlsx")
	return c.Send(buf.Bytes())
}

// ExportProducerScoresExcel exports producer scores with details to Excel
func ExportProducerScoresExcel(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Preload("Champs").Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	f := excelize.NewFile()
	sheet := "Scores Détaillés"
	f.NewSheet(sheet)

	f.SetCellValue(sheet, "A1", "SCORES DÉTAILLÉS DES PRODUCTEURS")
	f.SetCellValue(sheet, "A2", fmt.Sprintf("Généré le: %s", time.Now().Format("02/01/2006 15:04")))

	row := 4
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Nom")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "Zone")
	f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "Score Total")
	f.SetCellValue(sheet, fmt.Sprintf("D%d", row), "Environnemental")
	f.SetCellValue(sheet, fmt.Sprintf("E%d", row), "Expérience")
	f.SetCellValue(sheet, fmt.Sprintf("F%d", row), "Production")
	f.SetCellValue(sheet, fmt.Sprintf("G%d", row), "Gestion Risques")
	f.SetCellValue(sheet, fmt.Sprintf("H%d", row), "Accès Ressources")
	f.SetCellValue(sheet, fmt.Sprintf("I%d", row), "Soutien Instit.")

	row++
	var scored []struct {
		Producer models.Producer
		Score    float64
	}

	for _, p := range producers {
		scoreResult := ScoreProducer(p)
		scored = append(scored, struct {
			Producer models.Producer
			Score    float64
		}{p, scoreResult.Total})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	for _, sp := range scored {
		scoreDetail := ScoreProducer(sp.Producer)
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), sp.Producer.Nom)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), sp.Producer.Zone)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), math.Round(sp.Score*100)/100)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), math.Round(scoreDetail.Environmental*100)/100)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), math.Round(scoreDetail.Experience*100)/100)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), math.Round(scoreDetail.Production*100)/100)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), math.Round(scoreDetail.RiskManagement*100)/100)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), math.Round(scoreDetail.AccessToResources*100)/100)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), math.Round(scoreDetail.InstitutionalSupport*100)/100)
		row++
	}

	f.DeleteSheet("Sheet")

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate Excel file",
		})
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=producer_scores.xlsx")
	return c.Send(buf.Bytes())
}

// ExportUserPerformanceExcel exports user/agent performance metrics to Excel
func ExportUserPerformanceExcel(c *fiber.Ctx) error {
	db := database.DB

	var producers []models.Producer
	if err := db.Find(&producers).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch producers",
		})
	}

	userMap := make(map[string]struct {
		TotalProducers    int
		EligibleProducers int
		TotalScore        float64
		LastSurveyDate    *time.Time
	})

	for _, p := range producers {
		scoreResult := ScoreProducer(p)
		stats := userMap[p.AgentRecenseurUUID]
		stats.TotalProducers++
		if scoreResult.Total >= 60 {
			stats.EligibleProducers++
		}
		stats.TotalScore += scoreResult.Total
		if stats.LastSurveyDate == nil || p.DateRecensement.After(*stats.LastSurveyDate) {
			stats.LastSurveyDate = &p.DateRecensement
		}
		userMap[p.AgentRecenseurUUID] = stats
	}

	var users []models.User
	db.Find(&users)

	f := excelize.NewFile()
	sheet := "Performance des Agents"
	f.NewSheet(sheet)

	f.SetCellValue(sheet, "A1", "PERFORMANCE DES AGENTS DE RECENSEMENT")
	f.SetCellValue(sheet, "A2", fmt.Sprintf("Généré le: %s", time.Now().Format("02/01/2006 15:04")))

	row := 4
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Agent")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "Total Producteurs")
	f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "Producteurs Éligibles")
	f.SetCellValue(sheet, fmt.Sprintf("D%d", row), "Non-Éligibles")
	f.SetCellValue(sheet, fmt.Sprintf("E%d", row), "Score Moyen")
	f.SetCellValue(sheet, fmt.Sprintf("F%d", row), "Taux Réussite (%)")
	f.SetCellValue(sheet, fmt.Sprintf("G%d", row), "Dernier Recensement")

	row++
	for _, u := range users {
		if stats, exists := userMap[u.UUID]; exists {
			avgScore := 0.0
			completionRate := 0.0
			if stats.TotalProducers > 0 {
				avgScore = stats.TotalScore / float64(stats.TotalProducers)
				completionRate = (float64(stats.EligibleProducers) / float64(stats.TotalProducers)) * 100
			}

			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), u.Fullname)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), stats.TotalProducers)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), stats.EligibleProducers)
			f.SetCellValue(sheet, fmt.Sprintf("D%d", row), stats.TotalProducers-stats.EligibleProducers)
			f.SetCellValue(sheet, fmt.Sprintf("E%d", row), math.Round(avgScore*100)/100)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", row), math.Round(completionRate*100)/100)
			if stats.LastSurveyDate != nil {
				f.SetCellValue(sheet, fmt.Sprintf("G%d", row), stats.LastSurveyDate.Format("02/01/2006 15:04"))
			}
			row++
		}
	}

	f.DeleteSheet("Sheet")

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate Excel file",
		})
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=user_performance.xlsx")
	return c.Send(buf.Bytes())
}
