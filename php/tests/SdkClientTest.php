<?php

declare(strict_types=1);

require_once __DIR__ . '/../src/Auth.php';
require_once __DIR__ . '/../src/TokenSource.php';
require_once __DIR__ . '/../src/StaticTokenSource.php';
require_once __DIR__ . '/../src/ClientCredentialsTokenSource.php';
require_once __DIR__ . '/../src/Metadata.php';
require_once __DIR__ . '/../src/Models.php';
require_once __DIR__ . '/../src/SdkConfig.php';
require_once __DIR__ . '/../src/ServiceClient.php';
require_once __DIR__ . '/../src/DomainClients.php';
require_once __DIR__ . '/../src/SdkClient.php';

use TwoFinance\Sdk\Auth;
use TwoFinance\Sdk\ClientCredentialsTokenSource;
use TwoFinance\Sdk\DomainOperationsCatalog;
use TwoFinance\Sdk\HttpRequest;
use TwoFinance\Sdk\HttpResponse;
use TwoFinance\Sdk\IdempotencyRecord;
use TwoFinance\Sdk\Metadata;
use TwoFinance\Sdk\PaginationResponse;
use TwoFinance\Sdk\RequestOptions;
use TwoFinance\Sdk\ResolvedOperation;
use TwoFinance\Sdk\SdkClient;
use TwoFinance\Sdk\SdkConfig;
use TwoFinance\Sdk\SdkErrorPayload;
use TwoFinance\Sdk\ServiceClient;
use TwoFinance\Sdk\ServiceCatalog;
use TwoFinance\Sdk\ServiceException;
use TwoFinance\Sdk\StaticTokenSource;

function assertSameValue(mixed $expected, mixed $actual, string $message): void
{
    if ($expected !== $actual) {
        fwrite(STDERR, $message . PHP_EOL);
        fwrite(STDERR, 'Expected: ' . var_export($expected, true) . PHP_EOL);
        fwrite(STDERR, 'Actual: ' . var_export($actual, true) . PHP_EOL);
        exit(1);
    }
}

function contractFixture(string $name): array
{
    $json = file_get_contents(__DIR__ . '/../../contracts/examples/' . $name);
    if ($json === false) {
        fwrite(STDERR, 'Could not read contract fixture: ' . $name . PHP_EOL);
        exit(1);
    }
    return json_decode($json, true, 512, JSON_THROW_ON_ERROR);
}

function contractOperation(array $fixture, string $domainName, string $operationName): array
{
    foreach ($fixture['domains'] as $domain) {
        if ($domain['name'] !== $domainName) {
            continue;
        }
        foreach ($domain['operations'] as $operation) {
            if ($operation['name'] === $operationName) {
                return $operation;
            }
        }
    }
    fwrite(STDERR, 'Missing contract operation: ' . $domainName . '.' . $operationName . PHP_EOL);
    exit(1);
}

$domainOperationsFixture = contractFixture('domain-operations.json');
$errorFixture = contractFixture('error.json');
$paginationFixture = contractFixture('pagination.json');
$idempotencyFixture = contractFixture('idempotency.json');

assertSameValue('2finance-sdk-client', Metadata::SDK_NAME, 'SDK name should be public');
assertSameValue('0.1.0', Metadata::SDK_VERSION, 'SDK version should be public');
assertSameValue(12, count(Metadata::SERVICE_CATALOG), 'SDK service catalog should expose canonical services');
assertSameValue('TWO_FINANCE_AUTH_URL', Metadata::serviceCatalog()->services[0]->env, 'SDK service catalog should expose env vars');

assertSameValue('sdk.domain_operations.v2', $domainOperationsFixture['schema'], 'domain operations fixture schema should match');
assertSameValue(
    '/portfolio-manager/balances/{account_id}',
    contractOperation($domainOperationsFixture, 'analytics', 'balances')['path'],
    'analytics balances contract should define path',
);
assertSameValue(
    ['account_id'],
    contractOperation($domainOperationsFixture, 'analytics', 'balances')['path_params'],
    'analytics balances contract should define path params',
);
assertSameValue(
    'planner.trading_plan.request.v1',
    contractOperation($domainOperationsFixture, 'planner', 'trading_plan')['request_schema'],
    'planner trading contract should define request schema',
);
assertSameValue('rate_limited', $errorFixture['error'], 'error fixture should define rate limit error');
assertSameValue('HTTP_429', $errorFixture['code'], 'error fixture should define HTTP 429 code');
assertSameValue('cursor-next', $paginationFixture['next_cursor'], 'pagination fixture should define next cursor');
assertSameValue('idem-001', $idempotencyFixture['idempotency_key'], 'idempotency fixture should define canonical key');

$sdkError = SdkErrorPayload::fromArray($errorFixture);
$pagination = PaginationResponse::fromArray($paginationFixture);
$idempotency = IdempotencyRecord::fromArray($idempotencyFixture);
$catalog = ServiceCatalog::fromArray(contractFixture('service-catalog.json'));
$operations = DomainOperationsCatalog::fromArray($domainOperationsFixture);
assertSameValue('HTTP_429', $sdkError->code, 'SDK error model should parse code');
assertSameValue('req_2finance_001', $sdkError->details['request_id'], 'SDK error model should parse details');
assertSameValue(25, $pagination->limit, 'pagination model should parse limit');
assertSameValue('cursor-next', $pagination->nextCursor, 'pagination model should parse next cursor');
assertSameValue('idem-001', $idempotency->idempotencyKey, 'idempotency model should parse key');
assertSameValue('auth', $catalog->services[0]->name, 'service catalog model should parse services');
assertSameValue('sdk.domain_operations.v2', $operations->schema, 'domain operations model should parse schema');
assertSameValue('auth.login.request.v1', $operations->domains[0]->operations[0]->requestSchema, 'domain operations model should parse request schema');
assertSameValue('/portfolio-manager/balances/{account_id}', $operations->operation('analytics', 'balances')->path, 'domain operations model should locate operation');
$resolvedBalances = $operations->operation('analytics', 'balances')->resolve(['account_id' => 'acct/1 ok']);
assertSameValue('GET', $resolvedBalances->method, 'resolved operation should normalize method');
assertSameValue('/portfolio-manager/balances/acct%2F1%20ok', $resolvedBalances->path, 'resolved operation should escape path params');
assertSameValue(
    $resolvedBalances->path,
    $operations->resolveOperation('analytics', 'balances', ['account_id' => 'acct/1 ok'])->path,
    'catalog should resolve operation by name',
);
$resolvedRisk = $operations->operation('analytics', 'black_scholes')->resolve([], [
    'symbol' => 'BTC/USD',
    'strike' => 100000,
    'ignored' => 'drop-me',
    'volatility' => 0.5,
]);
assertSameValue('/risk-manager/blackscholes?symbol=BTC%2FUSD&strike=100000&volatility=0.5', $resolvedRisk->path, 'resolved operation should filter query params');
$configForServiceUrl = SdkConfig::fromEnv([
    'TWO_FINANCE_ANALYTICS_URL' => 'https://analytics.example',
    'TWO_FINANCE_MATCHENGINE_WS_URL' => 'wss://matchengine.example/ws',
]);
assertSameValue('https://analytics.example', $configForServiceUrl->serviceUrl('analytics'), 'config should resolve analytics URL');
assertSameValue('wss://matchengine.example/ws', $configForServiceUrl->serviceUrl('match_engine'), 'config should resolve matchengine URL');
assertSameValue('https://analytics.example', $configForServiceUrl->serviceUrls()['analytics'], 'config should list analytics URL');
assertSameValue('wss://matchengine.example/ws', $configForServiceUrl->serviceUrls()['matchengine'], 'config should list matchengine URL');
assertSameValue(2, count($configForServiceUrl->configuredServices()), 'config should list configured services');
assertSameValue('matchengine', $configForServiceUrl->configuredServices()[1]->name, 'configured service should preserve catalog order');
assertSameValue(10, count($configForServiceUrl->missingServiceUrls()), 'config should list missing services');
assertSameValue('auth', $configForServiceUrl->missingServiceUrls()[0]->name, 'missing service should preserve catalog order');

assertSameValue('Bearer abc', Auth::bearerAuthorization('abc'), 'bearer token should be normalized');
assertSameValue('Bearer abc', Auth::bearerAuthorization('Bearer abc'), 'bearer token should not be duplicated');

$config = SdkConfig::fromEnv([
    'TWO_FINANCE_AUTH_URL' => 'https://auth.example',
    'TWO_FINANCE_ANALYTICS_URL' => 'https://analytics.example',
]);
assertSameValue('https://auth.example', $config->authUrl, 'auth URL should load from env map');
assertSameValue('2finance', $config->authRealm, 'auth realm should default');

$config->tokenSource = new StaticTokenSource('token-123');
$seen = [];
$transport = static function (HttpRequest $request) use (&$seen): HttpResponse {
    $seen = [
        'url' => $request->url,
        'authorization' => $request->headers['Authorization'] ?? '',
    ];
    return new HttpResponse(200, '{"ok":true}');
};

$client = new SdkClient($config, $transport);
$result = $client->analytics->indicators();

assertSameValue(['ok' => true], $result, 'analytics response should decode');
assertSameValue('https://analytics.example/analytics/indicators', $seen['url'], 'analytics URL should resolve');
assertSameValue('Bearer token-123', $seen['authorization'], 'bearer auth should be injected');

$authSeen = [];
$authConfig = SdkConfig::fromEnv(['TWO_FINANCE_AUTH_URL' => 'https://auth.example']);
$authClient = new SdkClient($authConfig, static function (HttpRequest $request) use (&$authSeen): HttpResponse {
    $authSeen[] = $request->method . ' ' . $request->url . ' ' . ($request->body ?? '');
    return new HttpResponse(200, '{"ok":true}');
});
$authClient->auth->jwks();
$authClient->auth->validateToken('token-1');
assertSameValue(
    'GET https://auth.example/realms/2finance/protocol/openid-connect/certs ',
    $authSeen[0],
    'auth JWKS helper should use OIDC certs endpoint',
);
assertSameValue(
    'POST https://auth.example/realms/2finance/protocol/openid-connect/token/introspect {"token":"token-1"}',
    $authSeen[1],
    'auth token validation helper should use OIDC introspection endpoint',
);

$optionSeen = [];
$optionAttempts = 0;
$optionTransport = static function (HttpRequest $request) use (&$optionSeen, &$optionAttempts): HttpResponse {
    $optionAttempts++;
    $optionSeen = [
        'url' => $request->url,
        'trace' => $request->headers['X-Trace-ID'] ?? '',
        'idempotency' => $request->headers['Idempotency-Key'] ?? '',
        'timeout' => $request->timeoutSeconds,
    ];
    if ($optionAttempts === 1) {
        return new HttpResponse(500, 'temporary');
    }
    return new HttpResponse(200, '{"ok":true}');
};
$optionService = new ServiceClient('https://analytics.example', $optionTransport);
$optionService->post(
    '/analytics/candles:upsert',
    ['symbol' => 'BTC-USDT'],
    new RequestOptions(['X-Trace-ID' => 'trace-1'], ' idem-1 ', ['symbol' => 'BTC-USDT'], 1.5, 1, 2, 25),
);
assertSameValue(2, $optionAttempts, 'request options should retry retryable responses');
assertSameValue('https://analytics.example/analytics/candles:upsert?symbol=BTC-USDT&page=2&limit=25', $optionSeen['url'], 'request options should add query params and pagination');
assertSameValue('trace-1', $optionSeen['trace'], 'request options should add custom headers');
assertSameValue('idem-1', $optionSeen['idempotency'], 'request options should add idempotency key');
assertSameValue(1.5, $optionSeen['timeout'], 'request options should add timeout seconds');

$optionService->requestOperation(new ResolvedOperation('GET', '/portfolio-manager/balances/acct%2Fresolved'));
assertSameValue('https://analytics.example/portfolio-manager/balances/acct%2Fresolved', $optionSeen['url'], 'resolved operation should use resolved path');
$optionService->requestCatalogOperation($operations, 'analytics', 'balances', ['account_id' => 'acct/1 ok']);
assertSameValue('https://analytics.example/portfolio-manager/balances/acct%2F1%20ok', $optionSeen['url'], 'catalog operation should use resolved path');

$failingService = new ServiceClient(
    'https://analytics.example',
    static fn(HttpRequest $request): HttpResponse => new HttpResponse(429, 'rate limited'),
);
try {
    $failingService->get('/analytics/indicators');
    fwrite(STDERR, 'ServiceException should be thrown for non-2xx responses' . PHP_EOL);
    exit(1);
} catch (ServiceException $exception) {
    assertSameValue('GET', $exception->method, 'service exception should keep method');
    assertSameValue('https://analytics.example/analytics/indicators', $exception->url, 'service exception should keep URL');
    assertSameValue(429, $exception->statusCode, 'service exception should keep status code');
    assertSameValue('rate limited', $exception->body, 'service exception should keep body');
}

$tokenCalls = 0;
$tokenTransport = static function (HttpRequest $request) use (&$tokenCalls): HttpResponse {
    $tokenCalls++;
    assertSameValue('https://auth.example/token', $request->url, 'token URL should be used');
    assertSameValue('POST', $request->method, 'token request should use POST');
    if (!str_contains($request->body ?? '', 'grant_type=client_credentials')) {
        fwrite(STDERR, 'Token request body should include client credentials grant' . PHP_EOL);
        exit(1);
    }
    return new HttpResponse(200, '{"access_token":"cc-token","expires_in":3600}');
};
$tokenSource = new ClientCredentialsTokenSource(
    'https://auth.example/token',
    'client-id',
    'client-secret',
    ['analytics:read', 'mcp:invoke'],
    $tokenTransport,
);
assertSameValue('cc-token', $tokenSource->token(), 'client credentials token should parse');
assertSameValue('cc-token', $tokenSource->token(), 'client credentials token should cache');
assertSameValue(1, $tokenCalls, 'client credentials token should only fetch once while cached');

$matchConfig = SdkConfig::fromEnv([
    'TWO_FINANCE_MATCHENGINE_WS_URL' => 'wss://matchengine.example/ws',
]);
$matchClient = new SdkClient($matchConfig, $transport);
$command = $matchClient->matchEngine->orderCommand([
    'client_order_id' => 'co-1',
    'idempotency_key' => 'idem-1',
    'symbol' => 'BTC-USDT',
    'side' => 'buy',
    'type' => 'limit',
    'quantity' => '0.01',
]);
assertSameValue('wss://matchengine.example/ws', $matchClient->matchEngine->webSocketUrl, 'matchengine URL should load');
assertSameValue('matchengine.order_command.v2', $command['schema'], 'matchengine schema should default');
assertSameValue('BTC-USDT', $command['symbol'], 'matchengine symbol should pass through');
$subscription = $matchClient->matchEngine->marketDataSubscribe([
    'symbols' => ['BTC-USDT'],
    'channels' => ['book'],
]);
assertSameValue('matchengine.market_data_subscribe.v2', $subscription['schema'], 'matchengine market data schema should default');
assertSameValue(['BTC-USDT'], $subscription['symbols'], 'matchengine market data symbols should pass through');
$matchMessages = [];
$sender = static function (array $message) use (&$matchMessages): array {
    $matchMessages[] = $message;
    return ['ok' => true];
};
assertSameValue(['ok' => true], $matchClient->matchEngine->sendOrder($sender, $command), 'matchengine sendOrder should return sender result');
assertSameValue(['ok' => true], $matchClient->matchEngine->subscribeMarketData($sender, $subscription), 'matchengine subscribeMarketData should return sender result');
assertSameValue('matchengine.order_command.v2', $matchMessages[0]['schema'], 'matchengine sendOrder should send order schema');
assertSameValue('matchengine.market_data_subscribe.v2', $matchMessages[1]['schema'], 'matchengine subscribeMarketData should send market data schema');

$domainSeen = [];
$domainTransport = static function (HttpRequest $request) use (&$domainSeen): HttpResponse {
    $domainSeen[] = $request->method . ' ' . $request->url;
    return new HttpResponse(200, '{"ok":true}');
};
$domainConfig = SdkConfig::fromEnv([
    'TWO_FINANCE_ANALYTICS_URL' => 'https://analytics.example',
    'TWO_FINANCE_NETWORK_URL' => 'https://network.example',
    'TWO_FINANCE_TRADING_CONTROL_URL' => 'https://trading.example',
    'TWO_FINANCE_KEYSTORE_URL' => 'https://keys.example',
    'TWO_FINANCE_HUMMINGBOT_URL' => 'https://hbot.example',
    'TWO_FINANCE_MCP_URL' => 'https://mcp.example',
    'TWO_FINANCE_ORCHESTRATOR_URL' => 'https://orchestrator.example',
    'TWO_FINANCE_WISE_URL' => 'https://wise.example',
    'TWO_FINANCE_AIRWALLEX_URL' => 'https://airwallex.example',
]);
$domainClient = new SdkClient($domainConfig, $domainTransport);
$domainClient->analytics->upsertCandles(['symbol' => 'BTC-USDT']);
$domainClient->analytics->rankings();
$domainClient->analytics->balances('acct/1 ok');
$domainClient->analytics->blackScholes('symbol=BTC');
$domainClient->analytics->staking();
$domainClient->network->marketCandles('BTC/USDT spot', 'limit=10');
$domainClient->network->bonds();
$domainClient->network->createBond(['symbol' => 'BOND1']);
$domainClient->network->loans();
$domainClient->network->createLoan(['loan' => 'ln1']);
$domainClient->network->swaps();
$domainClient->network->createSwap(['pair' => 'BTC-USDT']);
$domainClient->network->stakingProducts();
$domainClient->network->createStakingProduct(['asset' => 'TWO']);
$domainClient->network->syntheticAssets();
$domainClient->network->createSyntheticAsset(['asset' => 'sBTC']);
$domainClient->network->liquidityPools();
$domainClient->network->createLiquidityPool(['pool' => 'BTC-USDT']);
$domainClient->tradingControl->startRobot('robot/1 ok');
$domainClient->tradingControl->pauseRobot('robot/1 ok');
$domainClient->tradingControl->riskPolicy('robot/1 ok');
$domainClient->tradingControl->riskView('robot/1 ok');
$domainClient->tradingControl->strategies();
$domainClient->tradingControl->createStrategy(['name' => 'mean-reversion']);
$domainClient->tradingControl->directives();
$domainClient->tradingControl->createDirective(['action' => 'rebalance']);
$domainClient->tradingControl->audit();
$domainClient->tradingControl->activity();
$domainClient->tradingControl->mcpTools();
$domainClient->keystore->health();
$domainClient->keystore->readiness();
$domainClient->keystore->startSigning(['key' => 'k1']);
$domainClient->keystore->keys('pub/1 ok');
$domainClient->keystore->signatures('pub/1 ok');
$domainClient->keystore->metrics();
$domainClient->hummingbot->balances();
$domainClient->hummingbot->connectorConfig(['connector' => '2finance']);
$domainClient->mcp->listTools();
$domainClient->mcp->listPrompts();
$domainClient->mcp->readResource('resource://portfolio');
$domainClient->orchestrator->tools();
$domainClient->orchestrator->resources();
$domainClient->orchestrator->approvals();
$domainClient->orchestrator->deleteSession('session/1 ok');
$domainClient->wise->profiles();
$domainClient->wise->profile('profile/1 ok');
$domainClient->wise->createQuote('profile/1 ok', ['source' => 'USD']);
$domainClient->wise->createTransfer(['target' => 'BRL']);
$domainClient->airwallex->accounts();
$domainClient->airwallex->payments();
$domainClient->airwallex->createPayment(['amount' => 10]);
$domainClient->airwallex->beneficiaries();
$domainClient->airwallex->createBeneficiary(['name' => 'beneficiary']);

foreach ([
    'POST https://analytics.example/analytics/candles:upsert',
    'GET https://analytics.example/portfolio-manager/rankings',
    'GET https://analytics.example/portfolio-manager/balances/acct%2F1%20ok',
    'GET https://analytics.example/risk-manager/blackscholes?symbol=BTC',
    'GET https://analytics.example/staking',
    'GET https://network.example/v2/2finance-network/markets/BTC%2FUSDT%20spot/candles?limit=10',
    'GET https://network.example/v2/2finance-network/products/bonds',
    'POST https://network.example/v2/2finance-network/products/bonds',
    'GET https://network.example/v2/2finance-network/products/loans',
    'POST https://network.example/v2/2finance-network/products/loans',
    'GET https://network.example/v2/2finance-network/products/swaps',
    'POST https://network.example/v2/2finance-network/products/swaps',
    'GET https://network.example/v2/2finance-network/products/staking',
    'POST https://network.example/v2/2finance-network/products/staking',
    'GET https://network.example/v2/2finance-network/products/synthetic-assets',
    'POST https://network.example/v2/2finance-network/products/synthetic-assets',
    'GET https://network.example/v2/2finance-network/products/liquidity-pools',
    'POST https://network.example/v2/2finance-network/products/liquidity-pools',
    'POST https://trading.example/robots/robot%2F1%20ok:start',
    'POST https://trading.example/robots/robot%2F1%20ok:pause',
    'GET https://trading.example/robots/robot%2F1%20ok/risk-policy',
    'GET https://trading.example/risk-view/robot%2F1%20ok',
    'GET https://trading.example/strategies',
    'POST https://trading.example/strategies',
    'GET https://trading.example/directives',
    'POST https://trading.example/directives',
    'GET https://trading.example/audit',
    'GET https://trading.example/activity',
    'GET https://trading.example/mcp/tools',
    'GET https://keys.example/healthz',
    'GET https://keys.example/readyz',
    'POST https://keys.example/keystore/signing/start',
    'GET https://keys.example/keystore/keys/pub%2F1%20ok',
    'GET https://keys.example/keystore/signatures/pub%2F1%20ok',
    'GET https://keys.example/keystore/tss/metrics',
    'GET https://hbot.example/api/v1/balances',
    'POST https://hbot.example/api/v1/connectors/2finance/config',
    'POST https://mcp.example/mcp',
    'GET https://orchestrator.example/v1/mcphost/tools',
    'GET https://orchestrator.example/v1/mcphost/resources',
    'GET https://orchestrator.example/v1/mcphost/approvals',
    'DELETE https://orchestrator.example/v1/mcphost/sessions/session%2F1%20ok',
    'GET https://wise.example/v1/profiles',
    'GET https://wise.example/v1/profiles/profile%2F1%20ok',
    'POST https://wise.example/v3/profiles/profile%2F1%20ok/quotes',
    'POST https://wise.example/v1/transfers',
    'GET https://airwallex.example/api/v1/accounts',
    'GET https://airwallex.example/api/v1/payments',
    'POST https://airwallex.example/api/v1/payments',
    'GET https://airwallex.example/api/v1/beneficiaries',
    'POST https://airwallex.example/api/v1/beneficiaries',
] as $expectedRequest) {
    if (!in_array($expectedRequest, $domainSeen, true)) {
        fwrite(STDERR, 'Missing expected domain request: ' . $expectedRequest . PHP_EOL);
        exit(1);
    }
}

$plannerSeen = [];
$plannerTransport = static function (HttpRequest $request) use (&$plannerSeen): HttpResponse {
    $plannerSeen[] = [
        'method' => $request->method,
        'url' => $request->url,
        'body' => json_decode($request->body ?? 'null', true, 512, JSON_THROW_ON_ERROR),
    ];
    if ($request->url === 'https://trading.example/robots') {
        return new HttpResponse(200, '{"robots":[{"id":"robot-1","status":"running"}]}');
    }
    if ($request->url === 'https://analytics.example/analytics/indicators') {
        return new HttpResponse(200, '{"indicators":[{"symbol":"BTC-USDT","rsi":55}]}');
    }
    if ($request->url === 'https://orchestrator.example/v1/mcphost/messages') {
        return new HttpResponse(200, '{"ok":true}');
    }
    if ($request->url === 'https://mcp.example/mcp') {
        return new HttpResponse(200, '{"jsonrpc":"2.0","id":1,"result":{"plan_id":"plan-1"}}');
    }
    return new HttpResponse(404, '{"error":"unexpected request"}');
};
$plannerConfig = SdkConfig::fromEnv([
    'TWO_FINANCE_ANALYTICS_URL' => 'https://analytics.example',
    'TWO_FINANCE_TRADING_CONTROL_URL' => 'https://trading.example',
    'TWO_FINANCE_MCP_URL' => 'https://mcp.example',
]);
$plannerClient = new SdkClient($plannerConfig, $plannerTransport);
$operationalResult = $plannerClient->planner->operationalPlan([
    'session_id' => 'session-1',
    'message' => 'operate',
]);
assertSameValue(['ok' => true], $operationalResult, 'planner operational plan should use orchestrator');
$plannerResult = $plannerClient->planner->tradingPlan([
    'goal' => 'protect portfolio',
    'use_trading' => true,
    'use_analytics' => true,
    'context' => ['risk_profile' => 'balanced'],
]);
assertSameValue('plan-1', $plannerResult['result']['plan_id'], 'planner trading plan should return MCP result');
assertSameValue('POST', $plannerSeen[0]['method'], 'planner operational plan should post orchestrator message');
assertSameValue('https://orchestrator.example/v1/mcphost/messages', $plannerSeen[0]['url'], 'planner operational plan should use orchestrator messages endpoint');
assertSameValue('GET', $plannerSeen[1]['method'], 'planner should read trading robots first');
assertSameValue('https://trading.example/robots', $plannerSeen[1]['url'], 'planner should use trading control robots endpoint');
assertSameValue('GET', $plannerSeen[2]['method'], 'planner should read analytics indicators');
assertSameValue('https://analytics.example/analytics/indicators', $plannerSeen[2]['url'], 'planner should use analytics indicators endpoint');
assertSameValue('POST', $plannerSeen[3]['method'], 'planner should call MCP after enrichment');
assertSameValue('https://mcp.example/mcp', $plannerSeen[3]['url'], 'planner should call MCP endpoint');
assertSameValue('tools/call', $plannerSeen[3]['body']['method'], 'planner should call MCP tools/call');
assertSameValue(
    'finance_assistant.conversation.plan',
    $plannerSeen[3]['body']['params']['name'],
    'planner should call conversation plan tool',
);
assertSameValue(
    'balanced',
    $plannerSeen[3]['body']['params']['arguments']['context']['risk_profile'],
    'planner should preserve caller context',
);
assertSameValue(
    'robot-1',
    $plannerSeen[3]['body']['params']['arguments']['context']['trading_robots']['robots'][0]['id'],
    'planner should enrich with trading robots',
);
assertSameValue(
    'BTC-USDT',
    $plannerSeen[3]['body']['params']['arguments']['context']['analytics_indicators']['indicators'][0]['symbol'],
    'planner should enrich with analytics indicators',
);

echo "PHP SDK smoke test passed\n";
