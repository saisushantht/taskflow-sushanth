INSERT INTO users (id, name, email, password) VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'Test User',
    'test@example.com',
    '$2a$12$558kPSqXsIqJdnkfTxKIvOS9.L32zWRzYfSibchIy.O3U9kyW.PGK'
);

INSERT INTO projects (id, name, description, owner_id) VALUES (
    'b0000000-0000-0000-0000-000000000001',
    'Demo Project',
    'A sample project to get started',
    'a0000000-0000-0000-0000-000000000001'
);

INSERT INTO tasks (title, description, status, priority, project_id, assignee_id) VALUES
    ('Setup repository', 'Initialize the project repo', 'done', 'high', 'b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001'),
    ('Build API', 'Implement REST endpoints', 'in_progress', 'high', 'b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001'),
    ('Write tests', 'Add integration tests', 'todo', 'medium', 'b0000000-0000-0000-0000-000000000001', NULL);
