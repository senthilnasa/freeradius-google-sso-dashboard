<?php

namespace App\Services;

use App\Database\Database;

class UserService
{
    private Database $db;
    private array $config;

    public function __construct(Database $db, array $config)
    {
        $this->db = $db;
        $this->config = $config;
    }

    public function getAllUsers(): array
    {
        return $this->db->fetchAll(
            "SELECT id, username, email, full_name, role, is_active,
                    force_password_change, last_login, last_login_ip,
                    failed_login_attempts, locked_until, created_at
             FROM admin_users
             ORDER BY created_at DESC"
        );
    }

    public function getUserById(int $id): ?array
    {
        return $this->db->fetchOne(
            "SELECT * FROM admin_users WHERE id = ?",
            [$id]
        );
    }

    public function createUser(array $data, int $createdBy): int
    {
        $passwordHash = password_hash($data['password'], PASSWORD_BCRYPT);

        return $this->db->insert('admin_users', [
            'username' => $data['username'],
            'email' => $data['email'],
            'password_hash' => $passwordHash,
            'full_name' => $data['full_name'],
            'role' => $data['role'],
            'is_active' => 1,
            'force_password_change' => 1, // Force password change on first login
            'created_by' => $createdBy,
        ]);
    }

    public function updateUser(int $id, array $data, int $updatedBy): void
    {
        $updateData = [
            'email' => $data['email'],
            'full_name' => $data['full_name'],
            'role' => $data['role'],
            'is_active' => $data['is_active'],
            'force_password_change' => $data['force_password_change'],
        ];

        // Update password if provided
        if (isset($data['password']) && !empty($data['password'])) {
            $updateData['password_hash'] = password_hash($data['password'], PASSWORD_BCRYPT);
        }

        $this->db->update('admin_users', $updateData, 'id = ?', [$id]);

        // Log the action
        $this->logAudit($updatedBy, 'UPDATE_USER', 'admin_user', $id, 'Updated user');
    }

    public function deleteUser(int $id, int $deletedBy): void
    {
        // Get user info for audit
        $user = $this->getUserById($id);

        // Delete user
        $this->db->delete('admin_users', 'id = ?', [$id]);

        // Log the action
        $this->logAudit($deletedBy, 'DELETE_USER', 'admin_user', $id, "Deleted user: {$user['username']}");
    }

    public function validate(array $data): array
    {
        $errors = [];

        // Username
        if (empty($data['username'])) {
            $errors[] = 'Username is required';
        } elseif (strlen($data['username']) < 3) {
            $errors[] = 'Username must be at least 3 characters';
        } elseif ($this->usernameExists($data['username'])) {
            $errors[] = 'Username already exists';
        }

        // Email
        if (empty($data['email'])) {
            $errors[] = 'Email is required';
        } elseif (!filter_var($data['email'], FILTER_VALIDATE_EMAIL)) {
            $errors[] = 'Invalid email format';
        } elseif ($this->emailExists($data['email'])) {
            $errors[] = 'Email already exists';
        }

        // Password
        if (empty($data['password'])) {
            $errors[] = 'Password is required';
        } elseif (strlen($data['password']) < $this->config['security']['password_min_length']) {
            $errors[] = 'Password must be at least ' . $this->config['security']['password_min_length'] . ' characters';
        }

        // Role
        if (empty($data['role'])) {
            $errors[] = 'Role is required';
        } elseif (!isset($this->config['roles'][$data['role']])) {
            $errors[] = 'Invalid role';
        }

        return $errors;
    }

    private function usernameExists(string $username): bool
    {
        $count = $this->db->fetchColumn(
            "SELECT COUNT(*) FROM admin_users WHERE username = ?",
            [$username]
        );
        return $count > 0;
    }

    private function emailExists(string $email): bool
    {
        $count = $this->db->fetchColumn(
            "SELECT COUNT(*) FROM admin_users WHERE email = ?",
            [$email]
        );
        return $count > 0;
    }

    private function logAudit(int $userId, string $action, string $resourceType, int $resourceId, string $description): void
    {
        $user = $this->getUserById($userId);

        $this->db->insert('admin_audit_logs', [
            'admin_user_id' => $userId,
            'admin_username' => $user['username'],
            'action' => $action,
            'resource_type' => $resourceType,
            'resource_id' => (string)$resourceId,
            'description' => $description,
            'ip_address' => $_SERVER['REMOTE_ADDR'] ?? null,
            'user_agent' => $_SERVER['HTTP_USER_AGENT'] ?? null,
        ]);
    }
}
