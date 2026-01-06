<?php
$hash = password_hash("admin123", PASSWORD_DEFAULT);
echo "Generated hash: $hash\n";

$pdo = new PDO("mysql:host=mysql;dbname=radius", "radius", "krea_radius_password_2024");
$stmt = $pdo->prepare("UPDATE admin_users SET password_hash = ? WHERE username = ?");

$stmt->execute([$hash, "admin"]);
$stmt->execute([$hash, "netadmin"]);
$stmt->execute([$hash, "helpdesk"]);

echo "Updated all admin users successfully\n";

// Verify
$verify = $pdo->query("SELECT username, password_hash FROM admin_users")->fetchAll(PDO::FETCH_ASSOC);
foreach ($verify as $user) {
    $matches = password_verify("admin123", $user['password_hash']) ? "YES" : "NO";
    echo "{$user['username']}: password_verify = $matches\n";
}
