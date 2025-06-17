<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Database Connection Error</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            margin: 0;
            padding: 0;
            background-color: #f8f9fa;
            color: #333;
            line-height: 1.6;
        }
        .container {
            max-width: 800px;
            margin: 50px auto;
            padding: 30px;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .error-icon {
            text-align: center;
            font-size: 48px;
            color: #dc3545;
            margin-bottom: 20px;
        }
        h1 {
            color: #dc3545;
            text-align: center;
            margin-bottom: 30px;
            font-size: 24px;
        }
        .info-box {
            background: #f8f9fa;
            border-left: 4px solid #007bff;
            padding: 15px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .database-info {
            background: #e9ecef;
            padding: 15px;
            border-radius: 4px;
            margin: 20px 0;
        }
        .database-info p {
            margin: 5px 0;
        }
        .solution {
            background: #d4edda;
            border-left: 4px solid #28a745;
            padding: 15px;
            margin: 20px 0;
            border-radius: 4px;
        }
        code {
            background: #f8f9fa;
            padding: 2px 6px;
            border-radius: 4px;
            font-family: monospace;
            color: #e83e8c;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="error-icon">⚠️</div>
        <h1>Database Connection Error</h1>
        
        <div class="info-box">
            <p>The system detected a database connection failure. Please check your database configuration.</p>
        </div>

        <div class="database-info">
            <h2>Current Database Configuration:</h2>
            <p><strong>Database Type:</strong> {{ $database['driver'] }}</p>
            <p><strong>Database Name:</strong> {{ $database['database'] }}</p>
            @if($database['url'])
                <p><strong>Database URL:</strong> {{ $database['url'] }}</p>
            @endif
        </div>

        <div class="solution">
            <h2>Solution:</h2>
            @if(strtolower($database['driver']) === 'sqlite')
                <p>Since you are using SQLite database:</p>
                <ol>
                    <li>Ensure the database file exists and has proper permissions</li>
                    <li>Run the following command to create the database and execute migrations:</li>
                    <code>php artisan migrate --force</code>
                </ol>
            @else
                <p>For your database type ({{ $database['driver'] }}):</p>
                <ol>
                    <li>Check if the database service is running</li>
                    <li>Verify database connection information is correct</li>
                    <li>Ensure the database user has appropriate permissions</li>
                </ol>
            @endif
        </div>
    </div>
</body>
</html>
