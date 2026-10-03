<?php

declare(strict_types=1);

use Swoole\Http\Request;
use Swoole\Http\Response;
use Swoole\Http\Server;

$port = (int) (getenv('APP_PORT') ?: 8080);

$server = new Server('0.0.0.0', $port);
$server->set([
    'worker_num' => 1,
    'reload_async' => true,
]);

$server->on('request', function (Request $request, Response $response): void {
    $response->header('Content-Type', 'text/plain; charset=utf-8');
    $response->end("hello from swoole-api\n");
});

echo "swoole-api listening on :{$port}\n";
$server->start();
