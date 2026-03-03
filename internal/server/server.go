// Package server provides the Gin HTTP API for HITL checkpoint management.
// Listens on :28080 (configurable). Compatible with CLI and agent callers.
package server

import (
	"net/http"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/store"
	"github.com/gin-gonic/gin"
)

// Server wraps the Gin engine and its dependencies.
type Server struct {
	router *gin.Engine
	jobs   store.JobRepository
	cps    store.CheckpointRepository
}

// New creates a Server with the given repositories.
func New(jobs store.JobRepository, cps store.CheckpointRepository) *Server {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	s := &Server{router: r, jobs: jobs, cps: cps}
	s.registerRoutes()
	return s
}

// Router returns the underlying gin.Engine (for testing with httptest).
func (s *Server) Router() http.Handler { return s.router }

// Run starts the HTTP server on the given address.
func (s *Server) Run(addr string) error {
	gin.SetMode(gin.ReleaseMode)
	return s.router.Run(addr)
}

func (s *Server) registerRoutes() {
	s.router.GET("/jobs/:id", s.getJob)
	s.router.GET("/checkpoints/:id", s.getCheckpoint)
	s.router.POST("/checkpoints/:id/approve", s.approveCheckpoint)
	s.router.POST("/checkpoints/:id/reject", s.rejectCheckpoint)
}

func (s *Server) getJob(c *gin.Context) {
	job, err := s.jobs.GetByID(c.Param("id"))
	if err == store.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, job)
}

func (s *Server) getCheckpoint(c *gin.Context) {
	cp, err := s.cps.GetByID(c.Param("id"))
	if err == store.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "checkpoint not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cp)
}

type updateRequest struct {
	Notes string `json:"notes"`
}

func (s *Server) approveCheckpoint(c *gin.Context) {
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if err := s.cps.UpdateStatus(c.Param("id"), domain.CheckpointStatusApproved, req.Notes); err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "checkpoint not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "approved"})
}

func (s *Server) rejectCheckpoint(c *gin.Context) {
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if err := s.cps.UpdateStatus(c.Param("id"), domain.CheckpointStatusRejected, req.Notes); err != nil {
		if err == store.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "checkpoint not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}
