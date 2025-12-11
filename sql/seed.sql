USE ecommerce;

-- Seed an admin user. You may want to replace the password with a bcrypt hash.
-- To make this admin usable with the code's bcrypt check, insert a bcrypt-hashed password for 'admin123'.

INSERT INTO users (name,email,phone,password,role) VALUES
('Admin User','admin@example.com','081234567890','changeme','admin');

-- Create store for admin explicitly
INSERT INTO stores (user_id,name) VALUES (LAST_INSERT_ID(), 'Admin Store');
