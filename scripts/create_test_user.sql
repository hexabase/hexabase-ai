-- Create test user and organization for local testing
-- Run this script with: docker exec -i hexabase-kaas-postgres-1 psql -U postgres -d hexabase < scripts/create_test_user.sql

-- Create a test user
INSERT INTO users (id, external_id, provider, email, display_name, created_at, updated_at)
VALUES (
    'test-user-001',
    'external-test-001', 
    'test',
    'test@hexabase.local',
    'Test User',
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Create a test organization
INSERT INTO organizations (id, name, created_at, updated_at)
VALUES (
    'test-org-001',
    'Test Organization',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Link user to organization as admin
INSERT INTO organization_users (organization_id, user_id, role, joined_at)
VALUES (
    'test-org-001',
    'test-user-001',
    'admin',
    NOW()
) ON CONFLICT (organization_id, user_id) DO NOTHING;

-- Display the created data
SELECT 'Test user created:' as message;
SELECT id, email, display_name FROM users WHERE id = 'test-user-001';

SELECT 'Test organization created:' as message;
SELECT id, name FROM organizations WHERE id = 'test-org-001';

SELECT 'User organization membership:' as message;
SELECT 
    u.email,
    o.name as organization,
    ou.role
FROM organization_users ou
JOIN users u ON u.id = ou.user_id
JOIN organizations o ON o.id = ou.organization_id
WHERE u.id = 'test-user-001';