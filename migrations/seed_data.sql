-- migrations/seed_data.sql

-- Create a default 'admin' role
INSERT INTO roles (code, description) VALUES ('admin', 'Administrator with full system access') ON CONFLICT (code) DO NOTHING;

-- Create some basic permissions
INSERT INTO permissions (resource, action, description) VALUES
    ('system', 'read', 'Read system information'),
    ('system', 'write', 'Modify system settings'),
    ('users', 'create', 'Create new users'),
    ('users', 'read', 'Read user information'),
    ('users', 'update', 'Update user information'),
    ('users', 'delete', 'Delete users'),
    ('roles', 'assign', 'Assign roles to users')
ON CONFLICT (resource, action) DO NOTHING;

-- Assign all permissions to the 'admin' role
DO $$
DECLARE
    admin_role_id INT;
    perm_id INT;
BEGIN
    SELECT id INTO admin_role_id FROM roles WHERE code = 'admin';

    FOR perm_id IN SELECT id FROM permissions
    LOOP
        INSERT INTO role_permissions (role_id, permission_id) VALUES (admin_role_id, perm_id) ON CONFLICT DO NOTHING;
    END LOOP;
END $$;
