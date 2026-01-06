<?php

namespace App\Controllers;

use App\Services\ExportService;

class ReportsController extends BaseController
{
    public function dailySummary()
    {
        $this->requireAuth();

        $date = $_GET['date'] ?? date('Y-m-d');

        // Get daily stats
        $stats = [
            'total_attempts' => $this->db->fetchColumn(
                "SELECT COUNT(*) FROM radpostauth WHERE DATE(authdate) = ?",
                [$date]
            ),
            'successful' => $this->db->fetchColumn(
                "SELECT COUNT(*) FROM radpostauth WHERE DATE(authdate) = ? AND reply = 'Access-Accept'",
                [$date]
            ),
            'failed' => $this->db->fetchColumn(
                "SELECT COUNT(*) FROM radpostauth WHERE DATE(authdate) = ? AND reply != 'Access-Accept'",
                [$date]
            ),
            'unique_users' => $this->db->fetchColumn(
                "SELECT COUNT(DISTINCT username) FROM radpostauth WHERE DATE(authdate) = ?",
                [$date]
            ),
        ];

        // Hourly breakdown
        $hourlyStats = $this->db->fetchAll(
            "SELECT HOUR(authdate) as hour, 
                    COUNT(*) as total,
                    SUM(CASE WHEN reply = 'Access-Accept' THEN 1 ELSE 0 END) as successful,
                    SUM(CASE WHEN reply != 'Access-Accept' THEN 1 ELSE 0 END) as failed
             FROM radpostauth 
             WHERE DATE(authdate) = ?
             GROUP BY HOUR(authdate)
             ORDER BY hour",
            [$date]
        );

        // VLAN distribution
        $vlanStats = $this->db->fetchAll(
            "SELECT vlan_id, user_type, COUNT(*) as count
             FROM radpostauth 
             WHERE DATE(authdate) = ? AND reply = 'Access-Accept'
             GROUP BY vlan_id, user_type
             ORDER BY count DESC",
            [$date]
        );

        $this->render('reports/daily-summary.twig', [
            'date' => $date,
            'stats' => $stats,
            'hourlyStats' => $hourlyStats,
            'vlanStats' => $vlanStats,
        ]);
    }

    public function monthlyUsage()
    {
        $this->requireAuth();

        $month = $_GET['month'] ?? date('Y-m');
        $startDate = $month . '-01';
        $endDate = date('Y-m-t', strtotime($startDate));

        // Daily breakdown for the month
        $dailyStats = $this->db->fetchAll(
            "SELECT DATE(authdate) as date,
                    COUNT(*) as total,
                    SUM(CASE WHEN reply = 'Access-Accept' THEN 1 ELSE 0 END) as successful,
                    COUNT(DISTINCT username) as unique_users
             FROM radpostauth 
             WHERE DATE(authdate) BETWEEN ? AND ?
             GROUP BY DATE(authdate)
             ORDER BY date",
            [$startDate, $endDate]
        );

        // Monthly totals
        $totals = [
            'total_attempts' => $this->db->fetchColumn(
                "SELECT COUNT(*) FROM radpostauth WHERE DATE(authdate) BETWEEN ? AND ?",
                [$startDate, $endDate]
            ),
            'successful' => $this->db->fetchColumn(
                "SELECT COUNT(*) FROM radpostauth WHERE DATE(authdate) BETWEEN ? AND ? AND reply = 'Access-Accept'",
                [$startDate, $endDate]
            ),
            'unique_users' => $this->db->fetchColumn(
                "SELECT COUNT(DISTINCT username) FROM radpostauth WHERE DATE(authdate) BETWEEN ? AND ?",
                [$startDate, $endDate]
            ),
        ];

        $this->render('reports/monthly-usage.twig', [
            'month' => $month,
            'dailyStats' => $dailyStats,
            'totals' => $totals,
        ]);
    }

    public function failedLogins()
    {
        $this->requireAuth();

        $date = $_GET['date'] ?? date('Y-m-d');
        $threshold = (int)($_GET['threshold'] ?? 3);

        // Get failed logins grouped by username
        $failedLogins = $this->db->fetchAll(
            "SELECT username, COUNT(*) as attempts, 
                    MAX(authdate) as last_attempt,
                    GROUP_CONCAT(DISTINCT reply) as error_types
             FROM radpostauth 
             WHERE DATE(authdate) = ? AND reply != 'Access-Accept'
             GROUP BY username
             HAVING attempts >= ?
             ORDER BY attempts DESC",
            [$date, $threshold]
        );

        // Recent failed attempts
        $recentFailed = $this->db->fetchAll(
            "SELECT username, reply, authdate, nasipaddress
             FROM radpostauth 
             WHERE DATE(authdate) = ? AND reply != 'Access-Accept'
             ORDER BY authdate DESC
             LIMIT 100",
            [$date]
        );

        $this->render('reports/failed-logins.twig', [
            'date' => $date,
            'threshold' => $threshold,
            'failedLogins' => $failedLogins,
            'recentFailed' => $recentFailed,
        ]);
    }

    public function userDistribution()
    {
        $this->requireAuth();

        $period = $_GET['period'] ?? 'today';
        
        switch ($period) {
            case 'week':
                $dateCondition = "DATE(authdate) >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)";
                break;
            case 'month':
                $dateCondition = "DATE(authdate) >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)";
                break;
            default:
                $dateCondition = "DATE(authdate) = CURDATE()";
        }

        // User type distribution
        $distribution = $this->db->fetchAll(
            "SELECT user_type, vlan_id, COUNT(*) as count,
                    COUNT(DISTINCT username) as unique_users
             FROM radpostauth 
             WHERE $dateCondition AND reply = 'Access-Accept'
             GROUP BY user_type, vlan_id
             ORDER BY count DESC"
        );

        // Domain distribution
        $domainStats = $this->db->fetchAll(
            "SELECT SUBSTRING_INDEX(username, '@', -1) as domain,
                    COUNT(*) as count,
                    COUNT(DISTINCT username) as unique_users
             FROM radpostauth 
             WHERE $dateCondition AND reply = 'Access-Accept'
             GROUP BY domain
             ORDER BY count DESC"
        );

        $this->render('reports/user-distribution.twig', [
            'period' => $period,
            'distribution' => $distribution,
            'domainStats' => $domainStats,
        ]);
    }

    public function systemHealth()
    {
        $this->requireAuth();

        // Database stats
        $dbStats = [
            'total_auth_records' => $this->db->fetchColumn("SELECT COUNT(*) FROM radpostauth"),
            'total_sessions' => $this->db->fetchColumn("SELECT COUNT(*) FROM sessions"),
            'active_sessions' => $this->db->fetchColumn("SELECT COUNT(*) FROM sessions WHERE is_active = 1 AND expires_at > NOW()"),
            'vlan_rules' => $this->db->fetchColumn("SELECT COUNT(*) FROM vlan_rules WHERE enabled = 1"),
        ];

        // NAS devices
        $nasDevices = $this->db->fetchAll("SELECT * FROM nas ORDER BY shortname");

        // Recent activity
        $recentActivity = $this->db->fetchAll(
            "SELECT DATE_FORMAT(authdate, '%Y-%m-%d %H:00:00') as auth_hour, COUNT(*) as count
             FROM radpostauth 
             WHERE authdate >= DATE_SUB(NOW(), INTERVAL 24 HOUR)
             GROUP BY DATE_FORMAT(authdate, '%Y-%m-%d %H:00:00')
             ORDER BY auth_hour DESC
             LIMIT 24"
        );

        $this->render('reports/system-health.twig', [
            'dbStats' => $dbStats,
            'nasDevices' => $nasDevices,
            'recentActivity' => $recentActivity,
        ]);
    }

    public function exportCSV()
    {
        $this->requireAuth();

        $type = $_GET['type'] ?? 'daily';
        $date = $_GET['date'] ?? date('Y-m-d');

        $exportService = new ExportService($this->db, $this->config);
        $exportService->exportCSV($type, ['date' => $date]);
    }

    public function exportExcel()
    {
        $this->requireAuth();

        $type = $_GET['type'] ?? 'daily';
        $date = $_GET['date'] ?? date('Y-m-d');

        $exportService = new ExportService($this->db, $this->config);
        $exportService->exportExcel($type, ['date' => $date]);
    }

    private function getExportData(string $type, string $date): array
    {
        switch ($type) {
            case 'daily':
                return $this->db->fetchAll(
                    "SELECT username, reply, vlan_id, user_type, authdate, nasipaddress
                     FROM radpostauth 
                     WHERE DATE(authdate) = ?
                     ORDER BY authdate DESC",
                    [$date]
                );
            case 'failed':
                return $this->db->fetchAll(
                    "SELECT username, reply, authdate, nasipaddress
                     FROM radpostauth 
                     WHERE DATE(authdate) = ? AND reply != 'Access-Accept'
                     ORDER BY authdate DESC",
                    [$date]
                );
            default:
                return [];
        }
    }
}
