<?php

namespace App\Controllers;

class AuthController extends BaseController
{
    public function showLogin()
    {
        // Redirect if already authenticated
        if ($this->auth->check()) {
            $this->redirect('/');
        }

        $this->render('auth/login.twig', [
            'title' => 'Login',
            'flash' => $this->getFlashMessages(),
        ]);
    }

    public function login()
    {
        $username = $this->getInput('username');
        $password = $this->getInput('password');

        if (!$username || !$password) {
            $this->flashMessage('error', 'Username and password are required');
            $this->redirect('/login');
        }

        if ($this->auth->attempt($username, $password)) {
            $this->flashMessage('success', 'Login successful');
            $this->redirect('/');
        } else {
            $this->flashMessage('error', 'Invalid credentials or account locked');
            $this->redirect('/login');
        }
    }

    public function logout()
    {
        $this->auth->logout();
        $this->flashMessage('success', 'Logged out successfully');
        $this->redirect('/login');
    }

    public function showChangePassword()
    {
        $this->requireAuth();

        $this->render('auth/change-password.twig', [
            'title' => 'Change Password',
            'flash' => $this->getFlashMessages(),
            'required' => $this->auth->mustChangePassword(),
        ]);
    }

    public function changePassword()
    {
        $this->requireAuth();

        $currentPassword = $this->getInput('current_password');
        $newPassword = $this->getInput('new_password');
        $confirmPassword = $this->getInput('confirm_password');

        // Validation
        if (!$currentPassword || !$newPassword || !$confirmPassword) {
            $this->flashMessage('error', 'All fields are required');
            $this->back();
            return;
        }

        if ($newPassword !== $confirmPassword) {
            $this->flashMessage('error', 'New passwords do not match');
            $this->back();
            return;
        }

        if (strlen($newPassword) < $this->config['security']['password_min_length']) {
            $this->flashMessage('error', 'Password must be at least ' . $this->config['security']['password_min_length'] . ' characters');
            $this->back();
            return;
        }

        // Verify current password
        $user = $this->auth->user();
        $userData = $this->db->fetchOne(
            'SELECT password_hash FROM admin_users WHERE id = ?',
            [$user->id]
        );

        if (!password_verify($currentPassword, $userData['password_hash'])) {
            $this->flashMessage('error', 'Current password is incorrect');
            $this->back();
            return;
        }

        // Change password
        if ($this->auth->changePassword($user->id, $newPassword)) {
            $this->flashMessage('success', 'Password changed successfully');
            $this->redirect('/');
        } else {
            $this->flashMessage('error', 'Failed to change password');
            $this->back();
        }
    }
}
