-- Migration: Create Work Orders Table
-- Description: Stores work order/ticket information
-- Date: Initial migration

CREATE TABLE IF NOT EXISTS work_orders (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Pending',
    unit VARCHAR(255) NOT NULL,
    photo_url VARCHAR(500),
    requester_id INT UNSIGNED NOT NULL,
    assignee_id INT UNSIGNED NULL,
    taken_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    completed_by_id INT UNSIGNED NULL,
    completion_note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (requester_id) REFERENCES users(id),
    FOREIGN KEY (assignee_id) REFERENCES users(id),
    FOREIGN KEY (completed_by_id) REFERENCES users(id),
    INDEX idx_status (status),
    INDEX idx_unit (unit),
    INDEX idx_requester (requester_id),
    INDEX idx_assignee (assignee_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
