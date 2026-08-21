import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:test/test.dart';
import 'package:two_finance_sdk_client/two_finance_sdk_client.dart';

void main() {
  test('exposes SDK metadata', () {
    expect(sdkName, '2finance-sdk-client');
    expect(sdkVersion, '0.1.0');
    expect(defaultServiceCatalog.services, hasLength(12));
    expect(defaultServiceCatalog.services.first.name, 'auth');
    expect(defaultServiceCatalog.services.first.env, 'TWO_FINANCE_AUTH_URL');
  });

  test('shared contract fixtures describe public SDK operations', () {
    final domains = contractFixture('domain-operations.json');
    final error = contractFixture('error.json');
    final pagination = contractFixture('pagination.json');
    final idempotency = contractFixture('idempotency.json');

    expect(domains['schema'], 'sdk.domain_operations.v2');
    expect(
      contractOperation(domains, 'analytics', 'balances')['path'],
      '/portfolio-manager/balances/{account_id}',
    );
    expect(contractOperation(domains, 'analytics', 'balances')['path_params'], [
      'account_id',
    ]);
    expect(
      contractOperation(domains, 'planner', 'trading_plan')['request_schema'],
      'planner.trading_plan.request.v1',
    );
    expect(error['error'], 'rate_limited');
    expect(error['code'], 'HTTP_429');
    expect(pagination['next_cursor'], 'cursor-next');
    expect(idempotency['idempotency_key'], 'idem-001');
  });

  test('shared SDK models parse contract fixtures', () {
    final error = SdkError.fromJson(contractFixture('error.json'));
    final pagination = PaginationResponse.fromJson(
      contractFixture('pagination.json'),
    );
    final idempotency = IdempotencyRecord.fromJson(
      contractFixture('idempotency.json'),
    );
    final catalog = ServiceCatalog.fromJson(
      contractFixture('service-catalog.json'),
    );
    final operations = DomainOperationsCatalog.fromJson(
      contractFixture('domain-operations.json'),
    );

    expect(error.code, 'HTTP_429');
    expect(error.details['request_id'], 'req_2finance_001');
    expect(pagination.limit, 25);
    expect(pagination.nextCursor, 'cursor-next');
    expect(idempotency.idempotencyKey, 'idem-001');
    expect(catalog.services.first.name, 'auth');
    expect(operations.schema, 'sdk.domain_operations.v2');
    expect(
      operations.domains.first.operations.first.requestSchema,
      'auth.login.request.v1',
    );
    expect(
      operations.operation('analytics', 'balances')?.path,
      '/portfolio-manager/balances/{account_id}',
    );
    final resolvedBalances = operations
        .operation('analytics', 'balances')!
        .resolve(pathParams: {'account_id': 'acct/1 ok'});
    expect(resolvedBalances.method, 'GET');
    expect(resolvedBalances.path, '/portfolio-manager/balances/acct%2F1%20ok');
    expect(
      operations
          .resolveOperation(
            'analytics',
            'balances',
            pathParams: {'account_id': 'acct/1 ok'},
          )
          .path,
      resolvedBalances.path,
    );
    final resolvedRisk = operations
        .operation('analytics', 'black_scholes')!
        .resolve(
          queryParameters: {
            'symbol': 'BTC/USD',
            'strike': 100000,
            'ignored': 'drop-me',
            'volatility': 0.5,
          },
        );
    expect(
      resolvedRisk.path,
      '/risk-manager/blackscholes?symbol=BTC%2FUSD&strike=100000&volatility=0.5',
    );
  });

  test('SdkConfig.fromMap loads standard URLs and defaults', () {
    final config = SdkConfig.fromMap({
      'TWO_FINANCE_AUTH_URL': 'https://auth.example',
      'TWO_FINANCE_ANALYTICS_URL': 'https://analytics.example',
      'TWO_FINANCE_MATCHENGINE_WS_URL': 'wss://matchengine.example/ws',
    });

    expect(config.authUrl, 'https://auth.example');
    expect(config.analyticsUrl, 'https://analytics.example');
    expect(config.serviceUrl('analytics'), 'https://analytics.example');
    expect(config.serviceUrl('match_engine'), 'wss://matchengine.example/ws');
    expect(config.serviceUrls()['analytics'], 'https://analytics.example');
    expect(config.serviceUrls()['matchengine'], 'wss://matchengine.example/ws');
    expect(config.configuredServices(), hasLength(3));
    expect(config.configuredServices()[1].name, 'analytics');
    expect(config.configuredServices()[1].url, 'https://analytics.example');
    expect(config.missingServiceUrls(), hasLength(9));
    expect(config.missingServiceUrls().first.name, 'network');
    expect(config.missingServiceUrls().first.env, 'TWO_FINANCE_NETWORK_URL');
    expect(config.authRealm, '2finance');
  });

  test('TwoFinanceClient.fromEnvironment builds default client', () {
    final client = TwoFinanceClient.fromEnvironment();

    expect(client.config.authRealm, '2finance');
    expect(client.config.authClientId, '2finance-network');
    expect(client.analytics.service.baseUrl, '');
  });

  test('bearerAuthorization normalizes tokens', () {
    expect(bearerAuthorization('abc'), 'Bearer abc');
    expect(bearerAuthorization('Bearer abc'), 'Bearer abc');
    expect(bearerAuthorization(''), '');
  });

  test('service client injects bearer auth and decodes response', () async {
    ServiceRequest? seen;
    final client = TwoFinanceClient(
      const SdkConfig(analyticsUrl: 'https://analytics.example'),
      tokenSource: StaticTokenSource('token-123'),
      transport: (request) async {
        seen = request;
        return const ServiceResponse(statusCode: 200, body: '{"ok":true}');
      },
    );

    final response = await client.analytics.indicators();

    expect(response, {'ok': true});
    expect(
      seen?.url.toString(),
      'https://analytics.example/analytics/indicators',
    );
    expect(seen?.headers['Authorization'], 'Bearer token-123');
  });

  test('auth client exposes JWKS and token validation helpers', () async {
    final seen = <String>[];
    final client = TwoFinanceClient(
      const SdkConfig(authUrl: 'https://auth.example'),
      transport: (request) async {
        seen.add('${request.method} ${request.url} ${request.body ?? ''}');
        return const ServiceResponse(statusCode: 200, body: '{"ok":true}');
      },
    );

    await client.auth.jwks();
    await client.auth.validateToken('token-1');

    expect(
      seen,
      contains(
        'GET https://auth.example/realms/2finance/protocol/openid-connect/certs ',
      ),
    );
    expect(
      seen,
      contains(
        'POST https://auth.example/realms/2finance/protocol/openid-connect/token/introspect {"token":"token-1"}',
      ),
    );
  });

  test('service client applies request options and idempotency key', () async {
    final fixture = requestOptionsFixture();
    final request = fixture['request']! as Map<String, Object?>;
    final expected = fixture['expected']! as Map<String, Object?>;
    final expectedHeaders = expected['headers']! as Map<String, Object?>;
    final pagination = request['pagination']! as Map<String, Object?>;
    ServiceRequest? seen;
    final service = ServiceClient(
      request['base_url']! as String,
      transport: (request) async {
        seen = request;
        return const ServiceResponse(statusCode: 200, body: '{"ok":true}');
      },
    );

    await service.post(
      request['path']! as String,
      {'symbol': 'BTC-USDT'},
      RequestOptions(
        headers: Map<String, String>.from(request['headers']! as Map),
        idempotencyKey: ' ${request['idempotency_key']} ',
        queryParameters: Map<String, String>.from(request['query']! as Map),
        page: pagination['page']! as int,
        limit: pagination['limit']! as int,
      ),
    );

    expect(seen?.url.toString(), expected['url']);
    expect(seen?.headers['X-Trace-ID'], expectedHeaders['X-Trace-ID']);
    expect(
      seen?.headers['Idempotency-Key'],
      expectedHeaders['Idempotency-Key'],
    );

    await service.requestOperation(
      const ResolvedOperation(
        method: 'GET',
        path: '/portfolio-manager/balances/acct%2Fresolved',
      ),
    );

    expect(
      seen?.url.toString(),
      'https://analytics.example/portfolio-manager/balances/acct%2Fresolved',
    );
    expect(seen?.method, 'GET');

    final operations = DomainOperationsCatalog.fromJson(
      contractFixture('domain-operations.json'),
    );
    await service.requestCatalogOperation(
      operations,
      'analytics',
      'balances',
      pathParams: {'account_id': 'acct/1 ok'},
    );

    expect(
      seen?.url.toString(),
      'https://analytics.example/portfolio-manager/balances/acct%2F1%20ok',
    );
  });

  test('service client applies per-request timeout', () async {
    final service = ServiceClient(
      'https://analytics.example',
      transport: (request) async {
        await Future<void>.delayed(const Duration(milliseconds: 50));
        return const ServiceResponse(statusCode: 200, body: '{"ok":true}');
      },
    );

    expect(
      service.get(
        '/analytics/indicators',
        const RequestOptions(timeout: Duration(milliseconds: 1)),
      ),
      throwsA(isA<TimeoutException>()),
    );
  });

  test('service client retries retryable responses', () async {
    var attempts = 0;
    final service = ServiceClient(
      'https://analytics.example',
      transport: (request) async {
        attempts++;
        if (attempts == 1) {
          return const ServiceResponse(statusCode: 500, body: 'temporary');
        }
        return const ServiceResponse(statusCode: 200, body: '{"ok":true}');
      },
    );

    final response = await service.get(
      '/analytics/indicators',
      const RequestOptions(maxRetries: 1),
    );

    expect(attempts, 2);
    expect(response, {'ok': true});
  });

  test('client credentials token source requests and caches tokens', () async {
    var calls = 0;
    AuthTokenRequest? seen;
    final source = ClientCredentialsTokenSource(
      tokenUrl: 'https://auth.example/token',
      clientId: 'client-id',
      clientSecret: 'client-secret',
      scopes: const ['analytics:read', 'mcp:invoke'],
      transport: (request) async {
        calls++;
        seen = request;
        return const AuthTokenResponse(
          statusCode: 200,
          body: '{"access_token":"cc-token","expires_in":3600}',
        );
      },
    );

    expect(await source.token(), 'cc-token');
    expect(await source.token(), 'cc-token');
    expect(calls, 1);
    expect(seen?.url.toString(), 'https://auth.example/token');
    expect(seen?.body, contains('grant_type=client_credentials'));
    expect(seen?.body, contains('scope=analytics%3Aread+mcp%3Ainvoke'));
  });

  test('planner delegates to MCP conversation plan tool', () async {
    ServiceRequest? seen;
    final client = TwoFinanceClient(
      const SdkConfig(mcpUrl: 'https://mcp.example'),
      transport: (request) async {
        seen = request;
        return const ServiceResponse(
          statusCode: 200,
          body: '{"jsonrpc":"2.0"}',
        );
      },
    );

    await client.planner.conversationPlan({'goal': 'trade plan'});

    expect(seen?.url.toString(), 'https://mcp.example/mcp');
    expect(seen?.body, contains('finance_assistant.conversation.plan'));
  });

  test(
    'planner tradingPlan enriches context from analytics and trading',
    () async {
      final seen = <String>[];
      final client = TwoFinanceClient(
        const SdkConfig(
          analyticsUrl: 'https://analytics.example',
          mcpUrl: 'https://mcp.example',
          tradingControlUrl: 'https://trading.example',
        ),
        transport: (request) async {
          seen.add('${request.method} ${request.url} ${request.body ?? ''}');
          if (request.url.toString() == 'https://trading.example/robots') {
            return const ServiceResponse(
              statusCode: 200,
              body: '{"robots":[{"id":"robot-1"}]}',
            );
          }
          if (request.url.toString() ==
              'https://analytics.example/analytics/indicators') {
            return const ServiceResponse(
              statusCode: 200,
              body: '{"indicators":["rsi"]}',
            );
          }
          return const ServiceResponse(statusCode: 200, body: '{"ok":true}');
        },
      );

      await client.planner.tradingPlan({
        'goal': 'rebalance BTC',
        'context': {'account_id': 'acct-1'},
        'useTrading': true,
        'useAnalytics': true,
      });

      final mcpRequest = seen.singleWhere((entry) {
        return entry.startsWith('POST https://mcp.example/mcp ');
      });
      expect(mcpRequest, contains('finance_assistant.conversation.plan'));
      expect(mcpRequest, contains('trading_robots'));
      expect(mcpRequest, contains('analytics_indicators'));
      expect(mcpRequest, contains('account_id'));
    },
  );

  test(
    'mcp and orchestrator expose tools prompts resources and sessions',
    () async {
      final seen = <String>[];
      final client = TwoFinanceClient(
        const SdkConfig(
          mcpUrl: 'https://mcp.example',
          orchestratorUrl: 'https://orchestrator.example',
        ),
        transport: (request) async {
          seen.add('${request.method} ${request.url} ${request.body ?? ''}');
          return const ServiceResponse(statusCode: 200, body: '{"ok":true}');
        },
      );

      await client.mcp.listTools();
      await client.mcp.listPrompts();
      await client.mcp.readResource('resource://portfolio');
      await client.orchestrator.tools();
      await client.orchestrator.resources();
      await client.orchestrator.approvals();
      await client.orchestrator.deleteSession('session/1');
      await client.planner.operationalPlan({
        'session_id': 'session-1',
        'message': 'operate',
      });

      expect(
        seen.any((entry) => entry.contains('"method":"tools/list"')),
        isTrue,
      );
      expect(
        seen.any((entry) => entry.contains('"method":"prompts/list"')),
        isTrue,
      );
      expect(
        seen.any((entry) => entry.contains('resource://portfolio')),
        isTrue,
      );
      expect(
        seen,
        contains('GET https://orchestrator.example/v1/mcphost/tools '),
      );
      expect(
        seen,
        contains('GET https://orchestrator.example/v1/mcphost/resources '),
      );
      expect(
        seen,
        contains('GET https://orchestrator.example/v1/mcphost/approvals '),
      );
      expect(
        seen,
        contains(
          'DELETE https://orchestrator.example/v1/mcphost/sessions/session%2F1 ',
        ),
      );
      expect(
        seen.any(
          (entry) => entry.startsWith(
            'POST https://orchestrator.example/v1/mcphost/messages ',
          ),
        ),
        isTrue,
      );
    },
  );

  test('domain clients expose core service endpoints', () async {
    final seen = <String>[];
    final client = TwoFinanceClient(
      const SdkConfig(
        analyticsUrl: 'https://analytics.example',
        networkUrl: 'https://network.example',
        tradingControlUrl: 'https://trading.example',
        keystoreUrl: 'https://keys.example',
        hummingbotUrl: 'https://hbot.example',
        wiseUrl: 'https://wise.example',
        airwallexUrl: 'https://airwallex.example',
      ),
      transport: (request) async {
        seen.add('${request.method} ${request.url}');
        return const ServiceResponse(statusCode: 200, body: '{"ok":true}');
      },
    );

    await client.analytics.upsertCandles({'symbol': 'BTC-USDT'});
    await client.analytics.rankings();
    await client.analytics.balances('acct/1');
    await client.analytics.blackScholes('symbol=BTC');
    await client.analytics.staking();
    await client.network.marketCandles('BTC/USDT', 'limit=10');
    await client.network.bonds();
    await client.network.createBond({'symbol': 'BOND1'});
    await client.network.loans();
    await client.network.createLoan({'loan': 'ln1'});
    await client.network.swaps();
    await client.network.createSwap({'pair': 'BTC-USDT'});
    await client.network.stakingProducts();
    await client.network.createStakingProduct({'asset': 'TWO'});
    await client.network.syntheticAssets();
    await client.network.createSyntheticAsset({'asset': 'sBTC'});
    await client.network.liquidityPools();
    await client.network.createLiquidityPool({'pool': 'BTC-USDT'});
    await client.tradingControl.startRobot('robot/1');
    await client.tradingControl.pauseRobot('robot/1');
    await client.tradingControl.riskPolicy('robot/1');
    await client.tradingControl.riskView('robot/1');
    await client.tradingControl.strategies();
    await client.tradingControl.createStrategy({'name': 'mean-reversion'});
    await client.tradingControl.directives();
    await client.tradingControl.createDirective({'action': 'rebalance'});
    await client.tradingControl.audit();
    await client.tradingControl.activity();
    await client.tradingControl.mcpTools();
    await client.keystore.health();
    await client.keystore.readiness();
    await client.keystore.startSigning({'key': 'k1'});
    await client.keystore.keys('pub/1');
    await client.keystore.signatures('pub/1');
    await client.keystore.metrics();
    await client.hummingbot.balances();
    await client.hummingbot.connectorConfig({'connector': '2finance'});
    await client.providers.wise.profiles();
    await client.providers.wise.profile('profile/1');
    await client.providers.wise.createQuote('profile/1', {'source': 'USD'});
    await client.providers.wise.createTransfer({'target': 'BRL'});
    await client.providers.airwallex.accounts();
    await client.providers.airwallex.payments();
    await client.providers.airwallex.createPayment({'amount': 10});
    await client.providers.airwallex.beneficiaries();
    await client.providers.airwallex.createBeneficiary({'name': 'beneficiary'});

    expect(
      seen,
      containsAll([
        'POST https://analytics.example/analytics/candles:upsert',
        'GET https://analytics.example/portfolio-manager/rankings',
        'GET https://analytics.example/portfolio-manager/balances/acct%2F1',
        'GET https://analytics.example/risk-manager/blackscholes?symbol=BTC',
        'GET https://analytics.example/staking',
        'GET https://network.example/v2/2finance-network/markets/BTC%2FUSDT/candles?limit=10',
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
        'POST https://trading.example/robots/robot%2F1:start',
        'POST https://trading.example/robots/robot%2F1:pause',
        'GET https://trading.example/robots/robot%2F1/risk-policy',
        'GET https://trading.example/risk-view/robot%2F1',
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
        'GET https://keys.example/keystore/keys/pub%2F1',
        'GET https://keys.example/keystore/signatures/pub%2F1',
        'GET https://keys.example/keystore/tss/metrics',
        'GET https://hbot.example/api/v1/balances',
        'POST https://hbot.example/api/v1/connectors/2finance/config',
        'GET https://wise.example/v1/profiles',
        'GET https://wise.example/v1/profiles/profile%2F1',
        'POST https://wise.example/v3/profiles/profile%2F1/quotes',
        'POST https://wise.example/v1/transfers',
        'GET https://airwallex.example/api/v1/accounts',
        'GET https://airwallex.example/api/v1/payments',
        'POST https://airwallex.example/api/v1/payments',
        'GET https://airwallex.example/api/v1/beneficiaries',
        'POST https://airwallex.example/api/v1/beneficiaries',
      ]),
    );
  });

  test('matchengine client prepares order commands', () {
    final client = TwoFinanceClient(
      const SdkConfig(matchEngineWsUrl: 'wss://matchengine.example/ws'),
    );
    final command = client.matchEngine.orderCommand({
      'client_order_id': 'co-1',
      'idempotency_key': 'idem-1',
      'symbol': 'BTC-USDT',
      'side': 'buy',
      'type': 'limit',
      'quantity': '0.01',
    });

    expect(client.matchEngine.webSocketUrl, 'wss://matchengine.example/ws');
    expect(command['schema'], 'matchengine.order_command.v2');
    expect(command['symbol'], 'BTC-USDT');
    final subscription = client.matchEngine.marketDataSubscribe({
      'symbols': ['BTC-USDT'],
      'channels': ['book'],
    });
    expect(subscription['schema'], 'matchengine.market_data_subscribe.v2');
    expect(subscription['symbols'], ['BTC-USDT']);
    final sent = <Map<String, Object?>>[];
    final sendResult = client.matchEngine.sendOrder((message) {
      sent.add(message);
      return {'accepted': true};
    }, command);
    final subscribeResult = client.matchEngine.subscribeMarketData((message) {
      sent.add(message);
      return {'subscribed': true};
    }, subscription);

    expect(sendResult, {'accepted': true});
    expect(subscribeResult, {'subscribed': true});
    expect(sent[0]['schema'], 'matchengine.order_command.v2');
    expect(sent[1]['schema'], 'matchengine.market_data_subscribe.v2');
  });
}

Map<String, Object?> requestOptionsFixture() {
  return contractFixture('request-options.json');
}

Map<String, Object?> contractFixture(String name) {
  final file = File('../../contracts/examples/$name');
  return jsonDecode(file.readAsStringSync()) as Map<String, Object?>;
}

Map<String, Object?> contractOperation(
  Map<String, Object?> fixture,
  String domainName,
  String operationName,
) {
  final domains = fixture['domains']! as List<Object?>;
  final domain = domains.cast<Map<String, Object?>>().firstWhere(
    (item) => item['name'] == domainName,
  );
  final operations = domain['operations']! as List<Object?>;
  return operations.cast<Map<String, Object?>>().firstWhere(
    (item) => item['name'] == operationName,
  );
}
