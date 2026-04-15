package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/phd-api/controllers/dashboard"
)

// SetupDashboardRoutes configure toutes les routes du dashboard
func SetupDashboardRoutes(api fiber.Router) {
	// ============================================================
	// DASHBOARD ROUTES
	// ============================================================
	dash := api.Group("/dashboard")

	// Main dashboard statistics
	dash.Get("/stats", dashboard.GetDashboardStats)

	// User/Agent performance
	dash.Get("/user-performance", dashboard.GetUserPerformance)

	// Zone performance
	dash.Get("/zone-performance", dashboard.GetZonePerformance)

	// Productivity metrics
	dash.Get("/productivity-metrics", dashboard.GetProductivityMetrics)

	// Environmental impact
	dash.Get("/environmental-impact", dashboard.GetEnvironmentalImpact)

	// Vulnerability index
	dash.Get("/vulnerability-index", dashboard.GetVulnerabilityIndex)

	// Survey time series data
	dash.Get("/survey-timeseries", dashboard.GetSurveyTimeSeries)

	// Detailed producer scores
	dash.Get("/producer-scores", dashboard.GetProducerScores)

	// ============================================================
	// PDF EXPORT ROUTES
	// ============================================================
	dash.Get("/export/pdf", dashboard.ExportDashboardPDF)
	dash.Get("/export/user-performance-pdf", dashboard.ExportUserPerformancePDF)

	// ============================================================
	// EXCEL EXPORT ROUTES
	// ============================================================
	dash.Get("/export/excel", dashboard.ExportDashboardExcel)
	dash.Get("/export/producer-scores-excel", dashboard.ExportProducerScoresExcel)
	dash.Get("/export/user-performance-excel", dashboard.ExportUserPerformanceExcel)
}
