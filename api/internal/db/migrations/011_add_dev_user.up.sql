-- Add development user for testing
INSERT INTO users (id, external_id, provider, email, display_name, created_at, updated_at)
VALUES (
    'dev-user-1',
    'test@hexabase.com',
    'credentials',
    'test@hexabase.com',
    'Test User',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;