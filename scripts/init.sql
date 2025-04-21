-- Make sure to use the correct database
USE user_db;

-- Create users table (GORM would create automatically, but we can set up some initial data here)
CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  email VARCHAR(100) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(50) NULL,
  last_name VARCHAR(50) NULL,
  INDEX idx_users_email (email),
  INDEX idx_users_deleted_at (deleted_at)
);

-- Create addresses table
CREATE TABLE IF NOT EXISTS addresses (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  street VARCHAR(255) NOT NULL,
  city VARCHAR(100) NOT NULL,
  state VARCHAR(100) NULL,
  postal_code VARCHAR(20) NOT NULL,
  country VARCHAR(100) NOT NULL,
  is_default BOOLEAN DEFAULT false,
  INDEX idx_addresses_user_id (user_id),
  INDEX idx_addresses_deleted_at (deleted_at),
  CONSTRAINT fk_addresses_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Insert test user data (password: test123, this is a bcrypt hash)
INSERT INTO users (email, password_hash, first_name, last_name, created_at, updated_at)
VALUES ('test@example.com', '$2a$10$rrSg5.VhICHpYMZYb0WMf.RjQWKprfKyGomIm5P5/YtgVQz81oE7G', 'Test', 'User', NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- Insert test address
INSERT INTO addresses (user_id, street, city, state, postal_code, country, is_default, created_at, updated_at)
VALUES (
  (SELECT id FROM users WHERE email = 'test@example.com'),
  '5 Zhongguancun South Street', 
  'Beijing', 
  'Haidian District', 
  '100081', 
  'China', 
  true, 
  NOW(), 
  NOW()
)
ON DUPLICATE KEY UPDATE updated_at = NOW(); 