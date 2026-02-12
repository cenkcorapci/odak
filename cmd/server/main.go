package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cenkcorapci/odak/internal/adapters/mlflow"
	"github.com/cenkcorapci/odak/internal/adapters/postgres"
	"github.com/cenkcorapci/odak/internal/application"
	"github.com/cenkcorapci/odak/internal/domain"
	"github.com/gin-gonic/gin"
)

func main() {
	// Get configuration from environment
	dbConnStr := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/odak?sslmode=disable")
	mlflowURL := getEnv("MLFLOW_URL", "http://localhost:5000")
	bucketingSalt := getEnv("BUCKETING_SALT", "odak-secret-salt")
	port := getEnv("PORT", "8080")

	// Initialize database
	db, err := postgres.NewDB(dbConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize schema
	if err := db.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Initialize repositories
	experimentRepo := postgres.NewExperimentRepository(db)
	assignmentRepo := postgres.NewAssignmentRepository(db)
	metricRepo := postgres.NewMetricRepository(db)

	// Initialize services
	bucketingService := domain.NewBucketingService(bucketingSalt)
	experimentService := application.NewExperimentService(
		experimentRepo,
		assignmentRepo,
		metricRepo,
		bucketingService,
	)

	// Initialize MLflow client
	mlflowClient := mlflow.NewClient(mlflowURL)

	// Setup HTTP server
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Experiment endpoints
	router.POST("/api/experiments", createExperimentHandler(experimentService))
	router.GET("/api/experiments", listExperimentsHandler(experimentService))
	router.GET("/api/experiments/:id", getExperimentHandler(experimentService))
	router.PUT("/api/experiments/:id", updateExperimentHandler(experimentService))

	// Assignment endpoint
	router.POST("/api/assign", assignVariantHandler(experimentService))

	// Metrics endpoint
	router.POST("/api/metrics", trackMetricHandler(experimentService))
	router.GET("/api/metrics/:experimentId/:variantId/:metricName", getMetricAggregationHandler(experimentService))

	// MLflow integration endpoint
	router.GET("/api/mlflow/models/:name/:version", func(c *gin.Context) {
		name := c.Param("name")
		version := c.Param("version")

		model, err := mlflowClient.GetModelVersion(name, version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, model)
	})

	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Handler functions
func createExperimentHandler(service *application.ExperimentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var exp domain.Experiment
		if err := c.ShouldBindJSON(&exp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := service.CreateExperiment(&exp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, exp)
	}
}

func listExperimentsHandler(service *application.ExperimentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		experiments, err := service.ListExperiments()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, experiments)
	}
}

func getExperimentHandler(service *application.ExperimentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		exp, err := service.GetExperiment(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, exp)
	}
}

func updateExperimentHandler(service *application.ExperimentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var exp domain.Experiment
		if err := c.ShouldBindJSON(&exp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exp.ID = id
		if err := service.UpdateExperiment(&exp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, exp)
	}
}

func assignVariantHandler(service *application.ExperimentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ExperimentID string `json:"experiment_id" binding:"required"`
			EntityID     string `json:"entity_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		assignment, err := service.AssignVariant(req.ExperimentID, req.EntityID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, assignment)
	}
}

func trackMetricHandler(service *application.ExperimentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var event domain.MetricEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := service.TrackMetric(&event); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"status": "success"})
	}
}

func getMetricAggregationHandler(service *application.ExperimentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		experimentID := c.Param("experimentId")
		variantID := c.Param("variantId")
		metricName := c.Param("metricName")

		avg, count, err := service.GetMetricAggregation(experimentID, variantID, metricName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"average": avg,
			"count":   count,
		})
	}
}
