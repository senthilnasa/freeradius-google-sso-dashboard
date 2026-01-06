<?php

namespace App\Auth;

use App\Database\Database;
use App\Models\AdminUser;

class Auth
{
    private Database $db;
    private array $config;
    private ?AdminUser $user = null;

    public function __construct(Database $db, array $config)
    {
        $this->db = $db;
        $this->config = $config;
        $this->loadUserFromSession();
    }

    public function attempt(string $username, string $password): bool
    {
        // Get user from database
        $userData = $this->db->fetchOne(
            'SELECT * FROM admin_users WHERE username = ? AND is_active = 1',
            [$username]
        );

        if (!$userData) {
            $this->logFailedAttempt($username, 'User not found');
            return false;
        }

        // Check if account is locked
        if ($this->isAccountLocked($userData)) {
            $this->logFailedAttempt($username, 'Account locked');
            return false;
        }

        // Verify password
        if (!password_verify($password, $userData['password_hash'])) {
            $this->incrementFailedAttempts($userData['id']);
            $this->logFailedAttempt($username, 'Invalid password');
            return false;
        }

        // Reset failed attempts
        $this->resetFailedAttempts($userData['id']);

        // Update last login
        $this->updateLastLogin($userData['id'], $_SERVER['REMOTE_ADDR']);

        // Create session
        $this->user = new AdminUser($userData);
        $this->createSession();

        // Log successful login
        $this->logAudit('LOGIN', 'User logged in successfully');

        return true;
    }

    public function logout(): void
    {
        if ($this->user) {
            $this->logAudit('LOGOUT', 'User logged out');
        }

        $this->destroySession();
        $this->user = null;
    }

    public function check(): bool
    {
        return $this->user !== null;
    }

    public function user(): ?AdminUser
    {
        return $this->user;
    }

    public function hasRole(string $role): bool
    {
        return $this->user && $this->user->role === $role;
    }

    public function hasPermission(string $permission): bool
    {
        if (!$this->user) {
            return false;
        }

        $rolePermissions = $this->config['roles'][$this->user->role]['permissions'] ?? [];

        // Superadmin has all permissions
        if (in_array('*', $rolePermissions)) {
            return true;
        }

        return in_array($permission, $rolePermissions);
    }

    public function requirePermission(string $permission): void
    {
        if (!$this->hasPermission($permission)) {
            http_response_code(403);
            die('Access denied: Insufficient permissions');
        }
    }

    public function requireRole(string $role): void
    {
        if (!$this->hasRole($role)) {
            http_response_code(403);
            die('Access denied: Insufficient role');
        }
    }

    private function loadUserFromSession(): void
    {
        if (!isset($_SESSION['user_id'])) {
            return;
        }

        $userData = $this->db->fetchOne(
            'SELECT * FROM admin_users WHERE id = ? AND is_active = 1',
            [$_SESSION['user_id']]
        );

        if ($userData) {
            $this->user = new AdminUser($userData);
        } else {
            $this->destroySession();
        }
    }

    private function createSession(): void
    {
        session_regenerate_id(true);
        $_SESSION['user_id'] = $this->user->id;
        $_SESSION['username'] = $this->user->username;
        $_SESSION['role'] = $this->user->role;
        $_SESSION['login_time'] = time();
    }

    private function destroySession(): void
    {
        $_SESSION = [];
        if (ini_get('session.use_cookies')) {
            $params = session_get_cookie_params();
            setcookie(
                session_name(),
                '',
                time() - 42000,
                $params['path'],
                $params['domain'],
                $params['secure'],
                $params['httponly']
            );
        }
        session_destroy();
    }

    private function isAccountLocked(array $userData): bool
    {
        if (!$userData['locked_until']) {
            return false;
        }

        $lockedUntil = strtotime($userData['locked_until']);
        if ($lockedUntil > time()) {
            return true;
        }

        // Unlock account if lock period has expired
        $this->db->update(
            'admin_users',
            ['locked_until' => null, 'failed_login_attempts' => 0],
            'id = ?',
            [$userData['id']]
        );

        return false;
    }

    private function incrementFailedAttempts(int $userId): void
    {
        $this->db->query(
            'UPDATE admin_users
             SET failed_login_attempts = failed_login_attempts + 1
             WHERE id = ?',
            [$userId]
        );

        // Check if account should be locked
        $attempts = $this->db->fetchColumn(
            'SELECT failed_login_attempts FROM admin_users WHERE id = ?',
            [$userId]
        );

        if ($attempts >= $this->config['security']['max_login_attempts']) {
            $lockUntil = date(
                'Y-m-d H:i:s',
                time() + $this->config['security']['lockout_duration']
            );

            $this->db->update(
                'admin_users',
                ['locked_until' => $lockUntil],
                'id = ?',
                [$userId]
            );
        }
    }

    private function resetFailedAttempts(int $userId): void
    {
        $this->db->update(
            'admin_users',
            ['failed_login_attempts' => 0, 'locked_until' => null],
            'id = ?',
            [$userId]
        );
    }

    private function updateLastLogin(int $userId, string $ipAddress): void
    {
        $this->db->update(
            'admin_users',
            [
                'last_login' => date('Y-m-d H:i:s'),
                'last_login_ip' => $ipAddress,
            ],
            'id = ?',
            [$userId]
        );
    }

    private function logFailedAttempt(string $username, string $reason): void
    {
        error_log("Failed login attempt for user: {$username}, reason: {$reason}");
    }

    private function logAudit(string $action, string $description): void
    {
        if (!$this->user) {
            return;
        }

        $this->db->insert('admin_audit_logs', [
            'admin_user_id' => $this->user->id,
            'admin_username' => $this->user->username,
            'action' => $action,
            'description' => $description,
            'ip_address' => $_SERVER['REMOTE_ADDR'] ?? null,
            'user_agent' => $_SERVER['HTTP_USER_AGENT'] ?? null,
        ]);
    }

    public function mustChangePassword(): bool
    {
        return $this->user && $this->user->force_password_change;
    }

    public function changePassword(int $userId, string $newPassword): bool
    {
        if (strlen($newPassword) < $this->config['security']['password_min_length']) {
            return false;
        }

        $hash = password_hash($newPassword, PASSWORD_BCRYPT);

        $this->db->update(
            'admin_users',
            [
                'password_hash' => $hash,
                'force_password_change' => 0,
            ],
            'id = ?',
            [$userId]
        );

        $this->logAudit('PASSWORD_CHANGE', 'Password changed');

        return true;
    }
}
