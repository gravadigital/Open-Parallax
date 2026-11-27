-- Test identities of the end-to-end tests (docker-compose.test.yml, service e2e-seed).
-- The roles admin and organizer can only be granted by writing users.role in the
-- database, so the test profile seeds one user of each. Password of both:
-- e2e-password-123 (bcrypt, cost 10). Local test values only: never in production.
INSERT INTO users (name, lastname, email, password_hash, role, created_at, updated_at)
VALUES
  ('Admin E2E', '', 'admin.e2e@openparallax.test',
   '$2a$10$Tg/sbLghBrlWEa7Yn1WLCO1WY70z.1Sl41obaZzjMQkccejWsnqHO', 'admin', now(), now()),
  ('Organizer E2E', '', 'organizer.e2e@openparallax.test',
   '$2a$10$Tg/sbLghBrlWEa7Yn1WLCO1WY70z.1Sl41obaZzjMQkccejWsnqHO', 'organizer', now(), now())
ON CONFLICT (email) DO UPDATE
  SET role = EXCLUDED.role, password_hash = EXCLUDED.password_hash;
