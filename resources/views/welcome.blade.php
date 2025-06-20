<!DOCTYPE html>
<html lang="{{ str_replace('_', '-', app()->getLocale()) }}">
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <link rel="icon" href="{{ asset('assets/images/favicon.svg') }}">

        <title>Mybili</title>
        @vite(['resources/css/app.css', 'resources/js/main.ts'])
        @if(config('app.website_id'))
            <script defer src="https://cloud.umami.is/script.js" data-website-id="{{ config('app.website_id') }}"></script>
        @endif
    </head>
    <body>
        <div id="app">
        </div>
    </body>
</html>
