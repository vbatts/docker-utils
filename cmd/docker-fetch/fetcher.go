package main

import (
	"github.com/docker/docker/graph"
	"github.com/docker/docker/registry"
)

func NewFetcher(dir string) *Fetcher {
	return &Fetcher{Root: dir, repositories: map[string]graph.Repository{}}
}

type Fetcher struct {
	Root         string
	sessions     map[string]*registry.Session
	repositories map[string]graph.Repository
}
