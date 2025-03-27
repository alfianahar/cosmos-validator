# Validator Delegation Tracking System

**Title:** Cosmos Validator Delegation Tracking System  
**Author:** Alfian Nahar Aswinda  
**Created:** 24 March 2024  
**Modified:** 27 March 2024

## Synopsis

The Cosmos Validator Delegation Tracking System is a monitoring tool designed to track delegation activities within the Cosmos blockchain network. It continuously collects, stores, and provides analytical data about validator-delegator relationships. The system captures hourly snapshots and aggregates them into daily summaries. Additionally, it offers RESTful APIs that allow users to monitor delegation changes, analyze historical trends, and track specific delegator activities across validators.

### Motivation

Tracking validator delegations is essential for multiple stakeholders in the Cosmos ecosystem:

- **Validators**: Helps them monitor delegation changes to assess their staking position and reputation.
- **Delegators**: Assists in making informed delegation decisions by tracking validator performance.
- **Network Analysts**: Provides insights into delegation trends and network health.

Without a robust tracking system, stakeholders lack visibility into delegation dynamics, making it difficult to detect anomalies or make data-driven decisions.

## Technical Specification

### System Architecture

The system is built on a modular architecture comprising four key components:

1. **Data Collection Service**

   - Polls the Cosmos API for delegation data at configurable intervals (default: hourly)
   - Implements a retry mechanism with exponential backoff and jitter to handle API failures
   - Uses a Watchlist model to track specific validator-delegator pairs
   - Computes hourly changes in delegation amounts

2. **Data Aggregation Service**

   - Runs daily to compile hourly snapshots into daily summaries
   - Ensures complete data aggregation by executing at midnight
   - Extracts the latest delegation amounts per day for trend analysis

3. **Database Layer**

   - Uses PostgreSQL (currently i used NEON) with GORM ORM
   - Implements optimized query patterns and proper indexing for efficient data retrieval
   - Ensures data integrity with foreign key constraints
   - Manages connection pooling to support high-concurrency scenarios

4. **API Service**
   - Provides RESTful endpoints using the Gin web framework
   - Standardizes API responses with pagination support
   - Includes health check endpoints for monitoring system status
   - Delivers both raw and aggregated delegation data

### Component Interaction Flow

```
┌─────────────────┐   Fetch Data   ┌─────────────────┐
│                 │◄──────────────►│                 │
│  Cosmos API     │                │ Data Collector  │
│                 │                │ Service         │
└─────────────────┘                └─────────┬───────┘
                                             │
                                             │ Store Data
                                             ▼
┌─────────────────┐   Query Data   ┌─────────────────┐
│                 │◄──────────────►│                 │
│ API Layer       │                │ Database        │
│                 │                │                 │
└────────┬────────┘                └──────────▲──────┘
         │                                    │
         │ Aggregate                          │
         │                                    │
┌────────┴────────┐                           │
│                 │                           │
│ Daily Aggregator│                           │
│ Service         │                           │
└─────────────────┘                           │
         │                                    │
         ▼                                    ▼
┌─────────────────┐                   ┌─────────────────┐
│ API Consumers   │                   │ Watchlist       │
└─────────────────┘                   └─────────────────┘
```

### Data Structures

The system defines the following core data models:

- **Watchlist Model**: Specifies which validator-delegator pairs to track.
- **Hourly Delegation Model**: Stores hourly snapshots of delegation amounts and calculates changes.
- **Daily Delegation Model**: Aggregates daily delegation data for trend analysis.

### Data Transfer Objects (DTOs)

- `HourlyDelegationDTO`: Transfers hourly delegation data to API consumers.
- `DailyDelegationDTO`: Transfers daily delegation data to API consumers.
- `DelegationResponse`: Standardizes all delegation-related API responses with pagination.
- `ErrorResponse`: Provides a consistent error format for API failures.

### API Specification

#### Delegation Endpoints

1. **Get Hourly Delegations**

   - **Endpoint**: `GET /api/v1/validators/:validator/delegations/hourly`
   - **Parameters**:
     - `validator` (required): Validator address
     - `page`: Page number (default: 1)
     - `limit`: Items per page (default: 50, max: 100)
   - **Response**: Hourly delegation data

2. **Get Daily Delegations**

   - **Endpoint**: `GET /api/v1/validators/:validator/delegations/daily`
   - **Parameters**: Similar to the hourly endpoint
   - **Response**: Daily delegation data

3. **Get Delegator History**
   - **Endpoint**: `GET /api/v1/validators/:validator/delegator/:delegator/history`
   - **Parameters**:
     - `validator` (required): Validator address
     - `delegator` (required): Delegator address
     - `page`, `limit`: Pagination
   - **Response**: Delegator-specific historical data

#### Watchlist Endpoints

1. **Add to Watchlist**

   - **Endpoint**: `POST /api/v1/watchlist`
   - **Request Body**: Validator and delegator details
   - **Response**: Confirmation of addition

2. **Get Watchlist**

   - **Endpoint**: `GET /api/v1/watchlist`
   - **Response**: List of tracked validator-delegator pairs

3. **Remove from Watchlist**
   - **Endpoint**: `DELETE /api/v1/watchlist/:id`
   - **Response**: Confirmation of removal

#### Health Check Endpoints

1. **System Health**

   - **Endpoint**: `GET /api/v1/health`
   - **Response**: Overall system status

2. **Data Health**
   - **Endpoint**: `GET /api/v1/health/data`
   - **Response**: Data freshness and availability status

### Error Handling

- **Custom Error Types**: Handles different error scenarios using structured error responses.
- **API Failure Retry Mechanism**:
  - **Exponential Backoff**: Retries with increasing intervals to prevent API overload.
  - **Jitter**: Adds randomness to retry intervals to avoid synchronized failures.
  - **Status Code Awareness**: Handles specific HTTP errors differently.
  - **Logging**: Logs retry attempts for debugging and analytics.

### Health Monitoring

- Database connectivity verification
- API service availability checks
- Data freshness monitoring with alerts for stale data

## Getting Started

### Prerequisites

- Go 1.20 or higher
- PostgreSQL 14.0 or higher

### Installation

1. Clone the repository
2. Configure environment variables in `.env`
3. Run `go mod tidy` to install dependencies
4. Start the application with `go run cmd/main.go`

### Configuration

- **Required Environment Variables**:
  - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `SSLMODE`
- **Optional**:
  - `DEBUG`: Enable debug mode
  - `SERVER_HOST`, `SERVER_PORT`: API server configuration

This README provides a detailed technical specification and deployment guide for the Cosmos Validator Delegation Tracking System.
