import 'dart:async';

import 'service_client.dart';

final class AuthClient {
  AuthClient(
    this.service, {
    this.realm = '2finance',
    this.clientId = '2finance-network',
    this.phoneClientId = '2finance-network-phone',
  });

  final ServiceClient service;
  final String realm;
  final String clientId;
  final String phoneClientId;

  Future<Object?> login(Map<String, Object?> input) {
    return service.post(_authPath(clientId, '/login'), input);
  }

  Future<Object?> refreshToken(String refreshToken) {
    return service.post(_authPath(clientId, '/refresh'), {
      'refresh_token': refreshToken,
    });
  }

  Future<Object?> phoneLogin(String phoneNumber, String code) {
    return service.post(_authPath(phoneClientId, '/phone/sms/login'), {
      'phone_number': phoneNumber,
      'code': code,
    });
  }

  Future<Object?> jwks() {
    return service.get(_oidcPath('/protocol/openid-connect/certs'));
  }

  Future<Object?> validateToken(String token) {
    return service.post(
      _oidcPath('/protocol/openid-connect/token/introspect'),
      {'token': token},
    );
  }

  String _authPath(String selectedClientId, String endpoint) {
    return '/v1/2finance-authenticator/$realm/$selectedClientId/${endpoint.replaceFirst(RegExp(r'^/+'), '')}';
  }

  String _oidcPath(String endpoint) {
    return '/realms/$realm/${endpoint.replaceFirst(RegExp(r'^/+'), '')}';
  }
}

final class AnalyticsClient {
  AnalyticsClient(this.service);

  final ServiceClient service;

  Future<Object?> indicators() => service.get('/analytics/indicators');

  Future<Object?> calculateTechnicalAnalysis(Map<String, Object?> request) {
    return service.post('/analytics/technical-analysis:calculate', request);
  }

  Future<Object?> optimizePortfolio(Map<String, Object?> request) {
    return service.post('/portfolio-manager/optimizer', request);
  }

  Future<Object?> upsertCandles(Map<String, Object?> request) {
    return service.post('/analytics/candles:upsert', request);
  }

  Future<Object?> rankings() => service.get('/portfolio-manager/rankings');

  Future<Object?> balances(String accountId) {
    return service.get(
      '/portfolio-manager/balances/${Uri.encodeComponent(accountId)}',
    );
  }

  Future<Object?> blackScholes([String query = '']) {
    final suffix = query.isEmpty ? '' : '?$query';
    return service.get('/risk-manager/blackscholes$suffix');
  }

  Future<Object?> staking() => service.get('/staking');
}

final class MCPClient {
  MCPClient(this.service);

  final ServiceClient service;
  int _nextId = 1;

  Future<Object?> call(String method, [Object? params]) {
    return service.post('/mcp', {
      'jsonrpc': '2.0',
      'id': _nextId++,
      'method': method,
      'params': params,
    });
  }

  Future<Object?> listTools() => call('tools/list');

  Future<Object?> listPrompts() => call('prompts/list');

  Future<Object?> listResources() => call('resources/list');

  Future<Object?> readResource(String uri) {
    return call('resources/read', {'uri': uri});
  }

  Future<Object?> callTool(String name, [Map<String, Object?>? arguments]) {
    return call('tools/call', {'name': name, 'arguments': arguments ?? {}});
  }

  Future<Object?> getPrompt(String name, [Map<String, Object?>? arguments]) {
    return call('prompts/get', {'name': name, 'arguments': arguments ?? {}});
  }

  Future<Object?> conversationPlan(Map<String, Object?> arguments) {
    return callTool('finance_assistant.conversation.plan', arguments);
  }
}

final class NetworkClient {
  NetworkClient(this.service);

  final ServiceClient service;

  Future<Object?> virtualMachine() {
    return service.post('/v2/2finance-network/query', {
      'method': 'get_blocks',
      'params': {'page': 1, 'limit': 1},
    });
  }

  Future<Object?> marketCandles(String market, [String query = '']) {
    final suffix = query.isEmpty ? '' : '?$query';
    return service.get(
      '/v2/2finance-network/markets/${Uri.encodeComponent(market)}/candles$suffix',
    );
  }

  Future<Object?> products(String productType) {
    return service.get(
      '/v2/2finance-network/products/${Uri.encodeComponent(productType)}',
    );
  }

  Future<Object?> createProduct(
    String productType,
    Map<String, Object?> request,
  ) {
    return service.post(
      '/v2/2finance-network/products/${Uri.encodeComponent(productType)}',
      request,
    );
  }

  Future<Object?> bonds() => products('bonds');

  Future<Object?> createBond(Map<String, Object?> request) {
    return createProduct('bonds', request);
  }

  Future<Object?> loans() => products('loans');

  Future<Object?> createLoan(Map<String, Object?> request) {
    return createProduct('loans', request);
  }

  Future<Object?> swaps() => products('swaps');

  Future<Object?> createSwap(Map<String, Object?> request) {
    return createProduct('swaps', request);
  }

  Future<Object?> stakingProducts() => products('staking');

  Future<Object?> createStakingProduct(Map<String, Object?> request) {
    return createProduct('staking', request);
  }

  Future<Object?> syntheticAssets() => products('synthetic-assets');

  Future<Object?> createSyntheticAsset(Map<String, Object?> request) {
    return createProduct('synthetic-assets', request);
  }

  Future<Object?> liquidityPools() => products('liquidity-pools');

  Future<Object?> createLiquidityPool(Map<String, Object?> request) {
    return createProduct('liquidity-pools', request);
  }
}

final class OrchestratorClient {
  OrchestratorClient(this.service);

  final ServiceClient service;

  Future<Object?> catalog() => service.get('/v1/mcphost/catalog/packages');

  Future<Object?> createSession(Map<String, Object?> request) {
    return service.post('/v1/mcphost/sessions', request);
  }

  Future<Object?> sendMessage(Map<String, Object?> request) {
    return service.post('/v1/mcphost/messages', request);
  }

  Future<Object?> tools() => service.get('/v1/mcphost/tools');

  Future<Object?> prompts() => service.get('/v1/mcphost/prompts');

  Future<Object?> resources() => service.get('/v1/mcphost/resources');

  Future<Object?> providers() => service.get('/v1/mcphost/providers');

  Future<Object?> approvals() => service.get('/v1/mcphost/approvals');

  Future<Object?> observability() => service.get('/v1/mcphost/observability');

  Future<Object?> deleteSession(String id) {
    return service.delete('/v1/mcphost/sessions/${Uri.encodeComponent(id)}');
  }
}

final class TradingControlClient {
  TradingControlClient(this.service);

  final ServiceClient service;

  Future<Object?> robots() => service.get('/robots');

  Future<Object?> createRobot(Map<String, Object?> request) {
    return service.post('/robots', request);
  }

  Future<Object?> robot(String id) {
    return service.get('/robots/${Uri.encodeComponent(id)}');
  }

  Future<Object?> startRobot(String id) {
    return service.post('/robots/${Uri.encodeComponent(id)}:start');
  }

  Future<Object?> pauseRobot(String id) {
    return service.post('/robots/${Uri.encodeComponent(id)}:pause');
  }

  Future<Object?> resumeRobot(String id) {
    return service.post('/robots/${Uri.encodeComponent(id)}:resume');
  }

  Future<Object?> stopRobot(String id) {
    return service.post('/robots/${Uri.encodeComponent(id)}:stop');
  }

  Future<Object?> riskPolicy(String id) {
    return service.get('/robots/${Uri.encodeComponent(id)}/risk-policy');
  }

  Future<Object?> setRiskPolicy(String id, Map<String, Object?> request) {
    return service.put(
      '/robots/${Uri.encodeComponent(id)}/risk-policy',
      request,
    );
  }

  Future<Object?> riskView(String id) {
    return service.get('/risk-view/${Uri.encodeComponent(id)}');
  }

  Future<Object?> strategies() => service.get('/strategies');

  Future<Object?> createStrategy(Map<String, Object?> request) {
    return service.post('/strategies', request);
  }

  Future<Object?> directives() => service.get('/directives');

  Future<Object?> createDirective(Map<String, Object?> request) {
    return service.post('/directives', request);
  }

  Future<Object?> audit() => service.get('/audit');

  Future<Object?> activity() => service.get('/activity');

  Future<Object?> mcpTools() => service.get('/mcp/tools');
}

final class KeyStoreClient {
  KeyStoreClient(this.service);

  final ServiceClient service;

  Future<Object?> health() => service.get('/healthz');

  Future<Object?> readiness() => service.get('/readyz');

  Future<Object?> startKeygen(Map<String, Object?> request) {
    return service.post('/keystore/keygen/start', request);
  }

  Future<Object?> keygenSignature(Map<String, Object?> request) {
    return service.post('/keystore/keygen/signature', request);
  }

  Future<Object?> startSigning(Map<String, Object?> request) {
    return service.post('/keystore/signing/start', request);
  }

  Future<Object?> signingSignature(Map<String, Object?> request) {
    return service.post('/keystore/signing/signature', request);
  }

  Future<Object?> startResharing(Map<String, Object?> request) {
    return service.post('/keystore/resharing/start', request);
  }

  Future<Object?> keys(String userPublicKey) {
    return service.get('/keystore/keys/${Uri.encodeComponent(userPublicKey)}');
  }

  Future<Object?> signatures(String userPublicKey) {
    return service.get(
      '/keystore/signatures/${Uri.encodeComponent(userPublicKey)}',
    );
  }

  Future<Object?> metrics() => service.get('/keystore/tss/metrics');
}

final class HummingbotClient {
  HummingbotClient(this.service);

  final ServiceClient service;

  Future<Object?> assets() => service.get('/api/v1/assets');

  Future<Object?> symbols() => service.get('/api/v1/symbols');

  Future<Object?> balances() => service.get('/api/v1/balances');

  Future<Object?> connectorConfig(Map<String, Object?> request) {
    return service.post('/api/v1/connectors/2finance/config', request);
  }
}

final class MatchEngineClient {
  MatchEngineClient(this.webSocketUrl);

  final String webSocketUrl;

  Map<String, Object?> orderCommand(Map<String, Object?> command) {
    return {
      'schema': 'matchengine.order_command.v2',
      'message_type': 'ORDER',
      'operation': 'ADD',
      ...command,
    };
  }

  Map<String, Object?> marketDataSubscribe(Map<String, Object?> request) {
    return {'schema': 'matchengine.market_data_subscribe.v2', ...request};
  }

  FutureOr<Object?> sendOrder(
    FutureOr<Object?> Function(Map<String, Object?> message) sender,
    Map<String, Object?> command,
  ) {
    return sender(orderCommand(command));
  }

  FutureOr<Object?> subscribeMarketData(
    FutureOr<Object?> Function(Map<String, Object?> message) sender,
    Map<String, Object?> request,
  ) {
    return sender(marketDataSubscribe(request));
  }
}

final class ProviderClient {
  ProviderClient(this.service);

  final ServiceClient service;

  Future<Object?> get(String path) => service.get(path);

  Future<Object?> post(String path, [Object? body]) => service.post(path, body);

  Future<Object?> put(String path, [Object? body]) => service.put(path, body);

  Future<Object?> delete(String path) => service.delete(path);
}

final class WiseClient extends ProviderClient {
  WiseClient(super.service);

  Future<Object?> profiles() => get('/v1/profiles');

  Future<Object?> profile(String profileId) {
    return get('/v1/profiles/${Uri.encodeComponent(profileId)}');
  }

  Future<Object?> createQuote(String profileId, Map<String, Object?> request) {
    return post(
      '/v3/profiles/${Uri.encodeComponent(profileId)}/quotes',
      request,
    );
  }

  Future<Object?> createTransfer(Map<String, Object?> request) {
    return post('/v1/transfers', request);
  }
}

final class AirwallexClient extends ProviderClient {
  AirwallexClient(super.service);

  Future<Object?> accounts() => get('/api/v1/accounts');

  Future<Object?> payments() => get('/api/v1/payments');

  Future<Object?> createPayment(Map<String, Object?> request) {
    return post('/api/v1/payments', request);
  }

  Future<Object?> beneficiaries() => get('/api/v1/beneficiaries');

  Future<Object?> createBeneficiary(Map<String, Object?> request) {
    return post('/api/v1/beneficiaries', request);
  }
}

final class ProvidersClient {
  ProvidersClient({required this.wise, required this.airwallex});

  final WiseClient wise;
  final AirwallexClient airwallex;
}

final class PlannerClient {
  PlannerClient({
    required this.mcp,
    required this.orchestrator,
    required this.analytics,
    required this.tradingControl,
  });

  final MCPClient mcp;
  final OrchestratorClient orchestrator;
  final AnalyticsClient analytics;
  final TradingControlClient tradingControl;

  Future<Object?> conversationPlan(Map<String, Object?> request) {
    return mcp.conversationPlan(request);
  }

  Future<Object?> orchestratedPlan(Map<String, Object?> request) {
    return orchestrator.sendMessage(request);
  }

  Future<Object?> operationalPlan(Map<String, Object?> request) {
    return orchestratedPlan(request);
  }

  Future<Object?> tradingPlan(Map<String, Object?> request) async {
    final existingContext = request['context'];
    final context = <String, Object?>{
      if (existingContext is Map<String, Object?>) ...existingContext,
    };
    if (request['useTrading'] == true || request['use_trading'] == true) {
      try {
        context['trading_robots'] = await tradingControl.robots();
      } catch (_) {
        // Best-effort enrichment keeps planning usable when trading is unavailable.
      }
    }
    if (request['useAnalytics'] == true || request['use_analytics'] == true) {
      try {
        context['analytics_indicators'] = await analytics.indicators();
      } catch (_) {
        // Best-effort enrichment keeps planning usable when analytics is unavailable.
      }
    }
    return conversationPlan({...request, 'context': context});
  }
}
