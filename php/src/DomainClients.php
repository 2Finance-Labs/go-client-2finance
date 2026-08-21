<?php

declare(strict_types=1);

namespace TwoFinance\Sdk;

final class AuthClient
{
    public function __construct(
        private readonly ServiceClient $service,
        private readonly string $realm = '2finance',
        private readonly string $clientId = '2finance-network',
        private readonly string $phoneClientId = '2finance-network-phone',
    ) {
    }

    /**
     * @return mixed
     */
    public function login(array $input): mixed
    {
        return $this->service->post($this->authPath($this->clientId, '/login'), $input);
    }

    /**
     * @return mixed
     */
    public function refreshToken(string $refreshToken): mixed
    {
        return $this->service->post($this->authPath($this->clientId, '/refresh'), ['refresh_token' => $refreshToken]);
    }

    /**
     * @return mixed
     */
    public function phoneLogin(string $phoneNumber, string $code): mixed
    {
        return $this->service->post(
            $this->authPath($this->phoneClientId, '/phone/sms/login'),
            ['phone_number' => $phoneNumber, 'code' => $code],
        );
    }

    public function jwks(): mixed
    {
        return $this->service->get($this->oidcPath('/protocol/openid-connect/certs'));
    }

    public function validateToken(string $token): mixed
    {
        return $this->service->post($this->oidcPath('/protocol/openid-connect/token/introspect'), ['token' => $token]);
    }

    private function authPath(string $clientId, string $endpoint): string
    {
        return sprintf('/v1/2finance-authenticator/%s/%s/%s', $this->realm, $clientId, ltrim($endpoint, '/'));
    }

    private function oidcPath(string $endpoint): string
    {
        return sprintf('/realms/%s/%s', $this->realm, ltrim($endpoint, '/'));
    }
}

final class AnalyticsClient
{
    public function __construct(private readonly ServiceClient $service)
    {
    }

    public function indicators(): mixed
    {
        return $this->service->get('/analytics/indicators');
    }

    public function get(string $path, ?RequestOptions $options = null): mixed
    {
        return $this->service->get($path, $options);
    }

    public function post(string $path, mixed $body = null, ?RequestOptions $options = null): mixed
    {
        return $this->service->post($path, $body, $options);
    }

    public function calculateTechnicalAnalysis(array $request): mixed
    {
        return $this->service->post('/analytics/technical-analysis:calculate', $request);
    }

    public function optimizePortfolio(array $request): mixed
    {
        return $this->service->post('/portfolio-manager/optimizer', $request);
    }

    public function upsertCandles(array $request): mixed
    {
        return $this->service->post('/analytics/candles:upsert', $request);
    }

    public function rankings(): mixed
    {
        return $this->service->get('/portfolio-manager/rankings');
    }

    public function balances(string $accountId): mixed
    {
        return $this->service->get('/portfolio-manager/balances/' . rawurlencode($accountId));
    }

    public function blackScholes(string $query = ''): mixed
    {
        return $this->service->get('/risk-manager/blackscholes' . ($query === '' ? '' : '?' . $query));
    }

    public function staking(): mixed
    {
        return $this->service->get('/staking');
    }
}

final class MCPClient
{
    private int $nextId = 1;

    public function __construct(private readonly ServiceClient $service)
    {
    }

    public function call(string $method, mixed $params = null): mixed
    {
        return $this->service->post('/mcp', [
            'jsonrpc' => '2.0',
            'id' => $this->nextId++,
            'method' => $method,
            'params' => $params,
        ]);
    }

    public function listTools(): mixed
    {
        return $this->call('tools/list');
    }

    public function listPrompts(): mixed
    {
        return $this->call('prompts/list');
    }

    public function listResources(): mixed
    {
        return $this->call('resources/list');
    }

    public function readResource(string $uri): mixed
    {
        return $this->call('resources/read', ['uri' => $uri]);
    }

    public function callTool(string $name, array $arguments = []): mixed
    {
        return $this->call('tools/call', ['name' => $name, 'arguments' => $arguments]);
    }

    public function getPrompt(string $name, array $arguments = []): mixed
    {
        return $this->call('prompts/get', ['name' => $name, 'arguments' => $arguments]);
    }

    public function conversationPlan(array $arguments): mixed
    {
        return $this->callTool('finance_assistant.conversation.plan', $arguments);
    }
}

final class OrchestratorClient
{
    public function __construct(private readonly ServiceClient $service)
    {
    }

    public function catalog(): mixed
    {
        return $this->service->get('/v1/mcphost/catalog/packages');
    }

    public function createSession(array $request): mixed
    {
        return $this->service->post('/v1/mcphost/sessions', $request);
    }

    public function sendMessage(array $request): mixed
    {
        return $this->service->post('/v1/mcphost/messages', $request);
    }

    public function tools(): mixed
    {
        return $this->service->get('/v1/mcphost/tools');
    }

    public function prompts(): mixed
    {
        return $this->service->get('/v1/mcphost/prompts');
    }

    public function resources(): mixed
    {
        return $this->service->get('/v1/mcphost/resources');
    }

    public function providers(): mixed
    {
        return $this->service->get('/v1/mcphost/providers');
    }

    public function approvals(): mixed
    {
        return $this->service->get('/v1/mcphost/approvals');
    }

    public function observability(): mixed
    {
        return $this->service->get('/v1/mcphost/observability');
    }

    public function deleteSession(string $id): mixed
    {
        return $this->service->delete('/v1/mcphost/sessions/' . rawurlencode($id));
    }
}

final class NetworkClient
{
    public function __construct(private readonly ServiceClient $service)
    {
    }

    public function virtualMachine(): mixed
    {
        return $this->service->post('/v2/2finance-network/query', [
            'method' => 'get_blocks',
            'params' => ['page' => 1, 'limit' => 1],
        ]);
    }

    public function marketCandles(string $market, string $query = ''): mixed
    {
        return $this->service->get(
            '/v2/2finance-network/markets/' . rawurlencode($market) . '/candles' . ($query === '' ? '' : '?' . $query),
        );
    }

    public function products(string $productType): mixed
    {
        return $this->service->get('/v2/2finance-network/products/' . rawurlencode($productType));
    }

    public function createProduct(string $productType, array $request): mixed
    {
        return $this->service->post('/v2/2finance-network/products/' . rawurlencode($productType), $request);
    }

    public function bonds(): mixed
    {
        return $this->products('bonds');
    }

    public function createBond(array $request): mixed
    {
        return $this->createProduct('bonds', $request);
    }

    public function loans(): mixed
    {
        return $this->products('loans');
    }

    public function createLoan(array $request): mixed
    {
        return $this->createProduct('loans', $request);
    }

    public function swaps(): mixed
    {
        return $this->products('swaps');
    }

    public function createSwap(array $request): mixed
    {
        return $this->createProduct('swaps', $request);
    }

    public function stakingProducts(): mixed
    {
        return $this->products('staking');
    }

    public function createStakingProduct(array $request): mixed
    {
        return $this->createProduct('staking', $request);
    }

    public function syntheticAssets(): mixed
    {
        return $this->products('synthetic-assets');
    }

    public function createSyntheticAsset(array $request): mixed
    {
        return $this->createProduct('synthetic-assets', $request);
    }

    public function liquidityPools(): mixed
    {
        return $this->products('liquidity-pools');
    }

    public function createLiquidityPool(array $request): mixed
    {
        return $this->createProduct('liquidity-pools', $request);
    }
}

final class TradingControlClient
{
    public function __construct(private readonly ServiceClient $service)
    {
    }

    public function robots(): mixed
    {
        return $this->service->get('/robots');
    }

    public function createRobot(array $request): mixed
    {
        return $this->service->post('/robots', $request);
    }

    public function robot(string $id): mixed
    {
        return $this->service->get('/robots/' . rawurlencode($id));
    }

    public function startRobot(string $id): mixed
    {
        return $this->service->post('/robots/' . rawurlencode($id) . ':start');
    }

    public function pauseRobot(string $id): mixed
    {
        return $this->service->post('/robots/' . rawurlencode($id) . ':pause');
    }

    public function resumeRobot(string $id): mixed
    {
        return $this->service->post('/robots/' . rawurlencode($id) . ':resume');
    }

    public function stopRobot(string $id): mixed
    {
        return $this->service->post('/robots/' . rawurlencode($id) . ':stop');
    }

    public function riskPolicy(string $id): mixed
    {
        return $this->service->get('/robots/' . rawurlencode($id) . '/risk-policy');
    }

    public function setRiskPolicy(string $id, array $request): mixed
    {
        return $this->service->put('/robots/' . rawurlencode($id) . '/risk-policy', $request);
    }

    public function riskView(string $id): mixed
    {
        return $this->service->get('/risk-view/' . rawurlencode($id));
    }

    public function strategies(): mixed
    {
        return $this->service->get('/strategies');
    }

    public function createStrategy(array $request): mixed
    {
        return $this->service->post('/strategies', $request);
    }

    public function directives(): mixed
    {
        return $this->service->get('/directives');
    }

    public function createDirective(array $request): mixed
    {
        return $this->service->post('/directives', $request);
    }

    public function audit(): mixed
    {
        return $this->service->get('/audit');
    }

    public function activity(): mixed
    {
        return $this->service->get('/activity');
    }

    public function mcpTools(): mixed
    {
        return $this->service->get('/mcp/tools');
    }
}

final class KeyStoreClient
{
    public function __construct(private readonly ServiceClient $service)
    {
    }

    public function health(): mixed
    {
        return $this->service->get('/healthz');
    }

    public function readiness(): mixed
    {
        return $this->service->get('/readyz');
    }

    public function startKeygen(array $request): mixed
    {
        return $this->service->post('/keystore/keygen/start', $request);
    }

    public function keygenSignature(array $request): mixed
    {
        return $this->service->post('/keystore/keygen/signature', $request);
    }

    public function startSigning(array $request): mixed
    {
        return $this->service->post('/keystore/signing/start', $request);
    }

    public function signingSignature(array $request): mixed
    {
        return $this->service->post('/keystore/signing/signature', $request);
    }

    public function startResharing(array $request): mixed
    {
        return $this->service->post('/keystore/resharing/start', $request);
    }

    public function keys(string $userPublicKey): mixed
    {
        return $this->service->get('/keystore/keys/' . rawurlencode($userPublicKey));
    }

    public function signatures(string $userPublicKey): mixed
    {
        return $this->service->get('/keystore/signatures/' . rawurlencode($userPublicKey));
    }

    public function metrics(): mixed
    {
        return $this->service->get('/keystore/tss/metrics');
    }
}

final class HummingbotClient
{
    public function __construct(private readonly ServiceClient $service)
    {
    }

    public function assets(): mixed
    {
        return $this->service->get('/api/v1/assets');
    }

    public function symbols(): mixed
    {
        return $this->service->get('/api/v1/symbols');
    }

    public function balances(): mixed
    {
        return $this->service->get('/api/v1/balances');
    }

    public function connectorConfig(array $request): mixed
    {
        return $this->service->post('/api/v1/connectors/2finance/config', $request);
    }
}

final class ProviderClient
{
    public function __construct(private readonly ServiceClient $service)
    {
    }

    public function get(string $path): mixed
    {
        return $this->service->get($path);
    }

    public function post(string $path, mixed $body = null): mixed
    {
        return $this->service->post($path, $body);
    }

    public function put(string $path, mixed $body = null): mixed
    {
        return $this->service->put($path, $body);
    }

    public function delete(string $path): mixed
    {
        return $this->service->delete($path);
    }
}

final class WiseClient
{
    public function __construct(private readonly ProviderClient $provider)
    {
    }

    public function get(string $path): mixed
    {
        return $this->provider->get($path);
    }

    public function post(string $path, mixed $body = null): mixed
    {
        return $this->provider->post($path, $body);
    }

    public function profiles(): mixed
    {
        return $this->get('/v1/profiles');
    }

    public function profile(string $profileId): mixed
    {
        return $this->get('/v1/profiles/' . rawurlencode($profileId));
    }

    public function createQuote(string $profileId, array $request): mixed
    {
        return $this->post('/v3/profiles/' . rawurlencode($profileId) . '/quotes', $request);
    }

    public function createTransfer(array $request): mixed
    {
        return $this->post('/v1/transfers', $request);
    }
}

final class AirwallexClient
{
    public function __construct(private readonly ProviderClient $provider)
    {
    }

    public function get(string $path): mixed
    {
        return $this->provider->get($path);
    }

    public function post(string $path, mixed $body = null): mixed
    {
        return $this->provider->post($path, $body);
    }

    public function accounts(): mixed
    {
        return $this->get('/api/v1/accounts');
    }

    public function payments(): mixed
    {
        return $this->get('/api/v1/payments');
    }

    public function createPayment(array $request): mixed
    {
        return $this->post('/api/v1/payments', $request);
    }

    public function beneficiaries(): mixed
    {
        return $this->get('/api/v1/beneficiaries');
    }

    public function createBeneficiary(array $request): mixed
    {
        return $this->post('/api/v1/beneficiaries', $request);
    }
}

final class MatchEngineClient
{
    public function __construct(public readonly string $webSocketUrl)
    {
    }

    /**
     * @param array<string,mixed> $command
     * @return array<string,mixed>
     */
    public function orderCommand(array $command): array
    {
        return array_merge([
            'schema' => 'matchengine.order_command.v2',
            'message_type' => 'ORDER',
            'operation' => 'ADD',
        ], $command);
    }

    /**
     * @param array<string,mixed> $request
     * @return array<string,mixed>
     */
    public function marketDataSubscribe(array $request): array
    {
        return ['schema' => 'matchengine.market_data_subscribe.v2'] + $request;
    }

    public function sendOrder(callable $sender, array $command): mixed
    {
        return $sender($this->orderCommand($command));
    }

    public function subscribeMarketData(callable $sender, array $request): mixed
    {
        return $sender($this->marketDataSubscribe($request));
    }
}

final class PlannerClient
{
    public function __construct(
        private readonly MCPClient $mcp,
        private readonly OrchestratorClient $orchestrator,
        private readonly AnalyticsClient $analytics,
        private readonly TradingControlClient $tradingControl,
    ) {
    }

    public function conversationPlan(array $request): mixed
    {
        return $this->mcp->conversationPlan($request);
    }

    public function orchestratedPlan(array $request): mixed
    {
        return $this->orchestrator->sendMessage($request);
    }

    public function operationalPlan(array $request): mixed
    {
        return $this->orchestratedPlan($request);
    }

    public function tradingPlan(array $request): mixed
    {
        $context = is_array($request['context'] ?? null) ? $request['context'] : [];
        if (($request['use_trading'] ?? false) === true) {
            try {
                $context['trading_robots'] = $this->tradingControl->robots();
            } catch (\Throwable) {
                // Best-effort enrichment keeps planning usable when trading is unavailable.
            }
        }
        if (($request['use_analytics'] ?? false) === true) {
            try {
                $context['analytics_indicators'] = $this->analytics->indicators();
            } catch (\Throwable) {
                // Best-effort enrichment keeps planning usable when analytics is unavailable.
            }
        }
        $request['context'] = $context;
        return $this->conversationPlan($request);
    }
}
