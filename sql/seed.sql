USE ecommerce;

-- Seed an admin user with bcrypt-hashed password for 'admin123'.
-- Admin credentials: email='admin@example.com', password='admin123'

INSERT INTO users (name,email,phone,password,role) VALUES
('Admin User','admin@example.com','081234567890','$2a$10$5fR35C1vwGu9lYNE.hkREOVo4ZFK520mJSXaBIdzJ4xJ7LQ1orqbO','admin');

-- Create store for admin explicitly
INSERT INTO stores (user_id,name) VALUES (LAST_INSERT_ID(), 'Admin Store');
