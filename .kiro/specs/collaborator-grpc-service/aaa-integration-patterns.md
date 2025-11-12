# AAA Service Integration Patterns for Collaborator gRPC

## Overview

This document defines the integration patterns between the Collaborator gRPC service and the AAA v2 service for address management, authentication, and authorization.

## Architecture Overview

```
┌─────────────────────┐
│  gRPC Client        │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Collaborator       │
│  gRPC Service       │
└──────────┬──────────┘
           │
    ┌──────┴──────┐
    │             │
    ▼             ▼
┌─────────┐  ┌─────────┐
│  AAA v2 │  │Database │
│ Service │  │         │
└─────────┘  └─────────┘
```

## AAA Service Client Implementation

### 1. Connection Pool Management

```go
package aaa

import (
    "context"
    "sync"
    "time"

    aaaPb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/keepalive"
    "google.golang.org/grpc/balancer/roundrobin"
)

// ClientConfig for AAA service connection
type ClientConfig struct {
    Addresses       []string      // Multiple AAA service addresses for load balancing
    MaxConnections  int           // Maximum connections in pool
    ConnTimeout     time.Duration // Connection timeout
    RequestTimeout  time.Duration // Default request timeout
    KeepAlive       time.Duration // Keep-alive interval
    MaxRetries      int           // Maximum retry attempts
    EnableTLS       bool          // Enable TLS
    CertFile        string        // TLS certificate file
}

// ConnectionPool manages gRPC connections to AAA service
type ConnectionPool struct {
    config      ClientConfig
    connections []*grpc.ClientConn
    clients     []aaaPb.AddressServiceClient
    authClients []aaaPb.AuthServiceClient
    currentIdx  int
    mu          sync.RWMutex
}

// NewConnectionPool creates a new AAA connection pool
func NewConnectionPool(config ClientConfig) (*ConnectionPool, error) {
    pool := &ConnectionPool{
        config:      config,
        connections: make([]*grpc.ClientConn, 0, config.MaxConnections),
        clients:     make([]aaaPb.AddressServiceClient, 0, config.MaxConnections),
        authClients: make([]aaaPb.AuthServiceClient, 0, config.MaxConnections),
    }

    // Create connections
    for i := 0; i < config.MaxConnections; i++ {
        conn, err := pool.createConnection()
        if err != nil {
            // Clean up already created connections
            pool.Close()
            return nil, fmt.Errorf("failed to create connection %d: %w", i, err)
        }

        pool.connections = append(pool.connections, conn)
        pool.clients = append(pool.clients, aaaPb.NewAddressServiceClient(conn))
        pool.authClients = append(pool.authClients, aaaPb.NewAuthServiceClient(conn))
    }

    // Start health check goroutine
    go pool.healthCheck()

    return pool, nil
}

func (p *ConnectionPool) createConnection() (*grpc.ClientConn, error) {
    opts := []grpc.DialOption{
        grpc.WithDefaultServiceConfig(fmt.Sprintf(`{
            "loadBalancingPolicy": "%s",
            "methodConfig": [{
                "name": [{"service": ""}],
                "timeout": "%s",
                "retryPolicy": {
                    "maxAttempts": %d,
                    "initialBackoff": "0.1s",
                    "maxBackoff": "1s",
                    "backoffMultiplier": 2,
                    "retryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
                }
            }]
        }`, roundrobin.Name, p.config.RequestTimeout, p.config.MaxRetries)),
        grpc.WithKeepaliveParams(keepalive.ClientParameters{
            Time:                p.config.KeepAlive,
            Timeout:             p.config.ConnTimeout,
            PermitWithoutStream: true,
        }),
    }

    // Add TLS if enabled
    if p.config.EnableTLS {
        creds, err := credentials.NewClientTLSFromFile(p.config.CertFile, "")
        if err != nil {
            return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
        }
        opts = append(opts, grpc.WithTransportCredentials(creds))
    } else {
        opts = append(opts, grpc.WithInsecure())
    }

    // Select address (round-robin for initial connection)
    address := p.config.Addresses[p.currentIdx%len(p.config.Addresses)]

    ctx, cancel := context.WithTimeout(context.Background(), p.config.ConnTimeout)
    defer cancel()

    conn, err := grpc.DialContext(ctx, address, opts...)
    if err != nil {
        return nil, fmt.Errorf("failed to dial %s: %w", address, err)
    }

    return conn, nil
}

// GetAddressClient returns an address service client from the pool
func (p *ConnectionPool) GetAddressClient() aaaPb.AddressServiceClient {
    p.mu.RLock()
    defer p.mu.RUnlock()

    idx := p.currentIdx % len(p.clients)
    p.currentIdx++

    return p.clients[idx]
}

// GetAuthClient returns an auth service client from the pool
func (p *ConnectionPool) GetAuthClient() aaaPb.AuthServiceClient {
    p.mu.RLock()
    defer p.mu.RUnlock()

    idx := p.currentIdx % len(p.authClients)
    p.currentIdx++

    return p.authClients[idx]
}

// healthCheck monitors connection health
func (p *ConnectionPool) healthCheck() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        p.mu.Lock()
        for i, conn := range p.connections {
            state := conn.GetState()
            if state == connectivity.TransientFailure || state == connectivity.Shutdown {
                // Recreate connection
                conn.Close()
                newConn, err := p.createConnection()
                if err != nil {
                    // Log error and continue
                    continue
                }
                p.connections[i] = newConn
                p.clients[i] = aaaPb.NewAddressServiceClient(newConn)
                p.authClients[i] = aaaPb.NewAuthServiceClient(newConn)
            }
        }
        p.mu.Unlock()
    }
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() error {
    p.mu.Lock()
    defer p.mu.Unlock()

    var errs []error
    for _, conn := range p.connections {
        if err := conn.Close(); err != nil {
            errs = append(errs, err)
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("failed to close %d connections", len(errs))
    }

    return nil
}
```

### 2. Address Service Integration

```go
package aaa

import (
    "context"
    "fmt"
    "time"

    aaaPb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
    "github.com/sony/gobreaker"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// AddressService handles address operations with AAA
type AddressService struct {
    pool          *ConnectionPool
    circuitBreaker *gobreaker.CircuitBreaker
    cache         *AddressCache
    logger        *logrus.Logger
}

// NewAddressService creates a new address service
func NewAddressService(pool *ConnectionPool, cache *AddressCache, logger *logrus.Logger) *AddressService {
    // Configure circuit breaker
    cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "AAA-Address-Service",
        MaxRequests: 100,
        Interval:    10 * time.Second,
        Timeout:     60 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 10 && failureRatio >= 0.6
        },
        OnStateChange: func(name string, from, to gobreaker.State) {
            logger.WithFields(logrus.Fields{
                "service": name,
                "from":    from,
                "to":      to,
            }).Warn("Circuit breaker state changed")
        },
    })

    return &AddressService{
        pool:           pool,
        circuitBreaker: cb,
        cache:          cache,
        logger:         logger,
    }
}

// CreateAddress creates a new address in AAA service
func (s *AddressService) CreateAddress(ctx context.Context, req *CreateAddressRequest) (*Address, error) {
    // Prepare AAA request
    aaaReq := &aaaPb.CreateAddressRequest{
        EntityId:   req.EntityID,
        EntityType: req.EntityType,
        Address: &aaaPb.Address{
            Line1:      req.Line1,
            Line2:      req.Line2,
            Line3:      req.Line3,
            City:       req.City,
            State:      req.State,
            Country:    req.Country,
            PostalCode: req.PostalCode,
            Type:       mapAddressType(req.Type),
            Landmark:   req.Landmark,
            IsPrimary:  req.IsPrimary,
        },
    }

    // Add coordinates if present
    if req.Coordinates != nil {
        aaaReq.Address.Latitude = req.Coordinates.Latitude
        aaaReq.Address.Longitude = req.Coordinates.Longitude
    }

    // Execute with circuit breaker
    result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
        client := s.pool.GetAddressClient()

        // Set timeout
        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()

        return client.CreateAddress(ctx, aaaReq)
    })

    if err != nil {
        return nil, s.handleError(err, "CreateAddress")
    }

    resp := result.(*aaaPb.CreateAddressResponse)

    // Convert to domain model
    address := s.toDomainAddress(resp.Address)

    // Cache the address
    s.cache.Set(address.ID, address, 5*time.Minute)

    return address, nil
}

// UpdateAddress updates an existing address
func (s *AddressService) UpdateAddress(ctx context.Context, req *UpdateAddressRequest) (*Address, error) {
    aaaReq := &aaaPb.UpdateAddressRequest{
        Id: req.ID,
        Address: &aaaPb.Address{
            Line1:      req.Line1,
            Line2:      req.Line2,
            Line3:      req.Line3,
            City:       req.City,
            State:      req.State,
            Country:    req.Country,
            PostalCode: req.PostalCode,
            Type:       mapAddressType(req.Type),
            Landmark:   req.Landmark,
        },
    }

    result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
        client := s.pool.GetAddressClient()
        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()
        return client.UpdateAddress(ctx, aaaReq)
    })

    if err != nil {
        return nil, s.handleError(err, "UpdateAddress")
    }

    resp := result.(*aaaPb.UpdateAddressResponse)
    address := s.toDomainAddress(resp.Address)

    // Update cache
    s.cache.Set(address.ID, address, 5*time.Minute)

    return address, nil
}

// GetAddress retrieves an address by ID
func (s *AddressService) GetAddress(ctx context.Context, id string) (*Address, error) {
    // Check cache first
    if cached, found := s.cache.Get(id); found {
        s.logger.WithField("address_id", id).Debug("Address cache hit")
        return cached, nil
    }

    result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
        client := s.pool.GetAddressClient()
        ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
        defer cancel()

        return client.GetAddress(ctx, &aaaPb.GetAddressRequest{Id: id})
    })

    if err != nil {
        return nil, s.handleError(err, "GetAddress")
    }

    resp := result.(*aaaPb.GetAddressResponse)
    address := s.toDomainAddress(resp.Address)

    // Cache the address
    s.cache.Set(address.ID, address, 5*time.Minute)

    return address, nil
}

// GetAddressesByEntity retrieves all addresses for an entity
func (s *AddressService) GetAddressesByEntity(ctx context.Context, entityID, entityType string) ([]*Address, error) {
    // Check cache
    cacheKey := fmt.Sprintf("%s:%s", entityType, entityID)
    if cached, found := s.cache.GetList(cacheKey); found {
        return cached, nil
    }

    result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
        client := s.pool.GetAddressClient()
        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()

        return client.ListAddresses(ctx, &aaaPb.ListAddressesRequest{
            EntityId:   entityID,
            EntityType: entityType,
        })
    })

    if err != nil {
        return nil, s.handleError(err, "GetAddressesByEntity")
    }

    resp := result.(*aaaPb.ListAddressesResponse)
    addresses := make([]*Address, 0, len(resp.Addresses))

    for _, addr := range resp.Addresses {
        addresses = append(addresses, s.toDomainAddress(addr))
    }

    // Cache the list
    s.cache.SetList(cacheKey, addresses, 5*time.Minute)

    return addresses, nil
}

// DeleteAddress deletes an address
func (s *AddressService) DeleteAddress(ctx context.Context, id string) error {
    _, err := s.circuitBreaker.Execute(func() (interface{}, error) {
        client := s.pool.GetAddressClient()
        ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
        defer cancel()

        return client.DeleteAddress(ctx, &aaaPb.DeleteAddressRequest{Id: id})
    })

    if err != nil {
        return s.handleError(err, "DeleteAddress")
    }

    // Remove from cache
    s.cache.Delete(id)

    return nil
}

// BulkCreateAddresses creates multiple addresses
func (s *AddressService) BulkCreateAddresses(ctx context.Context, requests []*CreateAddressRequest) ([]*Address, error) {
    // Prepare batch request
    aaaRequests := make([]*aaaPb.CreateAddressRequest, 0, len(requests))
    for _, req := range requests {
        aaaRequests = append(aaaRequests, &aaaPb.CreateAddressRequest{
            EntityId:   req.EntityID,
            EntityType: req.EntityType,
            Address: &aaaPb.Address{
                Line1:      req.Line1,
                Line2:      req.Line2,
                City:       req.City,
                State:      req.State,
                Country:    req.Country,
                PostalCode: req.PostalCode,
                Type:       mapAddressType(req.Type),
                IsPrimary:  req.IsPrimary,
            },
        })
    }

    result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
        client := s.pool.GetAddressClient()
        ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
        defer cancel()

        return client.BulkCreateAddresses(ctx, &aaaPb.BulkCreateAddressesRequest{
            Requests: aaaRequests,
        })
    })

    if err != nil {
        return nil, s.handleError(err, "BulkCreateAddresses")
    }

    resp := result.(*aaaPb.BulkCreateAddressesResponse)
    addresses := make([]*Address, 0, len(resp.Addresses))

    for _, addr := range resp.Addresses {
        address := s.toDomainAddress(addr)
        addresses = append(addresses, address)

        // Cache each address
        s.cache.Set(address.ID, address, 5*time.Minute)
    }

    return addresses, nil
}

// handleError converts AAA errors to domain errors
func (s *AddressService) handleError(err error, operation string) error {
    if err == gobreaker.ErrOpenState {
        s.logger.WithField("operation", operation).Error("Circuit breaker is open")
        return ErrServiceUnavailable
    }

    if err == gobreaker.ErrTooManyRequests {
        s.logger.WithField("operation", operation).Error("Too many requests")
        return ErrTooManyRequests
    }

    st, ok := status.FromError(err)
    if !ok {
        return fmt.Errorf("AAA service error: %w", err)
    }

    switch st.Code() {
    case codes.NotFound:
        return ErrAddressNotFound
    case codes.InvalidArgument:
        return ErrInvalidAddress
    case codes.AlreadyExists:
        return ErrAddressExists
    case codes.Unauthenticated:
        return ErrUnauthenticated
    case codes.PermissionDenied:
        return ErrUnauthorized
    case codes.ResourceExhausted:
        return ErrRateLimited
    case codes.Unavailable:
        return ErrServiceUnavailable
    default:
        s.logger.WithFields(logrus.Fields{
            "operation": operation,
            "error":     st.Message(),
            "code":      st.Code(),
        }).Error("AAA service error")
        return fmt.Errorf("AAA service error: %s", st.Message())
    }
}

// toDomainAddress converts AAA address to domain model
func (s *AddressService) toDomainAddress(aaaAddr *aaaPb.Address) *Address {
    if aaaAddr == nil {
        return nil
    }

    addr := &Address{
        ID:         aaaAddr.Id,
        Line1:      aaaAddr.Line1,
        Line2:      aaaAddr.Line2,
        Line3:      aaaAddr.Line3,
        City:       aaaAddr.City,
        State:      aaaAddr.State,
        Country:    aaaAddr.Country,
        PostalCode: aaaAddr.PostalCode,
        Type:       mapFromAAAAddressType(aaaAddr.Type),
        Landmark:   aaaAddr.Landmark,
        IsPrimary:  aaaAddr.IsPrimary,
        IsVerified: aaaAddr.IsVerified,
        CreatedAt:  aaaAddr.CreatedAt.AsTime(),
        UpdatedAt:  aaaAddr.UpdatedAt.AsTime(),
    }

    if aaaAddr.Latitude != 0 || aaaAddr.Longitude != 0 {
        addr.Coordinates = &GeoCoordinates{
            Latitude:  aaaAddr.Latitude,
            Longitude: aaaAddr.Longitude,
        }
    }

    return addr
}
```

### 3. Address Cache Implementation

```go
package aaa

import (
    "sync"
    "time"

    "github.com/patrickmn/go-cache"
)

// AddressCache provides caching for AAA addresses
type AddressCache struct {
    cache *cache.Cache
    mu    sync.RWMutex
}

// NewAddressCache creates a new address cache
func NewAddressCache(defaultExpiration, cleanupInterval time.Duration) *AddressCache {
    return &AddressCache{
        cache: cache.New(defaultExpiration, cleanupInterval),
    }
}

// Set caches an address
func (c *AddressCache) Set(id string, address *Address, duration time.Duration) {
    c.cache.Set(id, address, duration)
}

// Get retrieves an address from cache
func (c *AddressCache) Get(id string) (*Address, bool) {
    if item, found := c.cache.Get(id); found {
        return item.(*Address), true
    }
    return nil, false
}

// SetList caches a list of addresses
func (c *AddressCache) SetList(key string, addresses []*Address, duration time.Duration) {
    c.cache.Set(key, addresses, duration)
}

// GetList retrieves a list of addresses from cache
func (c *AddressCache) GetList(key string) ([]*Address, bool) {
    if item, found := c.cache.Get(key); found {
        return item.([]*Address), true
    }
    return nil, false
}

// Delete removes an address from cache
func (c *AddressCache) Delete(id string) {
    c.cache.Delete(id)
}

// Clear clears all cached addresses
func (c *AddressCache) Clear() {
    c.cache.Flush()
}
```

### 4. GST Deduplication Service

```go
package services

import (
    "context"
    "fmt"
    "strings"
    "sync"
    "time"

    "github.com/patrickmn/go-cache"
)

// GSTDeduplicationService handles GST-based deduplication
type GSTDeduplicationService struct {
    repo   CollaboratorRepository
    cache  *cache.Cache
    mu     sync.RWMutex
    logger *logrus.Logger
}

// NewGSTDeduplicationService creates a new GST deduplication service
func NewGSTDeduplicationService(repo CollaboratorRepository, logger *logrus.Logger) *GSTDeduplicationService {
    return &GSTDeduplicationService{
        repo:   repo,
        cache:  cache.New(5*time.Minute, 10*time.Minute),
        logger: logger,
    }
}

// ValidateGST validates GST number format
func (s *GSTDeduplicationService) ValidateGST(gst string) error {
    // GST format: 22AAAAA0000A1Z5
    // 2 digits (state code) + 10 alphanumeric (PAN) + 1 digit (entity number) +
    // 1 letter (Z by default) + 1 alphanumeric (check digit)

    if len(gst) != 15 {
        return fmt.Errorf("GST number must be 15 characters")
    }

    // Validate state code (01-37)
    stateCode := gst[0:2]
    if !isValidStateCode(stateCode) {
        return fmt.Errorf("invalid state code: %s", stateCode)
    }

    // Validate PAN format (5 letters + 4 digits + 1 letter)
    pan := gst[2:12]
    if !isValidPAN(pan) {
        return fmt.Errorf("invalid PAN in GST: %s", pan)
    }

    // Validate check digit
    if !s.validateCheckDigit(gst) {
        return fmt.Errorf("invalid GST check digit")
    }

    return nil
}

// CheckGSTExists checks if GST already exists
func (s *GSTDeduplicationService) CheckGSTExists(ctx context.Context, gst string) (bool, *Collaborator, error) {
    // Normalize GST
    gst = strings.ToUpper(strings.TrimSpace(gst))

    // Validate format
    if err := s.ValidateGST(gst); err != nil {
        return false, nil, err
    }

    // Check cache
    cacheKey := fmt.Sprintf("gst:%s", gst)
    if cached, found := s.cache.Get(cacheKey); found {
        if cached == "not_found" {
            return false, nil, nil
        }
        return true, cached.(*Collaborator), nil
    }

    // Check database
    collaborator, err := s.repo.GetByGST(ctx, gst)
    if err != nil {
        if err == ErrNotFound {
            // Cache negative result
            s.cache.Set(cacheKey, "not_found", 1*time.Minute)
            return false, nil, nil
        }
        return false, nil, fmt.Errorf("failed to check GST: %w", err)
    }

    // Cache positive result
    s.cache.Set(cacheKey, collaborator, 5*time.Minute)

    return true, collaborator, nil
}

// ExtractBusinessInfo extracts business information from GST
func (s *GSTDeduplicationService) ExtractBusinessInfo(gst string) *BusinessInfo {
    // Extract PAN from GST
    pan := gst[2:12]

    // Determine business type from PAN (5th character)
    businessType := s.getBusinessTypeFromPAN(pan)

    return &BusinessInfo{
        PANNumber:    pan,
        BusinessType: businessType,
    }
}

// getBusinessTypeFromPAN determines business type from PAN format
func (s *GSTDeduplicationService) getBusinessTypeFromPAN(pan string) BusinessType {
    if len(pan) < 5 {
        return BusinessTypeIndividual
    }

    // 5th character of PAN indicates entity type
    switch pan[4] {
    case 'C': // Company
        return BusinessTypePrivateLimited
    case 'P': // Person
        return BusinessTypeIndividual
    case 'H': // HUF
        return BusinessTypeProprietorship
    case 'F': // Firm/Partnership
        return BusinessTypePartnership
    case 'A': // Association
        return BusinessTypeCooperative
    case 'T': // Trust
        return BusinessTypeTrust
    default:
        return BusinessTypeIndividual
    }
}

// validateCheckDigit validates GST check digit using Luhn algorithm variant
func (s *GSTDeduplicationService) validateCheckDigit(gst string) bool {
    // Simplified check - in production, use proper GST checksum algorithm
    return true
}

// BulkCheckGST checks multiple GST numbers
func (s *GSTDeduplicationService) BulkCheckGST(ctx context.Context, gstNumbers []string) (map[string]bool, error) {
    results := make(map[string]bool)
    mu := sync.Mutex{}
    wg := sync.WaitGroup{}

    // Process in parallel with limit
    sem := make(chan struct{}, 10) // Limit to 10 concurrent checks

    for _, gst := range gstNumbers {
        wg.Add(1)
        go func(g string) {
            defer wg.Done()

            sem <- struct{}{} // Acquire semaphore
            defer func() { <-sem }() // Release semaphore

            exists, _, err := s.CheckGSTExists(ctx, g)
            if err != nil {
                s.logger.WithError(err).WithField("gst", g).Error("Failed to check GST")
                return
            }

            mu.Lock()
            results[g] = exists
            mu.Unlock()
        }(gst)
    }

    wg.Wait()
    return results, nil
}
```

### 5. Integration with Collaborator Service

```go
package grpc

import (
    "context"
    "fmt"

    pb "kisanlink-ecom/proto/gen/go/collaborator/v1"
    "kisanlink-ecom/internal/services"
    "kisanlink-ecom/internal/aaa"
)

// CollaboratorGRPCServer implements the gRPC service
type CollaboratorGRPCServer struct {
    pb.UnimplementedCollaboratorServiceServer

    service         services.CollaboratorServiceInterface
    addressService  *aaa.AddressService
    gstService      *services.GSTDeduplicationService
    logger          *logrus.Logger
}

// CreateCollaborator with AAA integration
func (s *CollaboratorGRPCServer) CreateCollaborator(
    ctx context.Context,
    req *pb.CreateCollaboratorRequest,
) (*pb.CollaboratorResponse, error) {

    // Extract user claims from context
    claims, err := getUserClaims(ctx)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "unauthorized")
    }

    // Validate GST if provided
    if req.BusinessInfo != nil && req.BusinessInfo.GstNumber != "" {
        // Check for duplicate GST
        exists, existingCollab, err := s.gstService.CheckGSTExists(ctx, req.BusinessInfo.GstNumber)
        if err != nil {
            s.logger.WithError(err).Error("Failed to check GST")
            return nil, status.Error(codes.Internal, "GST validation failed")
        }

        if exists {
            return nil, status.Errorf(codes.AlreadyExists,
                "GST number already registered to collaborator: %s", existingCollab.ID)
        }

        // Extract additional business info from GST
        extractedInfo := s.gstService.ExtractBusinessInfo(req.BusinessInfo.GstNumber)
        if req.BusinessInfo.PanNumber == "" {
            req.BusinessInfo.PanNumber = extractedInfo.PANNumber
        }
    }

    // Create address in AAA if provided
    var addressIDs []string
    if req.Address != nil {
        addressReq := &aaa.CreateAddressRequest{
            EntityID:    req.UserId, // Will be replaced with collaborator ID
            EntityType:  "collaborator",
            Line1:       req.Address.Line1,
            Line2:       req.Address.Line2,
            Line3:       req.Address.Line3,
            City:        req.Address.City,
            State:       req.Address.State,
            Country:     req.Address.Country,
            PostalCode:  req.Address.PostalCode,
            Type:        mapAddressType(req.Address.Type),
            Landmark:    req.Address.Landmark,
            IsPrimary:   req.Address.IsPrimary,
            Coordinates: mapCoordinates(req.Address.Coordinates),
        }

        address, err := s.addressService.CreateAddress(ctx, addressReq)
        if err != nil {
            s.logger.WithError(err).Error("Failed to create address in AAA")
            // Don't fail the entire request for address failure
            // Log and continue
        } else {
            addressIDs = append(addressIDs, address.ID)
        }
    }

    // Convert to domain request
    domainReq := &services.CreateCollaboratorRequest{
        UserID:          req.UserId,
        Username:        req.Username,
        Email:           req.Email,
        FirstName:       req.FirstName,
        LastName:        req.LastName,
        MiddleName:      req.MiddleName,
        Phone:           req.Phone,
        Type:            mapCollaboratorType(req.Type),
        OrganizationID:  req.OrganizationId,
        AddressIDs:      addressIDs,
    }

    // Add business info if present
    if req.BusinessInfo != nil {
        domainReq.BusinessInfo = &services.BusinessInfo{
            BusinessName:    req.BusinessInfo.BusinessName,
            BusinessType:    mapBusinessType(req.BusinessInfo.BusinessType),
            GSTNumber:       req.BusinessInfo.GstNumber,
            PANNumber:       req.BusinessInfo.PanNumber,
            TaxID:           req.BusinessInfo.TaxId,
            BusinessLicense: req.BusinessInfo.BusinessLicense,
        }
    }

    // Create collaborator
    collaborator, err := s.service.CreateCollaborator(ctx, domainReq, claims.UserID)
    if err != nil {
        return nil, mapDomainError(err)
    }

    // Update address entity ID to collaborator ID
    if len(addressIDs) > 0 {
        go func() {
            // Update asynchronously
            for _, addrID := range addressIDs {
                updateReq := &aaa.UpdateAddressEntityRequest{
                    AddressID:  addrID,
                    EntityID:   collaborator.ID,
                    EntityType: "collaborator",
                }
                if err := s.addressService.UpdateEntity(context.Background(), updateReq); err != nil {
                    s.logger.WithError(err).WithFields(logrus.Fields{
                        "address_id":      addrID,
                        "collaborator_id": collaborator.ID,
                    }).Error("Failed to update address entity")
                }
            }
        }()
    }

    // Convert to proto response
    resp := s.toProtoResponse(collaborator)

    // Expand addresses if requested
    if req.Address != nil && len(addressIDs) > 0 {
        address, err := s.addressService.GetAddress(ctx, addressIDs[0])
        if err == nil {
            resp.Collaborator.PrimaryAddress = s.toProtoAddress(address)
        }
    }

    return resp, nil
}

// GetCollaborator with address expansion
func (s *CollaboratorGRPCServer) GetCollaborator(
    ctx context.Context,
    req *pb.GetCollaboratorRequest,
) (*pb.CollaboratorResponse, error) {

    // Get collaborator from service
    collaborator, err := s.service.GetCollaboratorByID(ctx, req.Id)
    if err != nil {
        return nil, mapDomainError(err)
    }

    // Convert to proto
    resp := s.toProtoResponse(collaborator)

    // Expand addresses if requested
    if req.ExpandAddress && len(collaborator.AddressIDs) > 0 {
        addresses, err := s.addressService.GetAddressesByEntity(ctx, collaborator.ID, "collaborator")
        if err != nil {
            s.logger.WithError(err).Error("Failed to fetch addresses from AAA")
        } else {
            // Find primary address
            for _, addr := range addresses {
                if addr.IsPrimary {
                    resp.Collaborator.PrimaryAddress = s.toProtoAddress(addr)
                    break
                }
            }
            // If no primary, use first address
            if resp.Collaborator.PrimaryAddress == nil && len(addresses) > 0 {
                resp.Collaborator.PrimaryAddress = s.toProtoAddress(addresses[0])
            }
        }
    }

    return resp, nil
}
```

## Error Handling Patterns

```go
package aaa

import (
    "errors"
)

var (
    // Service errors
    ErrServiceUnavailable = errors.New("AAA service unavailable")
    ErrTooManyRequests    = errors.New("too many requests")
    ErrRateLimited        = errors.New("rate limited")

    // Address errors
    ErrAddressNotFound = errors.New("address not found")
    ErrInvalidAddress  = errors.New("invalid address")
    ErrAddressExists   = errors.New("address already exists")

    // Auth errors
    ErrUnauthenticated = errors.New("unauthenticated")
    ErrUnauthorized    = errors.New("unauthorized")

    // GST errors
    ErrInvalidGST      = errors.New("invalid GST number")
    ErrDuplicateGST    = errors.New("GST number already registered")
)

// ErrorMapper maps AAA errors to domain errors
type ErrorMapper struct {
    logger *logrus.Logger
}

func (m *ErrorMapper) MapError(err error, operation string) error {
    // Log the original error
    m.logger.WithFields(logrus.Fields{
        "operation": operation,
        "error":     err.Error(),
    }).Error("AAA operation failed")

    // Map to domain error
    switch {
    case errors.Is(err, context.DeadlineExceeded):
        return ErrServiceUnavailable
    case strings.Contains(err.Error(), "connection refused"):
        return ErrServiceUnavailable
    case strings.Contains(err.Error(), "not found"):
        return ErrAddressNotFound
    default:
        return err
    }
}
```

## Monitoring and Observability

```yaml
# Prometheus metrics for AAA integration

# Connection pool metrics
aaa_connection_pool_size{state="active|idle"}
aaa_connection_pool_acquisitions_total
aaa_connection_pool_timeouts_total

# Circuit breaker metrics
aaa_circuit_breaker_state{service="address|auth", state="closed|open|half_open"}
aaa_circuit_breaker_requests_total{service="address|auth", result="success|failure"}

# Operation metrics
aaa_operation_duration_seconds{service="address", operation="create|update|get|delete"}
aaa_operation_errors_total{service="address", operation="create|update|get|delete", error_type=""}

# Cache metrics
aaa_cache_hits_total{cache="address|gst"}
aaa_cache_misses_total{cache="address|gst"}
aaa_cache_evictions_total{cache="address|gst"}

# GST deduplication metrics
gst_validations_total{result="valid|invalid"}
gst_duplicates_detected_total
```

## Testing Strategy

```go
package aaa_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "google.golang.org/grpc"
)

func TestAddressServiceIntegration(t *testing.T) {
    // Setup mock AAA server
    mockServer := setupMockAAAServer()
    defer mockServer.Stop()

    // Create connection pool
    config := ClientConfig{
        Addresses:      []string{mockServer.Address()},
        MaxConnections: 2,
        ConnTimeout:    1 * time.Second,
        RequestTimeout: 1 * time.Second,
    }

    pool, err := NewConnectionPool(config)
    assert.NoError(t, err)
    defer pool.Close()

    // Create address service
    cache := NewAddressCache(5*time.Minute, 10*time.Minute)
    service := NewAddressService(pool, cache, logrus.New())

    t.Run("CreateAddress", func(t *testing.T) {
        req := &CreateAddressRequest{
            EntityID:   "collab-123",
            EntityType: "collaborator",
            Line1:      "123 Main St",
            City:       "Mumbai",
            State:      "Maharashtra",
            Country:    "India",
            PostalCode: "400001",
        }

        address, err := service.CreateAddress(context.Background(), req)
        assert.NoError(t, err)
        assert.NotEmpty(t, address.ID)
        assert.Equal(t, req.Line1, address.Line1)
    })

    t.Run("CircuitBreaker", func(t *testing.T) {
        // Stop mock server to trigger failures
        mockServer.Stop()

        // Make requests until circuit opens
        for i := 0; i < 15; i++ {
            _, err := service.GetAddress(context.Background(), "addr-123")
            assert.Error(t, err)
        }

        // Circuit should be open now
        _, err := service.GetAddress(context.Background(), "addr-123")
        assert.Equal(t, ErrServiceUnavailable, err)
    })
}

func TestGSTDeduplication(t *testing.T) {
    repo := &MockCollaboratorRepository{}
    service := NewGSTDeduplicationService(repo, logrus.New())

    t.Run("ValidateGST", func(t *testing.T) {
        tests := []struct {
            gst   string
            valid bool
        }{
            {"22AAAAA0000A1Z5", true},
            {"22AAAAA0000A1Z", false},  // Too short
            {"99AAAAA0000A1Z5", false}, // Invalid state code
            {"22aaaaa0000a1z5", true},  // Case insensitive
        }

        for _, tt := range tests {
            err := service.ValidateGST(tt.gst)
            if tt.valid {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
            }
        }
    })

    t.Run("CheckDuplicate", func(t *testing.T) {
        gst := "22AAAAA0000A1Z5"

        // First check - not found
        repo.On("GetByGST", mock.Anything, gst).Return(nil, ErrNotFound).Once()

        exists, collab, err := service.CheckGSTExists(context.Background(), gst)
        assert.NoError(t, err)
        assert.False(t, exists)
        assert.Nil(t, collab)

        // Second check - found (should hit cache)
        exists, collab, err = service.CheckGSTExists(context.Background(), gst)
        assert.NoError(t, err)
        assert.False(t, exists) // Still false from cache
    })
}
```