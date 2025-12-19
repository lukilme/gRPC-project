CREATE DATABASE IF NOT EXISTS `order`;
CREATE DATABASE IF NOT EXISTS `payment`;

USE `order`;

CREATE TABLE IF NOT EXISTS orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    customer_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at BIGINT NOT NULL,
    INDEX idx_customer (customer_id),
    INDEX idx_created (created_at)
);

CREATE TABLE IF NOT EXISTS order_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    unit_price FLOAT NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    INDEX idx_order (order_id),
    INDEX idx_product (product_id)
);

CREATE DATABASE IF NOT EXISTS `payment`;
USE `payment`;

CREATE TABLE IF NOT EXISTS payments (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_id BIGINT NOT NULL,
    total_price FLOAT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    bill_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_order (order_id),
    INDEX idx_created (created_at)
);

INSERT INTO `order`.orders (customer_id, status, created_at) VALUES
(1001, 'completed', UNIX_TIMESTAMP(NOW())),
(1002, 'pending', UNIX_TIMESTAMP(NOW()));

INSERT INTO `order`.order_items (order_id, product_id, quantity, unit_price) VALUES
(1, 101, 2, 29.99),
(1, 102, 1, 49.99),
(2, 103, 3, 9.99);