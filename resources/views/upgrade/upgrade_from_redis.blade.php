<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Redis Database Migration Required</title>
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
        .info-icon {
            text-align: center;
            font-size: 48px;
            color: #17a2b8;
            margin-bottom: 20px;
        }
        h1 {
            color: #17a2b8;
            text-align: center;
            margin-bottom: 30px;
            font-size: 24px;
        }
        .info-box {
            background: #f8f9fa;
            border-left: 4px solid #17a2b8;
            padding: 15px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .migration-steps {
            background: #e9ecef;
            padding: 15px;
            border-radius: 4px;
            margin: 20px 0;
        }
        .migration-steps p {
            margin: 5px 0;
        }
        .command-box {
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
        .note {
            background: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 15px;
            margin: 20px 0;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="info-icon">🔄</div>
        <h1>Redis Database Migration Required</h1>
        
        <div class="info-box">
            <p>The system has detected that you were using an old Redis database. A new database connection has been successfully established. To complete the setup, you need to migrate your data from Redis to the new database.</p>
        </div>

        <div class="migration-steps">
            <h2>Migration Steps:</h2>
            <ol>
                <li>Open the terminal in your application directory or Docker container</li>
                <li>Execute the following commands in sequence:</li>
            </ol>
        </div>

        <div class="command-box">
            <h3>Step 1: Migrate Redis Data</h3>
            <code>php artisan app:upgrade-redis-to-sqlite --all</code>
            
            <h3>Step 2: Scan Video Images</h3>
            <code>php artisan app:scan-video-image</code>
            
            <h3>Step 3: Scan Video Files</h3>
            <code>php artisan app:scan-video-file</code>
        </div>

        <div class="note">
            <h2>Important Note:</h2>
            <p>This process will migrate all your data from Redis to the new database and scan your video resources. Please ensure you have backed up your data before proceeding with the migration.</p>
        </div>
    </div>
</body>
</html>
