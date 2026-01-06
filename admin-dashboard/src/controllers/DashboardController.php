<?php

namespace App\Controllers;

class DashboardController extends BaseController
{
    public function index()
    {
        $this->requireAuth();

        // Get quick stats
        $stats = [
            'active_sessions' => $this->db->fetchColumn(
                "SELECT COUNT(*) FROM sessions WHERE is_active = 1 AND expires_at > NOW()"
            ),
            'total_users_today' => $this->db->fetchColumn(
                "SELECT COUNT(DISTINCT username) FROM radpostauth WHERE DATE(authdate) = CURDATE()"
            ),
            'total_auth_today' => $this->db->fetchColumn(
                "SELECT COUNT(*) FROM radpostauth WHERE DATE(authdate) = CURDATE()"
            ),
            'failed_auth_today' => $this->db->fetchColumn(
                "SELECT COUNT(*) FROM radpostauth WHERE DATE(authdate) = CURDATE() AND reply != 'Access-Accept'"
            ),
        ];

        // Recent authentications
        $recentAuths = $this->db->fetchAll(
            "SELECT username, reply, vlan_id, user_type, authdate, nasipaddress
             FROM radpostauth
             ORDER BY authdate DESC
             LIMIT 10"
        );

        // Active sessions
        $activeSessions = $this->db->fetchAll(
            "SELECT session_id, username, email, vlan_id, user_type, ip_address, created_at, last_activity
             FROM sessions
             WHERE is_active = 1 AND expires_at > NOW()
             ORDER BY last_activity DESC
             LIMIT 10"
        );

        // VLAN distribution
        $vlanStats = $this->db->fetchAll(
            "SELECT vlan_id, user_type, COUNT(*) as count
             FROM sessions
             WHERE is_active = 1
             GROUP BY vlan_id, user_type
             ORDER BY count DESC"
        );

        $this->render('dashboard/index.twig', [
            'title' => 'Dashboard',
            'stats' => $stats,
            'recent_auths' => $recentAuths,
            'active_sessions' => $activeSessions,
            'vlan_stats' => $vlanStats,
            'flash' => $this->getFlashMessages(),
        ]);
    }
}
