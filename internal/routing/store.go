package routing

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// RouteStore provides k8s-compatible route storage and management
// Implementations can use files (docker-compose), etcd (k8s), or memory (testing)
type RouteStore interface {
	// CRUD Operations (kubectl-like API)
	Apply(ctx context.Context, config *RouteConfig) error
	Get(ctx context.Context, namespace, name string) (*RouteConfig, error)
	List(ctx context.Context, namespace string) ([]*RouteConfig, error)
	Delete(ctx context.Context, namespace, name string) error
	
	// Watch for changes (k8s controller pattern)
	Watch(ctx context.Context, namespace string) (<-chan RouteEvent, error)
	
	// Lifecycle management
	Start(ctx context.Context) error
	Stop() error
}

// RouteEvent represents a change to a route configuration
type RouteEvent struct {
	Type     EventType    `json:"type"`
	Object   *RouteConfig `json:"object"`
	OldObject *RouteConfig `json:"oldObject,omitempty"`
}

type EventType string

const (
	EventTypeAdded    EventType = "ADDED"
	EventTypeModified EventType = "MODIFIED"
	EventTypeDeleted  EventType = "DELETED"
)

// FileRouteStore implements RouteStore using file system storage
// This provides k8s-style API over YAML files for docker-compose deployments
type FileRouteStore struct {
	mu        sync.RWMutex
	routesDir string
	routes    map[string]*RouteConfig // key: namespace/name
	watchers  map[string][]chan RouteEvent
	watcher   *fsnotify.Watcher
	ctx       context.Context
	cancel    context.CancelFunc
	router    *gin.Engine
	registry  *HandlerRegistry
}

// NewFileRouteStore creates a new file-based route store
func NewFileRouteStore(routesDir string, router *gin.Engine, registry *HandlerRegistry) *FileRouteStore {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &FileRouteStore{
		routesDir: routesDir,
		routes:    make(map[string]*RouteConfig),
		watchers:  make(map[string][]chan RouteEvent),
		ctx:       ctx,
		cancel:    cancel,
		router:    router,
		registry:  registry,
	}
}

// Start initializes the file watcher and loads existing routes
func (s *FileRouteStore) Start(ctx context.Context) error {
	log.Printf("Starting file route store, watching directory: %s", s.routesDir)
	
	// Initialize file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	s.watcher = watcher
	
	// Load existing route files
	if err := s.loadAllRoutes(); err != nil {
		return fmt.Errorf("failed to load existing routes: %w", err)
	}
	
	// Watch routes directory and subdirectories
	if err := s.watchDirectory(s.routesDir); err != nil {
		return fmt.Errorf("failed to watch routes directory: %w", err)
	}
	
	// Start file watching goroutine
	go s.handleFileEvents()
	
	// Apply all loaded routes to the router
	if err := s.reconcileRouter(); err != nil {
		log.Printf("Warning: Failed to reconcile router: %v", err)
	}
	
	log.Printf("File route store started, loaded %d route configurations", len(s.routes))
	return nil
}

// Stop shuts down the file watcher
func (s *FileRouteStore) Stop() error {
	log.Println("Stopping file route store...")
	s.cancel()
	if s.watcher != nil {
		return s.watcher.Close()
	}
	return nil
}

// Apply creates or updates a route configuration by writing to file
func (s *FileRouteStore) Apply(ctx context.Context, config *RouteConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid route configuration: %w", err)
	}
	
	// Create directory structure
	namespace := config.Metadata.Namespace
	if namespace == "" {
		namespace = "default"
	}
	
	namespaceDir := filepath.Join(s.routesDir, namespace)
	if err := os.MkdirAll(namespaceDir, 0755); err != nil {
		return fmt.Errorf("failed to create namespace directory: %w", err)
	}
	
	// Write to YAML file
	filename := config.Metadata.Name + ".yaml"
	filePath := filepath.Join(namespaceDir, filename)
	
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}
	
	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write route file: %w", err)
	}
	
	log.Printf("Applied route configuration to file: %s", filePath)
	return nil
}

// Get retrieves a route configuration by namespace and name
func (s *FileRouteStore) Get(ctx context.Context, namespace, name string) (*RouteConfig, error) {
	key := s.makeKey(namespace, name)
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	config, exists := s.routes[key]
	if !exists {
		return nil, fmt.Errorf("route configuration %s not found", key)
	}
	
	// Return a copy to prevent external modifications
	configCopy := *config
	return &configCopy, nil
}

// List returns all route configurations in a namespace
func (s *FileRouteStore) List(ctx context.Context, namespace string) ([]*RouteConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var configs []*RouteConfig
	prefix := namespace + "/"
	
	for key, config := range s.routes {
		if namespace == "" || strings.HasPrefix(key, prefix) {
			configCopy := *config
			configs = append(configs, &configCopy)
		}
	}
	
	return configs, nil
}

// Delete removes a route configuration by deleting the file
func (s *FileRouteStore) Delete(ctx context.Context, namespace, name string) error {
	// Find and delete the file
	if namespace == "" {
		namespace = "default"
	}
	
	filename := name + ".yaml"
	filePath := filepath.Join(s.routesDir, namespace, filename)
	
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("route configuration %s/%s not found", namespace, name)
		}
		return fmt.Errorf("failed to delete route file: %w", err)
	}
	
	log.Printf("Deleted route configuration file: %s", filePath)
	return nil
}

// Watch returns a channel that receives route change events
func (s *FileRouteStore) Watch(ctx context.Context, namespace string) (<-chan RouteEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	ch := make(chan RouteEvent, 100) // Buffered channel
	s.watchers[namespace] = append(s.watchers[namespace], ch)
	
	// Send existing configurations as ADDED events
	go func() {
		configs, _ := s.List(ctx, namespace)
		for _, config := range configs {
			select {
			case ch <- RouteEvent{Type: EventTypeAdded, Object: config}:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	// Clean up when context is done
	go func() {
		<-ctx.Done()
		s.mu.Lock()
		defer s.mu.Unlock()
		
		watchers := s.watchers[namespace]
		for i, watcher := range watchers {
			if watcher == ch {
				s.watchers[namespace] = append(watchers[:i], watchers[i+1:]...)
				close(ch)
				break
			}
		}
	}()
	
	return ch, nil
}

// loadAllRoutes scans the routes directory and loads all YAML files
func (s *FileRouteStore) loadAllRoutes() error {
	return filepath.Walk(s.routesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-YAML files
		if info.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}
		
		return s.loadRouteFile(path)
	})
}

// loadRouteFile loads a single YAML route file
func (s *FileRouteStore) loadRouteFile(filePath string) error {
	// Use Viper to load and parse the YAML file
	v := viper.New()
	v.SetConfigFile(filePath)
	
	if err := v.ReadInConfig(); err != nil {
		log.Printf("Warning: Failed to read route file %s: %v", filePath, err)
		return nil // Don't fail on individual file errors
	}
	
	var config RouteConfig
	if err := v.Unmarshal(&config); err != nil {
		log.Printf("Warning: Failed to unmarshal route file %s: %v", filePath, err)
		return nil
	}
	
	// Validate the configuration
	if err := config.Validate(); err != nil {
		log.Printf("Warning: Invalid route configuration in %s: %v", filePath, err)
		return nil
	}
	
	// Store the configuration
	key := s.makeKey(config.Metadata.Namespace, config.Metadata.Name)
	
	s.mu.Lock()
	oldConfig := s.routes[key]
	s.routes[key] = &config
	s.mu.Unlock()
	
	// Notify watchers
	eventType := EventTypeAdded
	if oldConfig != nil {
		eventType = EventTypeModified
	}
	
	s.notifyWatchers(config.Metadata.Namespace, RouteEvent{
		Type:      eventType,
		Object:    &config,
		OldObject: oldConfig,
	})
	
	log.Printf("Loaded route configuration: %s (%s)", key, filepath.Base(filePath))
	return nil
}

// watchDirectory adds the directory and subdirectories to the file watcher
func (s *FileRouteStore) watchDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return s.watcher.Add(path)
		}
		
		return nil
	})
}

// handleFileEvents processes file system events
func (s *FileRouteStore) handleFileEvents() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			
			s.handleFileEvent(event)
			
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)
			
		case <-s.ctx.Done():
			return
		}
	}
}

// handleFileEvent processes a single file system event
func (s *FileRouteStore) handleFileEvent(event fsnotify.Event) {
	// Only process YAML files
	if !strings.HasSuffix(event.Name, ".yaml") && !strings.HasSuffix(event.Name, ".yml") {
		return
	}
	
	switch {
	case event.Has(fsnotify.Write) || event.Has(fsnotify.Create):
		log.Printf("Route file changed: %s", event.Name)
		if err := s.loadRouteFile(event.Name); err != nil {
			log.Printf("Error loading changed route file %s: %v", event.Name, err)
		}
		// Trigger router reconciliation
		if err := s.reconcileRouter(); err != nil {
			log.Printf("Error reconciling router after file change: %v", err)
		}
		
	case event.Has(fsnotify.Remove):
		log.Printf("Route file deleted: %s", event.Name)
		s.handleFileDelete(event.Name)
		// Trigger router reconciliation
		if err := s.reconcileRouter(); err != nil {
			log.Printf("Error reconciling router after file deletion: %v", err)
		}
	}
}

// handleFileDelete removes a route configuration when its file is deleted
func (s *FileRouteStore) handleFileDelete(filePath string) {
	// Extract namespace and name from file path
	relPath, err := filepath.Rel(s.routesDir, filePath)
	if err != nil {
		log.Printf("Error getting relative path for %s: %v", filePath, err)
		return
	}
	
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) < 2 {
		log.Printf("Invalid file path structure: %s", filePath)
		return
	}
	
	namespace := parts[0]
	filename := parts[len(parts)-1]
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	
	key := s.makeKey(namespace, name)
	
	s.mu.Lock()
	config := s.routes[key]
	delete(s.routes, key)
	s.mu.Unlock()
	
	if config != nil {
		s.notifyWatchers(namespace, RouteEvent{
			Type:   EventTypeDeleted,
			Object: config,
		})
		log.Printf("Removed deleted route configuration: %s", key)
	}
}

// reconcileRouter applies all current route configurations to the gin router
func (s *FileRouteStore) reconcileRouter() error {
	s.mu.RLock()
	configs := make([]*RouteConfig, 0, len(s.routes))
	for _, config := range s.routes {
		configs = append(configs, config)
	}
	s.mu.RUnlock()
	
	// TODO: In a full implementation, we'd need to remove old routes
	// For now, we just register all current routes (Gin handles duplicates)
	
	for _, config := range configs {
		if err := s.applyRouteConfig(config); err != nil {
			log.Printf("Error applying route config %s/%s: %v", 
				config.Metadata.Namespace, config.Metadata.Name, err)
			continue
		}
	}
	
	return nil
}

// applyRouteConfig registers the routes from a configuration with the gin router
func (s *FileRouteStore) applyRouteConfig(config *RouteConfig) error {
	if !config.Metadata.Enabled {
		return nil // Skip disabled configurations
	}
	
	// Create route group with prefix
	routeGroup := s.router.Group(config.Spec.Prefix)
	
	// Apply middleware to group
	for _, middlewareName := range config.Spec.Middleware {
		middleware, err := s.registry.GetMiddleware(middlewareName)
		if err != nil {
			log.Printf("Warning: Middleware %s not found for route group %s", 
				middlewareName, config.Metadata.Name)
			continue
		}
		routeGroup.Use(middleware)
	}
	
	// Register individual routes
	for _, route := range config.Spec.Routes {
		if err := s.registerRoute(routeGroup, &route); err != nil {
			methods := route.GetMethods()
			methodStr := "GET"
			if len(methods) > 0 {
				methodStr = methods[0]
			}
			log.Printf("Error registering route %s %s: %v", methodStr, route.Path, err)
			continue
		}
	}
	
	return nil
}

// registerRoute registers a single route with the gin router
func (s *FileRouteStore) registerRoute(group *gin.RouterGroup, route *RouteDefinition) error {
	// Get handler
	var handler gin.HandlerFunc
	var err error
	
	if len(route.Handlers) > 0 {
		// Use the first handler for now - TODO: implement method-specific handlers
		for _, handlerName := range route.Handlers {
			handler, err = s.registry.Get(handlerName)
			if err == nil {
				break
			}
		}
	} else if route.Handler != "" {
		handler, err = s.registry.Get(route.Handler)
	} else {
		return fmt.Errorf("no handler specified for route %s", route.Path)
	}
	
	if err != nil {
		// Use placeholder handler for missing handlers
		handler = func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": fmt.Sprintf("Handler %s not implemented", route.Handler),
				"route":   route.Path,
			})
		}
		log.Printf("Using placeholder for missing handler: %s", route.Handler)
	}
	
	// Register route
	methods := route.GetMethods()
	if len(methods) == 0 {
		methods = []string{"GET"} // Default
	}
	
	for _, method := range methods {
		switch strings.ToUpper(method) {
		case "GET":
			group.GET(route.Path, handler)
		case "POST":
			group.POST(route.Path, handler)
		case "PUT":
			group.PUT(route.Path, handler)
		case "DELETE":
			group.DELETE(route.Path, handler)
		case "PATCH":
			group.PATCH(route.Path, handler)
		case "HEAD":
			group.HEAD(route.Path, handler)
		case "OPTIONS":
			group.OPTIONS(route.Path, handler)
		default:
			return fmt.Errorf("unsupported HTTP method: %s", method)
		}
	}
	
	return nil
}

// Helper methods

func (s *FileRouteStore) makeKey(namespace, name string) string {
	if namespace == "" {
		namespace = "default"
	}
	return namespace + "/" + name
}

func (s *FileRouteStore) notifyWatchers(namespace string, event RouteEvent) {
	// Notify namespace-specific watchers
	for _, ch := range s.watchers[namespace] {
		select {
		case ch <- event:
		default:
			log.Printf("Warning: Watcher channel full, dropping event")
		}
	}
	
	// Notify global watchers (namespace = "")
	for _, ch := range s.watchers[""] {
		select {
		case ch <- event:
		default:
			log.Printf("Warning: Global watcher channel full, dropping event")
		}
	}
}

// SimpleRouteManager provides a simplified interface for using file-based routes
type SimpleRouteManager struct {
	store RouteStore
}

// NewSimpleRouteManager creates a route manager that uses file-based storage
func NewSimpleRouteManager(routesDir string, router *gin.Engine, registry *HandlerRegistry) *SimpleRouteManager {
	store := NewFileRouteStore(routesDir, router, registry)
	
	return &SimpleRouteManager{
		store: store,
	}
}

// Start initializes the route manager
func (m *SimpleRouteManager) Start(ctx context.Context) error {
	return m.store.Start(ctx)
}

// Stop shuts down the route manager
func (m *SimpleRouteManager) Stop() error {
	return m.store.Stop()
}

// GetStore returns the underlying route store for advanced operations
func (m *SimpleRouteManager) GetStore() RouteStore {
	return m.store
}