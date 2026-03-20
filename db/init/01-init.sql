-- Multi-Tenant SaaS Database Schema
-- This script sets up the base database with Row-Level Security (RLS)

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) NOT NULL DEFAULT 'MEMBER',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    email_verified BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(tenant_id, email)
);

-- Roles table (for custom tenant roles)
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

-- User role assignments
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id),
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);

-- Subscription plans
CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    price_cents INTEGER NOT NULL DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    interval VARCHAR(20) NOT NULL, -- monthly, yearly
    features JSONB DEFAULT '[]'::jsonb,
    limits JSONB DEFAULT '{}'::jsonb,
    stripe_price_id VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Subscriptions
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id),
    status VARCHAR(50) NOT NULL DEFAULT 'incomplete',
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    canceled_at TIMESTAMPTZ,
    trial_start TIMESTAMPTZ,
    trial_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id)
);

-- Usage/Events table
CREATE TABLE IF NOT EXISTS usage_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,
    event_name VARCHAR(255) NOT NULL,
    properties JSONB DEFAULT '{}'::jsonb,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_tenant_email ON users(tenant_id, email);
CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);
CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_tenant_id ON user_roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant_id ON subscriptions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_usage_events_tenant_id ON usage_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_usage_events_timestamp ON usage_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_usage_events_type ON usage_events(event_type);

-- Enable Row-Level Security on tenant-aware tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_events ENABLE ROW LEVEL SECURITY;

-- RLS Policies: These policies use the app.tenant_id setting
-- which should be set at connection time after authentication

-- Users policy: Users can only see users from their tenant
DROP POLICY IF EXISTS users_tenant_policy ON users;
CREATE POLICY users_tenant_policy ON users
    FOR ALL
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- Roles policy
DROP POLICY IF EXISTS roles_tenant_policy ON roles;
CREATE POLICY roles_tenant_policy ON roles
    FOR ALL
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- User roles policy
DROP POLICY IF EXISTS user_roles_tenant_policy ON user_roles;
CREATE POLICY user_roles_tenant_policy ON user_roles
    FOR ALL
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- Subscriptions policy
DROP POLICY IF EXISTS subscriptions_tenant_policy ON subscriptions;
CREATE POLICY subscriptions_tenant_policy ON subscriptions
    FOR ALL
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- Usage events policy
DROP POLICY IF EXISTS usage_events_tenant_policy ON usage_events;
CREATE POLICY usage_events_tenant_policy ON usage_events
    FOR ALL
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- Function to set tenant context
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_id UUID)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.tenant_id', tenant_id::text, false);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get current tenant from context
CREATE OR REPLACE FUNCTION get_current_tenant_id()
RETURNS UUID AS $$
BEGIN
    RETURN current_setting('app.tenant_id', true)::UUID;
END;
$$ LANGUAGE plpgsql STABLE;

-- Insert default plans
INSERT INTO plans (name, slug, description, price_cents, currency, interval, features, limits) VALUES
    (
        'Free',
        'free',
        'Free plan for small teams',
        0,
        'USD',
        'monthly',
        '["Up to 5 users", "Basic analytics", "Community support"]'::jsonb,
        '{"users": 5, "storage_mb": 100}'::jsonb
    ),
    (
        'Pro',
        'pro',
        'Pro plan for growing teams',
        4900,
        'USD',
        'monthly',
        '["Up to 50 users", "Advanced analytics", "Priority support", "API access"]'::jsonb,
        '{"users": 50, "storage_mb": 10000}'::jsonb
    ),
    (
        'Enterprise',
        'enterprise',
        'Enterprise plan for large organizations',
        0,
        'USD',
        'monthly',
        '["Unlimited users", "Custom analytics", "24/7 support", "SSO", "SLA"]'::jsonb,
        '{"users": -1, "storage_mb": -1}'::jsonb
    )
ON CONFLICT (slug) DO NOTHING;

-- Grant necessary permissions (adjust as needed)
GRANT USAGE ON SCHEMA public TO postgres;
GRANT ALL ON ALL TABLES IN SCHEMA public TO postgres;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO postgres;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO postgres;

-- Create admin function for creating first tenant/user
CREATE OR REPLACE FUNCTION create_initial_tenant(
    tenant_name VARCHAR,
    tenant_slug VARCHAR,
    user_email VARCHAR,
    user_password_hash VARCHAR,
    user_first_name VARCHAR DEFAULT 'Admin',
    user_last_name VARCHAR DEFAULT 'User'
)
RETURNS TABLE(tenant_id UUID, user_id UUID) AS $$
DECLARE
    new_tenant_id UUID;
    new_user_id UUID;
    admin_role_id UUID;
BEGIN
    -- Create tenant
    INSERT INTO tenants (name, slug)
    VALUES (tenant_name, tenant_slug)
    RETURNING tenants.id INTO new_tenant_id;

    -- Create owner role for this tenant
    INSERT INTO roles (tenant_id, name, description, permissions)
    VALUES (
        new_tenant_id,
        'OWNER',
        'Full access to all resources',
        '["*:*"]'::jsonb
    )
    RETURNING roles.id INTO admin_role_id;

    -- Create user with OWNER role
    INSERT INTO users (tenant_id, email, password_hash, first_name, last_name, role)
    VALUES (new_tenant_id, user_email, user_password_hash, user_first_name, user_last_name, 'OWNER')
    RETURNING users.id INTO new_user_id;

    -- Assign role to user
    INSERT INTO user_roles (user_id, tenant_id, role_id)
    VALUES (new_user_id, new_tenant_id, admin_role_id);

    RETURN QUERY SELECT new_tenant_id, new_user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION create_initial_tenant IS 'Creates initial tenant, admin role, and first user. Use this to bootstrap the system.';
