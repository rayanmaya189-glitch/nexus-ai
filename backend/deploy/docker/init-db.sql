-- AeroXe Nexus AI Platform - Database Initialization
-- PostgreSQL 18 with extensions

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create service databases
CREATE DATABASE nexus_identity;
CREATE DATABASE nexus_gateway;
CREATE DATABASE nexus_agent;
CREATE DATABASE nexus_rag;
CREATE DATABASE nexus_audit;

GRANT ALL PRIVILEGES ON DATABASE nexus_identity TO nexus;
GRANT ALL PRIVILEGES ON DATABASE nexus_gateway TO nexus;
GRANT ALL PRIVILEGES ON DATABASE nexus_agent TO nexus;
GRANT ALL PRIVILEGES ON DATABASE nexus_rag TO nexus;
GRANT ALL PRIVILEGES ON DATABASE nexus_audit TO nexus;
