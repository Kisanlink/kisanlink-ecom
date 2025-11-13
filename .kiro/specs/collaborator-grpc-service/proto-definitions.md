# Collaborator gRPC Service Proto Definitions

## Directory Structure

```
proto/
├── collaborator/
│   └── v1/
│       ├── collaborator_service.proto
│       ├── collaborator_messages.proto
│       └── collaborator_types.proto
├── shared/
│   ├── common.proto
│   ├── pagination.proto
│   └── validation.proto
└── Makefile
```

## Proto Files

### 1. collaborator_types.proto

```protobuf
syntax = "proto3";

package kisanlink.collaborator.v1;

option go_package = "kisanlink-ecom/proto/gen/go/collaborator/v1;collaboratorpb";

import "google/protobuf/timestamp.proto";

// CollaboratorType defines the type of collaborator
enum CollaboratorType {
  COLLABORATOR_TYPE_UNSPECIFIED = 0;
  COLLABORATOR_TYPE_BUYER = 1;
  COLLABORATOR_TYPE_VENDOR = 2;
  COLLABORATOR_TYPE_SERVICE_PROVIDER = 3;
  COLLABORATOR_TYPE_ADMIN = 4;
  COLLABORATOR_TYPE_MANAGER = 5;
}

// CollaboratorStatus defines the status of a collaborator
enum CollaboratorStatus {
  COLLABORATOR_STATUS_UNSPECIFIED = 0;
  COLLABORATOR_STATUS_ACTIVE = 1;
  COLLABORATOR_STATUS_INACTIVE = 2;
  COLLABORATOR_STATUS_SUSPENDED = 3;
  COLLABORATOR_STATUS_PENDING_VERIFICATION = 4;
  COLLABORATOR_STATUS_VERIFIED = 5;
  COLLABORATOR_STATUS_REJECTED = 6;
}

// BusinessType defines the type of business
enum BusinessType {
  BUSINESS_TYPE_UNSPECIFIED = 0;
  BUSINESS_TYPE_INDIVIDUAL = 1;
  BUSINESS_TYPE_PROPRIETORSHIP = 2;
  BUSINESS_TYPE_PARTNERSHIP = 3;
  BUSINESS_TYPE_PRIVATE_LIMITED = 4;
  BUSINESS_TYPE_PUBLIC_LIMITED = 5;
  BUSINESS_TYPE_LLP = 6;
  BUSINESS_TYPE_COOPERATIVE = 7;
  BUSINESS_TYPE_TRUST = 8;
  BUSINESS_TYPE_FPO = 9;
}

// AddressType defines the type of address
enum AddressType {
  ADDRESS_TYPE_UNSPECIFIED = 0;
  ADDRESS_TYPE_HOME = 1;
  ADDRESS_TYPE_BUSINESS = 2;
  ADDRESS_TYPE_BILLING = 3;
  ADDRESS_TYPE_SHIPPING = 4;
  ADDRESS_TYPE_WAREHOUSE = 5;
}

// Collaborator represents a complete collaborator entity
message Collaborator {
  // System fields
  string id = 1;
  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Timestamp updated_at = 3;
  optional google.protobuf.Timestamp deleted_at = 4;

  // User identification
  string user_id = 5;
  string username = 6;
  string email = 7;
  optional string phone = 8;

  // Personal information
  string first_name = 9;
  string last_name = 10;
  optional string middle_name = 11;
  optional string profile_picture = 12;
  optional string bio = 13;

  // Business information
  optional BusinessInfo business_info = 14;

  // Type and status
  CollaboratorType type = 15;
  CollaboratorStatus status = 16;

  // Organization
  optional string organization_id = 17;
  optional string organization_name = 18;

  // Location
  optional string location = 19;
  optional GeoCoordinates coordinates = 20;

  // Preferences
  optional string language_preference = 21;
  optional string timezone_preference = 22;
  optional string currency_preference = 23;

  // Metadata
  map<string, string> metadata = 24;
  repeated string tags = 25;

  // Audit
  string created_by = 26;
  optional string updated_by = 27;

  // Addresses (IDs from AAA service)
  repeated string address_ids = 28;
  optional Address primary_address = 29; // Expanded address from AAA
}

// BusinessInfo contains business-related information
message BusinessInfo {
  string business_name = 1;
  BusinessType business_type = 2;
  optional string gst_number = 3;  // GST format: 22AAAAA0000A1Z5
  optional string pan_number = 4;  // PAN format: AAAAA0000A
  optional string tax_id = 5;
  optional string business_license = 6;
  optional string business_phone = 7;
  optional string business_email = 8;
  optional string business_website = 9;
  optional string business_description = 10;

  // Financial information
  optional string bank_account_number = 11;
  optional string bank_name = 12;
  optional string bank_ifsc_code = 13;
  optional string bank_branch = 14;

  // Verification
  bool is_verified = 15;
  optional google.protobuf.Timestamp verified_at = 16;
  optional string verified_by = 17;
  optional string verification_notes = 18;
}

// GeoCoordinates represents geographical coordinates
message GeoCoordinates {
  double latitude = 1;
  double longitude = 2;
  optional float accuracy = 3; // in meters
}

// Address represents an address (synced with AAA service)
message Address {
  string id = 1;  // AAA address ID
  string line1 = 2;
  optional string line2 = 3;
  optional string line3 = 4;
  string city = 5;
  string state = 6;
  string country = 7;
  string postal_code = 8;
  AddressType type = 9;
  optional string landmark = 10;
  optional GeoCoordinates coordinates = 11;
  bool is_primary = 12;
  bool is_verified = 13;
}
```

### 2. collaborator_messages.proto

```protobuf
syntax = "proto3";

package kisanlink.collaborator.v1;

option go_package = "kisanlink-ecom/proto/gen/go/collaborator/v1;collaboratorpb";

import "collaborator/v1/collaborator_types.proto";
import "shared/pagination.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/field_mask.proto";
import "validate/validate.proto";

// CreateCollaboratorRequest for creating a new collaborator
message CreateCollaboratorRequest {
  // Required user information
  string user_id = 1 [(validate.rules).string = {
    min_len: 1,
    max_len: 50,
    pattern: "^[a-zA-Z0-9-]+$"
  }];

  string username = 2 [(validate.rules).string = {
    min_len: 3,
    max_len: 50,
    pattern: "^[a-zA-Z0-9_-]+$"
  }];

  string email = 3 [(validate.rules).string = {
    email: true,
    max_len: 100
  }];

  string first_name = 4 [(validate.rules).string = {
    min_len: 1,
    max_len: 100,
    pattern: "^[a-zA-Z\\s'-]+$"
  }];

  string last_name = 5 [(validate.rules).string = {
    min_len: 1,
    max_len: 100,
    pattern: "^[a-zA-Z\\s'-]+$"
  }];

  // Optional personal information
  optional string middle_name = 6 [(validate.rules).string = {
    max_len: 100,
    pattern: "^[a-zA-Z\\s'-]*$"
  }];

  optional string phone = 7 [(validate.rules).string = {
    pattern: "^\\+?[1-9]\\d{1,14}$" // E.164 format
  }];

  optional string profile_picture = 8 [(validate.rules).string = {
    max_len: 500,
    pattern: "^https?://.*"
  }];

  optional string bio = 9 [(validate.rules).string = {
    max_len: 1000
  }];

  // Type and organization
  CollaboratorType type = 10 [(validate.rules).enum = {
    defined_only: true,
    not_in: [0]
  }];

  optional string organization_id = 11 [(validate.rules).string = {
    max_len: 50,
    pattern: "^[a-zA-Z0-9-]*$"
  }];

  // Business information (required for VENDOR/BUYER types)
  optional CreateBusinessInfoRequest business_info = 12;

  // Address (will be created in AAA service)
  optional CreateAddressRequest address = 13;

  // Location and preferences
  optional string location = 14 [(validate.rules).string = {
    max_len: 200
  }];

  optional GeoCoordinates coordinates = 15;

  optional string language_preference = 16 [(validate.rules).string = {
    max_len: 10,
    pattern: "^[a-z]{2}(-[A-Z]{2})?$" // ISO 639-1
  }];

  optional string timezone_preference = 17 [(validate.rules).string = {
    max_len: 50
  }];

  optional string currency_preference = 18 [(validate.rules).string = {
    len: 3,
    pattern: "^[A-Z]{3}$" // ISO 4217
  }];

  // Metadata
  map<string, string> metadata = 19;
  repeated string tags = 20 [(validate.rules).repeated = {
    max_items: 20,
    items: {
      string: {
        min_len: 1,
        max_len: 50
      }
    }
  }];
}

// CreateBusinessInfoRequest for business information
message CreateBusinessInfoRequest {
  string business_name = 1 [(validate.rules).string = {
    min_len: 1,
    max_len: 200
  }];

  BusinessType business_type = 2 [(validate.rules).enum = {
    defined_only: true,
    not_in: [0]
  }];

  optional string gst_number = 3 [(validate.rules).string = {
    pattern: "^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$"
  }];

  optional string pan_number = 4 [(validate.rules).string = {
    pattern: "^[A-Z]{5}[0-9]{4}[A-Z]{1}$"
  }];

  optional string tax_id = 5 [(validate.rules).string = {
    max_len: 50
  }];

  optional string business_license = 6 [(validate.rules).string = {
    max_len: 100
  }];

  optional string business_phone = 7 [(validate.rules).string = {
    pattern: "^\\+?[1-9]\\d{1,14}$"
  }];

  optional string business_email = 8 [(validate.rules).string = {
    email: true,
    max_len: 100
  }];

  optional string business_website = 9 [(validate.rules).string = {
    max_len: 200,
    pattern: "^https?://.*"
  }];

  optional string business_description = 10 [(validate.rules).string = {
    max_len: 2000
  }];

  // Banking information
  optional string bank_account_number = 11 [(validate.rules).string = {
    max_len: 50
  }];

  optional string bank_name = 12 [(validate.rules).string = {
    max_len: 100
  }];

  optional string bank_ifsc_code = 13 [(validate.rules).string = {
    pattern: "^[A-Z]{4}0[A-Z0-9]{6}$" // Indian IFSC
  }];

  optional string bank_branch = 14 [(validate.rules).string = {
    max_len: 100
  }];
}

// CreateAddressRequest for address creation
message CreateAddressRequest {
  string line1 = 1 [(validate.rules).string = {
    min_len: 1,
    max_len: 200
  }];

  optional string line2 = 2 [(validate.rules).string = {
    max_len: 200
  }];

  optional string line3 = 3 [(validate.rules).string = {
    max_len: 200
  }];

  string city = 4 [(validate.rules).string = {
    min_len: 1,
    max_len: 100
  }];

  string state = 5 [(validate.rules).string = {
    min_len: 1,
    max_len: 100
  }];

  string country = 6 [(validate.rules).string = {
    min_len: 2,
    max_len: 100
  }];

  string postal_code = 7 [(validate.rules).string = {
    min_len: 3,
    max_len: 20
  }];

  AddressType type = 8 [(validate.rules).enum = {
    defined_only: true,
    not_in: [0]
  }];

  optional string landmark = 9 [(validate.rules).string = {
    max_len: 200
  }];

  optional GeoCoordinates coordinates = 10;

  bool is_primary = 11;
}

// UpdateCollaboratorRequest for updating collaborator
message UpdateCollaboratorRequest {
  string id = 1 [(validate.rules).string = {
    min_len: 1,
    max_len: 50
  }];

  // Field mask for partial updates
  google.protobuf.FieldMask field_mask = 2;

  // Fields that can be updated
  optional string email = 3 [(validate.rules).string = {
    email: true,
    max_len: 100
  }];

  optional string phone = 4 [(validate.rules).string = {
    pattern: "^\\+?[1-9]\\d{1,14}$"
  }];

  optional string first_name = 5 [(validate.rules).string = {
    min_len: 1,
    max_len: 100
  }];

  optional string last_name = 6 [(validate.rules).string = {
    min_len: 1,
    max_len: 100
  }];

  optional string middle_name = 7;
  optional string profile_picture = 8;
  optional string bio = 9;
  optional string location = 10;
  optional GeoCoordinates coordinates = 11;
  optional UpdateBusinessInfoRequest business_info = 12;
  optional string language_preference = 13;
  optional string timezone_preference = 14;
  optional string currency_preference = 15;
  map<string, string> metadata = 16;
  repeated string tags = 17;
}

// UpdateBusinessInfoRequest for updating business information
message UpdateBusinessInfoRequest {
  optional string business_name = 1;
  optional BusinessType business_type = 2;
  optional string business_phone = 3;
  optional string business_email = 4;
  optional string business_website = 5;
  optional string business_description = 6;
  optional string bank_account_number = 7;
  optional string bank_name = 8;
  optional string bank_ifsc_code = 9;
  optional string bank_branch = 10;
}

// DeactivateCollaboratorRequest for deactivating collaborator
message DeactivateCollaboratorRequest {
  string id = 1 [(validate.rules).string = {
    min_len: 1,
    max_len: 50
  }];

  optional string reason = 2 [(validate.rules).string = {
    max_len: 500
  }];

  optional CollaboratorStatus new_status = 3 [(validate.rules).enum = {
    defined_only: true,
    in: [2, 3] // Only INACTIVE or SUSPENDED
  }];
}

// GetCollaboratorRequest for retrieving collaborator
message GetCollaboratorRequest {
  string id = 1 [(validate.rules).string = {
    min_len: 1,
    max_len: 50
  }];

  // Whether to expand address from AAA service
  bool expand_address = 2;

  // Whether to include deleted collaborators
  bool include_deleted = 3;
}

// ListCollaboratorsRequest for listing collaborators
message ListCollaboratorsRequest {
  // Pagination
  optional shared.PaginationRequest pagination = 1;

  // Filters
  optional CollaboratorFilter filter = 2;

  // Sorting
  optional string sort_by = 3 [(validate.rules).string = {
    in: ["created_at", "updated_at", "username", "email", "first_name", "last_name"]
  }];

  optional shared.SortOrder sort_order = 4;

  // Whether to expand addresses from AAA service
  bool expand_addresses = 5;
}

// CollaboratorFilter for filtering collaborators
message CollaboratorFilter {
  repeated CollaboratorType types = 1;
  repeated CollaboratorStatus statuses = 2;
  optional string organization_id = 3;
  optional string search_query = 4; // Search in username, email, name
  optional bool has_business_info = 5;
  optional bool is_verified = 6;
  optional string location = 7;
  optional google.protobuf.Timestamp created_after = 8;
  optional google.protobuf.Timestamp created_before = 9;
  repeated string tags = 10;
  map<string, string> metadata = 11;
}

// CollaboratorResponse for single collaborator
message CollaboratorResponse {
  Collaborator collaborator = 1;

  // Additional computed fields
  optional CollaboratorStats stats = 2;
  optional CollaboratorPermissions permissions = 3;
}

// ListCollaboratorsResponse for multiple collaborators
message ListCollaboratorsResponse {
  repeated Collaborator collaborators = 1;
  shared.PaginationResponse pagination = 2;
  optional CollaboratorStats aggregate_stats = 3;
}

// StatusResponse for status update operations
message StatusResponse {
  bool success = 1;
  string message = 2;
  optional CollaboratorStatus new_status = 3;
  optional google.protobuf.Timestamp updated_at = 4;
}

// CollaboratorStats for analytics
message CollaboratorStats {
  int32 total_transactions = 1;
  double total_revenue = 2;
  double average_rating = 3;
  int32 total_reviews = 4;
  google.protobuf.Timestamp last_active_at = 5;
  google.protobuf.Timestamp member_since = 6;
}

// CollaboratorPermissions for access control
message CollaboratorPermissions {
  bool can_edit = 1;
  bool can_delete = 2;
  bool can_view_sensitive_info = 3;
  bool can_manage_addresses = 4;
  bool can_verify = 5;
}
```

### 3. collaborator_service.proto

```protobuf
syntax = "proto3";

package kisanlink.collaborator.v1;

option go_package = "kisanlink-ecom/proto/gen/go/collaborator/v1;collaboratorpb";

import "collaborator/v1/collaborator_messages.proto";
import "google/api/annotations.proto";
import "google/protobuf/empty.proto";
import "protoc-gen-openapiv2/options/annotations.proto";

// CollaboratorService provides gRPC endpoints for collaborator management
service CollaboratorService {
  option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_tag) = {
    description: "Collaborator management service for KisanLink platform"
  };

  // CreateCollaborator creates a new collaborator
  rpc CreateCollaborator(CreateCollaboratorRequest) returns (CollaboratorResponse) {
    option (google.api.http) = {
      post: "/api/v1/collaborators"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Create a new collaborator"
      description: "Creates a new collaborator with optional AAA address integration"
      tags: "Collaborators"
      security: {
        security_requirement: {
          key: "JWT"
          value: {}
        }
      }
    };
  }

  // UpdateCollaborator updates an existing collaborator
  rpc UpdateCollaborator(UpdateCollaboratorRequest) returns (CollaboratorResponse) {
    option (google.api.http) = {
      patch: "/api/v1/collaborators/{id}"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Update a collaborator"
      description: "Updates collaborator details with field mask support"
      tags: "Collaborators"
      security: {
        security_requirement: {
          key: "JWT"
          value: {}
        }
      }
    };
  }

  // DeactivateCollaborator deactivates or suspends a collaborator
  rpc DeactivateCollaborator(DeactivateCollaboratorRequest) returns (StatusResponse) {
    option (google.api.http) = {
      post: "/api/v1/collaborators/{id}/deactivate"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Deactivate a collaborator"
      description: "Soft deletes or suspends a collaborator account"
      tags: "Collaborators"
      security: {
        security_requirement: {
          key: "JWT"
          value: {}
        }
      }
    };
  }

  // GetCollaborator retrieves a single collaborator by ID
  rpc GetCollaborator(GetCollaboratorRequest) returns (CollaboratorResponse) {
    option (google.api.http) = {
      get: "/api/v1/collaborators/{id}"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Get a collaborator by ID"
      description: "Retrieves collaborator details with optional address expansion"
      tags: "Collaborators"
      security: {
        security_requirement: {
          key: "JWT"
          value: {}
        }
      }
    };
  }

  // ListCollaborators lists collaborators with filtering and pagination
  rpc ListCollaborators(ListCollaboratorsRequest) returns (ListCollaboratorsResponse) {
    option (google.api.http) = {
      get: "/api/v1/collaborators"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "List collaborators"
      description: "Lists collaborators with advanced filtering and pagination"
      tags: "Collaborators"
      security: {
        security_requirement: {
          key: "JWT"
          value: {}
        }
      }
    };
  }

  // VerifyCollaborator verifies business information
  rpc VerifyCollaborator(VerifyCollaboratorRequest) returns (VerifyCollaboratorResponse) {
    option (google.api.http) = {
      post: "/api/v1/collaborators/{id}/verify"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Verify a collaborator"
      description: "Verifies collaborator business information and documents"
      tags: "Collaborators"
      security: {
        security_requirement: {
          key: "JWT"
          value: {}
        }
      }
    };
  }

  // GetCollaboratorByGST retrieves a collaborator by GST number
  rpc GetCollaboratorByGST(GetCollaboratorByGSTRequest) returns (CollaboratorResponse) {
    option (google.api.http) = {
      get: "/api/v1/collaborators/gst/{gst_number}"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Get collaborator by GST number"
      description: "Retrieves collaborator by GST for deduplication checks"
      tags: "Collaborators"
      security: {
        security_requirement: {
          key: "JWT"
          value: {}
        }
      }
    };
  }

  // Health check endpoint
  rpc HealthCheck(google.protobuf.Empty) returns (HealthCheckResponse) {
    option (google.api.http) = {
      get: "/api/v1/collaborators/health"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Health check"
      description: "Checks service health and dependencies"
      tags: "Health"
    };
  }
}

// Additional request/response messages

message VerifyCollaboratorRequest {
  string id = 1;
  bool is_verified = 2;
  optional string verification_notes = 3;
  repeated string verified_documents = 4;
}

message VerifyCollaboratorResponse {
  bool success = 1;
  string message = 2;
  CollaboratorStatus new_status = 3;
  google.protobuf.Timestamp verified_at = 4;
  string verified_by = 5;
}

message GetCollaboratorByGSTRequest {
  string gst_number = 1 [(validate.rules).string = {
    pattern: "^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$"
  }];
  bool expand_address = 2;
}

message HealthCheckResponse {
  enum Status {
    STATUS_UNSPECIFIED = 0;
    STATUS_HEALTHY = 1;
    STATUS_DEGRADED = 2;
    STATUS_UNHEALTHY = 3;
  }

  Status status = 1;
  string version = 2;
  google.protobuf.Timestamp timestamp = 3;

  message Dependency {
    string name = 1;
    Status status = 2;
    optional string error = 3;
    optional int64 latency_ms = 4;
  }

  repeated Dependency dependencies = 4;
}
```

### 4. shared/common.proto

```protobuf
syntax = "proto3";

package kisanlink.shared;

option go_package = "kisanlink-ecom/proto/gen/go/shared;sharedpb";

import "google/protobuf/any.proto";
import "google/protobuf/timestamp.proto";

// Error represents a standard error response
message Error {
  string code = 1;
  string message = 2;
  map<string, string> details = 3;
  optional string trace_id = 4;
  google.protobuf.Timestamp timestamp = 5;
}

// Response wrapper for standard responses
message Response {
  bool success = 1;
  optional Error error = 2;
  optional google.protobuf.Any data = 3;
  map<string, string> metadata = 4;
}
```

### 5. shared/pagination.proto

```protobuf
syntax = "proto3";

package kisanlink.shared;

option go_package = "kisanlink-ecom/proto/gen/go/shared;sharedpb";

// PaginationRequest for requesting paginated data
message PaginationRequest {
  int32 page = 1;  // 1-indexed
  int32 page_size = 2;
  optional string cursor = 3;  // For cursor-based pagination
}

// PaginationResponse contains pagination metadata
message PaginationResponse {
  int32 page = 1;
  int32 page_size = 2;
  int32 total_pages = 3;
  int64 total_items = 4;
  bool has_next = 5;
  bool has_previous = 6;
  optional string next_cursor = 7;
  optional string previous_cursor = 8;
}

// SortOrder for sorting
enum SortOrder {
  SORT_ORDER_UNSPECIFIED = 0;
  SORT_ORDER_ASC = 1;
  SORT_ORDER_DESC = 2;
}
```

## Proto Generation Makefile

```makefile
# Proto generation configuration
PROTO_DIR := proto
PROTO_GEN_DIR := proto/gen
PROTO_IMPORTS := -I$(PROTO_DIR) \
	-I$(GOPATH)/src \
	-I$(GOPATH)/src/github.com/googleapis/googleapis \
	-I$(GOPATH)/src/github.com/envoyproxy/protoc-gen-validate

# Go output configuration
GO_OUT := $(PROTO_GEN_DIR)/go
GO_OPT := paths=source_relative

# Proto files
COLLABORATOR_PROTOS := $(shell find $(PROTO_DIR)/collaborator -name "*.proto")
SHARED_PROTOS := $(shell find $(PROTO_DIR)/shared -name "*.proto")
ALL_PROTOS := $(COLLABORATOR_PROTOS) $(SHARED_PROTOS)

.PHONY: proto
proto: clean-proto gen-proto

.PHONY: gen-proto
gen-proto: install-proto-deps
	@echo "Generating Go code from proto files..."
	@mkdir -p $(GO_OUT)

	# Generate Go code
	protoc $(PROTO_IMPORTS) \
		--go_out=$(GO_OUT) --go_opt=$(GO_OPT) \
		--go-grpc_out=$(GO_OUT) --go-grpc_opt=$(GO_OPT) \
		--validate_out="lang=go,paths=source_relative:$(GO_OUT)" \
		--grpc-gateway_out=$(GO_OUT) --grpc-gateway_opt=$(GO_OPT) \
		--openapiv2_out=$(PROTO_GEN_DIR)/openapi \
		$(ALL_PROTOS)

.PHONY: install-proto-deps
install-proto-deps:
	@echo "Installing proto dependencies..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

.PHONY: clean-proto
clean-proto:
	@echo "Cleaning generated proto files..."
	@rm -rf $(PROTO_GEN_DIR)

.PHONY: validate-proto
validate-proto:
	@echo "Validating proto files..."
	@for proto in $(ALL_PROTOS); do \
		protoc $(PROTO_IMPORTS) --lint_out=. $$proto || exit 1; \
	done
```

## Validation Rules Summary

| Field | Validation Rule | Purpose |
|-------|----------------|---------|
| user_id | Pattern: ^[a-zA-Z0-9-]+$, Max: 50 | Alphanumeric with hyphens |
| username | Pattern: ^[a-zA-Z0-9_-]+$, Min: 3, Max: 50 | Standard username format |
| email | Valid email, Max: 100 | RFC-compliant email |
| phone | E.164 format | International phone standard |
| GST Number | Indian GST format | Business validation |
| PAN Number | Indian PAN format | Tax identification |
| IFSC Code | Indian bank IFSC | Banking validation |
| Currency | ISO 4217 (3 letters) | Standard currency codes |
| Language | ISO 639-1 | Standard language codes |

## Generated Code Structure

```
proto/gen/
├── go/
│   ├── collaborator/
│   │   └── v1/
│   │       ├── collaborator_service.pb.go
│   │       ├── collaborator_service_grpc.pb.go
│   │       ├── collaborator_service.pb.gw.go
│   │       ├── collaborator_messages.pb.go
│   │       ├── collaborator_messages.pb.validate.go
│   │       ├── collaborator_types.pb.go
│   │       └── collaborator_types.pb.validate.go
│   └── shared/
│       ├── common.pb.go
│       ├── pagination.pb.go
│       └── validation.pb.go
└── openapi/
    └── collaborator/
        └── v1/
            └── collaborator_service.swagger.json
```