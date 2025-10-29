package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/MeHungr/peanut-butter/internal/api"
	"github.com/MeHungr/peanut-butter/internal/conversion"
	"github.com/MeHungr/peanut-butter/internal/storage"
)

// GetResultsHandler retrieves results from the db and responds with a slice of results
func (srv *Server) GetResultsHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	filter := storage.AgentFilter{}

	// ?all = true returns all agents
	if query.Get("all") == "true" {
		filter.All = true
	}

	// ?agent_id=123&agent_id=456
	if ids, ok := query["agent_id"]; ok {
		filter.IDs = ids
	}

	// ?os=linux&os=windows
	if oses, ok := query["os"]; ok {
		filter.OSes = oses
	}

	// ?status=active&status=inactive
	if statuses, ok := query["status"]; ok {
		filter.Statuses = statuses
	}

	// ?limit=10
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, "Invalid query parameter 'limit': %w", http.StatusBadRequest)
		return
	}

	// Get results from db
	results, err := srv.storage.GetResults(filter, limit)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Generate response
	resp := api.GetResultsResponse{
		Results: conversion.StoragetoAPIResults(results),
	}

	// Marshal response and send
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
