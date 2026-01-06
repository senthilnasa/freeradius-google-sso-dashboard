<?php

return [
    // Application
    'app' => [
        'name' => getenv('APP_NAME') ?: 'RADIUS Admin Dashboard',
        'env' => getenv('APP_ENV') ?: 'production',
        'url' => getenv('APP_URL') ?: 'https://admin.yourdomain.com',
        'debug' => getenv('APP_ENV') === 'development',
    ],

    // Database
    'database' => [
        'host' => getenv('DB_HOST') ?: 'mysql',
        'port' => getenv('DB_PORT') ?: 3306,
        'name' => getenv('DB_NAME') ?: 'radius',
        'user' => getenv('DB_USER') ?: 'radius',
        'password' => getenv('DB_PASSWORD') ?: 'radiuspassword',
        'charset' => 'utf8mb4',
    ],

    // Session
    'session' => [
        'name' => 'RADIUS_ADMIN_SESSION',
        'lifetime' => (int)(getenv('SESSION_LIFETIME') ?: 3600), // 1 hour
        'secure' => getenv('APP_ENV') === 'production',
        'httponly' => true,
        'samesite' => 'Lax',
    ],

    // Security
    'security' => [
        'session_secret' => getenv('SESSION_SECRET'),
        'password_min_length' => 8,
        'password_require_special' => true,
        'max_login_attempts' => 5,
        'lockout_duration' => 900, // 15 minutes
    ],

    // Pagination
    'pagination' => [
        'per_page' => 25,
        'max_per_page' => 100,
    ],

    // Date/Time
    'datetime' => [
        'timezone' => getenv('TZ') ?: 'UTC',
        'format' => 'Y-m-d H:i:s',
        'date_format' => 'Y-m-d',
        'time_format' => 'H:i:s',
    ],

    // Export
    'export' => [
        'max_rows' => 10000,
        'temp_dir' => '/tmp',
    ],

    // Roles and Permissions
    'roles' => [
        'superadmin' => [
            'name' => 'Super Administrator',
            'permissions' => ['*'], // All permissions
        ],
        'netadmin' => [
            'name' => 'Network Administrator',
            'permissions' => [
                'users.view',
                'users.create',
                'users.edit',
                'users.delete',
                'reports.view',
                'reports.export',
                'vlans.view',
                'vlans.edit',
                'sessions.view',
                'audit.view',
            ],
        ],
        'helpdesk' => [
            'name' => 'Helpdesk',
            'permissions' => [
                'reports.view',
                'sessions.view',
                'audit.view',
            ],
        ],
    ],

    // Charts
    'charts' => [
        'default_days' => 30,
        'colors' => [
            'primary' => '#667eea',
            'success' => '#10b981',
            'warning' => '#f59e0b',
            'danger' => '#ef4444',
            'info' => '#3b82f6',
        ],
    ],
];
