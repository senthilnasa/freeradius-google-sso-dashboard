<?php

require_once __DIR__ . '/../vendor/autoload.php';

use App\Database\Database;
use App\Auth\Auth;
use Dotenv\Dotenv;

// Load environment (if .env file exists)
if (file_exists(__DIR__ . '/../.env')) {
    try {
        $dotenv = Dotenv::createImmutable(__DIR__ . '/..');
        $dotenv->load();
    } catch (Exception $e) {
        // Ignore .env loading errors - environment variables may be set via docker-compose
    }
}

// Load configuration
$config = require __DIR__ . '/../config/config.php';

// Set timezone
date_default_timezone_set($config['datetime']['timezone']);

// Start session
session_set_cookie_params([
    'lifetime' => $config['session']['lifetime'],
    'path' => '/',
    'secure' => $config['session']['secure'],
    'httponly' => $config['session']['httponly'],
    'samesite' => $config['session']['samesite'],
]);
session_name($config['session']['name']);
session_start();

// Initialize database
$db = new Database($config['database']);

// Initialize auth
$auth = new Auth($db, $config);

// Simple router
$uri = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);
$method = $_SERVER['REQUEST_METHOD'];

// Route handling
$routes = [
    'GET /' => 'DashboardController@index',
    'GET /login' => 'AuthController@showLogin',
    'POST /login' => 'AuthController@login',
    'GET /logout' => 'AuthController@logout',
    'POST /logout' => 'AuthController@logout',
    'GET /change-password' => 'AuthController@showChangePassword',
    'POST /change-password' => 'AuthController@changePassword',

    // Dashboard
    'GET /dashboard' => 'DashboardController@index',

    // Reports
    'GET /reports/daily-summary' => 'ReportsController@dailySummary',
    'GET /reports/monthly-usage' => 'ReportsController@monthlyUsage',
    'GET /reports/failed-logins' => 'ReportsController@failedLogins',
    'GET /reports/user-distribution' => 'ReportsController@userDistribution',
    'GET /reports/system-health' => 'ReportsController@systemHealth',

    // Export
    'GET /export/csv' => 'ReportsController@exportCSV',
    'GET /export/excel' => 'ReportsController@exportExcel',

    // Users
    'GET /users' => 'UsersController@index',
    'GET /users/create' => 'UsersController@create',
    'POST /users/create' => 'UsersController@store',
];

// Dynamic routes with parameters
$dynamicRoutes = [
    '#^GET /users/(\d+)/edit$#' => ['UsersController', 'edit'],
    '#^POST /users/(\d+)/edit$#' => ['UsersController', 'update'],
    '#^POST /users/(\d+)/delete$#' => ['UsersController', 'delete'],
];

// Try exact match first
$routeKey = "$method $uri";
if (isset($routes[$routeKey])) {
    [$controller, $action] = explode('@', $routes[$routeKey]);
    $controllerClass = "App\\Controllers\\$controller";
    $ctrl = new $controllerClass($db, $auth, $config);
    $ctrl->$action();
    exit;
}

// Try dynamic routes
foreach ($dynamicRoutes as $pattern => $handler) {
    if (preg_match($pattern, $routeKey, $matches)) {
        array_shift($matches); // Remove full match
        [$controller, $action] = $handler;
        $controllerClass = "App\\Controllers\\$controller";
        $ctrl = new $controllerClass($db, $auth, $config);
        call_user_func_array([$ctrl, $action], $matches);
        exit;
    }
}

// 404
http_response_code(404);
echo "404 - Not Found";
