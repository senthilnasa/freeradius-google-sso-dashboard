<?php

namespace App\Controllers;

use App\Services\UserService;

class UsersController extends BaseController
{
    private UserService $userService;

    public function __construct($db, $auth, $config)
    {
        parent::__construct($db, $auth, $config);
        $this->userService = new UserService($db, $config);
    }

    public function index()
    {
        $this->requirePermission('users.view');

        $users = $this->userService->getAllUsers();

        $this->render('users/index.twig', [
            'title' => 'User Management',
            'users' => $users,
            'flash' => $this->getFlashMessages(),
        ]);
    }

    public function create()
    {
        $this->requirePermission('users.create');

        $this->render('users/create.twig', [
            'title' => 'Create User',
            'roles' => array_keys($this->config['roles']),
            'flash' => $this->getFlashMessages(),
        ]);
    }

    public function store()
    {
        $this->requirePermission('users.create');

        $data = [
            'username' => $this->getInput('username'),
            'email' => $this->getInput('email'),
            'full_name' => $this->getInput('full_name'),
            'role' => $this->getInput('role'),
            'password' => $this->getInput('password'),
        ];

        // Validation
        $errors = $this->userService->validate($data);
        if (!empty($errors)) {
            $this->flashMessage('error', implode('<br>', $errors));
            $this->back();
            return;
        }

        // Create user
        try {
            $userId = $this->userService->createUser($data, $this->auth->user()->id);
            $this->flashMessage('success', 'User created successfully');
            $this->redirect('/users');
        } catch (\Exception $e) {
            $this->flashMessage('error', 'Failed to create user: ' . $e->getMessage());
            $this->back();
        }
    }

    public function edit(string $id)
    {
        $this->requirePermission('users.edit');

        $user = $this->userService->getUserById((int)$id);
        if (!$user) {
            $this->flashMessage('error', 'User not found');
            $this->redirect('/users');
            return;
        }

        $this->render('users/edit.twig', [
            'title' => 'Edit User',
            'user' => $user,
            'roles' => array_keys($this->config['roles']),
            'flash' => $this->getFlashMessages(),
        ]);
    }

    public function update(string $id)
    {
        $this->requirePermission('users.edit');

        $data = [
            'email' => $this->getInput('email'),
            'full_name' => $this->getInput('full_name'),
            'role' => $this->getInput('role'),
            'is_active' => $this->getInput('is_active') ? 1 : 0,
            'force_password_change' => $this->getInput('force_password_change') ? 1 : 0,
        ];

        // Update password if provided
        $newPassword = $this->getInput('password');
        if ($newPassword) {
            $data['password'] = $newPassword;
        }

        try {
            $this->userService->updateUser((int)$id, $data, $this->auth->user()->id);
            $this->flashMessage('success', 'User updated successfully');
            $this->redirect('/users');
        } catch (\Exception $e) {
            $this->flashMessage('error', 'Failed to update user: ' . $e->getMessage());
            $this->back();
        }
    }

    public function delete(string $id)
    {
        $this->requirePermission('users.delete');

        // Prevent self-deletion
        if ((int)$id === $this->auth->user()->id) {
            $this->flashMessage('error', 'Cannot delete your own account');
            $this->back();
            return;
        }

        try {
            $this->userService->deleteUser((int)$id, $this->auth->user()->id);
            $this->flashMessage('success', 'User deleted successfully');
        } catch (\Exception $e) {
            $this->flashMessage('error', 'Failed to delete user: ' . $e->getMessage());
        }

        $this->redirect('/users');
    }
}
