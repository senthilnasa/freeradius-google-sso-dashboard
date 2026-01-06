<?php

namespace App\Services;

use App\Database\Database;
use PhpOffice\PhpSpreadsheet\Spreadsheet;
use PhpOffice\PhpSpreadsheet\Writer\Xlsx;
use League\Csv\Writer;

class ExportService
{
    private Database $db;
    private array $config;

    public function __construct(Database $db, array $config)
    {
        $this->db = $db;
        $this->config = $config;
    }

    public function exportCSV(string $type, array $params): void
    {
        $data = $this->getData($type, $params);

        if (empty($data)) {
            die('No data to export');
        }

        // Create CSV writer
        $csv = Writer::createFromFileObject(new \SplTempFileObject());

        // Add headers
        $csv->insertOne(array_keys($data[0]));

        // Add data
        $csv->insertAll($data);

        // Output
        $filename = $this->getFilename($type, 'csv');
        header('Content-Type: text/csv; charset=UTF-8');
        header("Content-Disposition: attachment; filename=\"{$filename}\"");
        echo $csv->toString();
        exit;
    }

    public function exportExcel(string $type, array $params): void
    {
        $data = $this->getData($type, $params);

        if (empty($data)) {
            die('No data to export');
        }

        // Create spreadsheet
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();

        // Set title
        $sheet->setTitle($this->getSheetTitle($type));

        // Add headers
        $headers = array_keys($data[0]);
        $col = 'A';
        foreach ($headers as $header) {
            $sheet->setCellValue($col . '1', ucwords(str_replace('_', ' ', $header)));
            $sheet->getStyle($col . '1')->getFont()->setBold(true);
            $col++;
        }

        // Add data
        $row = 2;
        foreach ($data as $record) {
            $col = 'A';
            foreach ($record as $value) {
                $sheet->setCellValue($col . $row, $value);
                $col++;
            }
            $row++;
        }

        // Auto-size columns
        foreach (range('A', $col) as $column) {
            $sheet->getColumnDimension($column)->setAutoSize(true);
        }

        // Output
        $filename = $this->getFilename($type, 'xlsx');
        header('Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');
        header("Content-Disposition: attachment; filename=\"{$filename}\"");
        header('Cache-Control: max-age=0');

        $writer = new Xlsx($spreadsheet);
        $writer->save('php://output');
        exit;
    }

    private function getData(string $type, array $params): array
    {
        return match($type) {
            'daily-summary' => $this->getDailySummaryData($params),
            'monthly-usage' => $this->getMonthlyUsageData($params),
            'failed-logins' => $this->getFailedLoginsData($params),
            'user-distribution' => $this->getUserDistributionData($params),
            'active-sessions' => $this->getActiveSessionsData($params),
            'audit-logs' => $this->getAuditLogsData($params),
            default => [],
        };
    }

    private function getDailySummaryData(array $params): array
    {
        $date = $params['date'] ?? date('Y-m-d');

        return $this->db->fetchAll(
            "SELECT
                username,
                reply,
                vlan_id,
                user_type,
                email_domain,
                nasipaddress,
                authdate
             FROM radpostauth
             WHERE DATE(authdate) = ?
             ORDER BY authdate DESC",
            [$date]
        );
    }

    private function getMonthlyUsageData(array $params): array
    {
        $month = $params['month'] ?? date('Y-m');

        return $this->db->fetchAll(
            "SELECT
                DATE(acctstarttime) as date,
                username,
                vlan_id,
                user_type,
                acctsessiontime,
                acctinputoctets,
                acctoutputoctets,
                nasipaddress
             FROM radacct
             WHERE DATE_FORMAT(acctstarttime, '%Y-%m') = ?
             ORDER BY acctstarttime DESC",
            [$month]
        );
    }

    private function getFailedLoginsData(array $params): array
    {
        $dateFrom = $params['date_from'] ?? date('Y-m-d', strtotime('-7 days'));
        $dateTo = $params['date_to'] ?? date('Y-m-d');

        return $this->db->fetchAll(
            "SELECT
                username,
                email,
                email_domain,
                failure_reason,
                error_type,
                ip_address,
                created_at
             FROM auth_failures
             WHERE DATE(created_at) BETWEEN ? AND ?
             ORDER BY created_at DESC",
            [$dateFrom, $dateTo]
        );
    }

    private function getUserDistributionData(array $params): array
    {
        return $this->db->fetchAll(
            "SELECT
                user_type,
                vlan_id,
                COUNT(DISTINCT username) as unique_users,
                COUNT(*) as total_auths,
                DATE(authdate) as date
             FROM radpostauth
             WHERE reply = 'Access-Accept'
             AND authdate >= DATE_SUB(NOW(), INTERVAL 30 DAY)
             GROUP BY user_type, vlan_id, DATE(authdate)
             ORDER BY date DESC, total_auths DESC"
        );
    }

    private function getActiveSessionsData(array $params): array
    {
        return $this->db->fetchAll(
            "SELECT
                session_id,
                username,
                email,
                vlan_id,
                user_type,
                ip_address,
                mac_address,
                created_at,
                last_activity,
                expires_at
             FROM sessions
             WHERE is_active = 1 AND expires_at > NOW()
             ORDER BY created_at DESC"
        );
    }

    private function getAuditLogsData(array $params): array
    {
        $dateFrom = $params['date_from'] ?? date('Y-m-d', strtotime('-30 days'));
        $dateTo = $params['date_to'] ?? date('Y-m-d');

        return $this->db->fetchAll(
            "SELECT
                admin_username,
                action,
                resource_type,
                resource_id,
                description,
                ip_address,
                created_at
             FROM admin_audit_logs
             WHERE DATE(created_at) BETWEEN ? AND ?
             ORDER BY created_at DESC",
            [$dateFrom, $dateTo]
        );
    }

    private function getFilename(string $type, string $extension): string
    {
        $date = date('Y-m-d_His');
        $typeSlug = str_replace('-', '_', $type);
        return "radius_{$typeSlug}_{$date}.{$extension}";
    }

    private function getSheetTitle(string $type): string
    {
        return match($type) {
            'daily-summary' => 'Daily Summary',
            'monthly-usage' => 'Monthly Usage',
            'failed-logins' => 'Failed Logins',
            'user-distribution' => 'User Distribution',
            'active-sessions' => 'Active Sessions',
            'audit-logs' => 'Audit Logs',
            default => 'Export',
        };
    }
}
