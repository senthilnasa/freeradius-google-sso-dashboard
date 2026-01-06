<?php

namespace App\Models;

class AdminUser
{
    public int $id;
    public string $username;
    public string $email;
    public string $full_name;
    public string $role;
    public bool $is_active;
    public bool $force_password_change;
    public ?string $last_login;
    public ?string $last_login_ip;
    public int $failed_login_attempts;
    public ?string $locked_until;
    public string $created_at;

    public function __construct(array $data)
    {
        $this->id = (int)$data['id'];
        $this->username = $data['username'];
        $this->email = $data['email'];
        $this->full_name = $data['full_name'] ?? '';
        $this->role = $data['role'];
        $this->is_active = (bool)$data['is_active'];
        $this->force_password_change = (bool)$data['force_password_change'];
        $this->last_login = $data['last_login'] ?? null;
        $this->last_login_ip = $data['last_login_ip'] ?? null;
        $this->failed_login_attempts = (int)($data['failed_login_attempts'] ?? 0);
        $this->locked_until = $data['locked_until'] ?? null;
        $this->created_at = $data['created_at'];
    }

    public function isSuperAdmin(): bool
    {
        return $this->role === 'superadmin';
    }

    public function isNetAdmin(): bool
    {
        return $this->role === 'netadmin';
    }

    public function isHelpdesk(): bool
    {
        return $this->role === 'helpdesk';
    }

    public function getRoleName(): string
    {
        return match($this->role) {
            'superadmin' => 'Super Administrator',
            'netadmin' => 'Network Administrator',
            'helpdesk' => 'Helpdesk',
            default => ucfirst($this->role),
        };
    }

    public function canManageUsers(): bool
    {
        return in_array($this->role, ['superadmin', 'netadmin']);
    }

    public function canEditVLANs(): bool
    {
        return in_array($this->role, ['superadmin', 'netadmin']);
    }

    public function canExportReports(): bool
    {
        return in_array($this->role, ['superadmin', 'netadmin']);
    }

    public function isLocked(): bool
    {
        if (!$this->locked_until) {
            return false;
        }

        return strtotime($this->locked_until) > time();
    }

    public function toArray(): array
    {
        return [
            'id' => $this->id,
            'username' => $this->username,
            'email' => $this->email,
            'full_name' => $this->full_name,
            'role' => $this->role,
            'role_name' => $this->getRoleName(),
            'is_active' => $this->is_active,
            'force_password_change' => $this->force_password_change,
            'last_login' => $this->last_login,
            'last_login_ip' => $this->last_login_ip,
            'failed_login_attempts' => $this->failed_login_attempts,
            'is_locked' => $this->isLocked(),
            'created_at' => $this->created_at,
        ];
    }
}
