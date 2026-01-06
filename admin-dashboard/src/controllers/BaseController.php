<?php

namespace App\Controllers;

use App\Database\Database;
use App\Auth\Auth;
use Twig\Environment;
use Twig\Loader\FilesystemLoader;

abstract class BaseController
{
    protected Database $db;
    protected Auth $auth;
    protected array $config;
    protected Environment $twig;

    public function __construct(Database $db, Auth $auth, array $config)
    {
        $this->db = $db;
        $this->auth = $auth;
        $this->config = $config;
        $this->initializeTwig();
    }

    protected function initializeTwig(): void
    {
        $loader = new FilesystemLoader(__DIR__ . '/../../views');
        $this->twig = new Environment($loader, [
            'cache' => $this->config['app']['env'] === 'production' ? '/tmp/twig_cache' : false,
            'debug' => $this->config['app']['debug'],
            'auto_reload' => true,
        ]);

        // Add global variables
        $this->twig->addGlobal('auth', $this->auth);
        $this->twig->addGlobal('user', $this->auth->user());
        $this->twig->addGlobal('config', $this->config);
        $this->twig->addGlobal('app_name', $this->config['app']['name']);
    }

    protected function render(string $template, array $data = []): void
    {
        echo $this->twig->render($template, $data);
    }

    protected function json(array $data, int $statusCode = 200): void
    {
        http_response_code($statusCode);
        header('Content-Type: application/json');
        echo json_encode($data);
    }

    protected function redirect(string $url, int $statusCode = 302): void
    {
        http_response_code($statusCode);
        header("Location: $url");
        exit;
    }

    protected function requireAuth(): void
    {
        if (!$this->auth->check()) {
            $this->redirect('/login');
        }

        // Check if password change required
        if ($this->auth->mustChangePassword() && $_SERVER['REQUEST_URI'] !== '/change-password') {
            $this->redirect('/change-password');
        }
    }

    protected function requirePermission(string $permission): void
    {
        $this->requireAuth();
        $this->auth->requirePermission($permission);
    }

    protected function requireRole(string $role): void
    {
        $this->requireAuth();
        $this->auth->requireRole($role);
    }

    protected function back(): void
    {
        $referer = $_SERVER['HTTP_REFERER'] ?? '/';
        $this->redirect($referer);
    }

    protected function getInput(string $key, $default = null)
    {
        return $_POST[$key] ?? $_GET[$key] ?? $default;
    }

    protected function flashMessage(string $type, string $message): void
    {
        $_SESSION['flash'][$type] = $message;
    }

    protected function getFlashMessages(): array
    {
        $messages = $_SESSION['flash'] ?? [];
        unset($_SESSION['flash']);
        return $messages;
    }
}
