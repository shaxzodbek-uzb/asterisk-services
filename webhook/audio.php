<?php
/**
 * Simple Audio File Server
 * Returns audio file for given path or 404 if not found
 */

// Configuration
$recordingsPath = '/var/spool/asterisk/monitor/';

// Get file parameter
$file = $_GET['file'] ?? '';

if (empty($file)) {
    http_response_code(400);
    echo json_encode(['error' => 'Missing file parameter']);
    exit;
}

// Build full file path
$filePath = $recordingsPath . $file;

// Check if file exists
if (!file_exists($filePath)) {
    http_response_code(404);
    echo json_encode(['error' => 'File not found']);
    exit;
}

// Get file info
$filesize = filesize($filePath);
$filename = basename($filePath);
$extension = strtolower(pathinfo($filePath, PATHINFO_EXTENSION));

// Set content type based on extension
$contentTypes = [
    'wav' => 'audio/wav',
    'mp3' => 'audio/mpeg',
    'ogg' => 'audio/ogg',
    'gsm' => 'audio/gsm',
];
$contentType = $contentTypes[$extension] ?? 'application/octet-stream';

// Set headers and serve file
header('Content-Type: ' . $contentType);
header('Content-Length: ' . $filesize);
header('Content-Disposition: inline; filename="' . $filename . '"');

// Serve the file
readfile($filePath);
?>